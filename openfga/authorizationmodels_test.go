package openfga

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestModels_Write(t *testing.T) {
	var received WriteAuthorizationModelRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/stores/s1/authorization-models" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		_, _ = w.Write([]byte(`{"authorization_model_id":"model-1"}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	out, _, err := c.AuthorizationModels.Write(context.Background(), &WriteAuthorizationModelRequest{SchemaVersion: "1.1"})
	if err != nil {
		t.Fatal(err)
	}
	if out.AuthorizationModelID != "model-1" {
		t.Errorf("id = %q", out.AuthorizationModelID)
	}
	if received.SchemaVersion != "1.1" {
		t.Errorf("body schema_version = %q, want 1.1", received.SchemaVersion)
	}
}

func TestModels_List_WithQueryParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/stores/s1/authorization-models" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("page_size") != "5" || r.URL.Query().Get("continuation_token") != "tok" {
			t.Errorf("query = %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"authorization_models":[{"id":"model-2","schema_version":"1.1"}],"continuation_token":"next"}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	page, resp, err := c.AuthorizationModels.List(context.Background(), &ReadModelsOptions{PageSize: 5, ContinuationToken: "tok"})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.AuthorizationModels) != 1 {
		t.Fatalf("len = %d, want 1", len(page.AuthorizationModels))
	}
	if page.AuthorizationModels[0].ID != "model-2" {
		t.Errorf("id = %q", page.AuthorizationModels[0].ID)
	}
	if resp.ContinuationToken != "next" {
		t.Errorf("token = %q", resp.ContinuationToken)
	}
}

func TestModels_Get(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/stores/s1/authorization-models/model-5" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"authorization_model":{"id":"model-5","schema_version":"1.1"}}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	m, _, err := c.AuthorizationModels.Get(context.Background(), "model-5")
	if err != nil {
		t.Fatal(err)
	}
	if m.ID != "model-5" {
		t.Errorf("id = %q", m.ID)
	}
}

func TestModels_ReadLatest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page_size") != "1" {
			t.Errorf("page_size = %s", r.URL.Query().Get("page_size"))
		}
		_, _ = w.Write([]byte(`{"authorization_models":[{"id":"model-9","schema_version":"1.1"}],"continuation_token":""}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	m, _, err := c.AuthorizationModels.ReadLatest(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if m.ID != "model-9" {
		t.Errorf("id = %q", m.ID)
	}
}

func TestModels_ReadLatest_EmptyList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"authorization_models":[],"continuation_token":""}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	_, _, err := c.AuthorizationModels.ReadLatest(context.Background())
	if err == nil {
		t.Fatal("expected error for empty list, got nil")
	}
}
