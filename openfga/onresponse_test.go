package openfga

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOnResponse_CapturesResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Fga-Request-Id", "req-123")
		_, _ = w.Write([]byte(`{"allowed":true}`))
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, WithStoreID(testStoreID))

	var got *Response
	out, err := c.Relationships.Check(context.Background(), &CheckRequest{
		TupleKey: CheckRequestTupleKey{User: "user:a", Relation: "reader", Object: "doc:1"},
	}, OnResponse(func(r *Response) { got = r }))
	if err != nil {
		t.Fatal(err)
	}
	if !out.Allowed {
		t.Error("want allowed=true")
	}
	if got == nil {
		t.Fatal("OnResponse callback did not fire")
	}
	if got.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", got.StatusCode)
	}
	if got.RequestID() != "req-123" {
		t.Errorf("RequestID = %q, want req-123", got.RequestID())
	}
}

func TestOnResponse_FiresOnAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Fga-Request-Id", "req-err")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":"undefined_endpoint","message":"nope"}`))
	}))
	defer srv.Close()

	c, _ := NewClient(srv.URL, WithStoreID(testStoreID))

	var got *Response
	_, err := c.Relationships.Check(context.Background(), &CheckRequest{
		TupleKey: CheckRequestTupleKey{User: "user:a", Relation: "reader", Object: "doc:1"},
	}, OnResponse(func(r *Response) { got = r }))
	if err == nil {
		t.Fatal("expected an API error")
	}
	if got == nil {
		t.Fatal("OnResponse callback did not fire on API error")
	}
	if got.RequestID() != "req-err" {
		t.Errorf("RequestID = %q, want req-err", got.RequestID())
	}
}
