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
	crt := &captureRT{}
	h := http.Header{}
	h.Set("X-Static", "yes")
	rt := &headerTransport{base: crt, header: h}
	req, _ := http.NewRequest(http.MethodGet, "https://x/", nil)
	if _, err := rt.RoundTrip(req); err != nil {
		t.Fatal(err)
	}
	if crt.last.Header.Get("X-Static") != "yes" {
		t.Error("static header not applied")
	}
}

func TestHeaderTransport_SendsAllValuesForMultiValueHeader(t *testing.T) {
	crt := &captureRT{}
	h := http.Header{}
	h.Add("X-Multi", "one")
	h.Add("X-Multi", "two")
	rt := &headerTransport{base: crt, header: h}
	req, _ := http.NewRequest(http.MethodGet, "https://x/", nil)
	if _, err := rt.RoundTrip(req); err != nil {
		t.Fatal(err)
	}
	if got := crt.last.Header["X-Multi"]; len(got) != 2 || got[0] != "one" || got[1] != "two" {
		t.Errorf("X-Multi = %v, want [one two]", got)
	}
}

func TestHeaderTransport_PerRequestHeaderTakesPrecedence(t *testing.T) {
	crt := &captureRT{}
	static := http.Header{}
	static.Set("X-Foo", "static")
	static.Set("X-Only-Static", "from-static")
	rt := &headerTransport{base: crt, header: static}

	req, _ := http.NewRequest(http.MethodGet, "https://x/", nil)
	req.Header.Set("X-Foo", "req") // per-request value should win

	if _, err := rt.RoundTrip(req); err != nil {
		t.Fatal(err)
	}

	// Per-request header wins; static must not overwrite or duplicate it.
	vals := crt.last.Header["X-Foo"]
	if len(vals) != 1 || vals[0] != "req" {
		t.Errorf("X-Foo = %v, want [req]", vals)
	}

	// Static header absent on request IS applied.
	if crt.last.Header.Get("X-Only-Static") != "from-static" {
		t.Errorf("X-Only-Static = %q, want from-static", crt.last.Header.Get("X-Only-Static"))
	}
}
