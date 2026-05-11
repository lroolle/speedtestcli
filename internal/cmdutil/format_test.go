package cmdutil

import "testing"

func TestFormatBitsPerSec(t *testing.T) {
	tests := []struct {
		bps  uint64
		want string
	}{
		{500, "500 bps"},
		{56_000, "56.00 Kbps"},
		{524_288_000, "524.29 Mbps"},
		{1_500_000_000, "1.50 Gbps"},
	}
	for _, tt := range tests {
		if got := FormatBitsPerSec(tt.bps); got != tt.want {
			t.Errorf("FormatBitsPerSec(%d) = %q, want %q", tt.bps, got, tt.want)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		b    uint64
		want string
	}{
		{500, "500 B"},
		{10_500_000, "10.50 MB"},
		{2_500_000_000, "2.50 GB"},
	}
	for _, tt := range tests {
		if got := FormatBytes(tt.b); got != tt.want {
			t.Errorf("FormatBytes(%d) = %q, want %q", tt.b, got, tt.want)
		}
	}
}

func TestFormatMs(t *testing.T) {
	tests := []struct {
		ms   float64
		want string
	}{
		{5.23, "5.23ms"},
		{2500, "2.50s"},
	}
	for _, tt := range tests {
		if got := FormatMs(tt.ms); got != tt.want {
			t.Errorf("FormatMs(%f) = %q, want %q", tt.ms, got, tt.want)
		}
	}
}
