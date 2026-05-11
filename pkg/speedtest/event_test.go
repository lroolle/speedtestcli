package speedtest

import "testing"

func TestEmit_NilSink(t *testing.T) {
	EmitMeta(nil, &ConnectionInfo{ClientIP: "1.2.3.4"})
	EmitLatency(nil, 0, 5.0)
	EmitSample(nil, &SampleResult{Direction: "download"})
	EmitResult(nil, &Result{})
	EmitReport(nil, &Report{})
	EmitError(nil, "test", "msg")
}

func TestEmitMeta(t *testing.T) {
	var got Event
	sink := func(e Event) { got = e }
	info := &ConnectionInfo{ClientIP: "1.2.3.4"}
	EmitMeta(sink, info)
	if got.Type != "meta" {
		t.Errorf("expected type 'meta', got %q", got.Type)
	}
	if got.Meta.ClientIP != "1.2.3.4" {
		t.Errorf("expected clientIP '1.2.3.4', got %q", got.Meta.ClientIP)
	}
	if got.Timestamp.IsZero() {
		t.Error("timestamp should be set")
	}
}

func TestEmitLatency(t *testing.T) {
	var got Event
	sink := func(e Event) { got = e }
	EmitLatency(sink, 3, 12.5)
	if got.Type != "latency" {
		t.Errorf("expected type 'latency', got %q", got.Type)
	}
	if got.Latency.Index != 3 || got.Latency.Ms != 12.5 {
		t.Errorf("unexpected latency: %+v", got.Latency)
	}
}

func TestEmitSample(t *testing.T) {
	var got Event
	sink := func(e Event) { got = e }
	EmitSample(sink, &SampleResult{Direction: "upload", BitsPerSec: 100})
	if got.Type != "sample" {
		t.Errorf("expected type 'sample', got %q", got.Type)
	}
	if got.Sample.Direction != "upload" {
		t.Errorf("expected direction 'upload', got %q", got.Sample.Direction)
	}
}

func TestEmitResult(t *testing.T) {
	var got Event
	sink := func(e Event) { got = e }
	EmitResult(sink, &Result{Backend: "test"})
	if got.Type != "result" {
		t.Errorf("expected type 'result', got %q", got.Type)
	}
}

func TestEmitReport(t *testing.T) {
	var got Event
	sink := func(e Event) { got = e }
	EmitReport(sink, &Report{Preset: "quick"})
	if got.Type != "report" {
		t.Errorf("expected type 'report', got %q", got.Type)
	}
}

func TestEmitError(t *testing.T) {
	var got Event
	sink := func(e Event) { got = e }
	EmitError(sink, "download", "connection refused")
	if got.Type != "error" {
		t.Errorf("expected type 'error', got %q", got.Type)
	}
	if got.Error.Phase != "download" {
		t.Errorf("expected phase 'download', got %q", got.Error.Phase)
	}
}

func TestTaggedSink(t *testing.T) {
	var got Event
	base := func(e Event) { got = e }
	tagged := TaggedSink(base, "cloudflare")
	EmitLatency(tagged, 0, 5.0)
	if got.Backend != "cloudflare" {
		t.Errorf("expected backend 'cloudflare', got %q", got.Backend)
	}
}

func TestTaggedSink_Nil(t *testing.T) {
	tagged := TaggedSink(nil, "test")
	if tagged != nil {
		t.Error("TaggedSink(nil) should return nil")
	}
}

func TestRedactProxyURL_NoAuth(t *testing.T) {
	got := RedactProxyURL("http://proxy.example.com:8080")
	if got != "http://proxy.example.com:8080" {
		t.Errorf("expected no change, got %q", got)
	}
}

func TestRedactProxyURL_WithAuth(t *testing.T) {
	got := RedactProxyURL("http://admin:s3cret@proxy.corp.com:8080")
	if got == "" {
		t.Fatal("expected non-empty result")
	}
	if contains(got, "admin") || contains(got, "s3cret") {
		t.Errorf("credentials should be redacted, got %q", got)
	}
	if !contains(got, "proxy.corp.com:8080") {
		t.Errorf("host should be preserved, got %q", got)
	}
}

func TestRedactProxyURL_Empty(t *testing.T) {
	got := RedactProxyURL("")
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestRedactProxyURL_Malformed(t *testing.T) {
	got := RedactProxyURL("://not-a-url")
	if got != "[set]" {
		t.Errorf("expected '[set]', got %q", got)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestGenerateID(t *testing.T) {
	id := GenerateID()
	if len(id) != 32 {
		t.Errorf("expected 32-char hex ID, got %d chars: %q", len(id), id)
	}
	id2 := GenerateID()
	if id == id2 {
		t.Error("two generated IDs should differ")
	}
}
