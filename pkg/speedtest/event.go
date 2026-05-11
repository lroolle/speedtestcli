package speedtest

import "time"

type EventSink func(event Event)

type Event struct {
	Type      string          `json:"type"`
	Backend   string          `json:"backend,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	Meta      *ConnectionInfo `json:"meta,omitempty"`
	Sample    *SampleResult   `json:"sample,omitempty"`
	Latency   *LatencyProbe   `json:"latency,omitempty"`
	Result    *Result         `json:"result,omitempty"`
	Report    *Report         `json:"report,omitempty"`
	Error     *ErrorInfo      `json:"error,omitempty"`
}

type LatencyProbe struct {
	Index int     `json:"index"`
	Ms    float64 `json:"ms"`
}

type ErrorInfo struct {
	Message string `json:"message"`
	Phase   string `json:"phase"`
}

func emit(sink EventSink, e Event) {
	if sink != nil {
		e.Timestamp = time.Now()
		sink(e)
	}
}

func EmitMeta(sink EventSink, info *ConnectionInfo) {
	emit(sink, Event{Type: "meta", Meta: info})
}

func EmitLatency(sink EventSink, index int, ms float64) {
	emit(sink, Event{Type: "latency", Latency: &LatencyProbe{Index: index, Ms: ms}})
}

func EmitSample(sink EventSink, s *SampleResult) {
	emit(sink, Event{Type: "sample", Sample: s})
}

func EmitResult(sink EventSink, r *Result) {
	emit(sink, Event{Type: "result", Result: r})
}

func EmitReport(sink EventSink, r *Report) {
	emit(sink, Event{Type: "report", Report: r})
}

func EmitError(sink EventSink, phase, msg string) {
	emit(sink, Event{Type: "error", Error: &ErrorInfo{Message: msg, Phase: phase}})
}

func TaggedSink(sink EventSink, backend string) EventSink {
	if sink == nil {
		return nil
	}
	return func(e Event) {
		e.Backend = backend
		sink(e)
	}
}
