package openfga

import (
	"net/http"
	"testing"
	"time"
)

func TestClient_PerRequestOverridesWin(t *testing.T) {
	c, err := NewClient("https://api.fga.example",
		WithStoreID(testStoreID),
		WithAuthorizationModelID(testModelID),
		WithDefaultConsistency(ConsistencyMinimizeLatency))
	if err != nil {
		t.Fatal(err)
	}

	var rc requestConfig
	WithStore("override-store")(&rc)
	WithAuthorizationModel("override-model")(&rc)
	WithConsistency(ConsistencyHigherConsistency)(&rc)

	if got, _ := c.storeFor(&rc); got != "override-store" {
		t.Errorf("storeFor = %q, want override", got)
	}
	if got := c.modelFor(&rc); got != "override-model" {
		t.Errorf("modelFor = %q, want override", got)
	}
	if got := c.consistencyFor(&rc); got != ConsistencyHigherConsistency {
		t.Errorf("consistencyFor = %q, want override", got)
	}
}

func TestClient_ResolversFallBackToClientDefaults(t *testing.T) {
	c, err := NewClient("https://api.fga.example",
		WithStoreID(testStoreID),
		WithAuthorizationModelID(testModelID),
		WithDefaultConsistency(ConsistencyMinimizeLatency))
	if err != nil {
		t.Fatal(err)
	}

	var rc requestConfig // no per-request overrides
	if got, _ := c.storeFor(&rc); got != testStoreID {
		t.Errorf("storeFor = %q, want client default", got)
	}
	if got := c.modelFor(&rc); got != testModelID {
		t.Errorf("modelFor = %q, want client default", got)
	}
	if got := c.consistencyFor(&rc); got != ConsistencyMinimizeLatency {
		t.Errorf("consistencyFor = %q, want client default", got)
	}
}

func TestWithHeaders_Applied(t *testing.T) {
	h := http.Header{}
	h.Set("X-Custom", "v")
	c, err := NewClient("https://api.fga.example", WithHeaders(h))
	if err != nil {
		t.Fatal(err)
	}
	if c.staticHeaders.Get("X-Custom") != "v" {
		t.Errorf("static header not applied: %v", c.staticHeaders)
	}
}

func TestWithRetry_And_WithoutRetry(t *testing.T) {
	c, err := NewClient("https://api.fga.example",
		WithRetry(RetryConfig{MaxAttempts: 5, MinWait: time.Second, MaxWait: 2 * time.Second}))
	if err != nil {
		t.Fatal(err)
	}
	if c.retry == nil || c.retry.MaxAttempts != 5 {
		t.Errorf("WithRetry not applied: %+v", c.retry)
	}

	c2, err := NewClient("https://api.fga.example", WithoutRetry())
	if err != nil {
		t.Fatal(err)
	}
	if c2.retry != nil {
		t.Errorf("WithoutRetry should clear retry config, got %+v", c2.retry)
	}
}
