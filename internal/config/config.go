// Package config loads OpenFGA client settings from the environment.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
)

// Config holds the environment-derived client settings. Field values come from
// the FGA_* environment variables read in Load.
type Config struct {
	APIURL       string   // FGA_API_URL
	StoreID      string   // FGA_STORE_ID
	AuthModelID  string   // FGA_MODEL_ID
	APIToken     string   // FGA_API_TOKEN
	ClientID     string   // FGA_CLIENT_ID
	ClientSecret string   // FGA_CLIENT_SECRET
	TokenIssuer  string   // FGA_API_TOKEN_ISSUER
	Audience     string   // FGA_API_AUDIENCE
	Scopes       []string // FGA_API_SCOPES (comma-separated)
}

// Load reads the FGA_* environment into a Config and rejects a configuration
// that sets both an API token and client-credentials variables. It does not
// validate URL/ULID/credential-completeness/retry — those are checked once on
// the merged client state in the openfga package.
func Load() (Config, error) {
	c := Config{
		APIURL:       os.Getenv("FGA_API_URL"),
		StoreID:      os.Getenv("FGA_STORE_ID"),
		AuthModelID:  os.Getenv("FGA_MODEL_ID"),
		APIToken:     os.Getenv("FGA_API_TOKEN"),
		ClientID:     os.Getenv("FGA_CLIENT_ID"),
		ClientSecret: os.Getenv("FGA_CLIENT_SECRET"),
		TokenIssuer:  os.Getenv("FGA_API_TOKEN_ISSUER"),
		Audience:     os.Getenv("FGA_API_AUDIENCE"),
		Scopes:       splitScopes(os.Getenv("FGA_API_SCOPES")),
	}
	if c.APIToken != "" && c.HasClientCredentials() {
		return Config{}, errors.New("config: set either FGA_API_TOKEN or FGA_CLIENT_* credentials, not both")
	}
	return c, nil
}

// splitScopes parses a comma-separated FGA_API_SCOPES value, trimming spaces
// and dropping empty entries. An empty input yields a nil slice.
func splitScopes(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	scopes := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			scopes = append(scopes, s)
		}
	}
	if len(scopes) == 0 {
		return nil
	}
	return scopes
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
