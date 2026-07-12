package openfga

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStores_Create(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/stores" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id":"01H","name":"demo","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	store, err := c.Stores.Create(context.Background(), &CreateStoreRequest{Name: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	if store.ID != "01H" || store.Name != "demo" {
		t.Errorf("store = %+v", store)
	}
}

func TestStores_List_WithQueryParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page_size") != "10" || r.URL.Query().Get("continuation_token") != "tok" {
			t.Errorf("query = %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"stores":[{"id":"01H","name":"demo"}],"continuation_token":""}`))
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	page, err := c.Stores.List(context.Background(), &ListStoresOptions{PageSize: 10, ContinuationToken: "tok"})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Stores) != 1 || page.ContinuationToken != "" {
		t.Errorf("page = %+v", page)
	}
}

func TestStores_GetAndDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path != "/stores/01H" {
				t.Errorf("path = %s", r.URL.Path)
			}
			_, _ = w.Write([]byte(`{"id":"01H","name":"demo"}`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer srv.Close()
	c := testClient(t, srv.URL)
	if _, err := c.Stores.Get(context.Background(), "01H"); err != nil {
		t.Fatal(err)
	}
	if err := c.Stores.Delete(context.Background(), "01H"); err != nil {
		t.Fatal(err)
	}
}
