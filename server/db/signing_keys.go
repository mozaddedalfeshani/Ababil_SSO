package db

import (
	"context"
	"errors"
	"time"

	"ababilx-sso/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SigningKeysRepo struct{ pool *pgxpool.Pool }

func NewSigningKeysRepo(pool *pgxpool.Pool) *SigningKeysRepo { return &SigningKeysRepo{pool: pool} }

func (r *SigningKeysRepo) Create(ctx context.Context, k *models.SigningKey) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO signing_keys (kid, alg, private_key_enc, public_jwk, status, activates_at, retires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, k.KID, k.Alg, k.PrivateKeyEnc, k.PublicJWK, k.Status, k.ActivatesAt, k.RetiresAt)
	return err
}

func (r *SigningKeysRepo) Active(ctx context.Context) (*models.SigningKey, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT kid, alg, private_key_enc, public_jwk, status, activates_at, retires_at, created_at
		FROM signing_keys WHERE status = 'active'
	`)
	return scanSigningKey(row)
}

// Published returns every key that should appear in JWKS: active, and
// any not-yet-retired or recently-retired key, so in-flight tokens
// signed by a key mid-rotation still validate. See
// docs/architecture.md "Key management".
func (r *SigningKeysRepo) Published(ctx context.Context) ([]*models.SigningKey, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT kid, alg, private_key_enc, public_jwk, status, activates_at, retires_at, created_at
		FROM signing_keys WHERE status IN ('active', 'next')
			OR (status = 'retired' AND retires_at > now())
		ORDER BY activates_at
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*models.SigningKey
	for rows.Next() {
		k, err := scanSigningKey(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

func (r *SigningKeysRepo) ByKID(ctx context.Context, kid string) (*models.SigningKey, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT kid, alg, private_key_enc, public_jwk, status, activates_at, retires_at, created_at
		FROM signing_keys WHERE kid = $1
	`, kid)
	return scanSigningKey(row)
}

// Rotate demotes the current active key to retired (published until
// retiresAt) and promotes the given kid to active — done in one
// transaction so JWKS never observes zero active keys.
func (r *SigningKeysRepo) Rotate(ctx context.Context, newActiveKID string, retiresAt time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		UPDATE signing_keys SET status = 'retired', retires_at = $1 WHERE status = 'active'
	`, retiresAt); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE signing_keys SET status = 'active' WHERE kid = $1
	`, newActiveKID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func scanSigningKey(row pgx.Row) (*models.SigningKey, error) {
	var k models.SigningKey
	err := row.Scan(&k.KID, &k.Alg, &k.PrivateKeyEnc, &k.PublicJWK, &k.Status, &k.ActivatesAt, &k.RetiresAt, &k.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}
