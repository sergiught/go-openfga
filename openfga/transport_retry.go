package openfga

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"syscall"
	"time"
)

// RetryConfig controls automatic retries. Defaults retry only HTTP 429 plus
// transient network failures (connection resets, refused dials, timeouts).
type RetryConfig struct {
	MaxAttempts     int           // total attempts including the first; default 3
	MinWait         time.Duration // base backoff; default 1s
	MaxWait         time.Duration // backoff ceiling; default 30s
	RetryableStatus []int         // default {429}; add 5xx to opt in
	HonorRetryAfter bool          // default true
	Jitter          bool          // default true (equal jitter)
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

// WithRetry tunes retry behavior. Only the fields you set take effect; a
// zero-valued MaxAttempts, MinWait, MaxWait, or RetryableStatus falls back to
// its default (3 attempts, 1s–30s, {429}). This means a partial config such as
//
//	WithRetry(RetryConfig{MaxAttempts: 5})
//
// keeps the default 429 retry set and the MaxWait ceiling instead of silently
// disabling them. Jitter and Retry-After honoring stay enabled; supply your own
// transport via WithBaseTransport if you need to turn them off. Add 5xx to
// RetryableStatus to opt those statuses in (note this retries non-idempotent
// writes on ambiguous 5xx). Transient network errors are always retried.
func WithRetry(cfg RetryConfig) Option {
	return func(c *Client) {
		merged := defaultRetryConfig()
		if cfg.MaxAttempts > 0 {
			merged.MaxAttempts = cfg.MaxAttempts
		}
		if cfg.MinWait > 0 {
			merged.MinWait = cfg.MinWait
		}
		if cfg.MaxWait > 0 {
			merged.MaxWait = cfg.MaxWait
		}
		if len(cfg.RetryableStatus) > 0 {
			merged.RetryableStatus = cfg.RetryableStatus
		}
		c.retry = merged
	}
}

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

// replayableBody returns a function that reconstructs req's body for each retry
// attempt, or nil when there is nothing to replay. It prefers the stdlib's
// GetBody (no extra copy) and falls back to buffering the body once. The
// original req.Body is closed.
func replayableBody(req *http.Request) (func() (io.ReadCloser, error), error) {
	if req.Body == nil || req.Body == http.NoBody {
		return req.GetBody, nil
	}
	if req.GetBody != nil {
		_ = req.Body.Close()
		return req.GetBody, nil
	}
	bodyBytes, err := io.ReadAll(req.Body)
	_ = req.Body.Close()
	if err != nil {
		return nil, err
	}
	return func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(bodyBytes)), nil
	}, nil
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

	getBody, err := replayableBody(req)
	if err != nil {
		return nil, err
	}

	var resp *http.Response
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
		lastAttempt := attempt == maxAttempts-1
		if err != nil {
			// Transient network failures carry no HTTP response, so they are
			// not covered by retryable(status); retry them like a retryable
			// status. A cancelled/expired context is never retried.
			if lastAttempt || !retryableError(err) {
				return resp, err
			}
			if werr := waitFn(req.Context(), t.backoff(attempt, nil)); werr != nil {
				return nil, werr
			}
			continue
		}
		if lastAttempt || !t.retryable(resp.StatusCode) {
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

// retryableError reports whether a transport-level error (one with no HTTP
// response) is worth retrying: transient connection failures such as resets,
// refused dials, unexpected EOFs on a reused keep-alive connection, and network
// timeouts. A cancelled or expired request context is never retried.
func retryableError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNREFUSED) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr)
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
	if t.cfg.HonorRetryAfter && resp != nil {
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
	if t.cfg.Jitter && d > 1 {
		// Equal jitter: wait in [d/2, d) so backoff never collapses to near
		// zero (as full jitter's [0, d) can) yet still spreads retries out.
		half := d / 2
		//nolint:gosec // G404: backoff jitter is not security-sensitive; a weak RNG is fine.
		d = half + time.Duration(rand.Int63n(int64(d-half)))
	}
	return d
}
