package openfga

import (
	"bytes"
	"context"
	"io"
	"math"
	"math/rand"
	"net/http"
	"time"
)

// RetryConfig controls automatic retries. Defaults retry only HTTP 429.
type RetryConfig struct {
	MaxAttempts     int           // total attempts including the first; default 3
	MinWait         time.Duration // base backoff; default 1s
	MaxWait         time.Duration // backoff ceiling; default 30s
	RetryableStatus []int         // default {429}; add 5xx to opt in
	HonorRetryAfter bool          // default true
	Jitter          bool          // default true (full jitter)
}

func defaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:     3,
		MinWait:         time.Second,
		MaxWait:         30 * time.Second,
		RetryableStatus: []int{http.StatusTooManyRequests},
		HonorRetryAfter: true,
		Jitter:          true,
	}
}

// WithRetry overrides retry behavior.
func WithRetry(cfg RetryConfig) Option { return func(c *Client) { c.retry = &cfg } }

// WithoutRetry disables retries entirely.
func WithoutRetry() Option { return func(c *Client) { c.retry = nil } }

type retryTransport struct {
	base http.RoundTripper
	cfg  RetryConfig
	// wait is injectable for tests; nil uses the default cancellable timer/select.
	// It should block for duration d (or until ctx is cancelled) and return ctx.Err() or nil.
	wait func(ctx context.Context, d time.Duration) error
}

// defaultWait is the production wait: blocks for d or until ctx is cancelled.
func defaultWait(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	select {
	case <-ctx.Done():
		timer.Stop()
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	maxAttempts := t.cfg.MaxAttempts
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	// No retries possible: pass through without buffering the body.
	if maxAttempts == 1 {
		return t.base.RoundTrip(req)
	}

	waitFn := t.wait
	if waitFn == nil {
		waitFn = defaultWait
	}

	// getBody reconstructs the request body for each attempt. Prefer the
	// stdlib-provided GetBody (no extra copy); fall back to buffering once when
	// it is absent.
	getBody := req.GetBody
	if req.Body != nil && req.Body != http.NoBody {
		if getBody == nil {
			bodyBytes, readErr := io.ReadAll(req.Body)
			_ = req.Body.Close()
			if readErr != nil {
				return nil, readErr
			}
			getBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(bodyBytes)), nil
			}
		} else {
			// We reconstruct via GetBody; the original body is not consumed.
			_ = req.Body.Close()
		}
	}

	var resp *http.Response
	var err error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		r2 := req.Clone(req.Context())
		if getBody != nil {
			body, berr := getBody()
			if berr != nil {
				return nil, berr
			}
			r2.Body = body
		}
		resp, err = t.base.RoundTrip(r2)
		if err != nil {
			return resp, err
		}
		if attempt == maxAttempts-1 || !t.retryable(resp.StatusCode) {
			return resp, nil
		}
		d := t.backoff(attempt, resp)
		// Drain and close the soon-to-be-discarded body.
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		if werr := waitFn(req.Context(), d); werr != nil {
			return nil, werr
		}
	}
	return resp, err
}

func (t *retryTransport) retryable(status int) bool {
	for _, s := range t.cfg.RetryableStatus {
		if s == status {
			return true
		}
	}
	return false
}

func (t *retryTransport) backoff(attempt int, resp *http.Response) time.Duration {
	if t.cfg.HonorRetryAfter {
		if ra := parseRetryAfter(resp); ra > 0 {
			// Clamp to MaxWait so a hostile or misconfigured server cannot pin
			// the client asleep for an unbounded interval.
			if t.cfg.MaxWait > 0 && ra > t.cfg.MaxWait {
				return t.cfg.MaxWait
			}
			return ra
		}
	}
	base := t.cfg.MinWait
	if base <= 0 {
		base = time.Second
	}
	d := time.Duration(float64(base) * math.Pow(2, float64(attempt)))
	if t.cfg.MaxWait > 0 && d > t.cfg.MaxWait {
		d = t.cfg.MaxWait
	}
	if t.cfg.Jitter && d > 0 {
		//nolint:gosec // G404: backoff jitter is not security-sensitive; a weak RNG is fine.
		d = time.Duration(rand.Int63n(int64(d)))
	}
	return d
}
