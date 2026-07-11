package openfga

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func testClient(t *testing.T, base string) *Client {
	t.Helper()
	c, err := NewClient(base)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestNewRequest_EncodesBodyAndHeaders(t *testing.T) {
	c := testClient(t, "https://api.fga.example")
	body := map[string]string{"name": "demo"}
	req, err := c.NewRequest(context.Background(), http.MethodPost, "/stores", body,
		WithRequestHeader("X-Trace", "abc"))
	if err != nil {
		t.Fatal(err)
	}
	if req.URL.String() != "https://api.fga.example/stores" {
		t.Errorf("url = %s", req.URL.String())
	}
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("content-type = %s", got)
	}
	if got := req.Header.Get("X-Trace"); got != "abc" {
		t.Errorf("x-trace = %s", got)
	}
	var decoded map[string]string
	if err := json.NewDecoder(req.Body).Decode(&decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["name"] != "demo" {
		t.Errorf("body = %v", decoded)
	}
}

func TestNewRequest_NilBodyHasNoContentType(t *testing.T) {
	c := testClient(t, "https://api.fga.example")
	req, err := c.NewRequest(context.Background(), http.MethodGet, "/stores", nil)
	if err != nil {
		t.Fatal(err)
	}
	if req.Header.Get("Content-Type") != "" {
		t.Error("expected no content-type for nil body")
	}
	if req.Body != nil && req.Body != http.NoBody {
		b, _ := io.ReadAll(req.Body)
		if len(b) != 0 {
			t.Error("expected empty body")
		}
	}
}

func TestChunkingOptions_SetConfig(t *testing.T) {
	rc := newRequestConfig()
	applyOptions(rc, []RequestOption{
		WithMaxParallel(7),
		WithMaxPerChunk(20),
		WithMaxChecksPerBatch(25),
		WithTransaction(),
		WithOnDuplicate(OnDuplicateIgnore),
		WithOnMissing(OnMissingIgnore),
	})
	if rc.maxParallel != 7 || rc.maxPerChunk != 20 || rc.maxChecksPerBatch != 25 {
		t.Fatalf("chunk sizes not set: %+v", rc)
	}
	if !rc.transaction {
		t.Fatal("transaction not set")
	}
	if rc.onDuplicate != OnDuplicateIgnore || rc.onMissing != OnMissingIgnore {
		t.Fatal("conflict options not set")
	}
}
