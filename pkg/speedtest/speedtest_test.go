package speedtest

import (
	"context"
	"sync"
	"testing"
)

type mockBackend struct {
	name string
}

func (m *mockBackend) Name() string        { return m.name }
func (m *mockBackend) Granularity() string  { return "per-sample" }

func (m *mockBackend) FetchMeta(ctx context.Context) (*ConnectionInfo, error) {
	return &ConnectionInfo{
		ClientIP: "1.2.3.4",
		ASN:      12345,
		ASOrg:    "Test ISP",
		Country:  "US",
		City:     "Test City",
	}, nil
}

func (m *mockBackend) MeasureLatency(ctx context.Context, count int, sink EventSink) (*LatencyResult, error) {
	samples := make([]float64, count)
	for i := range count {
		samples[i] = float64(5 + i)
		EmitLatency(sink, i, samples[i])
	}
	r := ComputeLatencyResult(samples, false)
	return &r, nil
}

func (m *mockBackend) MeasureDownload(ctx context.Context, steps []TestStep, sink EventSink) (*ThroughputResult, error) {
	var rawBps []uint64
	var totalBytes uint64
	for _, s := range steps {
		for range s.Count {
			bps := uint64(100_000_000)
			rawBps = append(rawBps, bps)
			totalBytes += s.Bytes
			EmitSample(sink, &SampleResult{
				Direction:  "download",
				Bytes:      s.Bytes,
				BitsPerSec: bps,
			})
		}
	}
	r := ComputeThroughputResult(rawBps, totalBytes, false)
	return &r, nil
}

func (m *mockBackend) MeasureUpload(ctx context.Context, steps []TestStep, sink EventSink) (*ThroughputResult, error) {
	var rawBps []uint64
	var totalBytes uint64
	for _, s := range steps {
		for range s.Count {
			bps := uint64(50_000_000)
			rawBps = append(rawBps, bps)
			totalBytes += s.Bytes
			EmitSample(sink, &SampleResult{
				Direction:  "upload",
				Bytes:      s.Bytes,
				BitsPerSec: bps,
			})
		}
	}
	r := ComputeThroughputResult(rawBps, totalBytes, false)
	return &r, nil
}

func TestRunner_QuickPlan(t *testing.T) {
	var events []Event
	sink := func(e Event) { events = append(events, e) }

	runner := NewRunner(
		WithBackend(&mockBackend{name: "mock"}),
		WithPlan(QuickPlan),
		WithSink(sink),
	)

	result, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.Backend != "mock" {
		t.Errorf("backend: expected 'mock', got %q", result.Backend)
	}
	if result.Preset != "quick" {
		t.Errorf("preset: expected 'quick', got %q", result.Preset)
	}
	if result.Connection.ClientIP != "1.2.3.4" {
		t.Errorf("client_ip: expected '1.2.3.4', got %q", result.Connection.ClientIP)
	}
	if result.Latency.Samples != QuickPlan.LatencyCount {
		t.Errorf("latency samples: expected %d, got %d", QuickPlan.LatencyCount, result.Latency.Samples)
	}
	if result.Status != "ok" {
		t.Errorf("status: expected 'ok', got %q", result.Status)
	}
	if result.Download.BitsPerSec == 0 {
		t.Error("download should have results")
	}
	if result.Upload.BitsPerSec != 0 {
		t.Error("quick plan should have no upload results")
	}
	if result.DurationS <= 0 {
		t.Error("duration should be > 0")
	}
	if result.Granularity != "per-sample" {
		t.Errorf("granularity: expected 'per-sample', got %q", result.Granularity)
	}

	hasMetaEvent := false
	hasResultEvent := false
	for _, e := range events {
		if e.Type == "meta" {
			hasMetaEvent = true
		}
		if e.Type == "result" {
			hasResultEvent = true
		}
	}
	if !hasMetaEvent {
		t.Error("should have emitted meta event")
	}
	if !hasResultEvent {
		t.Error("should have emitted result event")
	}
}

