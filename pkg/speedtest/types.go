package speedtest

import (
	"net/url"
	"time"
)

type Result struct {
	ID            string           `json:"id"`
	Timestamp     time.Time        `json:"timestamp"`
	DurationS     float64          `json:"duration_s"`
	Status        string           `json:"status"`
	Backend       string           `json:"backend"`
	Preset        string           `json:"preset"`
	Granularity   string           `json:"granularity"`
	Connection    ConnectionInfo   `json:"connection"`
	Latency       LatencyResult    `json:"latency"`
	Download      ThroughputResult `json:"download"`
	Upload        ThroughputResult `json:"upload"`
	Errors        []string         `json:"errors,omitempty"`
	ProxyDetected *ProxyInfo       `json:"proxy_detected,omitempty"`
}

type ProxyInfo struct {
	HTTPProxy  string `json:"http_proxy,omitempty"`
	HTTPSProxy string `json:"https_proxy,omitempty"`
	AllProxy   string `json:"all_proxy,omitempty"`
}

func RedactProxyURL(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "[set]"
	}
	if u.User != nil {
		u.User = url.UserPassword("***", "***")
	}
	return u.String()
}

type Report struct {
	Timestamp time.Time `json:"timestamp"`
	DurationS float64   `json:"duration_s"`
	Preset    string    `json:"preset"`
	Results   []Result  `json:"results"`
}

type ConnectionInfo struct {
	ClientIP  string   `json:"client_ip"`
	ASN       uint64   `json:"asn"`
	ASOrg     string   `json:"as_organization"`
	Country   string   `json:"country"`
	Region    string   `json:"region,omitempty"`
	City      string   `json:"city"`
	Latitude  float64  `json:"latitude,omitempty"`
	Longitude float64  `json:"longitude,omitempty"`
	Colo      ColoInfo `json:"colo,omitempty"`
}

type ColoInfo struct {
	IATA    string `json:"iata,omitempty"`
	City    string `json:"city,omitempty"`
	Country string `json:"country,omitempty"`
}

type LatencyResult struct {
	Samples  int          `json:"samples"`
	Stats    LatencyStats `json:"stats"`
	JitterMs float64      `json:"jitter_ms"`
	RawMs    []float64    `json:"raw_ms,omitempty"`
}

type LatencyStats struct {
	MinMs    float64 `json:"min_ms"`
	P25Ms    float64 `json:"p25_ms"`
	MedianMs float64 `json:"median_ms"`
	MeanMs   float64 `json:"mean_ms"`
	P75Ms    float64 `json:"p75_ms"`
	P90Ms    float64 `json:"p90_ms"`
	MaxMs    float64 `json:"max_ms"`
}

type ThroughputResult struct {
	BitsPerSec      uint64          `json:"bits_per_sec"`
	BytesTotal      uint64          `json:"bytes_total"`
	Samples         int             `json:"samples"`
	ExpectedSamples int             `json:"expected_samples,omitempty"`
	Truncated       bool            `json:"truncated,omitempty"`
	Stats           ThroughputStats `json:"stats"`
	RawBps          []uint64        `json:"raw_bps,omitempty"`
}

type ThroughputStats struct {
	MinBps    uint64 `json:"min_bps"`
	P10Bps    uint64 `json:"p10_bps"`
	MedianBps uint64 `json:"median_bps"`
	MeanBps   uint64 `json:"mean_bps"`
	P90Bps    uint64 `json:"p90_bps"`
	MaxBps    uint64 `json:"max_bps"`
}

type TimingTrace struct {
	TTFB         float64 `json:"ttfb_ms"`
	ServerTiming float64 `json:"server_timing_ms"`
	Latency      float64 `json:"latency_ms"`
	Total        float64 `json:"total_ms"`
	ConnReused   bool    `json:"conn_reused"`
}

type SampleResult struct {
	Direction     string      `json:"direction"`
	Bytes         uint64      `json:"bytes"`
	BitsPerSec    uint64      `json:"bits_per_sec"`
	Timing        TimingTrace `json:"timing"`
	RequestBytes  uint64      `json:"request_bytes"`
	ResponseBytes uint64      `json:"response_bytes"`
}
