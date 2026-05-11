package cfbackend

import "net/http"

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewHTTPClient() *http.Client {
	// No global timeout — each request uses the context deadline from the
	// test plan timeout. A global timeout here would conflict with large
	// payloads in thorough mode (100MB downloads on slow connections).
	return &http.Client{}
}
