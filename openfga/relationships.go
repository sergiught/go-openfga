package openfga

import (
	"context"
	"encoding/json"
	"iter"
	"net/http"
)

// fillDefaults populates modelID and consistency from the client's defaults
// and per-call options, but only when the caller left them empty. It operates
// on local copies, never on the caller's original request struct.
func (s *RelationshipsService) fillDefaults(opts []RequestOption, modelID *string, cons *ConsistencyPreference) {
	rc := newRequestConfig()
	applyOptions(rc, opts)
	if *modelID == "" {
		*modelID = s.client.modelFor(rc)
	}
	if *cons == "" && rc.consistency != "" {
		*cons = rc.consistency
	}
}

// Check tests whether a user has a specific relation on an object.
// It returns the check outcome, the raw HTTP response, and any error.
func (s *RelationshipsService) Check(ctx context.Context, req *CheckRequest, opts ...RequestOption) (*CheckResponse, *Response, error) {
	r := *req
	s.fillDefaults(opts, &r.AuthorizationModelID, &r.Consistency)
	out := new(CheckResponse)
	resp, err := s.client.doStorePost(ctx, "/check", &r, out, opts)
	return out, resp, err
}

// BatchCheck runs multiple relationship checks in a single request. Results in
// BatchCheckResponse.Result are keyed by the CorrelationID of each item.
func (s *RelationshipsService) BatchCheck(ctx context.Context, req *BatchCheckRequest, opts ...RequestOption) (*BatchCheckResponse, *Response, error) {
	r := *req
	s.fillDefaults(opts, &r.AuthorizationModelID, &r.Consistency)
	out := new(BatchCheckResponse)
	resp, err := s.client.doStorePost(ctx, "/batch-check", &r, out, opts)
	return out, resp, err
}

// Expand returns the userset tree that proves a relationship.
func (s *RelationshipsService) Expand(ctx context.Context, req *ExpandRequest, opts ...RequestOption) (*ExpandResponse, *Response, error) {
	r := *req
	s.fillDefaults(opts, &r.AuthorizationModelID, &r.Consistency)
	out := new(ExpandResponse)
	resp, err := s.client.doStorePost(ctx, "/expand", &r, out, opts)
	return out, resp, err
}

// ListObjects returns all objects of a given type that a user has a specific
// relation with.
func (s *RelationshipsService) ListObjects(ctx context.Context, req *ListObjectsRequest, opts ...RequestOption) (*ListObjectsResponse, *Response, error) {
	r := *req
	s.fillDefaults(opts, &r.AuthorizationModelID, &r.Consistency)
	out := new(ListObjectsResponse)
	resp, err := s.client.doStorePost(ctx, "/list-objects", &r, out, opts)
	return out, resp, err
}

// ListUsers returns all users who have a specific relation with a given object.
func (s *RelationshipsService) ListUsers(ctx context.Context, req *ListUsersRequest, opts ...RequestOption) (*ListUsersResponse, *Response, error) {
	r := *req
	s.fillDefaults(opts, &r.AuthorizationModelID, &r.Consistency)
	out := new(ListUsersResponse)
	resp, err := s.client.doStorePost(ctx, "/list-users", &r, out, opts)
	return out, resp, err
}

// StreamedListObjects streams matching objects, decoding the NDJSON response
// lazily. The HTTP connection stays open until iteration ends or the caller breaks.
// Each yielded error value is non-nil only on failure; on success it is nil.
func (s *RelationshipsService) StreamedListObjects(ctx context.Context, req *ListObjectsRequest, opts ...RequestOption) iter.Seq2[StreamedListObjectsResponse, error] {
	r := *req
	s.fillDefaults(opts, &r.AuthorizationModelID, &r.Consistency)
	return func(yield func(StreamedListObjectsResponse, error) bool) {
		rc := newRequestConfig()
		applyOptions(rc, opts)
		store, err := s.client.storeFor(rc)
		if err != nil {
			yield(StreamedListObjectsResponse{}, err)
			return
		}
		httpReq, err := s.client.newRequest(ctx, http.MethodPost, "/stores/"+store+"/streamed-list-objects", &r, rc.header)
		if err != nil {
			yield(StreamedListObjectsResponse{}, err)
			return
		}
		resp, err := s.client.BareDo(httpReq)
		if err != nil {
			yield(StreamedListObjectsResponse{}, err)
			return
		}
		defer func() { _ = resp.Body.Close() }()
		dec := json.NewDecoder(resp.Body)
		for dec.More() {
			var env streamedEnvelope
			if err := dec.Decode(&env); err != nil {
				yield(StreamedListObjectsResponse{}, err)
				return
			}
			if !yield(env.Result, nil) {
				return
			}
		}
	}
}
