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
