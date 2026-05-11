package cfbackend

import (
	"net/http"
	"testing"
	"time"
)

func TestParseServerTiming_Valid(t *testing.T) {
	resp := &http.Response{Header: http.Header{}}
	resp.Header.Set("Server-Timing", "cfRequestDuration;dur=12.345")
	got := parseServerTiming(resp)
	expected := time.Duration(12345 * time.Microsecond)
	if got != expected {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestParseServerTiming_Missing(t *testing.T) {
	resp := &http.Response{Header: http.Header{}}
	got := parseServerTiming(resp)
	if got != 0 {
		t.Errorf("expected 0, got %v", got)
	}
}

func TestParseServerTiming_Malformed(t *testing.T) {
	resp := &http.Response{Header: http.Header{}}
	resp.Header.Set("Server-Timing", "someOtherThing;dur=abc")
	got := parseServerTiming(resp)
	if got != 0 {
		t.Errorf("expected 0, got %v", got)
	}
}

func TestParseServerTiming_Zero(t *testing.T) {
	resp := &http.Response{Header: http.Header{}}
	resp.Header.Set("Server-Timing", "cfRequestDuration;dur=0")
	got := parseServerTiming(resp)
	if got != 0 {
		t.Errorf("expected 0, got %v", got)
	}
}
