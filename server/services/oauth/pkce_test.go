package oauth

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func challengeFor(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func TestVerifyPKCEAcceptsMatchingS256(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	if !VerifyPKCE(challengeFor(verifier), verifier) {
		t.Fatal("expected matching verifier/challenge to verify")
	}
}

func TestVerifyPKCERejectsMismatch(t *testing.T) {
	if VerifyPKCE(challengeFor("verifier-a"), "verifier-b") {
		t.Fatal("expected mismatched verifier to fail")
	}
}

func TestVerifyPKCERejectsEmpty(t *testing.T) {
	if VerifyPKCE("", "") {
		t.Fatal("expected empty challenge/verifier to fail")
	}
	if VerifyPKCE(challengeFor("x"), "") {
		t.Fatal("expected empty verifier to fail")
	}
}

// This is the security-critical assertion from the architecture
// review: `plain` must never be accepted, only `S256`. OAuth 2.1
// treats PKCE as mandatory; this deployment doesn't offer the
// downgrade path RFC 7636 otherwise permits.
func TestValidCodeChallengeMethodRejectsPlain(t *testing.T) {
	if ValidCodeChallengeMethod("plain") {
		t.Fatal("plain code_challenge_method must never be accepted")
	}
	if ValidCodeChallengeMethod("") {
		t.Fatal("empty code_challenge_method must never be accepted")
	}
	if !ValidCodeChallengeMethod("S256") {
		t.Fatal("S256 must be accepted")
	}
}
