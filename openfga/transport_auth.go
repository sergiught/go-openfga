package openfga

import (
	"context"
	"crypto"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
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

// tokenFetchTimeout bounds each out-of-band OAuth2 token request so a slow or
// hostile token endpoint cannot wedge every API call behind an unbounded fetch.
const tokenFetchTimeout = 30 * time.Second

// authSpec is a credential configuration that can be validated and turned into
// an auth RoundTripper. Options store a spec; NewClient validates and builds it.
//
// The transport method wraps base with credential handling and returns the
// result. Specs performing out-of-band token fetches use tokenClient for those
// requests, so token traffic honors the configured base transport and a bounded
// timeout.
type authSpec interface {
	validate() error
	transport(base http.RoundTripper, tokenClient *http.Client) http.RoundTripper
}

// Pre-shared API token.

type apiTokenSource struct{ token string }

func (s *apiTokenSource) transport(base http.RoundTripper, _ *http.Client) http.RoundTripper {
	return &bearerTransport{token: s.token, base: base}
}

func (s *apiTokenSource) validate() error {
	if s.token == "" {
		return errors.New("openfga: api token is empty")
	}
	return nil
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
	return func(c *Client) { c.auth = &apiTokenSource{token: token} }
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

type clientCredentialsSpec struct {
	tokenURL     string
	clientID     string
	clientSecret string
	audience     string
	scopes       []string
}

func (s *clientCredentialsSpec) validate() error {
	switch {
	case s.tokenURL == "":
		return errors.New("openfga: client credentials require a token URL")
	case s.clientID == "":
		return errors.New("openfga: client credentials require a client ID")
	case s.clientSecret == "":
		return errors.New("openfga: client credentials require a client secret")
	}
	return nil
}

func (s *clientCredentialsSpec) transport(base http.RoundTripper, tokenClient *http.Client) http.RoundTripper {
	params := map[string][]string{}
	if s.audience != "" {
		params["audience"] = []string{s.audience}
	}
	oc := &clientcredentials.Config{
		ClientID:       s.clientID,
		ClientSecret:   s.clientSecret,
		TokenURL:       s.tokenURL,
		Scopes:         s.scopes,
		EndpointParams: params,
	}
	// Route token fetches through tokenClient (bounded timeout, configured base
	// transport) rather than oauth2's default http.DefaultClient.
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, tokenClient)
	ts := oc.TokenSource(ctx)
	return &oauth2.Transport{Source: oauth2.ReuseTokenSource(nil, ts), Base: base}
}

// WithClientCredentials authenticates via the OAuth2 client-credentials grant.
func WithClientCredentials(cfg ClientCredentialsConfig) Option {
	return func(c *Client) {
		c.auth = &clientCredentialsSpec{
			tokenURL:     cfg.TokenURL,
			clientID:     cfg.ClientID,
			clientSecret: cfg.ClientSecret,
			audience:     cfg.Audience,
			scopes:       cfg.Scopes,
		}
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

type privateKeyJWTSource struct {
	cfg        PrivateKeyJWTConfig
	httpClient *http.Client // token-fetch client; set when built into a transport.
}

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
	hc := s.httpClient
	if hc == nil {
		// No transport-supplied client (e.g. Token called directly): still bound
		// the fetch with a timeout rather than falling back to no deadline.
		hc = &http.Client{Timeout: tokenFetchTimeout}
	}
	resp, err := hc.Do(req)
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

func (s *privateKeyJWTSource) validate() error {
	switch {
	case s.cfg.TokenURL == "":
		return errors.New("openfga: private key JWT requires a token URL")
	case s.cfg.ClientID == "":
		return errors.New("openfga: private key JWT requires a client ID")
	case s.cfg.SigningKey == nil:
		return errors.New("openfga: private key JWT requires a signing key")
	case s.cfg.SigningMethod == nil:
		return errors.New("openfga: private key JWT requires a signing method")
	}
	return nil
}

func (s *privateKeyJWTSource) transport(base http.RoundTripper, tokenClient *http.Client) http.RoundTripper {
	s.httpClient = tokenClient
	return &oauth2.Transport{Source: oauth2.ReuseTokenSource(nil, s), Base: base}
}

// WithPrivateKeyJWT authenticates using a signed JWT client assertion.
func WithPrivateKeyJWT(cfg PrivateKeyJWTConfig) Option {
	return func(c *Client) { c.auth = &privateKeyJWTSource{cfg: cfg} }
}

// tokenSourceSpec authenticates with a caller-supplied oauth2.TokenSource.
type tokenSourceSpec struct{ src oauth2.TokenSource }

func (s *tokenSourceSpec) validate() error {
	if s.src == nil {
		return errors.New("openfga: WithTokenSource requires a non-nil oauth2.TokenSource")
	}
	return nil
}

func (s *tokenSourceSpec) transport(base http.RoundTripper, _ *http.Client) http.RoundTripper {
	return &oauth2.Transport{Source: oauth2.ReuseTokenSource(nil, s.src), Base: base}
}

// WithTokenSource authenticates every request with a bearer token obtained from
// any oauth2.TokenSource, for credential sources beyond the built-in modes
// (e.g. Vault, workload identity, a pre-existing token source). The SDK caches
// tokens via oauth2.ReuseTokenSource and keeps its retry and header chain
// beneath the auth layer. Pass exactly one authentication option to NewClient.
func WithTokenSource(src oauth2.TokenSource) Option {
	return func(c *Client) { c.auth = &tokenSourceSpec{src: src} }
}
