package openfga

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"sync/atomic"
	"testing"
)

// correlationIDPattern mirrors the OpenFGA batch-check correlation ID contract.
var correlationIDPattern = regexp.MustCompile(`^[\w\d-]{1,36}$`)

func TestRelationships_Check(t *testing.T) {
	var gotBody CheckRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("want POST, got %s", r.Method)
		}
		if r.URL.Path != "/stores/s1/check" {
			t.Fatalf("want /stores/s1/check, got %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"allowed":true}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	out, err := c.Relationships.Check(context.Background(), &CheckRequest{
		TupleKey: CheckRequestTupleKey{User: "user:anne", Relation: "reader", Object: "doc:1"},
	}, WithConsistency(ConsistencyHigherConsistency))
	if err != nil {
		t.Fatal(err)
	}
	if !out.Allowed {
		t.Error("want allowed=true")
	}
	if gotBody.Consistency != ConsistencyHigherConsistency {
		t.Errorf("consistency in body = %q, want %q", gotBody.Consistency, ConsistencyHigherConsistency)
	}
}

func TestRelationships_Check_AppliesDefaultConsistency(t *testing.T) {
	var gotBody CheckRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"allowed":true}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	c.consistency = ConsistencyHigherConsistency
	_, err := c.Relationships.Check(context.Background(), &CheckRequest{
		TupleKey: CheckRequestTupleKey{User: "user:anne", Relation: "reader", Object: "doc:1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotBody.Consistency != ConsistencyHigherConsistency {
		t.Errorf("consistency in body = %q, want %q", gotBody.Consistency, ConsistencyHigherConsistency)
	}
}

func TestRelationships_Check_DoesNotMutateRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"allowed":false}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	c.authModelID = "model-default"
	req := &CheckRequest{
		TupleKey: CheckRequestTupleKey{User: "user:anne", Relation: "reader", Object: "doc:1"},
	}
	_, err := c.Relationships.Check(context.Background(), req, WithConsistency(ConsistencyMinimizeLatency))
	if err != nil {
		t.Fatal(err)
	}
	if req.AuthorizationModelID != "" {
		t.Errorf("Check mutated req.AuthorizationModelID = %q, want empty", req.AuthorizationModelID)
	}
	if req.Consistency != "" {
		t.Errorf("Check mutated req.Consistency = %q, want empty", req.Consistency)
	}
}

