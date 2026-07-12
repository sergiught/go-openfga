package config

import "testing"

func TestLoad_DecodesAllFields(t *testing.T) {
	t.Setenv("FGA_API_URL", "https://api.fga.example")
	t.Setenv("FGA_STORE_ID", "store-1")
	t.Setenv("FGA_MODEL_ID", "model-1")
	t.Setenv("FGA_CLIENT_ID", "cid")
	t.Setenv("FGA_CLIENT_SECRET", "csecret")
	t.Setenv("FGA_API_TOKEN_ISSUER", "https://issuer.example")
	t.Setenv("FGA_API_AUDIENCE", "https://api.fga.example")
	t.Setenv("FGA_API_SCOPES", "read,write")

	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.APIURL != "https://api.fga.example" || c.StoreID != "store-1" || c.AuthModelID != "model-1" {
		t.Errorf("core fields = %+v", c)
	}
	if c.ClientID != "cid" || c.ClientSecret != "csecret" || c.TokenIssuer != "https://issuer.example" {
		t.Errorf("cred fields = %+v", c)
	}
	if len(c.Scopes) != 2 || c.Scopes[0] != "read" || c.Scopes[1] != "write" {
		t.Errorf("scopes = %v", c.Scopes)
	}
}

func TestLoad_ConflictingAuthMethodsError(t *testing.T) {
	t.Setenv("FGA_API_TOKEN", "tok")
	t.Setenv("FGA_CLIENT_ID", "cid")
	if _, err := Load(); err == nil {
		t.Fatal("expected error when both api_token and client_credentials env vars are set")
	}
}

func TestLoad_APITokenAloneOK(t *testing.T) {
	t.Setenv("FGA_API_TOKEN", "tok")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.APIToken != "tok" || c.HasClientCredentials() {
		t.Errorf("unexpected config %+v", c)
	}
}
