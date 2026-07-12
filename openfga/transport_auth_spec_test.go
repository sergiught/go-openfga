package openfga

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestAuthSpecValidate(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	cases := []struct {
		name    string
		spec    authSpec
		wantErr bool
	}{
		{"api token ok", &apiTokenSource{token: "t"}, false},
		{"api token empty", &apiTokenSource{token: ""}, true},
		{"client creds ok", &clientCredentialsSpec{tokenURL: "https://t/x", clientID: "c", clientSecret: "s"}, false},
		{"client creds no secret", &clientCredentialsSpec{tokenURL: "https://t/x", clientID: "c"}, true},
		{"client creds no url", &clientCredentialsSpec{clientID: "c", clientSecret: "s"}, true},
		{"client creds no id", &clientCredentialsSpec{tokenURL: "https://t/x", clientSecret: "s"}, true},
		{"jwt ok", &privateKeyJWTSource{cfg: PrivateKeyJWTConfig{TokenURL: "https://t/x", ClientID: "c", SigningKey: key, SigningMethod: jwt.SigningMethodRS256}}, false},
		{"jwt no url", &privateKeyJWTSource{cfg: PrivateKeyJWTConfig{ClientID: "c", SigningKey: key, SigningMethod: jwt.SigningMethodRS256}}, true},
		{"jwt no client id", &privateKeyJWTSource{cfg: PrivateKeyJWTConfig{TokenURL: "https://t/x", SigningKey: key, SigningMethod: jwt.SigningMethodRS256}}, true},
		{"jwt no key", &privateKeyJWTSource{cfg: PrivateKeyJWTConfig{TokenURL: "https://t/x", ClientID: "c", SigningMethod: jwt.SigningMethodRS256}}, true},
		{"jwt no method", &privateKeyJWTSource{cfg: PrivateKeyJWTConfig{TokenURL: "https://t/x", ClientID: "c", SigningKey: key}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.spec.validate()
			if tc.wantErr != (err != nil) {
				t.Fatalf("validate() err = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

func TestWithClientCredentials_SetsAuthSpec(t *testing.T) {
	c := &Client{}
	WithClientCredentials(ClientCredentialsConfig{TokenURL: "https://t/x", ClientID: "c", ClientSecret: "s"})(c)
	if c.auth == nil {
		t.Fatal("auth spec not set")
	}
	if err := c.auth.validate(); err != nil {
		t.Fatalf("spec invalid: %v", err)
	}
}
