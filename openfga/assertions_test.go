package openfga

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAssertions_WriteAndRead(t *testing.T) {
	var capturedBody WriteAssertionsRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/stores/s1/assertions/model-1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		switch r.Method {
		case http.MethodPut:
			if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
				t.Fatalf("failed to decode Write body: %v", err)
			}
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"authorization_model_id":"model-1","assertions":[{"tuple_key":{"user":"user:a","relation":"reader","object":"doc:1"},"expectation":true}]}`))
		default:
			t.Fatalf("unexpected method: %s", r.Method)
		}
	}))
	defer srv.Close()

	c := testClient(t, srv.URL)
	c.storeID = "s1"

	writeReq := &WriteAssertionsRequest{
		Assertions: []Assertion{
			{
				TupleKey:    CheckRequestTupleKey{User: "user:a", Relation: "reader", Object: "doc:1"},
				Expectation: true,
			},
		},
	}
	if _, err := c.Assertions.Write(context.Background(), "model-1", writeReq); err != nil {
		t.Fatalf("Write: %v", err)
	}

	// Verify the request body reached the server.
	if len(capturedBody.Assertions) != 1 {
		t.Fatalf("server received %d assertions, want 1", len(capturedBody.Assertions))
	}
	got := capturedBody.Assertions[0]
	if got.TupleKey.User != "user:a" || got.TupleKey.Relation != "reader" || got.TupleKey.Object != "doc:1" {
		t.Errorf("server received tuple_key = %+v", got.TupleKey)
	}
	if !got.Expectation {
		t.Errorf("server received expectation = false, want true")
	}

	out, _, err := c.Assertions.Read(context.Background(), "model-1")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(out.Assertions) != 1 || !out.Assertions[0].Expectation {
		t.Errorf("assertions = %+v", out.Assertions)
	}
}
