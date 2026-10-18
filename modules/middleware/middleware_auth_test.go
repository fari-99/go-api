package middleware

import "testing"

func TestCheckAuthHeaderRejectsMalformedHeader(t *testing.T) {
	base := DefaultConfig()

	cases := map[string]string{
		"empty":           "",
		"scheme only":     "Bearer",
		"empty token":     "Bearer ",
		"wrong scheme":    "Basic dXNlcjpwYXNz",
		"no scheme":       "eyJhbGciOiJIUzI1NiJ9",
		"invalid token":   "Bearer not-a-jwt",
		"lowercase empty": "bearer ",
	}

	for name, header := range cases {
		t.Run(name, func(t *testing.T) {
			claims, next, err := base.checkAuthHeader(header, "access_token")
			if err == nil {
				t.Fatalf("checkAuthHeader(%q) error = nil, want error", header)
			}
			if next {
				t.Errorf("checkAuthHeader(%q) next = true, want false", header)
			}
			if claims == nil {
				t.Errorf("checkAuthHeader(%q) claims = nil, want empty claims", header)
			}
		})
	}
}
