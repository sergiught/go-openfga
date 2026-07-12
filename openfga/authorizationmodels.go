package openfga

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

// Write creates a new authorization model in the store and returns its ID.
func (s *AuthorizationModelsService) Write(ctx context.Context, req *WriteAuthorizationModelRequest, opts ...RequestOption) (*WriteAuthorizationModelResponse, error) {
	rc := newRequestConfig()
	applyOptions(rc, opts)
	store, err := s.client.storeFor(rc)
	if err != nil {
		return nil, err
	}
	httpReq, err := s.client.newRequest(ctx, http.MethodPost, "/stores/"+store+"/authorization-models", req, rc.header)
	if err != nil {
		return nil, err
	}
	out := new(WriteAuthorizationModelResponse)
	err = s.client.do(httpReq, out, rc)
	return out, err
}

// List returns a page of authorization models for the store.
func (s *AuthorizationModelsService) List(ctx context.Context, opts *ListModelsOptions, ropts ...RequestOption) (*ListAuthorizationModelsResponse, error) {
	rc := newRequestConfig()
	applyOptions(rc, ropts)
	store, err := s.client.storeFor(rc)
	if err != nil {
		return nil, err
	}
	path := "/stores/" + store + "/authorization-models"
	if opts != nil {
		q := url.Values{}
		if opts.PageSize > 0 {
			q.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.ContinuationToken != "" {
			q.Set("continuation_token", opts.ContinuationToken)
		}
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
	}
	httpReq, err := s.client.newRequest(ctx, http.MethodGet, path, nil, rc.header)
	if err != nil {
		return nil, err
	}
	out := new(ListAuthorizationModelsResponse)
	err = s.client.do(httpReq, out, rc)
	return out, err
}

// Get retrieves a single authorization model by ID.
func (s *AuthorizationModelsService) Get(ctx context.Context, id string, opts ...RequestOption) (*AuthorizationModel, error) {
	rc := newRequestConfig()
	applyOptions(rc, opts)
	store, err := s.client.storeFor(rc)
	if err != nil {
		return nil, err
	}
	httpReq, err := s.client.newRequest(ctx, http.MethodGet, "/stores/"+store+"/authorization-models/"+id, nil, rc.header)
	if err != nil {
		return nil, err
	}
	out := new(readAuthorizationModelResponse)
	if err := s.client.do(httpReq, out, rc); err != nil {
		return nil, err
	}
	return &out.AuthorizationModel, nil
}

// ReadLatest returns the most recently created authorization model by fetching
// one page of size 1. It returns an error if no models exist in the store.
func (s *AuthorizationModelsService) ReadLatest(ctx context.Context, opts ...RequestOption) (*AuthorizationModel, error) {
	page, err := s.List(ctx, &ListModelsOptions{PageSize: 1}, opts...)
	if err != nil {
		return nil, err
	}
	if len(page.AuthorizationModels) == 0 {
		return nil, errors.New("openfga: no authorization models found")
	}
	return &page.AuthorizationModels[0], nil
}
