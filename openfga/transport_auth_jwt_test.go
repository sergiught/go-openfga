package openfga

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/url"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestPrivateKeyJWT_AssertionIsSignedAndParses(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	src := &privateKeyJWTSource{cfg: PrivateKeyJWTConfig{
		TokenURL:      "https://issuer.example/oauth/token",
		ClientID:      "client-123",
		Audience:      "https://issuer.example/",
		SigningKey:     key,
		SigningMethod:  jwt.SigningMethodRS256,
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
