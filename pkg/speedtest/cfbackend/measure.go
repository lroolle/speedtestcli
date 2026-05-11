package cfbackend

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/lroolle/speedtestcli/pkg/speedtest"
)

func measureLatency(ctx context.Context, client HTTPDoer, baseURL, measID string, count int, sink speedtest.EventSink) (*speedtest.LatencyResult, error) {
	samples := make([]float64, 0, count)

	for i := range count {
		if ctx.Err() != nil {
			break
		}

		url := fmt.Sprintf("%s/__down?bytes=0&measId=%s", baseURL, measID)

		rt := &requestTiming{start: time.Now()}
		trace := newClientTrace(rt)
		req, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("creating latency request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			speedtest.EmitError(sink, "latency", err.Error())
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		serverTiming := parseServerTiming(resp)
		ttfb := rt.ttfb()
		latencyMs := float64(ttfb-serverTiming) / float64(time.Millisecond)
		if latencyMs < 0 {
			latencyMs = float64(ttfb) / float64(time.Millisecond)
		}

		samples = append(samples, latencyMs)
		speedtest.EmitLatency(sink, i, latencyMs)
	}

	if len(samples) == 0 {
		return nil, fmt.Errorf("all %d latency probes failed", count)
	}

	result := speedtest.ComputeLatencyResult(samples, false)
	return &result, nil
}

func measureDownload(ctx context.Context, client HTTPDoer, baseURL, measID string, steps []speedtest.TestStep, sink speedtest.EventSink) (*speedtest.ThroughputResult, error) {
	var rawBps []uint64
	var bytesTotal uint64

	for _, step := range steps {
		for range step.Count {
			if ctx.Err() != nil {
				break
			}

			url := fmt.Sprintf("%s/__down?bytes=%d&measId=%s", baseURL, step.Bytes, measID)

			rt := &requestTiming{start: time.Now()}
			trace := newClientTrace(rt)
			req, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), "GET", url, nil)
			if err != nil {
				return nil, fmt.Errorf("creating download request: %w", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				speedtest.EmitError(sink, "download", err.Error())
				continue
			}

			written, _ := io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			serverTiming := parseServerTiming(resp)
			totalDuration := time.Since(rt.connectBase()) - serverTiming
			if totalDuration <= 0 {
				totalDuration = time.Since(rt.start)
			}

			responseBytes := uint64(written) + estimateHeaderSize(resp)
			bps := uint64(math.Round(float64(responseBytes*8) / totalDuration.Seconds()))

			rawBps = append(rawBps, bps)
			bytesTotal += uint64(written)

			speedtest.EmitSample(sink, &speedtest.SampleResult{
				Direction:     "download",
				Bytes:         uint64(written),
				BitsPerSec:    bps,
				ResponseBytes: responseBytes,
				Timing: speedtest.TimingTrace{
					TTFB:         float64(rt.ttfb()) / float64(time.Millisecond),
					ServerTiming: float64(serverTiming) / float64(time.Millisecond),
					Total:        float64(time.Since(rt.start)) / float64(time.Millisecond),
					ConnReused:   rt.connReused,
				},
			})
		}
	}

	if len(rawBps) == 0 {
		return nil, fmt.Errorf("all download measurements failed")
	}

	result := speedtest.ComputeThroughputResult(rawBps, bytesTotal, false)
	return &result, nil
}

func measureUpload(ctx context.Context, client HTTPDoer, baseURL, measID string, steps []speedtest.TestStep, sink speedtest.EventSink) (*speedtest.ThroughputResult, error) {
	var rawBps []uint64
	var bytesTotal uint64

	for _, step := range steps {
		for range step.Count {
			if ctx.Err() != nil {
				break
			}

			url := fmt.Sprintf("%s/__up?measId=%s", baseURL, measID)
			body := bytes.NewReader(make([]byte, step.Bytes))

			rt := &requestTiming{start: time.Now()}
			trace := newClientTrace(rt)
			req, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), "POST", url, body)
			if err != nil {
				return nil, fmt.Errorf("creating upload request: %w", err)
			}
			req.ContentLength = int64(step.Bytes)

			resp, err := client.Do(req)
			if err != nil {
				speedtest.EmitError(sink, "upload", err.Error())
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			serverTiming := parseServerTiming(resp)
			uploadDuration := rt.uploadDuration() - serverTiming
			if uploadDuration <= 0 {
				uploadDuration = rt.uploadDuration()
			}
			if uploadDuration <= 0 {
				uploadDuration = time.Since(rt.start)
			}

			requestBytes := step.Bytes + estimateRequestHeaderSize(req)
			bps := uint64(math.Round(float64(requestBytes*8) / uploadDuration.Seconds()))

			rawBps = append(rawBps, bps)
			bytesTotal += step.Bytes

			speedtest.EmitSample(sink, &speedtest.SampleResult{
				Direction:    "upload",
				Bytes:        step.Bytes,
				BitsPerSec:   bps,
				RequestBytes: requestBytes,
				Timing: speedtest.TimingTrace{
					TTFB:         float64(rt.ttfb()) / float64(time.Millisecond),
					ServerTiming: float64(serverTiming) / float64(time.Millisecond),
					Total:        float64(time.Since(rt.start)) / float64(time.Millisecond),
					ConnReused:   rt.connReused,
				},
			})
		}
	}

	if len(rawBps) == 0 {
		return nil, fmt.Errorf("all upload measurements failed")
	}

	result := speedtest.ComputeThroughputResult(rawBps, bytesTotal, false)
	return &result, nil
}

func estimateHeaderSize(resp *http.Response) uint64 {
	size := uint64(len(resp.Proto + " " + resp.Status + "\r\n"))
	for k, vals := range resp.Header {
		for _, v := range vals {
			size += uint64(len(k) + 2 + len(v) + 2)
		}
	}
	return size + 2
}

func estimateRequestHeaderSize(req *http.Request) uint64 {
	size := uint64(len(req.Method + " " + req.URL.String() + " " + req.Proto + "\r\n"))
	for k, vals := range req.Header {
		for _, v := range vals {
			size += uint64(len(k) + 2 + len(v) + 2)
		}
	}
	return size + 2
}
