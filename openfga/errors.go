package openfga

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// OpenFGA error codes carried in ErrorResponse.Code. This is not the full set
// the server may return; match on ErrorResponse.Code directly for others. See
// https://openfga.dev/api/service for the authoritative list.
const (
	CodeValidationError            = "validation_error"
	CodeInvalidAuthorizationModel  = "invalid_authorization_model"
	CodeTypeNotFound               = "type_not_found"
	CodeRelationNotFound           = "relation_not_found"
	CodeStoreIDInvalidLength       = "store_id_invalid_length"
	CodeAuthorizationModelNotFound = "authorization_model_not_found"
	CodeInternalError              = "internal_error"
)

// ErrorResponse is the base error returned for any non-2xx API response.
type ErrorResponse struct {
	Response *http.Response `json:"-"`
	Code     string         `json:"code"`
	Message  string         `json:"message"`
}

// RequestID returns the OpenFGA request correlation ID from the response
// headers, or "" if absent. Quote it when reporting a server-side error.
func (e *ErrorResponse) RequestID() string {
	if e.Response == nil {
		return ""
	}
	return e.Response.Header.Get(requestIDHeader)
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

// ValidationError is returned for HTTP 400 responses.
type ValidationError struct{ *ErrorResponse }

// Unwrap allows errors.As to reach the embedded *ErrorResponse.
func (e *ValidationError) Unwrap() error { return e.ErrorResponse }

// AuthenticationError is returned for HTTP 401 and 403 responses.
type AuthenticationError struct{ *ErrorResponse }

// Unwrap allows errors.As to reach the embedded *ErrorResponse.
func (e *AuthenticationError) Unwrap() error { return e.ErrorResponse }

// NotFoundError is returned for HTTP 404 responses.
type NotFoundError struct{ *ErrorResponse }

// Unwrap allows errors.As to reach the embedded *ErrorResponse.
func (e *NotFoundError) Unwrap() error { return e.ErrorResponse }

// InternalError is returned for HTTP 5xx responses.
type InternalError struct{ *ErrorResponse }

// Unwrap allows errors.As to reach the embedded *ErrorResponse.
func (e *InternalError) Unwrap() error { return e.ErrorResponse }

// RateLimitError is returned on HTTP 429.
type RateLimitError struct {
	*ErrorResponse
	RetryAfter time.Duration
}

// Unwrap allows errors.As to reach the embedded *ErrorResponse.
func (e *RateLimitError) Unwrap() error { return e.ErrorResponse }

// classifyResponse maps an *http.Response to a typed error, or nil for 2xx.
// It consumes the response body.
func classifyResponse(r *http.Response) error {
	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}
	base := &ErrorResponse{Response: r}
	if data, _ := io.ReadAll(r.Body); len(data) > 0 {
		_ = json.Unmarshal(data, base) // Best-effort; leave fields empty on failure.
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

// maxRetryAfterSeconds caps a numeric Retry-After so the conversion to a
// nanosecond time.Duration cannot overflow int64 (~292 years).
const maxRetryAfterSeconds = 100 * 365 * 24 * 60 * 60

func parseRetryAfter(r *http.Response) time.Duration {
	v := r.Header.Get("Retry-After")
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil {
		if secs <= 0 {
			return 0
		}
		// Clamp before multiplying so a huge value cannot overflow the int64
		// nanosecond duration into a negative wait. The ceiling (100 years) is
		// far beyond any real Retry-After; callers clamp it to MaxWait anyway.
		if secs > maxRetryAfterSeconds {
			secs = maxRetryAfterSeconds
		}
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}
