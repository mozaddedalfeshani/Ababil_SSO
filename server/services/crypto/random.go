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

// RandomDigits returns an n-digit decimal string (leading zeros
// preserved) using rejection sampling so each digit is uniform.
func RandomDigits(n int) (string, error) {
	if n <= 0 || n > 18 {
		return "", fmt.Errorf("digit length out of range: %d", n)
	}
	out := make([]byte, n)
	for i := 0; i < n; {
		var b [1]byte
		if _, err := rand.Read(b[:]); err != nil {
			return "", fmt.Errorf("read random bytes: %w", err)
		}
		// 250..255 would bias %10 — reject them.
		if b[0] >= 250 {
			continue
		}
		out[i] = '0' + (b[0] % 10)
		i++
	}
	return string(out), nil
}
