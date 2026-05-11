package cfbackend

import (
	"context"

	"github.com/lroolle/speedtestcli/pkg/speedtest"
)

type Backend struct {
	client  HTTPDoer
	baseURL string
	measID  string
}

type Option func(*Backend)

func WithHTTPDoer(d HTTPDoer) Option { return func(b *Backend) { b.client = d } }
func WithBaseURL(u string) Option    { return func(b *Backend) { b.baseURL = u } }
func WithMeasID(id string) Option    { return func(b *Backend) { b.measID = id } }

func New(opts ...Option) *Backend {
	b := &Backend{
		client:  NewHTTPClient(),
		baseURL: "https://speed.cloudflare.com",
		measID:  speedtest.GenerateID(),
	}
	for _, o := range opts {
		o(b)
	}
	return b
}

func (b *Backend) Name() string        { return "cloudflare" }
func (b *Backend) Granularity() string { return "per-sample" }

func (b *Backend) FetchMeta(ctx context.Context) (*speedtest.ConnectionInfo, error) {
	return fetchMeta(ctx, b.client, b.baseURL)
}

func (b *Backend) MeasureLatency(ctx context.Context, count int, sink speedtest.EventSink) (*speedtest.LatencyResult, error) {
	return measureLatency(ctx, b.client, b.baseURL, b.measID, count, sink)
}

func (b *Backend) MeasureDownload(ctx context.Context, steps []speedtest.TestStep, sink speedtest.EventSink) (*speedtest.ThroughputResult, error) {
	return measureDownload(ctx, b.client, b.baseURL, b.measID, steps, sink)
}

func (b *Backend) MeasureUpload(ctx context.Context, steps []speedtest.TestStep, sink speedtest.EventSink) (*speedtest.ThroughputResult, error) {
	return measureUpload(ctx, b.client, b.baseURL, b.measID, steps, sink)
}
