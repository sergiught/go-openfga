package openfga

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

// scriptRT returns the queued responses/status codes in order.
type scriptRT struct {
	statuses []int
	calls    int
	bodies   []string
}

func (s *scriptRT) RoundTrip(r *http.Request) (*http.Response, error) {
	// Drain body to confirm it is replayable across attempts.
	if r.Body != nil {
		_, _ = io.Copy(io.Discard, r.Body)
	}
	i := s.calls
	s.calls++
	status := s.statuses[i]
	body := ""
	if i < len(s.bodies) {
		body = s.bodies[i]
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     http.Header{},
		Request:    r,
	}, nil
}

func noSleep(time.Duration) {}

func TestRetry_RetriesOn429ThenSucceeds(t *testing.T) {
	rt := &retryTransport{
		base:  &scriptRT{statuses: []int{429, 429, 200}},
		cfg:   RetryConfig{MaxAttempts: 3, MinWait: time.Millisecond, MaxWait: time.Millisecond, RetryableStatus: []int{429}},
		sleep: noSleep,
	}
	req, _ := http.NewRequest(http.MethodPost, "https://x/", bytes.NewBufferString(`{}`))
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("status = %d", resp.StatusCode)
	}
}

func TestRetry_DoesNotRetry5xxByDefault(t *testing.T) {
	base := &scriptRT{statuses: []int{500, 200}}
	rt := &retryTransport{base: base, cfg: RetryConfig{MaxAttempts: 3, RetryableStatus: []int{429}}, sleep: noSleep}
	req, _ := http.NewRequest(http.MethodGet, "https://x/", nil)
	resp, _ := rt.RoundTrip(req)
	if resp.StatusCode != 500 {
		t.Errorf("status = %d (should not retry 5xx by default)", resp.StatusCode)
	}
	if base.calls != 1 {
		t.Errorf("calls = %d (want 1)", base.calls)
	}
}

func TestRetry_RespectsMaxAttempts(t *testing.T) {
	base := &scriptRT{statuses: []int{429, 429, 429, 429}}
	rt := &retryTransport{base: base, cfg: RetryConfig{MaxAttempts: 3, MinWait: time.Millisecond, RetryableStatus: []int{429}}, sleep: noSleep}
	req, _ := http.NewRequest(http.MethodGet, "https://x/", nil)
	_, _ = rt.RoundTrip(req)
	if base.calls != 3 {
		t.Errorf("calls = %d (want 3 = MaxAttempts)", base.calls)
	}
}

func TestRetry_HonorsRetryAfterHeader(t *testing.T) {
	// Build a custom base that adds Retry-After on the first response.
	var sleptFor time.Duration
	recordSleep := func(d time.Duration) { sleptFor = d }

	innerBase := &retryAfterScriptRT{statuses: []int{429, 200}, retryAfter: "1"}
	rt := &retryTransport{
		base:  innerBase,
		cfg:   RetryConfig{MaxAttempts: 2, MinWait: time.Millisecond, RetryableStatus: []int{429}, HonorRetryAfter: true},
		sleep: recordSleep,
	}
	req, _ := http.NewRequest(http.MethodGet, "https://x/", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("status = %d", resp.StatusCode)
	}
	if sleptFor != time.Second {
		t.Errorf("sleptFor = %v, want 1s (from Retry-After header)", sleptFor)
	}
}

// retryAfterScriptRT is like scriptRT but injects a Retry-After header on first response.
type retryAfterScriptRT struct {
	statuses   []int
	calls      int
	retryAfter string
}

func (s *retryAfterScriptRT) RoundTrip(r *http.Request) (*http.Response, error) {
	i := s.calls
	s.calls++
	h := http.Header{}
	if i == 0 && s.retryAfter != "" {
		h.Set("Retry-After", s.retryAfter)
	}
	return &http.Response{
		StatusCode: s.statuses[i],
		Body:       io.NopCloser(bytes.NewBufferString("")),
		Header:     h,
		Request:    r,
	}, nil
}

func TestRetry_ContextCancellationReturnsCtxErr(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	base := &scriptRT{statuses: []int{429, 429, 200}}
	rt := &retryTransport{
		base:  base,
		cfg:   RetryConfig{MaxAttempts: 3, MinWait: time.Millisecond, RetryableStatus: []int{429}},
		sleep: noSleep,
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://x/", nil)
	_, err := rt.RoundTrip(req)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}
