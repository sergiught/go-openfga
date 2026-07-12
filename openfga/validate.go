package openfga

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ulidRE matches a Crockford base32 ULID (26 chars; first char 0-7).
var ulidRE = regexp.MustCompile(`^[0-7][0-9A-HJKMNP-TV-Za-hjkmnp-tv-z]{25}$`)

// validate checks the merged client state and, on success, parses rawBaseURL
// into baseURL. It is the single validation point for env- and option-supplied
// configuration.
func (c *Client) validate() error {
	u, err := parseBaseURL(c.rawBaseURL)
	if err != nil {
		return err
	}
	c.baseURL = u

	if err := c.validateIDs(); err != nil {
		return err
	}
	if c.auth != nil {
		if err := c.auth.validate(); err != nil {
			return err
		}
	}
	return c.validateRetry()
}

// parseBaseURL normalizes and parses an API base URL, requiring an http(s)
// scheme and a host. Error messages quote the caller's original input.
func parseBaseURL(raw string) (*url.URL, error) {
	if raw == "" {
		return nil, errors.New("openfga: no API URL; set FGA_API_URL, pass apiURL to NewClient, or use WithBaseURL")
	}
	normalized := raw
	if !strings.HasSuffix(normalized, "/") {
		normalized += "/"
	}
	u, err := url.Parse(normalized)
	if err != nil {
		return nil, fmt.Errorf("openfga: invalid API URL %q: %w", raw, err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("openfga: invalid API URL %q: need an http(s) scheme and host", raw)
	}
	return u, nil
}

// validateIDs checks that any configured store and model IDs are ULIDs.
func (c *Client) validateIDs() error {
	if c.storeID != "" && !ulidRE.MatchString(c.storeID) {
		return fmt.Errorf("openfga: invalid store ID %q: not a ULID", c.storeID)
	}
	if c.authModelID != "" && !ulidRE.MatchString(c.authModelID) {
		return fmt.Errorf("openfga: invalid authorization model ID %q: not a ULID", c.authModelID)
	}
	return nil
}

// validateRetry checks the retry configuration, if any.
func (c *Client) validateRetry() error {
	if c.retry == nil {
		return nil
	}
	if c.retry.MaxAttempts < 1 {
		return errors.New("openfga: retry MaxAttempts must be >= 1")
	}
	if c.retry.MinWait > c.retry.MaxWait {
		return errors.New("openfga: retry MinWait must be <= MaxWait")
	}
	return nil
}
