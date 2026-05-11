package speedtest

import (
	"math"
	"sort"
)

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}
	rank := p / 100.0 * float64(len(sorted)-1)
	lower := int(math.Floor(rank))
	upper := lower + 1
	if upper >= len(sorted) {
		return sorted[len(sorted)-1]
	}
	frac := rank - float64(lower)
	return sorted[lower] + frac*(sorted[upper]-sorted[lower])
}

func percentileUint64(sorted []uint64, p float64) uint64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}
	rank := p / 100.0 * float64(len(sorted)-1)
	lower := int(math.Floor(rank))
	upper := lower + 1
	if upper >= len(sorted) {
		return sorted[len(sorted)-1]
	}
	frac := rank - float64(lower)
	return uint64(float64(sorted[lower]) + frac*float64(sorted[upper]-sorted[lower]))
}

func meanFloat64(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	var sum float64
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func meanUint64(data []uint64) uint64 {
	if len(data) == 0 {
		return 0
	}
	var sum uint64
	for _, v := range data {
		sum += v
	}
	return sum / uint64(len(data))
}

func jitter(samples []float64) float64 {
	if len(samples) < 2 {
		return 0
	}
	var sum float64
	for i := 1; i < len(samples); i++ {
		sum += math.Abs(samples[i] - samples[i-1])
	}
	return sum / float64(len(samples)-1)
}

func ComputeLatencyStats(raw []float64) LatencyStats {
	if len(raw) == 0 {
		return LatencyStats{}
	}
	sorted := make([]float64, len(raw))
	copy(sorted, raw)
	sort.Float64s(sorted)
	return LatencyStats{
		MinMs:    sorted[0],
		P25Ms:    percentile(sorted, 25),
		MedianMs: percentile(sorted, 50),
		MeanMs:   meanFloat64(raw),
		P75Ms:    percentile(sorted, 75),
		P90Ms:    percentile(sorted, 90),
		MaxMs:    sorted[len(sorted)-1],
	}
}

func ComputeLatencyResult(raw []float64, includeRaw bool) LatencyResult {
	r := LatencyResult{
		Samples:  len(raw),
		Stats:    ComputeLatencyStats(raw),
		JitterMs: jitter(raw),
	}
	if includeRaw {
		r.RawMs = raw
	}
	return r
}

func ComputeThroughputStats(raw []uint64) ThroughputStats {
	if len(raw) == 0 {
		return ThroughputStats{}
	}
	sorted := make([]uint64, len(raw))
	copy(sorted, raw)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	return ThroughputStats{
		MinBps:    sorted[0],
		P10Bps:    percentileUint64(sorted, 10),
		MedianBps: percentileUint64(sorted, 50),
		MeanBps:   meanUint64(raw),
		P90Bps:    percentileUint64(sorted, 90),
		MaxBps:    sorted[len(sorted)-1],
	}
}

func ComputeThroughputResult(rawBps []uint64, bytesTotal uint64, includeRaw bool) ThroughputResult {
	stats := ComputeThroughputStats(rawBps)
	r := ThroughputResult{
		BitsPerSec: stats.P90Bps,
		BytesTotal: bytesTotal,
		Samples:    len(rawBps),
		Stats:      stats,
	}
	if includeRaw {
		r.RawBps = rawBps
	}
	return r
}
