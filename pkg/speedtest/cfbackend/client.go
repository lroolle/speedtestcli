package cfbackend

import "net/http"

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// No global timeout — each request uses the context deadline from the
// test plan timeout. A global timeout here would conflict with large
// payloads in thorough mode (100MB downloads on slow connections).
func NewHTTPClient() *http.Client {
	return &http.Client{}
}

func NewDirectHTTPClient() *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.Proxy = nil
	return &http.Client{Transport: t}
}
