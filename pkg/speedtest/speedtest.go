package speedtest

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"time"
)

type Runner struct {
	backends   []Backend
	plan       TestPlan
	sink       EventSink
	sequential bool
	proxyMode  string
}

type RunnerOption func(*Runner)

func WithBackend(b Backend) RunnerOption     { return func(r *Runner) { r.backends = []Backend{b} } }
func WithBackends(bs []Backend) RunnerOption { return func(r *Runner) { r.backends = bs } }
func WithPlan(p TestPlan) RunnerOption       { return func(r *Runner) { r.plan = p } }
func WithSink(s EventSink) RunnerOption      { return func(r *Runner) { r.sink = s } }
func WithSequential(v bool) RunnerOption     { return func(r *Runner) { r.sequential = v } }
func WithProxyMode(m string) RunnerOption    { return func(r *Runner) { r.proxyMode = m } }

func NewRunner(opts ...RunnerOption) *Runner {
	r := &Runner{
		plan: DefaultPlan,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

func (r *Runner) Run(ctx context.Context) (*Result, error) {
	if len(r.backends) == 0 {
		return nil, fmt.Errorf("no backends configured")
	}
	return r.runBackend(ctx, r.backends[0], r.sink)
}

func (r *Runner) runBackend(ctx context.Context, backend Backend, sink EventSink) (*Result, error) {
	start := time.Now()

	if r.plan.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.plan.Timeout)
		defer cancel()
	}

	proxyMode := r.proxyMode
	if proxyMode == "" {
		proxyMode = ProxyModeSystem
	}

	result := &Result{
		ID:          GenerateID(),
		Timestamp:   start,
		Status:      "ok",
		Backend:     backend.Name(),
		Preset:      r.plan.Name,
		Granularity: backend.Granularity(),
		ProxyMode:   proxyMode,
	}

	if proxy := detectProxy(); proxy != nil {
		result.ProxyDetected = proxy
	}

	meta, err := backend.FetchMeta(ctx)
	if err != nil {
		result.Status = "failed"
		result.Errors = append(result.Errors, fmt.Sprintf("meta: %v", err))
		result.DurationS = time.Since(start).Seconds()
		return result, err
	}
	result.Connection = *meta
	EmitMeta(sink, meta)

	latencyCtx := ctx
	if r.plan.LatencyTimeout > 0 {
		var cancel context.CancelFunc
		latencyCtx, cancel = context.WithTimeout(ctx, r.plan.LatencyTimeout)
		defer cancel()
	}
	latency, err := backend.MeasureLatency(latencyCtx, r.plan.LatencyCount, sink)
	if err != nil {
		result.Status = "partial"
		result.Errors = append(result.Errors, fmt.Sprintf("latency: %v", err))
	}
	if latency != nil {
		if !r.plan.IncludeRaw {
			latency.RawMs = nil
		}
		result.Latency = *latency
	}

	dlSteps := r.plan.DownloadSteps()
	if len(dlSteps) > 0 {
		dlCtx := ctx
		if r.plan.DownloadTimeout > 0 {
			var cancel context.CancelFunc
			dlCtx, cancel = context.WithTimeout(ctx, r.plan.DownloadTimeout)
			defer cancel()
		}
		dl, err := backend.MeasureDownload(dlCtx, dlSteps, sink)
		if err != nil {
			if result.Status == "ok" {
				result.Status = "partial"
			}
			result.Errors = append(result.Errors, fmt.Sprintf("download: %v", err))
		}
		if dl != nil {
			if !r.plan.IncludeRaw {
				dl.RawBps = nil
			}
			result.Download = *dl
		}
	}

	upSteps := r.plan.UploadSteps()
	if len(upSteps) > 0 {
		upCtx := ctx
		if r.plan.UploadTimeout > 0 {
			var cancel context.CancelFunc
			upCtx, cancel = context.WithTimeout(ctx, r.plan.UploadTimeout)
			defer cancel()
		}
		ul, err := backend.MeasureUpload(upCtx, upSteps, sink)
		if err != nil {
			if result.Status == "ok" {
				result.Status = "partial"
			}
			result.Errors = append(result.Errors, fmt.Sprintf("upload: %v", err))
		}
		if ul != nil {
			if !r.plan.IncludeRaw {
				ul.RawBps = nil
			}
			result.Upload = *ul
		}
	}

	result.DurationS = time.Since(start).Seconds()
	EmitResult(sink, result)
	return result, nil
}

func (r *Runner) RunAll(ctx context.Context) (*Report, error) {
	if len(r.backends) == 0 {
		return nil, fmt.Errorf("no backends configured")
	}

	start := time.Now()

	report := &Report{
		Timestamp: start,
		Preset:    r.plan.Name,
		Results:   make([]Result, 0, len(r.backends)),
	}

	if r.sequential {
		for _, b := range r.backends {
			result, _ := r.runBackend(ctx, b, TaggedSink(r.sink, b.Name()))
			if result != nil {
				report.Results = append(report.Results, *result)
			}
		}
	} else {
		type backendResult struct {
			result *Result
			err    error
		}
		ch := make(chan backendResult, len(r.backends))
		for _, b := range r.backends {
			go func(b Backend) {
				result, err := r.runBackend(ctx, b, TaggedSink(r.sink, b.Name()))
				ch <- backendResult{result: result, err: err}
			}(b)
		}
		for range r.backends {
			br := <-ch
			if br.result != nil {
				report.Results = append(report.Results, *br.result)
			}
		}
	}

	report.DurationS = time.Since(start).Seconds()

	allFailed := len(report.Results) > 0
	for _, res := range report.Results {
		if res.Status != "failed" {
			allFailed = false
			break
		}
	}
	if len(report.Results) == 0 {
		allFailed = true
	}

	EmitReport(r.sink, report)

	if allFailed {
		return report, fmt.Errorf("all backends failed")
	}
	return report, nil
}

func detectProxy() *ProxyInfo {
	httpProxy := firstNonEmpty("HTTP_PROXY", "http_proxy")
	httpsProxy := firstNonEmpty("HTTPS_PROXY", "https_proxy")
	allProxy := firstNonEmpty("ALL_PROXY", "all_proxy")
	if httpProxy == "" && httpsProxy == "" && allProxy == "" {
		return nil
	}
	return &ProxyInfo{
		HTTPProxy:  RedactProxyURL(httpProxy),
		HTTPSProxy: RedactProxyURL(httpsProxy),
		AllProxy:   RedactProxyURL(allProxy),
	}
}

func firstNonEmpty(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

func GenerateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", b)
}
