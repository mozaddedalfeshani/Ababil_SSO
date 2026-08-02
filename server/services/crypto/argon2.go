package crypto

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2Params are deliberately conservative but config-tunable — see
// docs/architecture.md for why 64MB/t=3/p=2 was chosen and why it
// requires the admission-control semaphore in semaphore.go.
type Argon2Params struct {
	MemoryKiB   uint32
	Iterations  uint32
	Parallelism uint8
	SaltLen     uint32
	KeyLen      uint32
}

func DefaultArgon2Params() Argon2Params {
	return Argon2Params{
		MemoryKiB:   64 * 1024,
		Iterations:  3,
		Parallelism: 2,
		SaltLen:     16,
		KeyLen:      32,
	}
}

type PasswordHasher struct {
	params Argon2Params
	sem    *HashSemaphore
}

func NewPasswordHasher(params Argon2Params, sem *HashSemaphore) *PasswordHasher {
	return &PasswordHasher{params: params, sem: sem}
}

// Hash produces a self-describing encoded hash (params + salt + hash),
// gated by the admission-control semaphore.
func (h *PasswordHasher) Hash(ctx context.Context, password string) (string, error) {
	if err := h.sem.Acquire(ctx); err != nil {
		return "", err
	}
	defer h.sem.Release()

	salt := make([]byte, h.params.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("read salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, h.params.Iterations, h.params.MemoryKiB, h.params.Parallelism, h.params.KeyLen)

	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		h.params.MemoryKiB, h.params.Iterations, h.params.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
	return encoded, nil
}

// Verify checks a password against an encoded hash, gated by the same
// semaphore as Hash — verification is exactly as expensive as hashing.
func (h *PasswordHasher) Verify(ctx context.Context, password, encoded string) (bool, error) {
	if err := h.sem.Acquire(ctx); err != nil {
		return false, err
	}
	defer h.sem.Release()

	params, salt, hash, err := decodeArgon2Hash(encoded)
	if err != nil {
		return false, err
	}

	computed := argon2.IDKey([]byte(password), salt, params.Iterations, params.MemoryKiB, params.Parallelism, uint32(len(hash)))
	return subtle.ConstantTimeCompare(computed, hash) == 1, nil
}

func decodeArgon2Hash(encoded string) (Argon2Params, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return Argon2Params{}, nil, nil, fmt.Errorf("invalid argon2id hash format")
	}

	var mem, iter uint32
	var par uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &mem, &iter, &par); err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("parse argon2id params: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("decode salt: %w", err)
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return Argon2Params{}, nil, nil, fmt.Errorf("decode hash: %w", err)
	}

	return Argon2Params{MemoryKiB: mem, Iterations: iter, Parallelism: par}, salt, hash, nil
}

// DummyHash is a fixed, valid-format encoded hash with no known
// password. Login handlers run Verify against this for unknown emails
// so the response timing/shape is identical to a real wrong-password
// case — the standard defense against user-enumeration via login.
const DummyHash = "$argon2id$v=19$m=65536,t=3,p=2$c29tZXNhbHRzb21lc2FsdA$aW52YWxpZGludmFsaWRpbnZhbGlkaW52YWxpZA"
