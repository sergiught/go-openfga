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
		_, _ = w.Write([]byte(`{"changes":[],"continuation_token":"next"}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	out, resp, err := c.Tuples.ReadChanges(context.Background(), &ReadChangesOptions{
		Type:              "document",
		PageSize:          5,
		ContinuationToken: "tok",
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
