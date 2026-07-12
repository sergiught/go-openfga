package openfga

import "net/http"

// headerTransport applies static headers to every request. Static headers are
// applied only when not already set on the request; per-request headers take
// precedence. It clones the request to avoid mutating the caller's request.
type headerTransport struct {
	base   http.RoundTripper
	header http.Header
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r2 := req.Clone(req.Context())
	for k, vs := range t.header {
		if r2.Header.Get(k) != "" {
			continue
		}
		for _, v := range vs {
			r2.Header.Add(k, v)
		}
	}
	return t.base.RoundTrip(r2)
}
