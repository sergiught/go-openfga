package openfga

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// RequestOption customizes a single request.
type RequestOption func(*requestConfig)

type requestConfig struct {
	header      http.Header
	storeID     string
	authModelID string
	consistency ConsistencyPreference

	// Bulk/parallel knobs; zero means "use the method default".
	maxParallel       int
	maxPerChunk       int
	maxChecksPerBatch int
	transaction       bool
	onDuplicate       OnDuplicate
	onMissing         OnMissing
}

func newRequestConfig() *requestConfig { return &requestConfig{header: http.Header{}} }

func applyOptions(rc *requestConfig, opts []RequestOption) {
	for _, o := range opts {
		o(rc)
	}
}

// WithRequestHeader sets a header on a single request.
func WithRequestHeader(key, value string) RequestOption {
	return func(rc *requestConfig) { rc.header.Set(key, value) }
}

// WithConsistency overrides the read consistency for one query call.
func WithConsistency(c ConsistencyPreference) RequestOption {
	return func(rc *requestConfig) { rc.consistency = c }
}

// WithStore overrides the store ID for one call.
func WithStore(storeID string) RequestOption {
	return func(rc *requestConfig) { rc.storeID = storeID }
}

// WithAuthorizationModel overrides the authorization model ID for one call.
func WithAuthorizationModel(id string) RequestOption {
	return func(rc *requestConfig) { rc.authModelID = id }
}

// NewRequest builds an *http.Request against the client base URL. It is public
// so callers can hit arbitrary endpoints while reusing the configured transport.
func (c *Client) NewRequest(ctx context.Context, method, path string, body any, opts ...RequestOption) (*http.Request, error) {
	rc := newRequestConfig()
	applyOptions(rc, opts)
	return c.newRequest(ctx, method, path, body, rc.header)
}

func (c *Client) newRequest(ctx context.Context, method, path string, body any, header http.Header) (*http.Request, error) {
	rel := strings.TrimPrefix(path, "/")
	u, err := c.baseURL.Parse(rel)
	if err != nil {
		return nil, err
	}
	var buf *bytes.Buffer
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(body); err != nil {
			return nil, err
		}
	}
	var reqBody io.ReadCloser = http.NoBody
	if buf != nil {
		reqBody = io.NopCloser(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	for k, vs := range header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	return req, nil
}

// BareDo executes a request through the transport chain, classifies errors,
// and returns the wrapped response without decoding the body.
func (c *Client) BareDo(req *http.Request) (*Response, error) {
	httpResp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	resp := newResponse(httpResp)
	if err := classifyResponse(httpResp); err != nil {
		_ = httpResp.Body.Close()
		return resp, err
	}
	return resp, nil
}

// Do executes a request and decodes a 2xx JSON body into v (which may be nil).
// If v implements continuationTokener, its token is lifted onto Response.
func (c *Client) Do(req *http.Request, v any) (*Response, error) {
	resp, err := c.BareDo(req)
	if err != nil {
		return resp, err
	}
	defer func() { _ = resp.Body.Close() }()
	if v == nil {
		return resp, nil
	}
	if w, ok := v.(io.Writer); ok {
		_, err = io.Copy(w, resp.Body)
		return resp, err
	}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(v); err != nil && err != io.EOF {
		return resp, err
	}
	if ct, ok := v.(continuationTokener); ok {
		resp.ContinuationToken = ct.continuationToken()
	}
	return resp, nil
}

// WithMaxParallel caps the number of concurrent HTTP requests issued by
// Tuples.WriteTuples, Tuples.DeleteTuples, and Relationships.BatchCheckAll.
// Non-positive values fall back to the default (10). Other methods ignore it.
func WithMaxParallel(n int) RequestOption {
	return func(rc *requestConfig) { rc.maxParallel = n }
}

// WithMaxPerChunk sets how many tuples go into each non-transactional request
// issued by Tuples.WriteTuples / Tuples.DeleteTuples. Non-positive values fall
// back to the default (1). Ignored by other methods and when WithTransaction is set.
func WithMaxPerChunk(n int) RequestOption {
	return func(rc *requestConfig) { rc.maxPerChunk = n }
}

// WithMaxChecksPerBatch sets how many checks go into each /batch-check request
// issued by Relationships.BatchCheckAll. Non-positive values fall back to the
// default (50, the server maximum). Other methods ignore it.
func WithMaxChecksPerBatch(n int) RequestOption {
	return func(rc *requestConfig) { rc.maxChecksPerBatch = n }
}

// WithTransaction makes Tuples.WriteTuples / Tuples.DeleteTuples send a single
// transactional /write request instead of chunking. Other methods ignore it.
func WithTransaction() RequestOption {
	return func(rc *requestConfig) { rc.transaction = true }
}

// WithOnDuplicate sets the on_duplicate conflict mode on the write requests
// issued by Tuples.WriteTuples. Other methods ignore it.
func WithOnDuplicate(v OnDuplicate) RequestOption {
	return func(rc *requestConfig) { rc.onDuplicate = v }
}

// WithOnMissing sets the on_missing conflict mode on the delete requests
// issued by Tuples.DeleteTuples. Other methods ignore it.
func WithOnMissing(v OnMissing) RequestOption {
	return func(rc *requestConfig) { rc.onMissing = v }
}
