package openfga

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPIToken_SetsBearer(t *testing.T) {
	crt := &captureRT{}
	rt := wrapAuth((&apiTokenSource{token: "sekret"}).transport(), crt)
	req, _ := http.NewRequest(http.MethodGet, "https://x/", nil)
	if _, err := rt.RoundTrip(req); err != nil {
		t.Fatal(err)
	}
	if got := crt.last.Header.Get("Authorization"); got != "Bearer sekret" {
		t.Errorf("authorization = %q", got)
	}
}

func TestWithAPIToken_WiresAuthTransport(t *testing.T) {
	c, _ := NewClient("https://api.fga.example", WithAPIToken("tok"))
	if c.authTransport == nil {
		t.Fatal("auth transport not set")
	}
}

func TestWithClientCredentials_SendsBearerToken(t *testing.T) {
	const issuedToken = "test-access-token-xyz"

	// Token endpoint: issues an access token.
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": issuedToken,
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	defer tokenSrv.Close()

	// API endpoint: not used directly; we capture via captureRT instead.
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer apiSrv.Close()

	crt := &captureRT{}
	rt := wrapAuth(
		func() http.RoundTripper {
			cfg := ClientCredentialsConfig{
				TokenURL:     tokenSrv.URL,
				ClientID:     "client-id",
				ClientSecret: "client-secret",
			}
			c := &Client{}
			WithClientCredentials(cfg)(c)
			return c.authTransport
		}(),
		crt,
	)

	req, _ := http.NewRequest(http.MethodGet, apiSrv.URL+"/", nil)
	if _, err := rt.RoundTrip(req); err != nil {
		t.Fatal(err)
	}
	if crt.last == nil {
		t.Fatal("no request reached base transport")
	}
	auth := crt.last.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		t.Fatalf("Authorization = %q, want Bearer token", auth)
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	if token != issuedToken {
		t.Errorf("token = %q, want %q", token, issuedToken)
	}
}
