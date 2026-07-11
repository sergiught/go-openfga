package openfga

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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
	if *cons == "" {
		*cons = s.client.consistencyFor(rc)
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

// newCorrelationID returns a random 16-byte hex string for auto-populating
// batch-check items that lack a correlation ID.
func newCorrelationID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

// BatchCheckAll runs many checks by splitting req.Checks into chunks of at most
// WithMaxChecksPerBatch (default 50, the server maximum) and issuing the native
// /batch-check requests concurrently (bounded by WithMaxParallel). Results from
// every chunk are merged into a single map keyed by correlation ID. Items with
// an empty CorrelationID get a generated one. Duplicate caller-supplied
// correlation IDs are rejected before any request, since the merged map would
// collide. Any failing chunk request aborts the call with that error.
func (s *RelationshipsService) BatchCheckAll(ctx context.Context, req *BatchCheckRequest, opts ...RequestOption) (*BatchCheckResponse, error) {
	merged := &BatchCheckResponse{Result: map[string]BatchCheckSingleResult{}}
	if len(req.Checks) == 0 {
		return merged, nil
	}

	rc := newRequestConfig()
	applyOptions(rc, opts)
	store, err := s.client.storeFor(rc)
	if err != nil {
		return nil, err
	}

	modelID := req.AuthorizationModelID
	cons := req.Consistency
	s.fillDefaults(opts, &modelID, &cons)

	// Copy checks, populating/validating correlation IDs.
	checks := make([]BatchCheckItem, len(req.Checks))
	seen := make(map[string]struct{}, len(req.Checks))
	for i, item := range req.Checks {
		if item.CorrelationID == "" {
			id, err := newCorrelationID()
			if err != nil {
				return nil, err
			}
			item.CorrelationID = id
		} else if _, dup := seen[item.CorrelationID]; dup {
			return nil, fmt.Errorf("openfga: duplicate correlation_id %q in BatchCheckAll", item.CorrelationID)
		}
		seen[item.CorrelationID] = struct{}{}
		checks[i] = item
	}

	perBatch := resolvePositive(rc.maxChecksPerBatch, defaultMaxChecksPerBatch)
	maxPar := resolvePositive(rc.maxParallel, defaultMaxParallel)

	type span struct{ lo, hi int }
	var spans []span
	for lo := 0; lo < len(checks); lo += perBatch {
		hi := lo + perBatch
		if hi > len(checks) {
			hi = len(checks)
		}
		spans = append(spans, span{lo, hi})
	}

	chunkResults := make([]BatchCheckResponse, len(spans))
	errs := runParallel(ctx, len(spans), maxPar, func(i int) error {
		body := &BatchCheckRequest{
			Checks:               checks[spans[i].lo:spans[i].hi],
			AuthorizationModelID: modelID,
			Consistency:          cons,
		}
		httpReq, err := s.client.newRequest(ctx, http.MethodPost, "/stores/"+store+"/batch-check", body, rc.header)
		if err != nil {
			return err
		}
		if _, err := s.client.Do(httpReq, &chunkResults[i]); err != nil {
			return err
		}
		return nil
	})
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	for _, cr := range chunkResults {
		for k, v := range cr.Result {
			merged.Result[k] = v
		}
	}
	return merged, nil
}

// ListRelations reports which of req.Relations the user has on the object. It
// issues the checks through BatchCheckAll (one native /batch-check request per
// chunk, bounded by WithMaxParallel) and returns the allowed relations in the
// order they were supplied. Duplicate relations are rejected. Because it builds
// on the native batch-check endpoint, it requires OpenFGA >= 1.8.0.
func (s *RelationshipsService) ListRelations(ctx context.Context, req *ListRelationsRequest, opts ...RequestOption) ([]string, error) {
	if len(req.Relations) == 0 {
		return nil, nil
	}

	checks := make([]BatchCheckItem, len(req.Relations))
	seen := make(map[string]struct{}, len(req.Relations))
	for i, rel := range req.Relations {
		if _, dup := seen[rel]; dup {
			return nil, fmt.Errorf("openfga: duplicate relation %q in ListRelations", rel)
		}
		seen[rel] = struct{}{}
		checks[i] = BatchCheckItem{
			TupleKey:         CheckRequestTupleKey{User: req.User, Relation: rel, Object: req.Object},
			ContextualTuples: req.ContextualTuples,
			Context:          req.Context,
			CorrelationID:    rel,
		}
	}

	resp, err := s.BatchCheckAll(ctx, &BatchCheckRequest{
		Checks:               checks,
		AuthorizationModelID: req.AuthorizationModelID,
		Consistency:          req.Consistency,
	}, opts...)
	if err != nil {
		return nil, err
	}

	allowed := make([]string, 0, len(req.Relations))
	for _, rel := range req.Relations {
		if resp.Result[rel].Allowed {
			allowed = append(allowed, rel)
		}
	}
	return allowed, nil
}
