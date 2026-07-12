package openfga

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"

	"github.com/sergiught/go-openfga/internal/config"
)

// modulePath is this module's import path, used to find its version in build info.
const modulePath = "github.com/sergiught/go-openfga"

// defaultUserAgent is the User-Agent sent when the caller does not override it.
// It embeds the module version recorded in the consuming binary's build info
// (e.g. "go-openfga/1.2.0"), falling back to the bare name when no version is
// available (local builds, replace directives, or running the SDK's own tests).
var defaultUserAgent = buildUserAgent()

func buildUserAgent() string {
	const name = "go-openfga"
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return name
	}
	for _, dep := range info.Deps {
		if dep.Path == modulePath && dep.Version != "" && dep.Version != "(devel)" {
			return name + "/" + strings.TrimPrefix(dep.Version, "v")
		}
	}
	return name
}

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
	baseTransport http.RoundTripper
	observer      RequestObserver
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
// Construction is explicit: apiURL and opts are the only inputs, applied in
// order so later opts win. It does not read the environment — use
// NewClientFromEnv or EnvOptions to opt into FGA_* configuration.
func NewClient(apiURL string, opts ...Option) (*Client, error) {
	c := &Client{
		userAgent:     defaultUserAgent,
		staticHeaders: http.Header{},
		retry:         defaultRetryConfig(),
	}
	if apiURL != "" {
		c.rawBaseURL = apiURL
	}
	for _, o := range opts {
		o(c)
	}

	if err := c.validate(); err != nil {
		return nil, err
	}
	return c.finish(), nil
}

// NewClientFromEnv creates a client from FGA_* environment variables, with opts
// overriding the environment-derived configuration. It is the explicit opt-in
// for environment configuration; NewClient alone never reads the environment.
func NewClientFromEnv(opts ...Option) (*Client, error) {
	envOpts, err := EnvOptions()
	if err != nil {
		return nil, err
	}
	return NewClient("", append(envOpts, opts...)...)
}

// EnvOptions resolves FGA_* environment variables into client options for
// NewClient. Place them ahead of your own options so explicit settings win:
//
//	envOpts, err := openfga.EnvOptions()
//	client, err := openfga.NewClient("", append(envOpts, opts...)...)
func EnvOptions() ([]Option, error) {
	envCfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return configOptions(envCfg)
}

// configOptions turns environment-derived settings into options, reusing the
// public option constructors so env and explicit configuration share one path.
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
	if c.client == nil {
		base := c.baseTransport
		if base == nil {
			base = http.DefaultTransport
		}
		c.client = &http.Client{Transport: c.buildTransport(base)}
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

// buildTransport composes: retry -> auth -> static-headers -> observer -> base.
// The observer sits innermost so it sees each attempt's fully-decorated request.
func (c *Client) buildTransport(base http.RoundTripper) http.RoundTripper {
	rt := base
	if c.observer != nil {
		rt = &observerTransport{base: rt, obs: c.observer}
	}
	if len(c.staticHeaders) > 0 {
		rt = &headerTransport{base: rt, header: c.staticHeaders}
	}
	if c.auth != nil {
		// Token fetches go through base (configured transport) with a bounded
		// timeout, but bypass the retry/header layers above.
		tokenClient := &http.Client{Transport: base, Timeout: tokenFetchTimeout}
		rt = c.auth.transport(rt, tokenClient)
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
// WithHeaders, WithRetry, and WithBaseTransport have no effect — configure auth, headers, and
// retries on the supplied client's Transport yourself.
func WithHTTPClient(hc *http.Client) Option { return func(c *Client) { c.client = hc } }

// WithBaseTransport sets the innermost http.RoundTripper beneath the SDK's
// retry, auth, and header layers. Use it to add tracing, metrics, logging, or a
// custom dialer while keeping the SDK's auth and retries — for example
// otelhttp.NewTransport(nil) for per-attempt spans. It also becomes the base
// for out-of-band OAuth2 token fetches. Defaults to http.DefaultTransport.
// Ignored when WithHTTPClient supplies a full client.
func WithBaseTransport(rt http.RoundTripper) Option {
	return func(c *Client) { c.baseTransport = rt }
}

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

// StoreID returns the client's default store ID (empty if unset).
func (c *Client) StoreID() string { return c.storeID }

// SetStoreID updates the client's default store ID, validating it as a ULID
// (an empty string clears it). Intended for reconfiguring a client between
// requests; it is not safe to call concurrently with in-flight requests. Use
// the per-call WithStore option for concurrent overrides.
func (c *Client) SetStoreID(id string) error {
	if id != "" && !ulidRE.MatchString(id) {
		return fmt.Errorf("openfga: invalid store ID %q: not a ULID", id)
	}
	c.storeID = id
	return nil
}

// AuthorizationModelID returns the client's default authorization model ID
// (empty if unset).
func (c *Client) AuthorizationModelID() string { return c.authModelID }

// SetAuthorizationModelID updates the client's default authorization model ID,
// validating it as a ULID (an empty string clears it). Same concurrency caveat
// as SetStoreID; use the per-call WithAuthorizationModel option for concurrent
// overrides.
func (c *Client) SetAuthorizationModelID(id string) error {
	if id != "" && !ulidRE.MatchString(id) {
		return fmt.Errorf("openfga: invalid authorization model ID %q: not a ULID", id)
	}
	c.authModelID = id
	return nil
}

// DefaultConsistency returns the client's default read consistency.
func (c *Client) DefaultConsistency() ConsistencyPreference { return c.consistency }

// SetDefaultConsistency updates the client's default read consistency. Same
// concurrency caveat as SetStoreID; use the per-call WithConsistency option for
// concurrent overrides.
func (c *Client) SetDefaultConsistency(cons ConsistencyPreference) { c.consistency = cons }

// BaseURL returns the API base URL the client targets.
func (c *Client) BaseURL() string { return c.baseURL.String() }

// Transport returns the http.RoundTripper the client uses: the assembled
// retry/auth/header chain, or the transport of a client supplied via
// WithHTTPClient. It lets callers reuse the SDK's configured transport, for
// example to wrap the whole logical request (across retries) in a span.
func (c *Client) Transport() http.RoundTripper { return c.client.Transport }
