package openfga

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/sergiught/go-openfga/internal/config"
)

const defaultUserAgent = "go-openfga"

// Client is an OpenFGA API client. Construct it with NewClient.
type Client struct {
	client     *http.Client
	baseURL    *url.URL
	rawBaseURL string
	userAgent  string

	storeID     string
	authModelID string
	consistency ConsistencyPreference

	// Transport-layer config assembled in NewClient.
	staticHeaders http.Header
	auth          authSpec
	authTransport http.RoundTripper
	retry         *RetryConfig

	common service

	Stores              *StoresService
	AuthorizationModels *AuthorizationModelsService
	Tuples              *TuplesService
	Relationships       *RelationshipsService
	Assertions          *AssertionsService
}

type service struct{ client *Client }

// Option configures a Client during NewClient.
type Option func(*Client)

// NewClient creates a client targeting apiURL (e.g. "https://api.fga.example").
// Configuration is layered lowest-to-highest: environment (FGA_*), then apiURL,
// then functional options — expressed as the order they are applied below.
func NewClient(apiURL string, opts ...Option) (*Client, error) {
	envCfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	envOpts, err := configOptions(envCfg)
	if err != nil {
		return nil, err
	}

	c := &Client{
		userAgent:     defaultUserAgent,
		staticHeaders: http.Header{},
		retry:         defaultRetryConfig(),
	}

	layered := envOpts
	if apiURL != "" {
		layered = append(layered, WithBaseURL(apiURL))
	}
	layered = append(layered, opts...)
	for _, o := range layered {
		o(c)
	}

	if err := c.validate(); err != nil {
		return nil, err
	}
	return c.finish(), nil
}

// configOptions turns environment-derived settings into the lowest-precedence
// option layer, reusing the public option constructors so env and explicit
// configuration share one code path.
func configOptions(cfg config.Config) ([]Option, error) {
	var opts []Option
	if cfg.APIURL != "" {
		opts = append(opts, WithBaseURL(cfg.APIURL))
	}
	if cfg.StoreID != "" {
		opts = append(opts, WithStoreID(cfg.StoreID))
	}
	if cfg.AuthModelID != "" {
		opts = append(opts, WithAuthorizationModelID(cfg.AuthModelID))
	}
	switch {
	case cfg.APIToken != "":
		opts = append(opts, WithAPIToken(cfg.APIToken))
	case cfg.HasClientCredentials():
		tokenURL, err := config.NormalizeTokenURL(cfg.TokenIssuer)
		if err != nil {
			return nil, err
		}
		opts = append(opts, WithClientCredentials(ClientCredentialsConfig{
			TokenURL:     tokenURL,
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Audience:     cfg.Audience,
			Scopes:       cfg.Scopes,
		}))
	}
	return opts, nil
}

// finish assembles the HTTP client and service handles after validation.
func (c *Client) finish() *Client {
	c.buildHTTPClient()
	c.wireServices()
	return c
}

// buildHTTPClient assembles the auth + transport chain, unless the caller
// supplied a full *http.Client via WithHTTPClient.
func (c *Client) buildHTTPClient() {
	if c.auth != nil {
		c.authTransport = c.auth.transport()
	}
	if c.client == nil {
		c.client = &http.Client{Transport: c.buildTransport(http.DefaultTransport)}
	}
}

// wireServices points the sub-service handles at the client.
func (c *Client) wireServices() {
	c.common.client = c
	c.Stores = (*StoresService)(&c.common)
	c.AuthorizationModels = (*AuthorizationModelsService)(&c.common)
	c.Tuples = (*TuplesService)(&c.common)
	c.Relationships = (*RelationshipsService)(&c.common)
	c.Assertions = (*AssertionsService)(&c.common)
}

// buildTransport composes: retry -> auth -> static-headers -> base.
func (c *Client) buildTransport(base http.RoundTripper) http.RoundTripper {
	rt := base
	if len(c.staticHeaders) > 0 {
		rt = &headerTransport{base: rt, header: c.staticHeaders}
	}
	if c.authTransport != nil {
		rt = wrapAuth(c.authTransport, rt)
	}
	if c.retry != nil {
		rt = &retryTransport{base: rt, cfg: *c.retry}
	}
	return rt
}

// WithStoreID sets the default OpenFGA store ID used by all requests.
func WithStoreID(id string) Option { return func(c *Client) { c.storeID = id } }

// WithAuthorizationModelID sets the default authorization model ID used by all requests.
func WithAuthorizationModelID(id string) Option { return func(c *Client) { c.authModelID = id } }

// WithDefaultConsistency sets the read consistency applied to all relationship
// query and tuple read requests. A per-call WithConsistency option overrides it.
func WithDefaultConsistency(cons ConsistencyPreference) Option {
	return func(c *Client) { c.consistency = cons }
}

// WithUserAgent overrides the User-Agent header sent on every request.
func WithUserAgent(ua string) Option { return func(c *Client) { c.userAgent = ua } }

// WithHTTPClient supplies a fully-configured *http.Client (escape hatch). When set, the SDK does NOT
// assemble its own transport chain, so WithAPIToken, WithClientCredentials, WithPrivateKeyJWT,
// WithHeaders, and WithRetry have no effect — configure auth, headers, and retries on the supplied
// client's Transport yourself.
func WithHTTPClient(hc *http.Client) Option { return func(c *Client) { c.client = hc } }

// WithBaseURL overrides the API base URL (highest precedence).
func WithBaseURL(raw string) Option {
	return func(c *Client) { c.rawBaseURL = raw }
}

// WithHeaders adds static headers applied to every request.
func WithHeaders(h http.Header) Option {
	return func(c *Client) {
		for k, vs := range h {
			for _, v := range vs {
				c.staticHeaders.Add(k, v)
			}
		}
	}
}

// storeFor resolves the effective store ID for a call (per-call override wins).
func (c *Client) storeFor(rc *requestConfig) (string, error) {
	id := rc.storeID
	if id == "" {
		id = c.storeID
	}
	if id == "" {
		return "", errors.New("no store ID set; use WithStoreID or WithStore")
	}
	return id, nil
}

// modelFor resolves the effective authorization model ID (may be empty).
func (c *Client) modelFor(rc *requestConfig) string {
	if rc.authModelID != "" {
		return rc.authModelID
	}
	return c.authModelID
}

// consistencyFor resolves the read consistency for a call (per-call override wins).
func (c *Client) consistencyFor(rc *requestConfig) ConsistencyPreference {
	if rc.consistency != "" {
		return rc.consistency
	}
	return c.consistency
}

// BaseURL returns the API base URL the client targets.
func (c *Client) BaseURL() string { return c.baseURL.String() }