func TestRelationships_BatchCheckKeyedByCorrelationID(t *testing.T) {
	var gotBody BatchCheckRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("want POST, got %s", r.Method)
		}
		if r.URL.Path != "/stores/s1/batch-check" {
			t.Fatalf("want /stores/s1/batch-check, got %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"result":{"id-1":{"allowed":true},"id-2":{"allowed":false}}}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	out, err := c.Relationships.BatchCheck(context.Background(), &BatchCheckRequest{
		Checks: []BatchCheckItem{
			{CorrelationID: "id-1"},
			{CorrelationID: "id-2"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Result["id-1"].Allowed {
		t.Errorf("result[id-1].Allowed = false, want true")
	}
	if out.Result["id-2"].Allowed {
		t.Errorf("result[id-2].Allowed = true, want false")
	}
	if len(gotBody.Checks) != 2 {
		t.Errorf("body.Checks len = %d, want 2", len(gotBody.Checks))
	}
}

func TestRelationships_BatchCheckGeneratesMissingCorrelationIDs(t *testing.T) {
	var gotBody BatchCheckRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		out := BatchCheckResponse{Result: map[string]BatchCheckSingleResult{}}
		for _, item := range gotBody.Checks {
			if item.CorrelationID == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			out.Result[item.CorrelationID] = BatchCheckSingleResult{Allowed: true}
		}
		_ = json.NewEncoder(w).Encode(out)
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, WithStoreID(testStoreID))
	out, err := c.Relationships.BatchCheck(context.Background(), &BatchCheckRequest{
		Checks: []BatchCheckItem{
			{TupleKey: CheckRequestTupleKey{User: "user:a", Relation: "reader", Object: "doc:1"}},
			{TupleKey: CheckRequestTupleKey{User: "user:b", Relation: "reader", Object: "doc:2"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for i, item := range gotBody.Checks {
		if item.CorrelationID == "" {
			t.Errorf("check %d sent an empty correlation_id", i)
		}
	}
	if len(out.Result) != 2 {
		t.Errorf("results = %d, want 2", len(out.Result))
	}
}

func TestRelationships_BatchCheckRejectsDuplicateCorrelationIDs(t *testing.T) {
	c, _ := NewClient("http://example.invalid", WithStoreID(testStoreID))
	_, err := c.Relationships.BatchCheck(context.Background(), &BatchCheckRequest{
		Checks: []BatchCheckItem{
			{CorrelationID: "dup"},
			{CorrelationID: "dup"},
		},
	})
	if err == nil {
		t.Fatal("expected error on duplicate correlation IDs")
	}
}

func TestRelationships_ListObjects(t *testing.T) {
	var gotBody ListObjectsRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("want POST, got %s", r.Method)
		}
		if r.URL.Path != "/stores/s1/list-objects" {
			t.Fatalf("want /stores/s1/list-objects, got %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"objects":["doc:1","doc:2"]}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	out, err := c.Relationships.ListObjects(context.Background(), &ListObjectsRequest{
		Type: "doc", Relation: "reader", User: "user:anne",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Objects) != 2 {
		t.Errorf("objects = %v, want 2 elements", out.Objects)
	}
	if gotBody.Type != "doc" || gotBody.Relation != "reader" || gotBody.User != "user:anne" {
		t.Errorf("body = %+v", gotBody)
	}
}

func TestRelationships_Expand(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("want POST, got %s", r.Method)
		}
		if r.URL.Path != "/stores/s1/expand" {
			t.Fatalf("want /stores/s1/expand, got %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"tree":{"root":{"name":"doc:1#reader"}}}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	out, err := c.Relationships.Expand(context.Background(), &ExpandRequest{
		TupleKey: CheckRequestTupleKey{User: "user:anne", Relation: "reader", Object: "doc:1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Tree) == 0 {
		t.Error("want ExpandResponse.Tree to be populated")
	}
}

func TestRelationships_ListUsers(t *testing.T) {
	var gotBody ListUsersRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("want POST, got %s", r.Method)
		}
		if r.URL.Path != "/stores/s1/list-users" {
			t.Fatalf("want /stores/s1/list-users, got %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"users":[{"object":{"type":"user","id":"anne"}}]}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	out, err := c.Relationships.ListUsers(context.Background(), &ListUsersRequest{
		Object:      FGAObjectRelation{Object: "doc:1"},
		Relation:    "reader",
		UserFilters: []UserTypeFilter{{Type: "user"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Users) != 1 {
		t.Errorf("users = %v, want 1 element", out.Users)
	}
	if gotBody.Relation != "reader" {
		t.Errorf("body.Relation = %q, want %q", gotBody.Relation, "reader")
	}
	if gotBody.Object.Object != "doc:1" {
		t.Errorf("body.Object.Object = %q, want %q", gotBody.Object.Object, "doc:1")
	}
}

func TestBatchCheckAll_ChunksAndMerges(t *testing.T) {
	var reqCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&reqCount, 1)
		var body BatchCheckRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		out := BatchCheckResponse{Result: map[string]BatchCheckSingleResult{}}
		for _, item := range body.Checks {
			out.Result[item.CorrelationID] = BatchCheckSingleResult{Allowed: item.TupleKey.Object == "doc:yes"}
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(out)
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, WithStoreID(testStoreID))
	var checks []BatchCheckItem
	for i := 0; i < 5; i++ {
		obj := "doc:no"
		if i%2 == 0 {
			obj = "doc:yes"
		}
		checks = append(checks, BatchCheckItem{
			TupleKey:      CheckRequestTupleKey{User: "user:a", Relation: "reader", Object: obj},
			CorrelationID: "c" + strconv.Itoa(i),
		})
	}
	resp, err := c.Relationships.BatchCheckAll(context.Background(), &BatchCheckRequest{Checks: checks}, WithMaxChecksPerBatch(2))
	if err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(&reqCount) != 3 { // ceil(5/2)
		t.Fatalf("request count = %d, want 3", atomic.LoadInt32(&reqCount))
	}
	if len(resp.Result) != 5 {
		t.Fatalf("merged result size = %d, want 5", len(resp.Result))
	}
	if !resp.Result["c0"].Allowed || resp.Result["c1"].Allowed {
		t.Fatal("merged results incorrect")
	}
}

func TestBatchCheckAll_GeneratesMissingCorrelationIDs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body BatchCheckRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		out := BatchCheckResponse{Result: map[string]BatchCheckSingleResult{}}
		for _, item := range body.Checks {
			if item.CorrelationID == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			out.Result[item.CorrelationID] = BatchCheckSingleResult{Allowed: true}
		}
		_ = json.NewEncoder(w).Encode(out)
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, WithStoreID(testStoreID))
	resp, err := c.Relationships.BatchCheckAll(context.Background(), &BatchCheckRequest{Checks: []BatchCheckItem{
		{TupleKey: CheckRequestTupleKey{User: "user:a", Relation: "reader", Object: "doc:1"}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Result) != 1 {
		t.Fatalf("want 1 result, got %d", len(resp.Result))
	}
}

func TestBatchCheckAll_DuplicateCorrelationIDsError(t *testing.T) {
	c, _ := NewClient("http://example.invalid", WithStoreID(testStoreID))
	_, err := c.Relationships.BatchCheckAll(context.Background(), &BatchCheckRequest{Checks: []BatchCheckItem{
		{TupleKey: CheckRequestTupleKey{User: "user:a", Relation: "reader", Object: "doc:1"}, CorrelationID: "dup"},
		{TupleKey: CheckRequestTupleKey{User: "user:b", Relation: "reader", Object: "doc:2"}, CorrelationID: "dup"},
	}})
	if err == nil {
		t.Fatal("expected error on duplicate correlation IDs")
	}
}

func TestBatchCheckAll_EmptyChecks(t *testing.T) {
	c, _ := NewClient("http://example.invalid", WithStoreID(testStoreID))
	resp, err := c.Relationships.BatchCheckAll(context.Background(), &BatchCheckRequest{})
	if err != nil || len(resp.Result) != 0 {
		t.Fatal("empty checks should return empty result and nil error")
	}
}

func TestListRelations_ReturnsAllowedInInputOrder(t *testing.T) {
	var reqCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&reqCount, 1)
		if r.URL.Path != "/stores/"+testStoreID+"/batch-check" {
			t.Fatalf("want /stores/%s/batch-check, got %s", testStoreID, r.URL.Path)
		}
		var body BatchCheckRequest
		_ = json.NewDecoder(r.Body).Decode(&body)
		out := BatchCheckResponse{Result: map[string]BatchCheckSingleResult{}}
		for _, item := range body.Checks {
			// Correlation IDs must satisfy the server's ^[\w\d-]{1,36}$ contract,
			// so they cannot simply be the (arbitrary) relation name.
			if !correlationIDPattern.MatchString(item.CorrelationID) {
				t.Fatalf("correlation id %q violates ^[\\w\\d-]{1,36}$", item.CorrelationID)
			}
			out.Result[item.CorrelationID] = BatchCheckSingleResult{Allowed: item.TupleKey.Relation != "can_delete"}
		}
		_ = json.NewEncoder(w).Encode(out)
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, WithStoreID(testStoreID))
	got, err := c.Relationships.ListRelations(context.Background(), &ListRelationsRequest{
		User:      "user:anne",
		Object:    "document:budget",
		Relations: []string{"can_view", "can_edit", "can_delete"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(&reqCount) != 1 {
		t.Fatalf("request count = %d, want 1", atomic.LoadInt32(&reqCount))
	}
	if len(got) != 2 || got[0] != "can_view" || got[1] != "can_edit" {
		t.Fatalf("got %v, want [can_view can_edit]", got)
	}
}

func TestListRelations_EmptyRelationsNoRequest(t *testing.T) {
	var reqCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&reqCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	c, _ := NewClient(srv.URL, WithStoreID(testStoreID))
	got, err := c.Relationships.ListRelations(context.Background(), &ListRelationsRequest{
		User: "user:anne", Object: "document:budget",
	})
	if err != nil || len(got) != 0 || atomic.LoadInt32(&reqCount) != 0 {
		t.Fatalf("empty relations should issue no request; got=%v err=%v reqs=%d", got, err, reqCount)
	}
}

func TestListRelations_DuplicateRelationsError(t *testing.T) {
	c, _ := NewClient("http://example.invalid", WithStoreID(testStoreID))
	_, err := c.Relationships.ListRelations(context.Background(), &ListRelationsRequest{
		User: "user:anne", Object: "document:budget",
		Relations: []string{"can_view", "can_view"},
	})
	if err == nil {
		t.Fatal("expected error on duplicate relations")
	}
}
