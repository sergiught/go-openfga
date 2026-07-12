package openfga

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestPrivateKeyJWT_AssertionIsSignedAndParses(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	src := &privateKeyJWTSource{cfg: PrivateKeyJWTConfig{
		TokenURL:      "https://issuer.example/oauth/token",
		ClientID:      "client-123",
		Audience:      "https://issuer.example/",
		SigningKey:    key,
		SigningMethod: jwt.SigningMethodRS256,
		KeyID:         "kid-1",
	}}

	values, err := src.assertionValues()
	if err != nil {
		t.Fatal(err)
	}
	if values.Get("client_assertion_type") != "urn:ietf:params:oauth:client-assertion-type:jwt-bearer" {
		t.Errorf("assertion type = %q", values.Get("client_assertion_type"))
	}
	raw := values.Get("client_assertion")
	tok, err := jwt.Parse(raw, func(*jwt.Token) (any, error) { return &key.PublicKey, nil })
	if err != nil || !tok.Valid {
		t.Fatalf("assertion not valid: %v", err)
	}
	claims := tok.Claims.(jwt.MapClaims)
	if claims["iss"] != "client-123" || claims["sub"] != "client-123" {
		t.Errorf("iss/sub = %v/%v", claims["iss"], claims["sub"])
	}
	if claims["aud"] != "https://issuer.example/" {
		t.Errorf("aud = %v", claims["aud"])
	}
	if tok.Header["kid"] != "kid-1" {
		t.Errorf("kid = %v", tok.Header["kid"])
	}
}

func TestWithPrivateKeyJWT_WiresAuthTransport(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	c, _ := NewClient("https://api.fga.example", WithPrivateKeyJWT(PrivateKeyJWTConfig{
		TokenURL: "https://issuer.example/oauth/token", ClientID: "x",
		SigningKey: key, SigningMethod: jwt.SigningMethodRS256,
	}))
	if c.authTransport == nil {
		t.Fatal("auth transport not set")
	}
	_ = context.Background
	_ = url.Values{}
}

func TestPrivateKeyJWT_TokenSuccess(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	var gotForm url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		gotForm = r.Form
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"access_token":"tok-123","token_type":"Bearer","expires_in":3600}`)
	}))
	defer srv.Close()

	src := &privateKeyJWTSource{cfg: PrivateKeyJWTConfig{
		TokenURL:      srv.URL,
		ClientID:      "cid",
		Audience:      "https://issuer.example",
		APIAudience:   "https://api.fga.example",
		Scopes:        []string{"read", "write"},
		SigningKey:    key,
		SigningMethod: jwt.SigningMethodRS256,
		KeyID:         "kid-1",
	}}
	tok, err := src.Token()
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "tok-123" {
		t.Errorf("access token = %q", tok.AccessToken)
	}
	if tok.Expiry.IsZero() {
		t.Error("expiry should be set from expires_in")
	}
	if gotForm.Get("client_assertion") == "" {
		t.Error("client_assertion not posted")
	}
	if gotForm.Get("audience") != "https://api.fga.example" {
		t.Errorf("audience = %q, want API audience", gotForm.Get("audience"))
	}
	if gotForm.Get("scope") != "read write" {
		t.Errorf("scope = %q, want space-joined", gotForm.Get("scope"))
	}
}

func TestPrivateKeyJWT_TokenReturnsErrorOnNon2xx(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprint(w, `{"error":"invalid_client"}`)
	}))
	defer srv.Close()

	src := &privateKeyJWTSource{cfg: PrivateKeyJWTConfig{
		TokenURL:      srv.URL,
		ClientID:      "client-123",
		Audience:      srv.URL,
		SigningKey:    key,
		SigningMethod: jwt.SigningMethodRS256,
	}}
	tok, err := src.Token()
	if err == nil {
		t.Fatal("expected error for non-2xx token endpoint response, got nil")
	}
	if tok != nil {
		t.Errorf("expected nil token, got %v", tok)
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error %q should mention status 401", err.Error())
	}
}
