package oauth

import "testing"

func TestScopeSubsetOf(t *testing.T) {
	granted := []string{"openid", "email", "offline_access"}

	if !scopeSubsetOf([]string{"openid"}, granted) {
		t.Error("expected narrower request to be a valid subset")
	}
	if !scopeSubsetOf(granted, granted) {
		t.Error("expected identical scopes to be a valid subset")
	}
	if scopeSubsetOf([]string{"openid", "profile"}, granted) {
		t.Error("expected a scope outside the original grant to be rejected — this is the refresh-time escalation guard")
	}
}
