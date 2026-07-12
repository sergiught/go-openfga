package openfga

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"
)

// noWait is the test no-op wait: returns immediately without sleeping.
func noWait(_ context.Context, _ time.Duration) error { return nil }

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

func TestRetry_RetriesOn429ThenSucceeds(t *testing.T) {
	rt := &retryTransport{
		base: &scriptRT{statuses: []int{429, 429, 200}},
		cfg:  RetryConfig{MaxAttempts: 3, MinWait: time.Millisecond, MaxWait: time.Millisecond, RetryableStatus: []int{429}},
		wait: noWait,
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
	rt := &retryTransport{base: base, cfg: RetryConfig{MaxAttempts: 3, RetryableStatus: []int{429}}, wait: noWait}
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
	rt := &retryTransport{base: base, cfg: RetryConfig{MaxAttempts: 3, MinWait: time.Millisecond, RetryableStatus: []int{429}}, wait: noWait}
	req, _ := http.NewRequest(http.MethodGet, "https://x/", nil)
	_, _ = rt.RoundTrip(req)
	if base.calls != 3 {
		t.Errorf("calls = %d (want 3 = MaxAttempts)", base.calls)
	}
}

func TestRetry_HonorsRetryAfterHeader(t *testing.T) {
	// Build a custom base that adds Retry-After on the first response.
	var waitedFor time.Duration
	recordWait := func(_ context.Context, d time.Duration) error {
		waitedFor = d
		return nil
	}

	innerBase := &retryAfterScriptRT{statuses: []int{429, 200}, retryAfter: "1"}
	rt := &retryTransport{
		base: innerBase,
		cfg:  RetryConfig{MaxAttempts: 2, MinWait: time.Millisecond, RetryableStatus: []int{429}, HonorRetryAfter: true},
		wait: recordWait,
	}
	req, _ := http.NewRequest(http.MethodGet, "https://x/", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("status = %d", resp.StatusCode)
	}
	if waitedFor != time.Second {
		t.Errorf("waitedFor = %v, want 1s (from Retry-After header)", waitedFor)
	}
}

func TestRetry_ClampsRetryAfterToMaxWait(t *testing.T) {
	var waitedFor time.Duration
	recordWait := func(_ context.Context, d time.Duration) error {
		waitedFor = d
		return nil
	}
	// Server asks for 3600s; MaxWait caps the client at 2s.
	innerBase := &retryAfterScriptRT{statuses: []int{429, 200}, retryAfter: "3600"}
	rt := &retryTransport{
		base: innerBase,
		cfg:  RetryConfig{MaxAttempts: 2, MinWait: time.Millisecond, MaxWait: 2 * time.Second, RetryableStatus: []int{429}, HonorRetryAfter: true},
		wait: recordWait,
	}
	req, _ := http.NewRequest(http.MethodGet, "https://x/", nil)
	if _, err := rt.RoundTrip(req); err != nil {
		t.Fatal(err)
	}
	if waitedFor != 2*time.Second {
		t.Errorf("waitedFor = %v, want 2s (Retry-After clamped to MaxWait)", waitedFor)
	}
}

func TestRetry_SingleAttemptSkipsBodyBuffering(t *testing.T) {
	base := &scriptRT{statuses: []int{200}}
	rt := &retryTransport{base: base, cfg: RetryConfig{MaxAttempts: 1, RetryableStatus: []int{429}}, wait: noWait}
	// GetBody returning an error would fail if the transport tried to buffer/replay.
	req, _ := http.NewRequest(http.MethodPost, "https://x/", bytes.NewBufferString(`{}`))
	req.GetBody = func() (io.ReadCloser, error) { return nil, errors.New("must not be called") }
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 || base.calls != 1 {
		t.Errorf("status=%d calls=%d", resp.StatusCode, base.calls)
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
	// Use the real defaultWait so cancellation is honoured during the wait itself.
	rt := &retryTransport{
		base: base,
		cfg:  RetryConfig{MaxAttempts: 3, MinWait: time.Millisecond, RetryableStatus: []int{429}},
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://x/", nil)
	_, err := rt.RoundTrip(req)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

// TestRetry_ContextCancelDuringBackoff verifies that a long backoff is interrupted
// when the caller's context is cancelled mid-wait.
func TestRetry_ContextCancelDuringBackoff(t *testing.T) {
	base := &scriptRT{statuses: []int{429, 200}}
	rt := &retryTransport{
		base: base,
		cfg: RetryConfig{
			MaxAttempts:     2,
			MinWait:         30 * time.Second, // would block for 30s without cancellation
			MaxWait:         30 * time.Second,
			RetryableStatus: []int{429},
			HonorRetryAfter: false,
			Jitter:          false,
		},
		// Use the real defaultWait so context cancellation is exercised.
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel the context ~20ms after RoundTrip starts (giving it time to hit the wait).
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://x/", nil)
	_, err := rt.RoundTrip(req)
	elapsed := time.Since(start)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if elapsed > 2*time.Second {
		t.Errorf("RoundTrip took %v, want < 2s (should have been cancelled during backoff)", elapsed)
	}
}
