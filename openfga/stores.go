package openfga

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// Create creates a new store.
func (s *StoresService) Create(ctx context.Context, req *CreateStoreRequest, opts ...RequestOption) (*Store, error) {
	rc := newRequestConfig()
	applyOptions(rc, opts)
	httpReq, err := s.client.newRequest(ctx, http.MethodPost, "/stores", req, rc.header)
	if err != nil {
		return nil, err
	}
	store := new(Store)
	err = s.client.do(httpReq, store, rc)
	return store, err
}

// List returns a page of stores.
func (s *StoresService) List(ctx context.Context, opts *ListStoresOptions, ropts ...RequestOption) (*ListStoresResponse, error) {
	rc := newRequestConfig()
	applyOptions(rc, ropts)
	path := "/stores"
	if opts != nil {
		q := url.Values{}
		if opts.PageSize > 0 {
			q.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.ContinuationToken != "" {
			q.Set("continuation_token", opts.ContinuationToken)
		}
		if opts.Name != "" {
			q.Set("name", opts.Name)
		}
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
	}
	httpReq, err := s.client.newRequest(ctx, http.MethodGet, path, nil, rc.header)
	if err != nil {
		return nil, err
	}
	out := new(ListStoresResponse)
	err = s.client.do(httpReq, out, rc)
	return out, err
}

// Get retrieves a store by ID.
func (s *StoresService) Get(ctx context.Context, storeID string, opts ...RequestOption) (*Store, error) {
	rc := newRequestConfig()
	applyOptions(rc, opts)
	httpReq, err := s.client.newRequest(ctx, http.MethodGet, "/stores/"+storeID, nil, rc.header)
	if err != nil {
		return nil, err
	}
	store := new(Store)
	err = s.client.do(httpReq, store, rc)
	return store, err
}

// Delete removes a store by ID.
func (s *StoresService) Delete(ctx context.Context, storeID string, opts ...RequestOption) error {
	rc := newRequestConfig()
	applyOptions(rc, opts)
	httpReq, err := s.client.newRequest(ctx, http.MethodDelete, "/stores/"+storeID, nil, rc.header)
	if err != nil {
		return err
	}
	return s.client.do(httpReq, nil, rc)
}
