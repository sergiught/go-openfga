package openfga

import "net/http"

// requestIDHeader carries the OpenFGA server-side request correlation ID.
const requestIDHeader = "Fga-Request-Id"

// Response wraps the raw *http.Response and adds OpenFGA pagination metadata.
type Response struct {
	*http.Response
	// ContinuationToken is populated from the decoded body of paginated
	// endpoints; empty when there are no further pages.
	ContinuationToken string
}

// RequestID returns the OpenFGA request correlation ID from the response
// headers, or "" if absent. Quote it when reporting an issue to correlate with
// server logs.
func (r *Response) RequestID() string {
	if r == nil || r.Response == nil {
		return ""
	}
	return r.Header.Get(requestIDHeader)
}

// continuationTokener lets Do lift the body's token onto Response.
type continuationTokener interface {
	continuationToken() string
}

func newResponse(r *http.Response) *Response { return &Response{Response: r} }
