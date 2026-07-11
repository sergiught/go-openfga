package openfga

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTuples_Write(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("want POST, got %s", r.Method)
		}
		if r.URL.Path != "/stores/s1/write" {
			t.Fatalf("want /stores/s1/write, got %s", r.URL.Path)
		}
		var body WriteRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body.Writes == nil || len(body.Writes.TupleKeys) != 1 {
			t.Errorf("writes = %+v", body.Writes)
		}
		if body.Writes.TupleKeys[0].User != "user:anne" {
			t.Errorf("user = %q", body.Writes.TupleKeys[0].User)
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	_, err := c.Tuples.Write(context.Background(), &WriteRequest{
		Writes: &WriteRequestTuples{TupleKeys: []TupleKey{{User: "user:anne", Relation: "reader", Object: "doc:1"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestTuples_Write_FillsAuthorizationModelID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body WriteRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body.AuthorizationModelID != "model-1" {
			t.Errorf("authorization_model_id = %q, want model-1", body.AuthorizationModelID)
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	c.authModelID = "model-1"
	_, err := c.Tuples.Write(context.Background(), &WriteRequest{
		Writes: &WriteRequestTuples{TupleKeys: []TupleKey{{User: "user:anne", Relation: "reader", Object: "doc:1"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestTuples_ReadPaginates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("want POST, got %s", r.Method)
		}
		if r.URL.Path != "/stores/s1/read" {
			t.Fatalf("want /stores/s1/read, got %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"tuples":[{"key":{"user":"user:anne","relation":"reader","object":"doc:1"},"timestamp":"2024-01-01T00:00:00Z"}],"continuation_token":"next"}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	page, resp, err := c.Tuples.Read(context.Background(), &ReadRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Tuples) != 1 {
		t.Errorf("tuples = %d, want 1", len(page.Tuples))
	}
	if resp.ContinuationToken != "next" {
		t.Errorf("token = %q, want next", resp.ContinuationToken)
	}
}

func TestTuples_Read_AppliesConsistency(t *testing.T) {
	var gotConsistency ConsistencyPreference
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body ReadRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		gotConsistency = body.Consistency
		_, _ = w.Write([]byte(`{"tuples":[],"continuation_token":""}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	_, _, err := c.Tuples.Read(context.Background(), &ReadRequest{}, WithConsistency(ConsistencyHigherConsistency))
	if err != nil {
		t.Fatal(err)
	}
	if gotConsistency != ConsistencyHigherConsistency {
		t.Errorf("consistency = %q, want %q", gotConsistency, ConsistencyHigherConsistency)
	}
}

func TestTuples_Read_AppliesDefaultConsistency(t *testing.T) {
	var gotConsistency ConsistencyPreference
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body ReadRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		gotConsistency = body.Consistency
		_, _ = w.Write([]byte(`{"tuples":[],"continuation_token":""}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	c.consistency = ConsistencyHigherConsistency
	_, _, err := c.Tuples.Read(context.Background(), &ReadRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if gotConsistency != ConsistencyHigherConsistency {
		t.Errorf("consistency = %q, want %q", gotConsistency, ConsistencyHigherConsistency)
	}
}

func TestTuples_Read_PerCallConsistencyOverridesDefault(t *testing.T) {
	var gotConsistency ConsistencyPreference
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body ReadRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		gotConsistency = body.Consistency
		_, _ = w.Write([]byte(`{"tuples":[],"continuation_token":""}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	c.consistency = ConsistencyHigherConsistency
	_, _, err := c.Tuples.Read(context.Background(), &ReadRequest{}, WithConsistency(ConsistencyMinimizeLatency))
	if err != nil {
		t.Fatal(err)
	}
	if gotConsistency != ConsistencyMinimizeLatency {
		t.Errorf("consistency = %q, want %q", gotConsistency, ConsistencyMinimizeLatency)
	}
}

func TestTuples_Write_DoesNotMutateRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	c.authModelID = "m1"
	req := &WriteRequest{
		Writes: &WriteRequestTuples{TupleKeys: []TupleKey{{User: "user:anne", Relation: "reader", Object: "doc:1"}}},
	}
	_, err := c.Tuples.Write(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if req.AuthorizationModelID != "" {
		t.Errorf("Write mutated req.AuthorizationModelID = %q, want empty", req.AuthorizationModelID)
	}
}

func TestTuples_Read_DoesNotMutateRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tuples":[],"continuation_token":""}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	req := &ReadRequest{}
	_, _, err := c.Tuples.Read(context.Background(), req, WithConsistency(ConsistencyHigherConsistency))
	if err != nil {
		t.Fatal(err)
	}
	if req.Consistency != "" {
		t.Errorf("Read mutated req.Consistency = %q, want empty", req.Consistency)
	}
}

func TestTuples_ReadChanges_QueryParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("want GET, got %s", r.Method)
		}
		if r.URL.Path != "/stores/s1/changes" {
			t.Fatalf("want /stores/s1/changes, got %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("type") != "document" {
			t.Errorf("type = %q, want document", q.Get("type"))
		}
		if q.Get("page_size") != "5" {
			t.Errorf("page_size = %q, want 5", q.Get("page_size"))
		}
		if q.Get("continuation_token") != "tok" {
			t.Errorf("continuation_token = %q, want tok", q.Get("continuation_token"))
		}
		if q.Get("start_time") != "2024-01-01T00:00:00Z" {
			t.Errorf("start_time = %q, want 2024-01-01T00:00:00Z", q.Get("start_time"))
		}
		_, _ = w.Write([]byte(`{"changes":[],"continuation_token":"next"}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	out, resp, err := c.Tuples.ReadChanges(context.Background(), &ReadChangesOptions{
		Type:              "document",
		PageSize:          5,
		ContinuationToken: "tok",
		StartTime:         "2024-01-01T00:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out == nil {
		t.Fatal("out is nil")
	}
	if resp.ContinuationToken != "next" {
		t.Errorf("token = %q, want next", resp.ContinuationToken)
	}
}

func TestWrite_ConflictOptionsSerialized(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, WithStoreID("store1"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Tuples.Write(context.Background(), &WriteRequest{
		Writes: &WriteRequestTuples{
			TupleKeys:   []TupleKey{{User: "user:anne", Relation: "reader", Object: "doc:1"}},
			OnDuplicate: OnDuplicateIgnore,
		},
		Deletes: &WriteRequestTuples{
			TupleKeys: []TupleKey{{User: "user:bob", Relation: "reader", Object: "doc:2"}},
			OnMissing: OnMissingIgnore,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	writes := got["writes"].(map[string]any)
	if writes["on_duplicate"] != "ignore" {
		t.Fatalf("writes.on_duplicate = %v, want ignore", writes["on_duplicate"])
	}
	deletes := got["deletes"].(map[string]any)
	if deletes["on_missing"] != "ignore" {
		t.Fatalf("deletes.on_missing = %v, want ignore", deletes["on_missing"])
	}
}

func TestWrite_ConflictOptionsOmittedWhenUnset(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, WithStoreID("store1"))
	_, err := c.Tuples.Write(context.Background(), &WriteRequest{
		Writes: &WriteRequestTuples{TupleKeys: []TupleKey{{User: "user:anne", Relation: "reader", Object: "doc:1"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	writes := got["writes"].(map[string]any)
	if _, present := writes["on_duplicate"]; present {
		t.Fatalf("on_duplicate should be omitted when unset")
	}
}
