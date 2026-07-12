// Package config loads OpenFGA client settings from the environment.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/caarlos0/env/v11"
)

// Config holds the environment-derived client settings. Field values come from
// the FGA_* environment variables named in the struct tags.
type Config struct {
	APIURL       string   `env:"FGA_API_URL"`
	StoreID      string   `env:"FGA_STORE_ID"`
	AuthModelID  string   `env:"FGA_MODEL_ID"`
	APIToken     string   `env:"FGA_API_TOKEN"`
	ClientID     string   `env:"FGA_CLIENT_ID"`
	ClientSecret string   `env:"FGA_CLIENT_SECRET"`
	TokenIssuer  string   `env:"FGA_API_TOKEN_ISSUER"`
	Audience     string   `env:"FGA_API_AUDIENCE"`
	Scopes       []string `env:"FGA_API_SCOPES" envSeparator:","`
}

// Load decodes the environment into a Config and rejects a configuration that
// sets both an API token and client-credentials variables. It does not validate
// URL/ULID/credential-completeness/retry — those are checked once on the merged
// client state in the openfga package.
func Load() (Config, error) {
	var c Config
	if err := env.Parse(&c); err != nil {
		return Config{}, err
	}
	if c.APIToken != "" && c.HasClientCredentials() {
		return Config{}, errors.New("config: set either FGA_API_TOKEN or FGA_CLIENT_* credentials, not both")
	}
	return c, nil
}

// HasClientCredentials reports whether any client-credentials field is set.
func (c Config) HasClientCredentials() bool {
	return c.ClientID != "" || c.ClientSecret != "" || c.TokenIssuer != "" ||
		c.Audience != "" || len(c.Scopes) > 0
}

// NormalizeTokenURL turns an OAuth2 issuer into a full token endpoint, mirroring
// the official SDK: a bare host gets an https scheme, a missing or root path
// becomes /oauth/token, and non-http(s) schemes are rejected. An empty issuer
// returns "" so the caller's completeness check reports the missing endpoint.
func NormalizeTokenURL(issuer string) (string, error) {
	if issuer == "" {
		return "", nil
	}
	if !strings.Contains(issuer, "://") {
		issuer = "https://" + issuer
	}
	u, err := url.Parse(issuer)
	if err != nil {
		return "", fmt.Errorf("config: invalid FGA_API_TOKEN_ISSUER %q: %w", issuer, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("config: FGA_API_TOKEN_ISSUER scheme %q must be http or https", u.Scheme)
	}
	if u.Path == "" || u.Path == "/" {
		u.Path = "/oauth/token"
	}
	return u.String(), nil
}
