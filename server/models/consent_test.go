package models

import (
	"testing"
	"time"
)

func TestConsentSatisfies(t *testing.T) {
	c := &Consent{Scopes: []string{"openid", "email"}}

	if !c.Satisfies([]string{"openid"}) {
		t.Error("expected narrower request to be satisfied by broader consent")
	}
	if c.Satisfies([]string{"openid", "offline_access"}) {
		t.Error("expected a scope outside the consent to require re-prompting")
	}
}

func TestConsentRevokedNeverSatisfies(t *testing.T) {
	now := time.Now()
	c := &Consent{Scopes: []string{"openid"}, RevokedAt: &now}

	if c.Satisfies([]string{"openid"}) {
		t.Error("a revoked consent must never satisfy any request")
	}
}
