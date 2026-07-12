package openfga

import (
	"net/http"
	"strconv"
	"time"
)

// requestIDHeader carries the OpenFGA server-side request correlation ID.
const requestIDHeader = "Fga-Request-Id"

// queryDurationHeader reports how long the server spent evaluating the query,
// in milliseconds (as a possibly-fractional number).
const queryDurationHeader = "Fga-Query-Duration-Ms"

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

// QueryDuration returns the time the OpenFGA server reports it spent evaluating
// the query, parsed from the Fga-Query-Duration-Ms response header. This is the
// server-side evaluation cost only — it excludes network round-trip and any
// client-side retries, which the RequestObserver's elapsed argument measures.
// The bool is false when the header is absent or unparseable (e.g. on endpoints
// that do not report it), letting callers distinguish that from a genuine 0.
//
// Reach it from the (result, error) surface via OnResponse:
//
//	res, err := client.Relationships.Check(ctx, req,
//		openfga.OnResponse(func(r *openfga.Response) {
//			if d, ok := r.QueryDuration(); ok {
//				metrics.Observe("fga.query", d)
//			}
//		}))
func (r *Response) QueryDuration() (time.Duration, bool) {
	if r == nil || r.Response == nil {
		return 0, false
	}
	v := r.Header.Get(queryDurationHeader)
	if v == "" {
		return 0, false
	}
	ms, err := strconv.ParseFloat(v, 64)
	if err != nil || ms < 0 {
		return 0, false
	}
	return time.Duration(ms * float64(time.Millisecond)), true
}

// continuationTokener lets Do lift the body's token onto Response.
type continuationTokener interface {
	continuationToken() string
}

func newResponse(r *http.Response) *Response { return &Response{Response: r} }
