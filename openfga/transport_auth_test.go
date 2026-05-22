package openfga

import (
	"net/http"
	"testing"
)

func TestAPIToken_SetsBearer(t *testing.T) {
	cap := &captureRT{}
	rt := wrapAuth((&apiTokenSource{token: "sekret"}).transport(), cap)
	req, _ := http.NewRequest(http.MethodGet, "https://x/", nil)
	if _, err := rt.RoundTrip(req); err != nil {
		t.Fatal(err)
	}
	if got := cap.last.Header.Get("Authorization"); got != "Bearer sekret" {
		t.Errorf("authorization = %q", got)
	}
}

func TestWithAPIToken_WiresAuthTransport(t *testing.T) {
	c, _ := NewClient("https://api.fga.example", WithAPIToken("tok"))
	if c.authTransport == nil {
		t.Fatal("auth transport not set")
	}
}
