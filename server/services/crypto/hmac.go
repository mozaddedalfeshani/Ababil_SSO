package crypto

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

// HashToken hashes high-entropy tokens (session tokens, email tokens,
// refresh tokens, client secrets, authorization codes) for storage.
// Plain SHA-256, not HMAC: these values already carry 128+ bits of
// server-generated entropy, so a keyed hash adds no real protection —
// it only adds a secret to manage. Deliberately NOT Argon2id either:
// hashing an already-unguessable token with a memory-hard function
// just adds a self-inflicted DoS surface on every token-endpoint
// request. Passwords are different — low-entropy, human-chosen — and
// use Argon2id (see argon2.go).
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// EqualTokenHash compares in constant time.
func EqualTokenHash(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
