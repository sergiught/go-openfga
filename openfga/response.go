package openfga

import "net/http"

// Response wraps the raw *http.Response and adds OpenFGA pagination metadata.
type Response struct {
	*http.Response
	// ContinuationToken is populated from the decoded body of paginated
	// endpoints; empty when there are no further pages.
	ContinuationToken string
}

// continuationTokener lets Do lift the body's token onto Response.
type continuationTokener interface {
	continuationToken() string
}

func newResponse(r *http.Response) *Response { return &Response{Response: r} }