func TestRunner_DefaultPlan(t *testing.T) {
	runner := NewRunner(
		WithBackend(&mockBackend{name: "mock"}),
		WithPlan(DefaultPlan),
	)

	result, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.Download.BitsPerSec == 0 {
		t.Error("download should have results")
	}
	if result.Upload.BitsPerSec == 0 {
		t.Error("upload should have results")
	}
}

func TestRunner_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	runner := NewRunner(
		WithBackend(&mockBackend{name: "mock"}),
		WithPlan(QuickPlan),
	)

	result, err := runner.Run(ctx)
	if err != nil {
		t.Logf("got expected error: %v", err)
	}
	if result == nil {
		t.Error("should return partial result even on error")
	}
}

func TestRunAll_MultipleBackends(t *testing.T) {
	var mu sync.Mutex
	var events []Event
	sink := func(e Event) {
		mu.Lock()
		events = append(events, e)
		mu.Unlock()
	}

	runner := NewRunner(
		WithBackends([]Backend{
			&mockBackend{name: "alpha"},
			&mockBackend{name: "beta"},
		}),
		WithPlan(QuickPlan),
		WithSink(sink),
	)

	report, err := runner.RunAll(context.Background())
	if err != nil {
		t.Fatalf("RunAll failed: %v", err)
	}

	if len(report.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(report.Results))
	}

	names := map[string]bool{}
	for _, r := range report.Results {
		names[r.Backend] = true
		if r.Download.BitsPerSec == 0 {
			t.Errorf("backend %s: download should have results", r.Backend)
		}
	}
	if !names["alpha"] || !names["beta"] {
		t.Errorf("expected both alpha and beta backends, got %v", names)
	}

	if report.DurationS <= 0 {
		t.Error("report duration should be > 0")
	}
	if report.Preset != "quick" {
		t.Errorf("preset: expected 'quick', got %q", report.Preset)
	}

	mu.Lock()
	defer mu.Unlock()
	hasReport := false
	taggedBackends := map[string]bool{}
	for _, e := range events {
		if e.Type == "report" {
			hasReport = true
		}
		if e.Backend != "" {
			taggedBackends[e.Backend] = true
		}
	}
	if !hasReport {
		t.Error("should have emitted report event")
	}
	if !taggedBackends["alpha"] || !taggedBackends["beta"] {
		t.Errorf("events should be tagged with backend names, got %v", taggedBackends)
	}
}

func TestRunAll_SingleBackendFallback(t *testing.T) {
	runner := NewRunner(
		WithBackend(&mockBackend{name: "solo"}),
		WithPlan(QuickPlan),
	)

	report, err := runner.RunAll(context.Background())
	if err != nil {
		t.Fatalf("RunAll failed: %v", err)
	}

	if len(report.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(report.Results))
	}
	if report.Results[0].Backend != "solo" {
		t.Errorf("expected 'solo', got %q", report.Results[0].Backend)
	}
}

func TestRunAll_Sequential(t *testing.T) {
	var order []string
	var mu sync.Mutex
	sink := func(e Event) {
		if e.Type == "meta" && e.Backend != "" {
			mu.Lock()
			order = append(order, e.Backend)
			mu.Unlock()
		}
	}

	runner := NewRunner(
		WithBackends([]Backend{
			&mockBackend{name: "first"},
			&mockBackend{name: "second"},
		}),
		WithPlan(QuickPlan),
		WithSink(sink),
		WithSequential(true),
	)

	report, err := runner.RunAll(context.Background())
	if err != nil {
		t.Fatalf("RunAll failed: %v", err)
	}
	if len(report.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(report.Results))
	}

	mu.Lock()
	defer mu.Unlock()
	if len(order) != 2 || order[0] != "first" || order[1] != "second" {
		t.Errorf("sequential execution should preserve order, got %v", order)
	}
}

func TestResult_StatusOk(t *testing.T) {
	runner := NewRunner(
		WithBackend(&mockBackend{name: "test"}),
		WithPlan(QuickPlan),
	)
	result, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", result.Status)
	}
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got %v", result.Errors)
	}
}
