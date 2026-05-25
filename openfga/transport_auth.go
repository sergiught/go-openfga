package openfga

import (
	"context"
	"crypto"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// wrapAuth places an auth transport in front of base. The auth transport is
// expected to be an *oauth2.Transport whose Base we set to base.
func wrapAuth(auth http.RoundTripper, base http.RoundTripper) http.RoundTripper {
	if ot, ok := auth.(*oauth2.Transport); ok {
		ot.Base = base
		return ot
	}
	if bt, ok := auth.(*bearerTransport); ok {
		bt.base = base
		return bt
	}
	return auth
}

// Pre-shared API token.

type apiTokenSource struct{ token string }

func (s *apiTokenSource) transport() http.RoundTripper {
	return &bearerTransport{token: s.token}
}

type bearerTransport struct {
	token string
	base  http.RoundTripper
}

func (t *bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r2 := req.Clone(req.Context())
	r2.Header.Set("Authorization", "Bearer "+t.token)
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(r2)
}

// WithAPIToken authenticates with a pre-shared key (Authorization: Bearer).
func WithAPIToken(token string) Option {
	return func(c *Client) { c.authTransport = (&apiTokenSource{token: token}).transport() }
}

// OAuth2 client credentials.

// ClientCredentialsConfig configures the OAuth2 client-credentials grant.
type ClientCredentialsConfig struct {
	TokenURL     string
	ClientID     string
	ClientSecret string
	Audience     string
	Scopes       []string
}

// WithClientCredentials authenticates via the OAuth2 client-credentials grant.
func WithClientCredentials(cfg ClientCredentialsConfig) Option {
	return func(c *Client) {
		params := map[string][]string{}
		if cfg.Audience != "" {
			params["audience"] = []string{cfg.Audience}
		}
		oc := &clientcredentials.Config{
			ClientID:       cfg.ClientID,
			ClientSecret:   cfg.ClientSecret,
			TokenURL:       cfg.TokenURL,
			Scopes:         cfg.Scopes,
			EndpointParams: params,
		}
		ts := oc.TokenSource(context.Background())
		c.authTransport = &oauth2.Transport{Source: oauth2.ReuseTokenSource(nil, ts)}
	}
}

// Private-key JWT (RFC 7523 client assertion).

// PrivateKeyJWTConfig configures client-credentials with a signed JWT assertion.
type PrivateKeyJWTConfig struct {
	TokenURL      string
	ClientID      string
	Audience      string // assertion "aud" (usually the token endpoint/issuer)
	APIAudience   string // OpenFGA API audience requested in the grant
	Scopes        []string
	SigningKey    crypto.PrivateKey // *rsa.PrivateKey or *ecdsa.PrivateKey
	SigningMethod jwt.SigningMethod
	KeyID         string
}

type privateKeyJWTSource struct{ cfg PrivateKeyJWTConfig }

// jwtBearerAssertionType is the RFC 7523 client-assertion type URN, not a
// credential.
//
//nolint:gosec // G101: this is a well-known OAuth2 URN, not a hardcoded secret.
const jwtBearerAssertionType = "urn:ietf:params:oauth:client-assertion-type:jwt-bearer"

// assertionValues builds the token-request form including a freshly signed JWT.
func (s *privateKeyJWTSource) assertionValues() (url.Values, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": s.cfg.ClientID,
		"sub": s.cfg.ClientID,
		"aud": s.cfg.Audience,
		"jti": newJTI(),
		"iat": now.Unix(),
		"exp": now.Add(5 * time.Minute).Unix(),
	}
	tok := jwt.NewWithClaims(s.cfg.SigningMethod, claims)
	if s.cfg.KeyID != "" {
		tok.Header["kid"] = s.cfg.KeyID
	}
	signed, err := tok.SignedString(s.cfg.SigningKey)
	if err != nil {
		return nil, err
	}
	v := url.Values{}
	v.Set("grant_type", "client_credentials")
	v.Set("client_assertion_type", jwtBearerAssertionType)
	v.Set("client_assertion", signed)
	if s.cfg.APIAudience != "" {
		v.Set("audience", s.cfg.APIAudience)
	}
	if len(s.cfg.Scopes) > 0 {
		v.Set("scope", strings.Join(s.cfg.Scopes, " "))
	}
	return v, nil
}

func newJTI() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Token implements oauth2.TokenSource: it POSTs the signed assertion and parses
// the access token. ReuseTokenSource (in WithPrivateKeyJWT) caches until expiry.
func (s *privateKeyJWTSource) Token() (*oauth2.Token, error) {
	values, err := s.assertionValues()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, s.cfg.TokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("openfga: token endpoint returned %d: %s", resp.StatusCode, snippet)
	}
	var tr struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}
	tok := &oauth2.Token{AccessToken: tr.AccessToken, TokenType: tr.TokenType}
	if tr.ExpiresIn > 0 {
		tok.Expiry = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	}
	return tok, nil
}

// WithPrivateKeyJWT authenticates using a signed JWT client assertion.
func WithPrivateKeyJWT(cfg PrivateKeyJWTConfig) Option {
	return func(c *Client) {
		src := &privateKeyJWTSource{cfg: cfg}
		c.authTransport = &oauth2.Transport{Source: oauth2.ReuseTokenSource(nil, src)}
	}
}
