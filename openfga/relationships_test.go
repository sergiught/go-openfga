package openfga

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
	out, _, err := c.Relationships.Check(context.Background(), &CheckRequest{
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
	_, _, err := c.Relationships.Check(context.Background(), req, WithConsistency(ConsistencyMinimizeLatency))
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
	out, _, err := c.Relationships.BatchCheck(context.Background(), &BatchCheckRequest{
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
	out, _, err := c.Relationships.ListObjects(context.Background(), &ListObjectsRequest{
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
