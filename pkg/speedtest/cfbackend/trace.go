package cfbackend

import (
	"crypto/tls"
	"net/http"
	"net/http/httptrace"
	"regexp"
	"strconv"
	"time"
)

var serverTimingRe = regexp.MustCompile(`cfRequestDuration;dur=([\d.]+)`)

type requestTiming struct {
	start         time.Time
	connectDone   time.Time
	tlsDone       time.Time
	wroteRequest  time.Time
	gotFirstByte  time.Time
	connReused    bool
}

func (rt *requestTiming) ttfb() time.Duration {
	base := rt.connectBase()
	if rt.gotFirstByte.IsZero() {
		return 0
	}
	return rt.gotFirstByte.Sub(base)
}

func (rt *requestTiming) connectBase() time.Time {
	if rt.connReused {
		return rt.start
	}
	if !rt.tlsDone.IsZero() {
		return rt.tlsDone
	}
	if !rt.connectDone.IsZero() {
		return rt.connectDone
	}
	return rt.start
}

func (rt *requestTiming) uploadDuration() time.Duration {
	base := rt.connectBase()
	if rt.wroteRequest.IsZero() {
		return 0
	}
	return rt.wroteRequest.Sub(base)
}

func (rt *requestTiming) totalDuration() time.Duration {
	return time.Since(rt.start)
}

func newClientTrace(rt *requestTiming) *httptrace.ClientTrace {
	rt.connReused = true
	return &httptrace.ClientTrace{
		ConnectDone: func(string, string, error) {
			rt.connectDone = time.Now()
			rt.connReused = false
		},
		TLSHandshakeDone: func(tls.ConnectionState, error) {
			rt.tlsDone = time.Now()
			rt.connReused = false
		},
		WroteRequest: func(httptrace.WroteRequestInfo) {
			rt.wroteRequest = time.Now()
		},
		GotFirstResponseByte: func() {
			rt.gotFirstByte = time.Now()
		},
	}
}

func parseServerTiming(resp *http.Response) time.Duration {
	header := resp.Header.Get("Server-Timing")
	if header == "" {
		return 0
	}
	match := serverTimingRe.FindStringSubmatch(header)
	if len(match) != 2 {
		return 0
	}
	ms, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0
	}
	return time.Duration(ms * float64(time.Millisecond))
}
