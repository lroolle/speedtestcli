package ooklabackend

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	stgo "github.com/showwin/speedtest-go/speedtest"

	st "github.com/lroolle/speedtestcli/pkg/speedtest"
)

type Backend struct {
	client *stgo.Speedtest
	server *stgo.Server
}

type Option func(*Backend)

func WithClient(c *stgo.Speedtest) Option {
	return func(b *Backend) { b.client = c }
}

func WithDirect() Option {
	return func(b *Backend) {
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.Proxy = nil
		b.client = stgo.New(stgo.WithDoer(&http.Client{Transport: t}))
	}
}

func New(opts ...Option) *Backend {
	b := &Backend{
		client: stgo.New(),
	}
	for _, o := range opts {
		o(b)
	}
	return b
}

func (b *Backend) Name() string        { return "ookla" }
func (b *Backend) Granularity() string { return "aggregate" }

func (b *Backend) FetchMeta(ctx context.Context) (*st.ConnectionInfo, error) {
	user, err := b.client.FetchUserInfoContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching user info: %w", err)
	}
	lat, _ := strconv.ParseFloat(user.Lat, 64)
	lon, _ := strconv.ParseFloat(user.Lon, 64)
	return &st.ConnectionInfo{
		ClientIP:  user.IP,
		ASOrg:     user.Isp,
		Latitude:  lat,
		Longitude: lon,
	}, nil
}

func (b *Backend) ensureServer(ctx context.Context) (*stgo.Server, error) {
	if b.server != nil {
		return b.server, nil
	}
	servers, err := b.client.FetchServerListContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching servers: %w", err)
	}
	targets, err := servers.FindServer([]int{})
	if err != nil || len(targets) == 0 {
		return nil, fmt.Errorf("no servers found")
	}
	b.server = targets[0]
	return b.server, nil
}

func (b *Backend) MeasureLatency(ctx context.Context, count int, sink st.EventSink) (*st.LatencyResult, error) {
	server, err := b.ensureServer(ctx)
	if err != nil {
		return nil, err
	}

	samples := make([]float64, 0, count)
	for i := range count {
		if ctx.Err() != nil {
			break
		}
		err := server.PingTestContext(ctx, nil)
		if err != nil {
			st.EmitError(sink, "latency", err.Error())
			continue
		}
		ms := float64(server.Latency) / float64(time.Millisecond)
		samples = append(samples, ms)
		st.EmitLatency(sink, i, ms)
	}

	if len(samples) == 0 {
		return nil, fmt.Errorf("all latency probes failed")
	}

	result := st.ComputeLatencyResult(samples, false)
	return &result, nil
}

func (b *Backend) MeasureDownload(ctx context.Context, steps []st.TestStep, sink st.EventSink) (*st.ThroughputResult, error) {
	server, err := b.ensureServer(ctx)
	if err != nil {
		return nil, err
	}

	err = server.DownloadTestContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("download test: %w", err)
	}

	bps := uint64(server.DLSpeed * 1_000_000)
	bytesTotal := uint64(server.Context.Manager.GetTotalDownload())

	st.EmitSample(sink, &st.SampleResult{
		Direction:  "download",
		Bytes:      bytesTotal,
		BitsPerSec: bps,
	})

	result := st.ThroughputResult{
		BitsPerSec: bps,
		BytesTotal: bytesTotal,
		Samples:    1,
		Stats: st.ThroughputStats{
			MinBps:    bps,
			P10Bps:    bps,
			MedianBps: bps,
			MeanBps:   bps,
			P90Bps:    bps,
			MaxBps:    bps,
		},
	}
	return &result, nil
}

func (b *Backend) MeasureUpload(ctx context.Context, steps []st.TestStep, sink st.EventSink) (*st.ThroughputResult, error) {
	server, err := b.ensureServer(ctx)
	if err != nil {
		return nil, err
	}

	err = server.UploadTestContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("upload test: %w", err)
	}

	bps := uint64(server.ULSpeed * 1_000_000)
	bytesTotal := uint64(server.Context.Manager.GetTotalUpload())

	st.EmitSample(sink, &st.SampleResult{
		Direction:  "upload",
		Bytes:      bytesTotal,
		BitsPerSec: bps,
	})

	result := st.ThroughputResult{
		BitsPerSec: bps,
		BytesTotal: bytesTotal,
		Samples:    1,
		Stats: st.ThroughputStats{
			MinBps:    bps,
			P10Bps:    bps,
			MedianBps: bps,
			MeanBps:   bps,
			P90Bps:    bps,
			MaxBps:    bps,
		},
	}
	return &result, nil
}
