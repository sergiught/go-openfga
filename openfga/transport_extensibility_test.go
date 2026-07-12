package openfga

import (
	"context"
	"net/http"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

// TestWithRequestObserver_FiresPerAttempt verifies the observer sees the
// fully-decorated request (auth applied) once per HTTP attempt.
func TestWithRequestObserver_FiresPerAttempt(t *testing.T) {
	crt := &captureRT{}
	var calls int
	var sawAuth string
	c, err := NewClient("https://api.fga.example",
		WithAPIToken("tok"),
		WithBaseTransport(crt),
		WithRequestObserver(func(req *http.Request, _ *http.Response, _ error, _ time.Duration) {
			calls++
			sawAuth = req.Header.Get("Authorization")
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	req, err := c.NewRequest(context.Background(), http.MethodGet, "/stores", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Transport().RoundTrip(req); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Errorf("observer calls = %d, want 1", calls)
	}
	if sawAuth != "Bearer tok" {
		t.Errorf("observer saw Authorization = %q, want Bearer tok", sawAuth)
	}
}

// TestWithTokenSource_AppliesBearerThroughChain verifies that a caller-supplied
// oauth2.TokenSource authenticates requests and that WithBaseTransport is used
// as the innermost transport beneath the auth layer.
func TestWithTokenSource_AppliesBearerThroughChain(t *testing.T) {
	crt := &captureRT{}
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "abc", TokenType: "Bearer"})

	c, err := NewClient("https://api.fga.example",
		WithTokenSource(src),
		WithBaseTransport(crt),
	)
	if err != nil {
		t.Fatal(err)
	}

	req, err := c.NewRequest(context.Background(), http.MethodGet, "/stores", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Transport().RoundTrip(req); err != nil {
		t.Fatal(err)
	}
	if got := crt.last.Header.Get("Authorization"); got != "Bearer abc" {
		t.Errorf("Authorization = %q, want Bearer abc", got)
	}
}

func TestWithTokenSource_NilSourceFailsValidation(t *testing.T) {
	if _, err := NewClient("https://api.fga.example", WithTokenSource(nil)); err == nil {
		t.Fatal("expected validation error for nil token source")
	}
}

// TestWithBaseTransport_IsInnermost verifies the base transport receives the
// request even when no auth is configured.
func TestWithBaseTransport_IsInnermost(t *testing.T) {
	crt := &captureRT{}
	c, err := NewClient("https://api.fga.example", WithBaseTransport(crt))
	if err != nil {
		t.Fatal(err)
	}
	req, err := c.NewRequest(context.Background(), http.MethodGet, "/stores", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Transport().RoundTrip(req); err != nil {
		t.Fatal(err)
	}
	if crt.last == nil {
		t.Fatal("base transport was not invoked")
	}
}
