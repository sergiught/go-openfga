package openfga

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStoresAll_PaginatesAndStopsOnEmptyToken(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch r.URL.Query().Get("continuation_token") {
		case "":
			_, _ = w.Write([]byte(`{"stores":[{"id":"1"},{"id":"2"}],"continuation_token":"p2"}`))
		case "p2":
			_, _ = w.Write([]byte(`{"stores":[{"id":"3"}],"continuation_token":""}`))
		default:
			t.Errorf("unexpected token %q", r.URL.Query().Get("continuation_token"))
		}
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)

	var ids []string
	for store, err := range c.Stores.All(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, store.ID)
	}
	if fmt.Sprint(ids) != "[1 2 3]" {
		t.Errorf("ids = %v", ids)
	}
	if calls != 2 {
		t.Errorf("calls = %d (want 2)", calls)
	}
}

func TestStoresAll_EarlyBreakStopsFetching(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = w.Write([]byte(`{"stores":[{"id":"1"},{"id":"2"}],"continuation_token":"p2"}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	for store, err := range c.Stores.All(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		_ = store
		break
	}
	if calls != 1 {
		t.Errorf("calls = %d (want 1 after break)", calls)
	}
}

// TestStoresAll_DoesNotMutateCallerOptions verifies that All copies the options
// struct so the caller's ContinuationToken is never overwritten.
func TestStoresAll_DoesNotMutateCallerOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"stores":[{"id":"1"}],"continuation_token":""}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)

	original := &ListStoresOptions{ContinuationToken: "orig"}
	for _, err := range c.Stores.All(context.Background(), original) {
		if err != nil {
			t.Fatal(err)
		}
	}
	if original.ContinuationToken != "orig" {
		t.Errorf("caller's ContinuationToken was mutated to %q", original.ContinuationToken)
	}
}

func TestAuthorizationModelsAll_Paginates(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch r.URL.Query().Get("continuation_token") {
		case "":
			_, _ = w.Write([]byte(`{"authorization_models":[{"id":"m1"},{"id":"m2"}],"continuation_token":"p2"}`))
		case "p2":
			_, _ = w.Write([]byte(`{"authorization_models":[{"id":"m3"}],"continuation_token":""}`))
		default:
			t.Errorf("unexpected token %q", r.URL.Query().Get("continuation_token"))
		}
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"

	var ids []string
	for m, err := range c.AuthorizationModels.All(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, m.ID)
	}
	if fmt.Sprint(ids) != "[m1 m2 m3]" {
		t.Errorf("ids = %v", ids)
	}
	if calls != 2 {
		t.Errorf("calls = %d (want 2)", calls)
	}
}

func TestChangesAll_Paginates(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch r.URL.Query().Get("continuation_token") {
		case "":
			_, _ = w.Write([]byte(`{"changes":[{"tuple_key":{"user":"user:a","relation":"reader","object":"doc:1"},"operation":"TUPLE_OPERATION_WRITE","timestamp":"2024-01-01T00:00:00Z"},{"tuple_key":{"user":"user:b","relation":"reader","object":"doc:2"},"operation":"TUPLE_OPERATION_WRITE","timestamp":"2024-01-01T00:00:00Z"}],"continuation_token":"p2"}`))
		case "p2":
			_, _ = w.Write([]byte(`{"changes":[{"tuple_key":{"user":"user:c","relation":"reader","object":"doc:3"},"operation":"TUPLE_OPERATION_DELETE","timestamp":"2024-01-02T00:00:00Z"}],"continuation_token":""}`))
		default:
			t.Errorf("unexpected token %q", r.URL.Query().Get("continuation_token"))
		}
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"

	var count int
	for _, err := range c.Tuples.ChangesAll(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		count++
	}
	if count != 3 {
		t.Errorf("count = %d (want 3)", count)
	}
	if calls != 2 {
		t.Errorf("calls = %d (want 2)", calls)
	}
}

func TestChangesAll_EarlyBreakStopsFetching(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = w.Write([]byte(`{"changes":[{"tuple_key":{"user":"user:a","relation":"reader","object":"doc:1"},"operation":"TUPLE_OPERATION_WRITE","timestamp":"2024-01-01T00:00:00Z"},{"tuple_key":{"user":"user:b","relation":"reader","object":"doc:2"},"operation":"TUPLE_OPERATION_WRITE","timestamp":"2024-01-01T00:00:00Z"}],"continuation_token":"p2"}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"

	for _, err := range c.Tuples.ChangesAll(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		break
	}
	if calls != 1 {
		t.Errorf("calls = %d (want 1 after break)", calls)
	}
}

// TestTuplesReadAll_PropagatesError verifies that when the server returns an
// error on page 2 the iterator yields a non-nil error and stops.
func TestTuplesReadAll_PropagatesError(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		// First page: one tuple, token pointing to page 2.
		// Second page: server-side error.
		switch calls {
		case 1:
			_, _ = w.Write([]byte(`{"tuples":[{"key":{"user":"user:a","relation":"reader","object":"doc:1"},"timestamp":"2024-01-01T00:00:00Z"}],"continuation_token":"p2"}`))
		default:
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"code":"internal","message":"boom"}`))
		}
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"

	var gotItems int
	var gotErr error
	for item, err := range c.Tuples.ReadAll(context.Background(), nil) {
		if err != nil {
			gotErr = err
			break
		}
		gotItems++
		_ = item
	}
	if gotErr == nil {
		t.Error("expected a non-nil error from the iterator, got nil")
	}
	if gotItems != 1 {
		t.Errorf("expected 1 item before error, got %d", gotItems)
	}
	if calls != 2 {
		t.Errorf("expected exactly 2 HTTP calls, got %d", calls)
	}
}
