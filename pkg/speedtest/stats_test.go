package speedtest

import (
	"math"
	"testing"
)

func approxEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestPercentile_SingleValue(t *testing.T) {
	got := percentile([]float64{42.0}, 50)
	if got != 42.0 {
		t.Errorf("expected 42.0, got %f", got)
	}
}

func TestPercentile_Empty(t *testing.T) {
	got := percentile([]float64{}, 50)
	if got != 0 {
		t.Errorf("expected 0, got %f", got)
	}
}

func TestPercentile_Median(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5}
	got := percentile(data, 50)
	if got != 3.0 {
		t.Errorf("expected 3.0, got %f", got)
	}
}

func TestPercentile_P90(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	got := percentile(data, 90)
	if !approxEqual(got, 9.1, 0.01) {
		t.Errorf("expected ~9.1, got %f", got)
	}
}

func TestPercentile_P25(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5}
	got := percentile(data, 25)
	if got != 2.0 {
		t.Errorf("expected 2.0, got %f", got)
	}
}

func TestMeanFloat64(t *testing.T) {
	got := meanFloat64([]float64{2, 4, 6})
	if got != 4.0 {
		t.Errorf("expected 4.0, got %f", got)
	}
}

func TestMeanFloat64_Empty(t *testing.T) {
	got := meanFloat64([]float64{})
	if got != 0 {
		t.Errorf("expected 0, got %f", got)
	}
}

func TestJitter(t *testing.T) {
	samples := []float64{5.0, 8.0, 6.0, 9.0}
	got := jitter(samples)
	expected := (3.0 + 2.0 + 3.0) / 3.0
	if !approxEqual(got, expected, 0.001) {
		t.Errorf("expected %f, got %f", expected, got)
	}
}

func TestJitter_SingleSample(t *testing.T) {
	got := jitter([]float64{5.0})
	if got != 0 {
		t.Errorf("expected 0, got %f", got)
	}
}

func TestComputeLatencyStats(t *testing.T) {
	raw := []float64{10.0, 5.0, 15.0, 8.0, 12.0}
	stats := ComputeLatencyStats(raw)
	if stats.MinMs != 5.0 {
		t.Errorf("min: expected 5.0, got %f", stats.MinMs)
	}
	if stats.MaxMs != 15.0 {
		t.Errorf("max: expected 15.0, got %f", stats.MaxMs)
	}
	if stats.MedianMs != 10.0 {
		t.Errorf("median: expected 10.0, got %f", stats.MedianMs)
	}
	if stats.MeanMs != 10.0 {
		t.Errorf("mean: expected 10.0, got %f", stats.MeanMs)
	}
}

func TestComputeLatencyResult_IncludeRaw(t *testing.T) {
	raw := []float64{5.0, 10.0, 15.0}
	r := ComputeLatencyResult(raw, true)
	if r.Samples != 3 {
		t.Errorf("samples: expected 3, got %d", r.Samples)
	}
	if len(r.RawMs) != 3 {
		t.Errorf("raw_ms length: expected 3, got %d", len(r.RawMs))
	}
}

func TestComputeLatencyResult_ExcludeRaw(t *testing.T) {
	raw := []float64{5.0, 10.0, 15.0}
	r := ComputeLatencyResult(raw, false)
	if r.RawMs != nil {
		t.Errorf("raw_ms should be nil when includeRaw=false")
	}
}

func TestComputeThroughputStats(t *testing.T) {
	raw := []uint64{100, 200, 300, 400, 500, 600, 700, 800, 900, 1000}
	stats := ComputeThroughputStats(raw)
	if stats.MinBps != 100 {
		t.Errorf("min: expected 100, got %d", stats.MinBps)
	}
	if stats.MaxBps != 1000 {
		t.Errorf("max: expected 1000, got %d", stats.MaxBps)
	}
	if stats.MedianBps != 550 {
		t.Errorf("median: expected 550, got %d", stats.MedianBps)
	}
	if stats.P90Bps != 910 {
		t.Errorf("p90: expected 910, got %d", stats.P90Bps)
	}
}

func TestComputeThroughputResult_HeadlineIsP90(t *testing.T) {
	raw := []uint64{100, 200, 300, 400, 500, 600, 700, 800, 900, 1000}
	r := ComputeThroughputResult(raw, 5500, false)
	if r.BitsPerSec != r.Stats.P90Bps {
		t.Errorf("headline should be p90: expected %d, got %d", r.Stats.P90Bps, r.BitsPerSec)
	}
	if r.BytesTotal != 5500 {
		t.Errorf("bytes_total: expected 5500, got %d", r.BytesTotal)
	}
}

func TestPercentileUint64_Empty(t *testing.T) {
	got := percentileUint64([]uint64{}, 50)
	if got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}
