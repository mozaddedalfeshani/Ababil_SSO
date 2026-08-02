package db

import (
	"context"
	"errors"
	"time"

	"ababilx-sso/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshTokensRepo struct{ pool *pgxpool.Pool }

func NewRefreshTokensRepo(pool *pgxpool.Pool) *RefreshTokensRepo {
	return &RefreshTokensRepo{pool: pool}
}

type CreateRefreshTokenParams struct {
	TokenHash    string
	FamilyID     string
	ClientID     string
	UserID       string
	SessionID    *string
	Scope        string
	SessionBound bool
	ExpiresAt    time.Time
}

func (r *RefreshTokensRepo) Create(ctx context.Context, p CreateRefreshTokenParams) (*models.RefreshToken, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO refresh_tokens (token_hash, family_id, client_id, user_id, session_id, scope, session_bound, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, token_hash, family_id, client_id, user_id, session_id, scope, session_bound,
			expires_at, consumed_at, rotated_to, revoked_at, revoke_reason, created_at
	`, p.TokenHash, p.FamilyID, p.ClientID, p.UserID, p.SessionID, p.Scope, p.SessionBound, p.ExpiresAt)
	return scanRefreshToken(row)
}

func (r *RefreshTokensRepo) ByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, token_hash, family_id, client_id, user_id, session_id, scope, session_bound,
			expires_at, consumed_at, rotated_to, revoked_at, revoke_reason, created_at
		FROM refresh_tokens WHERE token_hash = $1
	`, tokenHash)
	return scanRefreshToken(row)
}

// Rotate atomically consumes the presented token and links it to its
// replacement, returning false if it was already consumed/revoked —
// the caller (services/oauth) treats false as reuse and revokes the
// whole family.
func (r *RefreshTokensRepo) Rotate(ctx context.Context, oldID, newID string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE refresh_tokens SET consumed_at = now(), rotated_to = $2
		WHERE id = $1 AND consumed_at IS NULL AND revoked_at IS NULL
	`, oldID, newID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// RevokeFamily kills every token in a family — called on detected
// reuse (treat the whole chain as compromised) and on explicit
// "revoke this app's access" from the account UI.
func (r *RefreshTokensRepo) RevokeFamily(ctx context.Context, familyID, reason string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = now(), revoke_reason = $2
		WHERE family_id = $1 AND revoked_at IS NULL
	`, familyID, reason)
	return err
}

// RevokeSessionBound is the logout-time cleanup: tokens issued without
// offline_access die with the session; offline_access tokens
// (session_bound = false) are untouched by design.
func (r *RefreshTokensRepo) RevokeSessionBound(ctx context.Context, sessionID, reason string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = now(), revoke_reason = $2
		WHERE session_id = $1 AND session_bound AND revoked_at IS NULL
	`, sessionID, reason)
	return err
}

func (r *RefreshTokensRepo) RevokeAllForUserClient(ctx context.Context, userID, clientID, reason string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = now(), revoke_reason = $3
		WHERE user_id = $1 AND client_id = $2 AND revoked_at IS NULL
	`, userID, clientID, reason)
	return err
}

func scanRefreshToken(row pgx.Row) (*models.RefreshToken, error) {
	var t models.RefreshToken
	err := row.Scan(&t.ID, &t.TokenHash, &t.FamilyID, &t.ClientID, &t.UserID, &t.SessionID, &t.Scope,
		&t.SessionBound, &t.ExpiresAt, &t.ConsumedAt, &t.RotatedTo, &t.RevokedAt, &t.RevokeReason, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}
