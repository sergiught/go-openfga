package config

import "testing"

func TestNormalizeTokenURL(t *testing.T) {
	cases := []struct {
		name, in, want string
		wantErr        bool
	}{
		{"empty", "", "", false},
		{"bare host", "issuer.example", "https://issuer.example/oauth/token", false},
		{"scheme no path", "https://issuer.example", "https://issuer.example/oauth/token", false},
		{"root path", "https://issuer.example/", "https://issuer.example/oauth/token", false},
		{"explicit path kept", "https://issuer.example/custom/token", "https://issuer.example/custom/token", false},
		{"http allowed", "http://localhost:8080", "http://localhost:8080/oauth/token", false},
		{"bad scheme", "ftp://issuer.example", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeTokenURL(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want error for %q", tc.in)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("NormalizeTokenURL(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
