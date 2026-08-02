// Package crypto holds the low-level primitives everything else in the
// identity/OAuth layers builds on: random tokens, password hashing,
// HMAC verification, and AES-GCM sealing. No policy decisions here —
// callers decide token lengths, TTLs, and when to use which primitive.
package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// RandomToken returns a URL-safe token with n bytes of entropy, encoded
// for transport (cookies, URLs, API responses).
func RandomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// RandomHex returns n bytes of entropy hex-encoded — used for values
// that end up in user-facing codes (e.g. recovery codes) where
// URL-safe alphabet isn't required and hex is easier to read/type.
func RandomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}
