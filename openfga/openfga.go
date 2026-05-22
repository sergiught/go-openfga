package openfga

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
)

const defaultUserAgent = "go-openfga"

// Client is an OpenFGA API client. Construct it with NewClient.
type Client struct {
	client    *http.Client
	baseURL   *url.URL
	userAgent string

	storeID     string
	authModelID string

	// transport-layer config assembled in NewClient
	staticHeaders http.Header
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
func NewClient(apiURL string, opts ...Option) (*Client, error) {
	if !strings.HasSuffix(apiURL, "/") {
		apiURL += "/"
	}
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, errors.New("invalid api url: " + apiURL)
	}

	c := &Client{
		baseURL:       u,
		userAgent:     defaultUserAgent,
		staticHeaders: http.Header{},
		retry:         defaultRetryConfig(),
	}
	for _, o := range opts {
		o(c)
	}

	// Assemble the transport chain unless the caller supplied a full client.
	if c.client == nil {
		c.client = &http.Client{Transport: c.buildTransport(http.DefaultTransport)}
	}

	c.common.client = c
	c.Stores = (*StoresService)(&c.common)
	c.AuthorizationModels = (*AuthorizationModelsService)(&c.common)
	c.Tuples = (*TuplesService)(&c.common)
	c.Relationships = (*RelationshipsService)(&c.common)
	c.Assertions = (*AssertionsService)(&c.common)
	return c, nil
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

// WithUserAgent overrides the User-Agent header sent on every request.
func WithUserAgent(ua string) Option { return func(c *Client) { c.userAgent = ua } }

// WithHTTPClient supplies a fully-configured *http.Client (escape hatch). When
// set, the SDK does not assemble its own transport chain.
func WithHTTPClient(hc *http.Client) Option { return func(c *Client) { c.client = hc } }

// WithBaseURL overrides the API base URL after construction-time parsing.
func WithBaseURL(raw string) Option {
	return func(c *Client) {
		if !strings.HasSuffix(raw, "/") {
			raw += "/"
		}
		if u, err := url.Parse(raw); err == nil {
			c.baseURL = u
		}
	}
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

// BaseURL returns the API base URL the client targets.
func (c *Client) BaseURL() string { return c.baseURL.String() }
