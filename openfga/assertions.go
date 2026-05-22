package openfga

import (
	"context"
	"net/http"
)

// Write replaces the assertions for the given authorization model ID.
// It issues a PUT request to /stores/{store}/assertions/{modelID}.
func (s *AssertionsService) Write(ctx context.Context, modelID string, req *WriteAssertionsRequest, opts ...RequestOption) (*Response, error) {
	rc := newRequestConfig()
	applyOptions(rc, opts)
	store, err := s.client.storeFor(rc)
	if err != nil {
		return nil, err
	}
	httpReq, err := s.client.newRequest(ctx, http.MethodPut, "/stores/"+store+"/assertions/"+modelID, req, rc.header)
	if err != nil {
		return nil, err
	}
	return s.client.Do(httpReq, nil)
}

// Read retrieves the assertions for the given authorization model ID.
// It issues a GET request to /stores/{store}/assertions/{modelID}.
func (s *AssertionsService) Read(ctx context.Context, modelID string, opts ...RequestOption) (*ReadAssertionsResponse, *Response, error) {
	rc := newRequestConfig()
	applyOptions(rc, opts)
	store, err := s.client.storeFor(rc)
	if err != nil {
		return nil, nil, err
	}
	httpReq, err := s.client.newRequest(ctx, http.MethodGet, "/stores/"+store+"/assertions/"+modelID, nil, rc.header)
	if err != nil {
		return nil, nil, err
	}
	out := new(ReadAssertionsResponse)
	resp, err := s.client.Do(httpReq, out)
	return out, resp, err
}
