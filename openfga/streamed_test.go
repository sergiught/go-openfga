package openfga

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStreamedListObjects(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/stores/s1/streamed-list-objects" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte("{\"result\":{\"object\":\"doc:1\"}}\n{\"result\":{\"object\":\"doc:2\"}}\n"))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	var got []string
	for item, err := range c.Relationships.StreamedListObjects(context.Background(), &ListObjectsRequest{Type: "doc", Relation: "reader", User: "user:a"}) {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, item.Object)
	}
	if len(got) != 2 || got[0] != "doc:1" || got[1] != "doc:2" {
		t.Errorf("got = %v", got)
	}
}

func TestStreamedListObjects_EarlyBreak(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{\"result\":{\"object\":\"doc:1\"}}\n{\"result\":{\"object\":\"doc:2\"}}\n{\"result\":{\"object\":\"doc:3\"}}\n"))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	var got []string
	for item, err := range c.Relationships.StreamedListObjects(context.Background(), &ListObjectsRequest{Type: "doc", Relation: "reader", User: "user:a"}) {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, item.Object)
		break // early break; defer resp.Body.Close() still runs
	}
	if len(got) != 1 {
		t.Errorf("got = %v, want exactly 1 item", got)
	}
}

func TestStreamedListObjects_PropagatesHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"code":"internal","message":"oops"}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	c.storeID = "s1"
	var gotErr error
	for _, err := range c.Relationships.StreamedListObjects(context.Background(), &ListObjectsRequest{Type: "doc", Relation: "reader", User: "user:a"}) {
		gotErr = err
	}
	if gotErr == nil {
		t.Error("expected a non-nil error from a 500 response")
	}
}
