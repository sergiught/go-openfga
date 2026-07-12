package openfga

import (
	"net/http"
	"time"
)

// RequestObserver is invoked once per HTTP attempt, after the request has been
// fully decorated with auth and static headers and the response (or error) is
// available. The elapsed argument measures that single attempt; with retries
// enabled an observer fires once per attempt. Use it for logging, metrics, or debug
// tracing without implementing an http.RoundTripper. It must not modify req or
// resp, and must not read resp.Body (doing so would consume it).
type RequestObserver func(req *http.Request, resp *http.Response, err error, elapsed time.Duration)

// observerTransport calls obs around each RoundTrip of base.
type observerTransport struct {
	base http.RoundTripper
	obs  RequestObserver
}

func (t *observerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := t.base.RoundTrip(req)
	t.obs(req, resp, err, time.Since(start))
	return resp, err
}

// WithRequestObserver registers a callback invoked once per HTTP attempt. It is
// the lightweight alternative to a custom transport for logging, metrics, or
// debug output. Ignored when WithHTTPClient supplies a full client.
func WithRequestObserver(obs RequestObserver) Option {
	return func(c *Client) { c.observer = obs }
}
