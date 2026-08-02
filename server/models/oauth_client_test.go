package models

import "testing"

func testClient() *OAuthClient {
	return &OAuthClient{
		RedirectURIs:  []string{"https://app.example.com/callback"},
		AllowedScopes: []string{"openid", "email"},
		GrantTypes:    []string{"authorization_code", "refresh_token"},
	}
}

// AllowsRedirectURI must be exact-match — no prefix, substring, or
// query-string-appended variant may pass. This is the redirect_uri
// validation the architecture review specifically called out
// (substring, extra query, scheme downgrade, fragment).
func TestAllowsRedirectURIExactMatchOnly(t *testing.T) {
	c := testClient()

	cases := map[string]bool{
		"https://app.example.com/callback":         true,
		"https://app.example.com/callback/":        false, // trailing slash
		"https://app.example.com/callback?extra=1": false, // appended query
		"https://app.example.com/callback#frag":    false, // fragment
		"http://app.example.com/callback":          false, // scheme downgrade
		"https://evil.com/callback":                false,
		"https://app.example.com/callbackXYZ":      false, // substring
	}
	for uri, want := range cases {
		if got := c.AllowsRedirectURI(uri); got != want {
			t.Errorf("AllowsRedirectURI(%q) = %v, want %v", uri, got, want)
		}
	}
}

func TestAllowsScopesRejectsEscalation(t *testing.T) {
	c := testClient()

	if !c.AllowsScopes([]string{"openid"}) {
		t.Error("expected allowed scope to pass")
	}
	if c.AllowsScopes([]string{"openid", "admin"}) {
		t.Error("expected a scope outside allowed_scopes to be rejected")
	}
}

func TestSupportsGrant(t *testing.T) {
	c := testClient()
	if !c.SupportsGrant("authorization_code") {
		t.Error("expected authorization_code to be supported")
	}
	if c.SupportsGrant("client_credentials") {
		t.Error("expected unregistered grant type to be rejected")
	}
}
