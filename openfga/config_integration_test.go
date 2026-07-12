package openfga

import "testing"

func TestNewClient_SeedsFromEnv(t *testing.T) {
	t.Setenv("FGA_API_URL", "https://env.fga.example")
	t.Setenv("FGA_STORE_ID", testStoreID)
	t.Setenv("FGA_API_TOKEN", "env-token")

	c, err := NewClient("")
	if err != nil {
		t.Fatal(err)
	}
	if c.baseURL.String() != "https://env.fga.example/" {
		t.Errorf("baseURL = %q", c.baseURL.String())
	}
	if c.storeID != testStoreID {
		t.Errorf("storeID = %q", c.storeID)
	}
	if c.auth == nil {
		t.Fatal("expected auth seeded from FGA_API_TOKEN")
	}
}

func TestNewClient_ArgOverridesEnvURL(t *testing.T) {
	t.Setenv("FGA_API_URL", "https://env.fga.example")
	c, err := NewClient("https://arg.fga.example")
	if err != nil {
		t.Fatal(err)
	}
	if c.baseURL.String() != "https://arg.fga.example/" {
		t.Errorf("baseURL = %q, want arg to win", c.baseURL.String())
	}
}

func TestNewClient_OptionOverridesEnvStore(t *testing.T) {
	t.Setenv("FGA_API_URL", "https://env.fga.example")
	t.Setenv("FGA_STORE_ID", testStoreID)
	c, err := NewClient("", WithStoreID(testModelID))
	if err != nil {
		t.Fatal(err)
	}
	if c.storeID != testModelID {
		t.Errorf("storeID = %q, want option to win", c.storeID)
	}
}

func TestNewClient_EnvClientCredentialsNormalized(t *testing.T) {
	t.Setenv("FGA_API_URL", "https://env.fga.example")
	t.Setenv("FGA_CLIENT_ID", "cid")
	t.Setenv("FGA_CLIENT_SECRET", "csecret")
	t.Setenv("FGA_API_TOKEN_ISSUER", "issuer.example")

	c, err := NewClient("")
	if err != nil {
		t.Fatal(err)
	}
	spec, ok := c.auth.(*clientCredentialsSpec)
	if !ok {
		t.Fatalf("auth = %T, want *clientCredentialsSpec", c.auth)
	}
	if spec.tokenURL != "https://issuer.example/oauth/token" {
		t.Errorf("tokenURL = %q, want normalized", spec.tokenURL)
	}
}

func TestNewClient_ConflictingEnvAuthErrors(t *testing.T) {
	t.Setenv("FGA_API_URL", "https://env.fga.example")
	t.Setenv("FGA_API_TOKEN", "tok")
	t.Setenv("FGA_CLIENT_ID", "cid")
	if _, err := NewClient(""); err == nil {
		t.Fatal("expected error for conflicting env auth methods")
	}
}
