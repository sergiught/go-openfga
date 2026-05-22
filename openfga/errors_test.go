package openfga

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func newHTTPResp(status int, body, retryAfter string) *http.Response {
	req, _ := http.NewRequest(http.MethodPost, "https://api.fga.example/stores/s/check", nil)
	h := http.Header{}
	if retryAfter != "" {
		h.Set("Retry-After", retryAfter)
	}
	return &http.Response{
		StatusCode: status,
		Header:     h,
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
}

func TestCheckResponse_Success(t *testing.T) {
	if err := CheckResponse(newHTTPResp(200, "", "")); err != nil {
		t.Fatalf("want nil, got %v", err)
	}
}

func TestCheckResponse_RateLimit(t *testing.T) {
	err := CheckResponse(newHTTPResp(429, `{"code":"rate_limit","message":"slow down"}`, "3"))
	var rl *RateLimitError
	if !errors.As(err, &rl) {
		t.Fatalf("want *RateLimitError, got %T", err)
	}
	if rl.RetryAfter != 3*time.Second {
		t.Fatalf("want 3s, got %v", rl.RetryAfter)
	}
	if rl.Code != "rate_limit" {
		t.Fatalf("want code rate_limit, got %q", rl.Code)
	}
}

func TestCheckResponse_TypedStatuses(t *testing.T) {
	cases := []struct {
		status int
		check  func(error) bool
	}{
		{400, func(e error) bool { var t *ValidationError; return errors.As(e, &t) }},
		{401, func(e error) bool { var t *AuthenticationError; return errors.As(e, &t) }},
		{404, func(e error) bool { var t *NotFoundError; return errors.As(e, &t) }},
		{500, func(e error) bool { var t *InternalError; return errors.As(e, &t) }},
	}
	for _, c := range cases {
		err := CheckResponse(newHTTPResp(c.status, `{"code":"x","message":"y"}`, ""))
		if !c.check(err) {
			t.Errorf("status %d: wrong error type %T", c.status, err)
		}
	}
}

func TestCheckResponse_ErrorsAsReachesBase(t *testing.T) {
	err := CheckResponse(newHTTPResp(400, `{"code":"validation_error","message":"bad input"}`, ""))

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("want *ValidationError, got %T", err)
	}

	var base *ErrorResponse
	if !errors.As(err, &base) {
		t.Fatal("errors.As did not reach *ErrorResponse through Unwrap")
	}
	if base.Code != "validation_error" {
		t.Errorf("base.Code = %q, want %q", base.Code, "validation_error")
	}
	if base.Message != "bad input" {
		t.Errorf("base.Message = %q, want %q", base.Message, "bad input")
	}
}

func TestParseRetryAfter_HTTPDate(t *testing.T) {
	// Pick a time clearly in the future.
	future := time.Now().UTC().Add(10 * time.Second)
	header := future.Format(http.TimeFormat)
	r := &http.Response{Header: http.Header{"Retry-After": []string{header}}}
	d := parseRetryAfter(r)
	if d <= 0 {
		t.Errorf("expected positive duration for HTTP-date Retry-After, got %v", d)
	}
}
