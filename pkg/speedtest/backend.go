package speedtest

import "context"

type Backend interface {
	Name() string
	Granularity() string
	FetchMeta(ctx context.Context) (*ConnectionInfo, error)
	MeasureLatency(ctx context.Context, count int, sink EventSink) (*LatencyResult, error)
	MeasureDownload(ctx context.Context, steps []TestStep, sink EventSink) (*ThroughputResult, error)
	MeasureUpload(ctx context.Context, steps []TestStep, sink EventSink) (*ThroughputResult, error)
}
