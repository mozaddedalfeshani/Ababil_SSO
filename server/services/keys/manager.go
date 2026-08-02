package keys

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"time"

	"ababilx-sso/db"
	"ababilx-sso/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// signingKeyLockKey is a distinct advisory-lock key from migrations'
// (see db/migrate.go) so first-boot key generation and schema
// migration never contend with each other, while each still
// serializes correctly against concurrent replicas of itself.
const signingKeyLockKey = 727_310_002

type Manager struct {
	pool             *pgxpool.Pool
	repo             *db.SigningKeysRepo
	keyEncryptionKey []byte
}

func NewManager(pool *pgxpool.Pool, repo *db.SigningKeysRepo, kek []byte) *Manager {
	return &Manager{pool: pool, repo: repo, keyEncryptionKey: kek}
}

// EnsureActiveKey generates the first signing key if none exists yet.
// Runs under a session-level advisory lock so two replicas booting
// simultaneously can't both decide they're the one to mint the
// "active" key — the loser blocks, then sees the winner's key via the
// unique partial index on signing_keys and does nothing.
func (m *Manager) EnsureActiveKey(ctx context.Context) error {
	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", signingKeyLockKey); err != nil {
		return fmt.Errorf("acquire signing key lock: %w", err)
	}
	defer conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", signingKeyLockKey)

	if _, err := m.repo.Active(ctx); err == nil {
		return nil // already have one
	} else if !errors.Is(err, db.ErrNotFound) {
		return err
	}

	kid, sealedPrivate, publicJWK, err := generateKeyPair(m.keyEncryptionKey)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	return m.repo.Create(ctx, &models.SigningKey{
		KID:           kid,
		Alg:           Algorithm,
		PrivateKeyEnc: sealedPrivate,
		PublicJWK:     publicJWK,
		Status:        models.SigningKeyActive,
		ActivatesAt:   now,
	})
}

// ActiveSigningKey returns the current active key with its private
// half decrypted, ready for services/token to sign with.
func (m *Manager) ActiveSigningKey(ctx context.Context) (kid string, priv *ecdsa.PrivateKey, err error) {
	k, err := m.repo.Active(ctx)
	if err != nil {
		return "", nil, err
	}
	priv, err = unsealPrivateKey(k.PrivateKeyEnc, k.KID, m.keyEncryptionKey)
	if err != nil {
		return "", nil, err
	}
	return k.KID, priv, nil
}

// Rotate generates a fresh key, promotes it to active, and retires the
// previous one with a grace period long enough that tokens signed
// just before rotation still validate — see docs/architecture.md.
func (m *Manager) Rotate(ctx context.Context, retiredKeyGracePeriod time.Duration) error {
	kid, sealedPrivate, publicJWK, err := generateKeyPair(m.keyEncryptionKey)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	if err := m.repo.Create(ctx, &models.SigningKey{
		KID:           kid,
		Alg:           Algorithm,
		PrivateKeyEnc: sealedPrivate,
		PublicJWK:     publicJWK,
		Status:        models.SigningKeyNext,
		ActivatesAt:   now,
	}); err != nil {
		return err
	}

	return m.repo.Rotate(ctx, kid, now.Add(retiredKeyGracePeriod))
}
