package openfga

import (
	"bytes"
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
	base  http.RoundTripper
	cfg   RetryConfig
	sleep func(time.Duration) // injectable for tests; nil => time.Sleep
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	sleep := t.sleep
	if sleep == nil {
		sleep = time.Sleep
	}
	// Buffer body for replay across attempts.
	var bodyBytes []byte
	if req.Body != nil && req.Body != http.NoBody {
		var readErr error
		bodyBytes, readErr = io.ReadAll(req.Body)
		_ = req.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
	}

	max := t.cfg.MaxAttempts
	if max < 1 {
		max = 1
	}
	var resp *http.Response
	var err error
	for attempt := 0; attempt < max; attempt++ {
		r2 := req.Clone(req.Context())
		if bodyBytes != nil {
			r2.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			r2.ContentLength = int64(len(bodyBytes))
		}
		resp, err = t.base.RoundTrip(r2)
		if err != nil {
			return resp, err
		}
		if attempt == max-1 || !t.retryable(resp.StatusCode) {
			return resp, nil
		}
		wait := t.backoff(attempt, resp)
		// Drain and close the soon-to-be-discarded body.
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		default:
		}
		sleep(wait)
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
		d = time.Duration(rand.Int63n(int64(d)))
	}
	return d
}
