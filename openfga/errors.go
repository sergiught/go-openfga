package openfga

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// ErrorResponse is the base error returned for any non-2xx API response.
type ErrorResponse struct {
	Response *http.Response `json:"-"`
	Code     string         `json:"code"`
	Message  string         `json:"message"`
}

func (e *ErrorResponse) Error() string {
	u := ""
	if e.Response != nil && e.Response.Request != nil {
		u = e.Response.Request.Method + " " + e.Response.Request.URL.String()
	}
	return fmt.Sprintf("%s: %d %s %s", u, e.statusCode(), e.Code, e.Message)
}

func (e *ErrorResponse) statusCode() int {
	if e.Response == nil {
		return 0
	}
	return e.Response.StatusCode
}

// Typed errors. Each embeds *ErrorResponse so errors.As reaches the base too.
type ValidationError struct{ *ErrorResponse }
type AuthenticationError struct{ *ErrorResponse }
type NotFoundError struct{ *ErrorResponse }
type InternalError struct{ *ErrorResponse }

// RateLimitError is returned on HTTP 429.
type RateLimitError struct {
	*ErrorResponse
	RetryAfter time.Duration
}

// CheckResponse maps an *http.Response to a typed error, or nil for 2xx.
// It consumes the response body.
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}
	base := &ErrorResponse{Response: r}
	if data, _ := io.ReadAll(r.Body); len(data) > 0 {
		_ = json.Unmarshal(data, base) // best-effort; leave fields empty on failure
	}
	switch {
	case r.StatusCode == http.StatusTooManyRequests:
		return &RateLimitError{ErrorResponse: base, RetryAfter: parseRetryAfter(r)}
	case r.StatusCode == http.StatusBadRequest:
		return &ValidationError{base}
	case r.StatusCode == http.StatusUnauthorized, r.StatusCode == http.StatusForbidden:
		return &AuthenticationError{base}
	case r.StatusCode == http.StatusNotFound:
		return &NotFoundError{base}
	case r.StatusCode >= 500:
		return &InternalError{base}
	default:
		return base
	}
}

func parseRetryAfter(r *http.Response) time.Duration {
	v := r.Header.Get("Retry-After")
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}
