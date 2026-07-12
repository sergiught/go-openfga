package openfga

import "testing"

func TestNewClient_ValidationErrors(t *testing.T) {
	cases := []struct {
		name string
		url  string
		opts []Option
	}{
		{"missing url", "", nil},
		{"bad scheme", "ftp://api.fga.example", nil},
		{"malformed url", "://nope", nil},
		{"bad store ulid", "https://api.fga.example", []Option{WithStoreID("store1")}},
		{"bad model ulid", "https://api.fga.example", []Option{WithAuthorizationModelID("model1")}},
		{"incomplete client creds", "https://api.fga.example", []Option{WithClientCredentials(ClientCredentialsConfig{ClientID: "c"})}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewClient(tc.url, tc.opts...); err == nil {
				t.Fatalf("expected error for %s", tc.name)
			}
		})
	}
}

func TestNewClient_ValidStoreULIDAccepted(t *testing.T) {
	c, err := NewClient("https://api.fga.example", WithStoreID(testStoreID))
	if err != nil {
		t.Fatal(err)
	}
	if c.baseURL.String() != "https://api.fga.example/" {
		t.Errorf("baseURL = %q", c.baseURL.String())
	}
}
