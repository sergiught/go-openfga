package openfga

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// FuzzClassifyResponse feeds arbitrary status codes and bodies through the
// error-classification path that runs on every non-2xx server response.
func FuzzClassifyResponse(f *testing.F) {
	f.Add(200, []byte(`{}`))
	f.Add(400, []byte(`{"code":"validation_error","message":"bad input"}`))
	f.Add(401, []byte(`{"code":"auth_failed"}`))
	f.Add(404, []byte(`not json`))
	f.Add(429, []byte(`{"code":"rate_limited"}`))
	f.Add(500, []byte(``))
	f.Fuzz(func(t *testing.T, status int, body []byte) {
		if status < 100 || status > 599 {
			t.Skip()
		}
		resp := &http.Response{
			StatusCode: status,
			Header:     http.Header{},
			Body:       io.NopCloser(bytes.NewReader(body)),
			Request:    httptest.NewRequest(http.MethodGet, "http://example/", nil),
		}
		err := classifyResponse(resp) // must never panic
		if status >= 200 && status <= 299 && err != nil {
			t.Errorf("2xx classified as error: %v", err)
		}
		if (status < 200 || status > 299) && err == nil {
			t.Errorf("non-2xx %d classified as success", status)
		}
	})
}

// FuzzFGAObjectRelationCodec fuzzes the string<->struct JSON codec used by
// ListUsers targets. It asserts the codec never panics and reaches a fixed
// point: the first round-trip may normalize a value (e.g. the degenerate
// "type:" with an empty id collapses to "type", which encodes identically),
// but every round-trip thereafter must leave it unchanged.
func FuzzFGAObjectRelationCodec(f *testing.F) {
	f.Add([]byte(`"document:budget"`))
	f.Add([]byte(`{"type":"document","id":"budget"}`))
	f.Add([]byte(`{"type":"group","id":"eng","relation":"member"}`))
	f.Add([]byte(`{"type":"user"}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		var o FGAObjectRelation
		if err := json.Unmarshal(data, &o); err != nil {
			return // invalid input is fine
		}
		roundTrip := func(in FGAObjectRelation) FGAObjectRelation {
			b, err := json.Marshal(in)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var out FGAObjectRelation
			if err := json.Unmarshal(b, &out); err != nil {
				t.Fatalf("re-unmarshal of %q: %v", b, err)
			}
			return out
		}
		normalized := roundTrip(o)
		if again := roundTrip(normalized); !reflect.DeepEqual(normalized, again) {
			t.Errorf("codec not idempotent: %+v -> %+v -> %+v", o, normalized, again)
		}
	})
}

// FuzzStreamedEnvelopeDecode fuzzes the NDJSON envelope decoded per line of a
// StreamedListObjects response.
func FuzzStreamedEnvelopeDecode(f *testing.F) {
	f.Add([]byte(`{"result":{"object":"document:budget"}}`))
	f.Add([]byte(`{"result":{}}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		var env streamedEnvelope
		_ = json.Unmarshal(data, &env) // must never panic
	})
}

// FuzzParseRetryAfter fuzzes Retry-After header parsing (seconds or HTTP-date).
func FuzzParseRetryAfter(f *testing.F) {
	f.Add("120")
	f.Add("Wed, 21 Oct 2015 07:28:00 GMT")
	f.Add("-5")
	f.Add("not a date")
	f.Fuzz(func(t *testing.T, v string) {
		r := &http.Response{Header: http.Header{}}
		r.Header.Set("Retry-After", v)
		if d := parseRetryAfter(r); d < 0 {
			t.Errorf("negative duration %v for %q", d, v)
		}
	})
}
