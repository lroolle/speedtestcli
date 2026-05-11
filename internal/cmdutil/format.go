package cmdutil

import "fmt"

func FormatBitsPerSec(bps uint64) string {
	switch {
	case bps >= 1_000_000_000:
		return fmt.Sprintf("%.2f Gbps", float64(bps)/1_000_000_000)
	case bps >= 1_000_000:
		return fmt.Sprintf("%.2f Mbps", float64(bps)/1_000_000)
	case bps >= 1_000:
		return fmt.Sprintf("%.2f Kbps", float64(bps)/1_000)
	default:
		return fmt.Sprintf("%d bps", bps)
	}
}

func FormatBytes(b uint64) string {
	switch {
	case b >= 1_000_000_000:
		return fmt.Sprintf("%.2f GB", float64(b)/1_000_000_000)
	case b >= 1_000_000:
		return fmt.Sprintf("%.2f MB", float64(b)/1_000_000)
	case b >= 1_000:
		return fmt.Sprintf("%.2f KB", float64(b)/1_000)
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func FormatMs(ms float64) string {
	if ms >= 1000 {
		return fmt.Sprintf("%.2fs", ms/1000)
	}
	return fmt.Sprintf("%.2fms", ms)
}
