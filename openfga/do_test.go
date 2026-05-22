package openfga

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type sampleBody struct {
	Name  string `json:"name"`
	Token string `json:"continuation_token"`
}

func (s *sampleBody) continuationToken() string { return s.Token }

func TestDo_DecodesBodyAndLiftsToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"demo","continuation_token":"next"}`))
	}))
	defer srv.Close()

	c := testClient(t, srv.URL)
	req, _ := c.NewRequest(context.Background(), http.MethodGet, "/stores", nil)
	var out sampleBody
	resp, err := c.Do(req, &out)
	if err != nil {
		t.Fatal(err)
	}
	if out.Name != "demo" {
		t.Errorf("name = %q", out.Name)
	}
	if resp.ContinuationToken != "next" {
		t.Errorf("token = %q", resp.ContinuationToken)
	}
}

func TestDo_ReturnsTypedErrorOn4xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code":"validation_error","message":"bad"}`))
	}))
	defer srv.Close()

	c := testClient(t, srv.URL)
	req, _ := c.NewRequest(context.Background(), http.MethodPost, "/stores", map[string]string{})
	_, err := c.Do(req, nil)
	var ve *ValidationError
	if err == nil || !asValidation(err, &ve) {
		t.Fatalf("want *ValidationError, got %v", err)
	}
}

func asValidation(err error, target **ValidationError) bool {
	v, ok := err.(*ValidationError)
	if ok {
		*target = v
	}
	return ok
}
