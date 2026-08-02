package oauth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
)

// VerifyPKCE implements S256 only — `plain` is never accepted. RFC
// 7636 permits `plain` as a fallback for clients that can't compute
// SHA-256, but accepting it here would let an attacker who intercepts
// the authorization request (but not a code_verifier generated
// client-side) trivially satisfy the challenge. Every modern client
// can do SHA-256; there's no compatibility reason to weaken this.
func VerifyPKCE(codeChallenge, codeVerifier string) bool {
	if codeChallenge == "" || codeVerifier == "" {
		return false
	}
	sum := sha256.Sum256([]byte(codeVerifier))
	computed := base64.RawURLEncoding.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(computed), []byte(codeChallenge)) == 1
}

// ValidCodeChallengeMethod reports whether the client asked for a
// challenge method we accept — only "S256". Absence of the parameter,
// or "plain", is rejected: PKCE is mandatory in this deployment, not
// optional per OAuth 2.1.
func ValidCodeChallengeMethod(method string) bool {
	return method == "S256"
}
