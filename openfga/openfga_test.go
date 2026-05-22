package openfga

import (
	"net/http"
	"testing"
)

func TestNewClient_DefaultsAndOptions(t *testing.T) {
	c, err := NewClient("https://api.fga.example",
		WithStoreID("store1"),
		WithAuthorizationModelID("model1"),
		WithUserAgent("test-agent"))
	if err != nil {
		t.Fatal(err)
	}
	if c.storeID != "store1" || c.authModelID != "model1" {
		t.Errorf("ids = %q %q", c.storeID, c.authModelID)
	}
	if c.userAgent != "test-agent" {
		t.Errorf("ua = %q", c.userAgent)
	}
	if c.Stores == nil || c.Relationships == nil || c.Tuples == nil ||
		c.AuthorizationModels == nil || c.Assertions == nil {
		t.Error("services not wired")
	}
}

func TestNewClient_BaseURLTrailingSlash(t *testing.T) {
	c, _ := NewClient("https://api.fga.example")
	if c.baseURL.String() != "https://api.fga.example/" {
		t.Errorf("baseURL = %q (want trailing slash)", c.baseURL.String())
	}
}

func TestNewClient_RejectsBadURL(t *testing.T) {
	if _, err := NewClient("://nope"); err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestWithHTTPClient_Honored(t *testing.T) {
	hc := &http.Client{}
	c, _ := NewClient("https://api.fga.example", WithHTTPClient(hc))
	if c.client != hc {
		t.Error("WithHTTPClient not honored")
	}
}
