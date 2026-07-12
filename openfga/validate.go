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
	if c.rawBaseURL == "" {
		return errors.New("openfga: no API URL; set FGA_API_URL, pass apiURL to NewClient, or use WithBaseURL")
	}
	raw := c.rawBaseURL
	if !strings.HasSuffix(raw, "/") {
		raw += "/"
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("openfga: invalid API URL %q: %w", c.rawBaseURL, err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("openfga: invalid API URL %q: need an http(s) scheme and host", c.rawBaseURL)
	}
	c.baseURL = u

	if c.storeID != "" && !ulidRE.MatchString(c.storeID) {
		return fmt.Errorf("openfga: invalid store ID %q: not a ULID", c.storeID)
	}
	if c.authModelID != "" && !ulidRE.MatchString(c.authModelID) {
		return fmt.Errorf("openfga: invalid authorization model ID %q: not a ULID", c.authModelID)
	}
	if c.auth != nil {
		if err := c.auth.validate(); err != nil {
			return err
		}
	}
	if c.retry != nil {
		if c.retry.MaxAttempts < 1 {
			return errors.New("openfga: retry MaxAttempts must be >= 1")
		}
		if c.retry.MinWait > c.retry.MaxWait {
			return errors.New("openfga: retry MinWait must be <= MaxWait")
		}
	}
	return nil
}
