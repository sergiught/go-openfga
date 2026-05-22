package openfga

import (
	"context"
	"iter"
)

// All iterates every store across pages, fetching lazily one page at a time.
// It copies opts so the caller's struct is never mutated. Iteration stops when
// the server returns an empty continuation token, when the caller breaks, or
// when the server returns an error (which is yielded as the second value).
func (s *StoresService) All(ctx context.Context, opts *ListStoresOptions, ropts ...RequestOption) iter.Seq2[Store, error] {
	var o ListStoresOptions
	if opts != nil {
		o = *opts
	}
	return func(yield func(Store, error) bool) {
		for {
			page, _, err := s.List(ctx, &o, ropts...)
			if err != nil {
				yield(Store{}, err)
				return
			}
			for _, st := range page.Stores {
				if !yield(st, nil) {
					return
				}
			}
			if page.ContinuationToken == "" {
				return
			}
			o.ContinuationToken = page.ContinuationToken
		}
	}
}

// All iterates every authorization model across pages. It copies opts so the
// caller's struct is never mutated.
func (s *AuthorizationModelsService) All(ctx context.Context, opts *ReadModelsOptions, ropts ...RequestOption) iter.Seq2[AuthorizationModel, error] {
	var o ReadModelsOptions
	if opts != nil {
		o = *opts
	}
	return func(yield func(AuthorizationModel, error) bool) {
		for {
			page, _, err := s.List(ctx, &o, ropts...)
			if err != nil {
				yield(AuthorizationModel{}, err)
				return
			}
			for _, m := range page.AuthorizationModels {
				if !yield(m, nil) {
					return
				}
			}
			if page.ContinuationToken == "" {
				return
			}
			o.ContinuationToken = page.ContinuationToken
		}
	}
}

// ReadAll iterates every tuple matching req across pages. It copies req so the
// caller's struct is never mutated.
func (s *TuplesService) ReadAll(ctx context.Context, req *ReadRequest, ropts ...RequestOption) iter.Seq2[Tuple, error] {
	var r ReadRequest
	if req != nil {
		r = *req
	}
	return func(yield func(Tuple, error) bool) {
		for {
			page, _, err := s.Read(ctx, &r, ropts...)
			if err != nil {
				yield(Tuple{}, err)
				return
			}
			for _, tp := range page.Tuples {
				if !yield(tp, nil) {
					return
				}
			}
			if page.ContinuationToken == "" {
				return
			}
			r.ContinuationToken = page.ContinuationToken
		}
	}
}

// ChangesAll iterates every tuple change across pages. It copies opts so the
// caller's struct is never mutated.
func (s *TuplesService) ChangesAll(ctx context.Context, opts *ReadChangesOptions, ropts ...RequestOption) iter.Seq2[TupleChange, error] {
	var o ReadChangesOptions
	if opts != nil {
		o = *opts
	}
	return func(yield func(TupleChange, error) bool) {
		for {
			page, _, err := s.ReadChanges(ctx, &o, ropts...)
			if err != nil {
				yield(TupleChange{}, err)
				return
			}
			for _, ch := range page.Changes {
				if !yield(ch, nil) {
					return
				}
			}
			if page.ContinuationToken == "" {
				return
			}
			o.ContinuationToken = page.ContinuationToken
		}
	}
}
