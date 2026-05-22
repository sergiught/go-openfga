package openfga

import (
	"net/http"
	"testing"
)

type captureRT struct{ last *http.Request }

func (c *captureRT) RoundTrip(r *http.Request) (*http.Response, error) {
	c.last = r
	return &http.Response{StatusCode: 200, Body: http.NoBody, Header: http.Header{}, Request: r}, nil
}

func TestHeaderTransport_AddsHeaders(t *testing.T) {
	cap := &captureRT{}
	h := http.Header{}
	h.Set("X-Static", "yes")
	rt := &headerTransport{base: cap, header: h}
	req, _ := http.NewRequest(http.MethodGet, "https://x/", nil)
	if _, err := rt.RoundTrip(req); err != nil {
		t.Fatal(err)
	}
	if cap.last.Header.Get("X-Static") != "yes" {
		t.Error("static header not applied")
	}
}

func TestHeaderTransport_PerRequestHeaderTakesPrecedence(t *testing.T) {
	cap := &captureRT{}
	static := http.Header{}
	static.Set("X-Foo", "static")
	static.Set("X-Only-Static", "from-static")
	rt := &headerTransport{base: cap, header: static}

	req, _ := http.NewRequest(http.MethodGet, "https://x/", nil)
	req.Header.Set("X-Foo", "req") // per-request value should win

	if _, err := rt.RoundTrip(req); err != nil {
		t.Fatal(err)
	}

	// Per-request header wins; static must not overwrite or duplicate it.
	vals := cap.last.Header["X-Foo"]
	if len(vals) != 1 || vals[0] != "req" {
		t.Errorf("X-Foo = %v, want [req]", vals)
	}

	// Static header absent on request IS applied.
	if cap.last.Header.Get("X-Only-Static") != "from-static" {
		t.Errorf("X-Only-Static = %q, want from-static", cap.last.Header.Get("X-Only-Static"))
	}
}
