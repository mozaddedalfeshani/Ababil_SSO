package db

import (
	"context"
	"errors"
	"time"

	"ababilx-sso/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthorizationCodesRepo struct{ pool *pgxpool.Pool }

func NewAuthorizationCodesRepo(pool *pgxpool.Pool) *AuthorizationCodesRepo {
	return &AuthorizationCodesRepo{pool: pool}
}

type CreateCodeParams struct {
	CodeHash      string
	ClientID      string
	UserID        string
	SessionID     string
	RedirectURI   string
	Scope         string
	Nonce         *string
	CodeChallenge string
	AuthTime      time.Time
	ExpiresAt     time.Time
}

func (r *AuthorizationCodesRepo) Create(ctx context.Context, p CreateCodeParams) (*models.AuthorizationCode, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO authorization_codes (code_hash, client_id, user_id, session_id, redirect_uri,
			scope, nonce, code_challenge, auth_time, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, code_hash, client_id, user_id, session_id, redirect_uri, scope, nonce,
			code_challenge, auth_time, refresh_family_id, expires_at, consumed_at, created_at
	`, p.CodeHash, p.ClientID, p.UserID, p.SessionID, p.RedirectURI, p.Scope, p.Nonce,
		p.CodeChallenge, p.AuthTime, p.ExpiresAt)
	return scanCode(row)
}

func (r *AuthorizationCodesRepo) ByHash(ctx context.Context, codeHash string) (*models.AuthorizationCode, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, code_hash, client_id, user_id, session_id, redirect_uri, scope, nonce,
			code_challenge, auth_time, refresh_family_id, expires_at, consumed_at, created_at
		FROM authorization_codes WHERE code_hash = $1
	`, codeHash)
	return scanCode(row)
}

// Consume is the exactly-once redemption at the heart of replay
// detection: the row survives after consumption (see
// docs/architecture.md), so a second attempt at the same code finds
// consumed_at already set and the caller treats that as evidence of a
// leaked/replayed code rather than a client bug.
func (r *AuthorizationCodesRepo) Consume(ctx context.Context, id string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE authorization_codes SET consumed_at = now()
		WHERE id = $1 AND consumed_at IS NULL AND expires_at > now()
	`, id)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func scanCode(row pgx.Row) (*models.AuthorizationCode, error) {
	var c models.AuthorizationCode
	err := row.Scan(&c.ID, &c.CodeHash, &c.ClientID, &c.UserID, &c.SessionID, &c.RedirectURI,
		&c.Scope, &c.Nonce, &c.CodeChallenge, &c.AuthTime, &c.RefreshFamilyID, &c.ExpiresAt,
		&c.ConsumedAt, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}
