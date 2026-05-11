package cfbackend

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/lroolle/speedtestcli/pkg/speedtest"
)

func newTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/meta":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(testMetaJSON))

		case "/__down":
			bytesStr := r.URL.Query().Get("bytes")
			n, _ := strconv.Atoi(bytesStr)
			w.Header().Set("Server-Timing", "cfRequestDuration;dur=1.000")
			if n > 0 {
				w.Write(make([]byte, n))
			}

		case "/__up":
			w.Header().Set("Server-Timing", "cfRequestDuration;dur=1.000")
			w.WriteHeader(200)

		default:
			w.WriteHeader(404)
		}
	}))
}

func TestBackend_FetchMeta(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	b := New(WithHTTPDoer(srv.Client()), WithBaseURL(srv.URL))
	info, err := b.FetchMeta(context.Background())
	if err != nil {
		t.Fatalf("FetchMeta failed: %v", err)
	}
	if info.ClientIP != "203.0.113.42" {
		t.Errorf("expected 203.0.113.42, got %s", info.ClientIP)
	}
}

func TestBackend_MeasureLatency(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	b := New(WithHTTPDoer(srv.Client()), WithBaseURL(srv.URL))

	var events []speedtest.Event
	sink := func(e speedtest.Event) { events = append(events, e) }

	result, err := b.MeasureLatency(context.Background(), 5, sink)
	if err != nil {
		t.Fatalf("MeasureLatency failed: %v", err)
	}
	if result.Samples != 5 {
		t.Errorf("expected 5 samples, got %d", result.Samples)
	}

	latencyEvents := 0
	for _, e := range events {
		if e.Type == "latency" {
			latencyEvents++
		}
	}
	if latencyEvents != 5 {
		t.Errorf("expected 5 latency events, got %d", latencyEvents)
	}
}

func TestBackend_MeasureDownload(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	b := New(WithHTTPDoer(srv.Client()), WithBaseURL(srv.URL))

	steps := []speedtest.TestStep{
		{Direction: "download", Bytes: 1000, Count: 3},
	}

	var events []speedtest.Event
	sink := func(e speedtest.Event) { events = append(events, e) }

	result, err := b.MeasureDownload(context.Background(), steps, sink)
	if err != nil {
		t.Fatalf("MeasureDownload failed: %v", err)
	}
	if result.Samples != 3 {
		t.Errorf("expected 3 samples, got %d", result.Samples)
	}
	if result.BitsPerSec == 0 {
		t.Error("bits_per_sec should be > 0")
	}
	if result.BytesTotal == 0 {
		t.Error("bytes_total should be > 0")
	}

	sampleEvents := 0
	for _, e := range events {
		if e.Type == "sample" && e.Sample != nil && e.Sample.Direction == "download" {
			sampleEvents++
		}
	}
	if sampleEvents != 3 {
		t.Errorf("expected 3 download sample events, got %d", sampleEvents)
	}
}

func TestBackend_MeasureUpload(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	b := New(WithHTTPDoer(srv.Client()), WithBaseURL(srv.URL))

	steps := []speedtest.TestStep{
		{Direction: "upload", Bytes: 1000, Count: 3},
	}

	var events []speedtest.Event
	sink := func(e speedtest.Event) { events = append(events, e) }

	result, err := b.MeasureUpload(context.Background(), steps, sink)
	if err != nil {
		t.Fatalf("MeasureUpload failed: %v", err)
	}
	if result.Samples != 3 {
		t.Errorf("expected 3 samples, got %d", result.Samples)
	}
	if result.BitsPerSec == 0 {
		t.Error("bits_per_sec should be > 0")
	}
}

func TestBackend_Name(t *testing.T) {
	b := New()
	if b.Name() != "cloudflare" {
		t.Errorf("expected 'cloudflare', got %q", b.Name())
	}
}

func TestBackend_ContextCancellation(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	b := New(WithHTTPDoer(srv.Client()), WithBaseURL(srv.URL))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := b.MeasureLatency(ctx, 100, nil)
	if err == nil {
		t.Error("expected error when context is cancelled")
	}
	fmt.Println("got expected error:", err)
}
