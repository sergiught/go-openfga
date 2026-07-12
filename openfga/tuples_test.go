package openfga

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
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
	err := c.Tuples.Write(context.Background(), &WriteRequest{
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
	err := c.Tuples.Write(context.Background(), &WriteRequest{
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
	page, err := c.Tuples.Read(context.Background(), &ReadRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Tuples) != 1 {
		t.Errorf("tuples = %d, want 1", len(page.Tuples))
	}
	if page.ContinuationToken != "next" {
		t.Errorf("token = %q, want next", page.ContinuationToken)
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
	_, err := c.Tuples.Read(context.Background(), &ReadRequest{}, WithConsistency(ConsistencyHigherConsistency))
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
	_, err := c.Tuples.Read(context.Background(), &ReadRequest{})
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
	_, err := c.Tuples.Read(context.Background(), &ReadRequest{}, WithConsistency(ConsistencyMinimizeLatency))
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
	err := c.Tuples.Write(context.Background(), req)
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
	_, err := c.Tuples.Read(context.Background(), req, WithConsistency(ConsistencyHigherConsistency))
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
	out, err := c.Tuples.ReadChanges(context.Background(), &ReadChangesOptions{
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
	if out.ContinuationToken != "next" {
		t.Errorf("token = %q, want next", out.ContinuationToken)
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

	c, err := NewClient(srv.URL, WithStoreID(testStoreID))
	if err != nil {
		t.Fatal(err)
	}
	err = c.Tuples.Write(context.Background(), &WriteRequest{
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

	c, _ := NewClient(srv.URL, WithStoreID(testStoreID))
	err := c.Tuples.Write(context.Background(), &WriteRequest{
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

func TestWriteTuples_ChunksAndReportsPerTuple(t *testing.T) {
	var reqCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&reqCount, 1)
		var body struct {
			Writes struct {
				TupleKeys []TupleKey `json:"tuple_keys"`
			} `json:"writes"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		for _, k := range body.Writes.TupleKeys {
			if k.Object == "doc:bad" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"code":"validation_error","message":"bad"}`))
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, WithStoreID(testStoreID))
	keys := []TupleKey{
		{User: "user:a", Relation: "reader", Object: "doc:1"},
		{User: "user:b", Relation: "reader", Object: "doc:bad"},
		{User: "user:c", Relation: "reader", Object: "doc:3"},
	}
	resp, err := c.Tuples.WriteTuples(context.Background(), keys, WithMaxPerChunk(1), WithMaxParallel(2))
	if err != nil {
		t.Fatalf("top-level err = %v, want nil for partial failure", err)
	}
	if got := atomic.LoadInt32(&reqCount); got != 3 {
		t.Fatalf("request count = %d, want 3 (one per chunk)", got)
	}
	if len(resp.Writes) != 3 {
		t.Fatalf("len(Writes) = %d, want 3", len(resp.Writes))
	}
	if resp.Writes[0].Status != WriteStatusSuccess || resp.Writes[2].Status != WriteStatusSuccess {
		t.Fatal("chunks 0 and 2 should succeed")
	}
	if resp.Writes[1].Status != WriteStatusFailure || resp.Writes[1].Err == nil {
		t.Fatal("chunk 1 should fail with an error")
	}
	if resp.Writes[1].TupleKey.Object != "doc:bad" {
		t.Fatalf("result order not preserved: %+v", resp.Writes[1])
	}
}

func TestWriteTuples_TransactionSendsSingleRequest(t *testing.T) {
	var reqCount int32
	var lastBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&reqCount, 1)
		_ = json.NewDecoder(r.Body).Decode(&lastBody)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, WithStoreID(testStoreID))
	keys := []TupleKey{
		{User: "user:a", Relation: "reader", Object: "doc:1"},
		{User: "user:b", Relation: "reader", Object: "doc:2"},
	}
	resp, err := c.Tuples.WriteTuples(context.Background(), keys, WithTransaction())
	if err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(&reqCount) != 1 {
		t.Fatal("transaction mode should send exactly one request")
	}
	if len(resp.Writes) != 2 || resp.Writes[0].Status != WriteStatusSuccess {
		t.Fatal("all tuples should be success")
	}
}

func TestDeleteTuples_SendsDeletesWithOnMissing(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, WithStoreID(testStoreID))
	keys := []TupleKey{{User: "user:a", Relation: "reader", Object: "doc:1"}}
	resp, err := c.Tuples.DeleteTuples(context.Background(), keys, WithTransaction(), WithOnMissing(OnMissingIgnore))
	if err != nil {
		t.Fatal(err)
	}
	deletes := body["deletes"].(map[string]any)
	if deletes["on_missing"] != "ignore" {
		t.Fatalf("deletes.on_missing = %v", deletes["on_missing"])
	}
	if _, ok := body["writes"]; ok {
		t.Fatal("DeleteTuples must not send a writes block")
	}
	if len(resp.Deletes) != 1 || len(resp.Writes) != 0 {
		t.Fatal("DeleteTuples should populate Deletes only")
	}
}

func TestWriteTuples_EmptyKeysNoRequest(t *testing.T) {
	var reqCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&reqCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	c, _ := NewClient(srv.URL, WithStoreID(testStoreID))
	resp, err := c.Tuples.WriteTuples(context.Background(), nil)
	if err != nil || len(resp.Writes) != 0 || atomic.LoadInt32(&reqCount) != 0 {
		t.Fatal("empty keys should issue no request and return empty response")
	}
}
