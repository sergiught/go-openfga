package openfga

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// Write writes or deletes relationship tuples in the given store.
// If AuthorizationModelID is unset in req, it is filled from the client's
// configured model ID (or the per-call WithAuthorizationModel override).
func (s *TuplesService) Write(ctx context.Context, req *WriteRequest, opts ...RequestOption) (*Response, error) {
	rc := newRequestConfig()
	applyOptions(rc, opts)
	store, err := s.client.storeFor(rc)
	if err != nil {
		return nil, err
	}
	r := *req
	if r.AuthorizationModelID == "" {
		r.AuthorizationModelID = s.client.modelFor(rc)
	}
	httpReq, err := s.client.newRequest(ctx, http.MethodPost, "/stores/"+store+"/write", &r, rc.header)
	if err != nil {
		return nil, err
	}
	return s.client.Do(httpReq, nil)
}

// Read returns a page of relationship tuples from the given store.
// If Consistency is unset in req, it is filled from the per-call
// WithConsistency option.
func (s *TuplesService) Read(ctx context.Context, req *ReadRequest, opts ...RequestOption) (*ReadResponse, *Response, error) {
	rc := newRequestConfig()
	applyOptions(rc, opts)
	r := *req
	if r.Consistency == "" {
		r.Consistency = s.client.consistencyFor(rc)
	}
	store, err := s.client.storeFor(rc)
	if err != nil {
		return nil, nil, err
	}
	httpReq, err := s.client.newRequest(ctx, http.MethodPost, "/stores/"+store+"/read", &r, rc.header)
	if err != nil {
		return nil, nil, err
	}
	out := new(ReadResponse)
	resp, err := s.client.Do(httpReq, out)
	return out, resp, err
}

// ReadChanges returns a page of changelog entries (tuple writes/deletes) for
// the given store. Filtering and pagination are controlled via opts.
// Parameters are sent as query-string parameters: type, page_size,
// continuation_token, start_time.
func (s *TuplesService) ReadChanges(ctx context.Context, opts *ReadChangesOptions, ropts ...RequestOption) (*ReadChangesResponse, *Response, error) {
	rc := newRequestConfig()
	applyOptions(rc, ropts)
	store, err := s.client.storeFor(rc)
	if err != nil {
		return nil, nil, err
	}
	path := "/stores/" + store + "/changes"
	if opts != nil {
		q := url.Values{}
		if opts.Type != "" {
			q.Set("type", opts.Type)
		}
		if opts.PageSize > 0 {
			q.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.ContinuationToken != "" {
			q.Set("continuation_token", opts.ContinuationToken)
		}
		if opts.StartTime != "" {
			q.Set("start_time", opts.StartTime)
		}
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
	}
	httpReq, err := s.client.newRequest(ctx, http.MethodGet, path, nil, rc.header)
	if err != nil {
		return nil, nil, err
	}
	out := new(ReadChangesResponse)
	resp, err := s.client.Do(httpReq, out)
	return out, resp, err
}
