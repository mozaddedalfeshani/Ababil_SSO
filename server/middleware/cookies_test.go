package middleware

import "testing"

// This test exists because of a real bug caught during manual testing:
// __Host- cookies require the Secure attribute, and browsers/curl
// silently drop them without it — which broke every plain-HTTP local
// dev login until the prefix was gated on Secure as well as Domain.
func TestCookiePolicyHostPrefixRequiresSecure(t *testing.T) {
	cases := []struct {
		name       string
		policy     CookiePolicy
		wantPrefix bool
	}{
		{"secure, no domain -> __Host-", CookiePolicy{Domain: "", Secure: true}, true},
		{"insecure, no domain -> plain (dev over http)", CookiePolicy{Domain: "", Secure: false}, false},
		{"secure, with domain -> plain (subdomain deployment)", CookiePolicy{Domain: "example.com", Secure: true}, false},
		{"insecure, with domain -> plain", CookiePolicy{Domain: "example.com", Secure: false}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			name := tc.policy.SessionCookieName()
			hasPrefix := len(name) >= 7 && name[:7] == "__Host-"
			if hasPrefix != tc.wantPrefix {
				t.Errorf("SessionCookieName() = %q, wantPrefix=%v", name, tc.wantPrefix)
			}
		})
	}
}
