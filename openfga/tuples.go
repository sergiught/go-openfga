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

// WriteTuples writes many tuples, chunking them into non-transactional /write
// requests issued in parallel. Use WithMaxPerChunk, WithMaxParallel,
// WithOnDuplicate, and WithTransaction to tune behavior. The returned response
// carries a per-tuple result (order matches keys); the top-level error is
// non-nil only when no request could be issued.
func (s *TuplesService) WriteTuples(ctx context.Context, keys []TupleKey, opts ...RequestOption) (*WriteTuplesResponse, error) {
	results, err := s.bulkWrite(ctx, keys, false, opts)
	return &WriteTuplesResponse{Writes: results}, err
}

// DeleteTuples deletes many tuples, chunking them into non-transactional
// /write requests issued in parallel. Use WithMaxPerChunk, WithMaxParallel,
// WithOnMissing, and WithTransaction to tune behavior.
func (s *TuplesService) DeleteTuples(ctx context.Context, keys []TupleKey, opts ...RequestOption) (*WriteTuplesResponse, error) {
	results, err := s.bulkWrite(ctx, keys, true, opts)
	return &WriteTuplesResponse{Deletes: results}, err
}

// bulkWrite backs WriteTuples/DeleteTuples. When del is true, keys are sent as
// deletes; otherwise as writes.
func (s *TuplesService) bulkWrite(ctx context.Context, keys []TupleKey, del bool, opts []RequestOption) ([]TupleResult, error) {
	results := make([]TupleResult, len(keys))
	for i, k := range keys {
		results[i] = TupleResult{TupleKey: k, Status: WriteStatusSuccess}
	}
	if len(keys) == 0 {
		return results, nil
	}

	rc := newRequestConfig()
	applyOptions(rc, opts)
	store, err := s.client.storeFor(rc)
	if err != nil {
		return results, err
	}
	modelID := s.client.modelFor(rc)

	buildBody := func(chunk []TupleKey) *WriteRequest {
		block := &WriteRequestTuples{TupleKeys: chunk}
		body := &WriteRequest{AuthorizationModelID: modelID}
		if del {
			block.OnMissing = rc.onMissing
			body.Deletes = block
		} else {
			block.OnDuplicate = rc.onDuplicate
			body.Writes = block
		}
		return body
	}

	send := func(chunk []TupleKey) error {
		req, err := s.client.newRequest(ctx, http.MethodPost, "/stores/"+store+"/write", buildBody(chunk), rc.header)
		if err != nil {
			return err
		}
		_, err = s.client.Do(req, nil)
		return err
	}

	// Transaction mode: a single request for all keys.
	if rc.transaction {
		if err := send(keys); err != nil {
			for i := range results {
				results[i].Status = WriteStatusFailure
				results[i].Err = err
			}
			return results, err
		}
		return results, nil
	}

	// Non-transactional: chunk and parallelize.
	perChunk := resolvePositive(rc.maxPerChunk, defaultMaxPerChunk)
	maxPar := resolvePositive(rc.maxParallel, defaultMaxParallel)

	type span struct{ lo, hi int }
	var spans []span
	for lo := 0; lo < len(keys); lo += perChunk {
		hi := lo + perChunk
		if hi > len(keys) {
			hi = len(keys)
		}
		spans = append(spans, span{lo, hi})
	}

	errs := runParallel(ctx, len(spans), maxPar, func(i int) error {
		return send(keys[spans[i].lo:spans[i].hi])
	})
	for i, e := range errs {
		if e == nil {
			continue
		}
		for j := spans[i].lo; j < spans[i].hi; j++ {
			results[j].Status = WriteStatusFailure
			results[j].Err = e
		}
	}
	return results, nil
}
