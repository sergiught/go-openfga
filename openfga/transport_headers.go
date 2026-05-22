package openfga

import "net/http"

// headerTransport applies static headers to every request. It clones the
// request to avoid mutating the caller's request.
type headerTransport struct {
	base   http.RoundTripper
	header http.Header
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r2 := req.Clone(req.Context())
	for k, vs := range t.header {
		for _, v := range vs {
			if r2.Header.Get(k) == "" {
				r2.Header.Set(k, v)
			} else {
				r2.Header.Add(k, v)
			}
		}
	}
	return t.base.RoundTrip(r2)
}
