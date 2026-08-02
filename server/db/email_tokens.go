package db

import (
	"context"
	"errors"
	"time"

	"ababilx-sso/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmailTokensRepo struct{ pool *pgxpool.Pool }

func NewEmailTokensRepo(pool *pgxpool.Pool) *EmailTokensRepo { return &EmailTokensRepo{pool: pool} }

func (r *EmailTokensRepo) Create(ctx context.Context, userID string, purpose models.EmailTokenPurpose, tokenHash string, expiresAt time.Time) (*models.EmailToken, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO email_tokens (user_id, purpose, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, purpose, token_hash, expires_at, consumed_at, created_at
	`, userID, purpose, tokenHash, expiresAt)
	return scanEmailToken(row)
}

func (r *EmailTokensRepo) ByTokenHash(ctx context.Context, tokenHash string) (*models.EmailToken, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, purpose, token_hash, expires_at, consumed_at, created_at
		FROM email_tokens WHERE token_hash = $1
	`, tokenHash)
	return scanEmailToken(row)
}

// Consume atomically marks the token used and reports whether this
// call was the one that consumed it — the same single-statement
// pattern as authorization codes, so a resend/replay race can't
// double-spend a token.
func (r *EmailTokensRepo) Consume(ctx context.Context, id string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE email_tokens SET consumed_at = now()
		WHERE id = $1 AND consumed_at IS NULL AND expires_at > now()
	`, id)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// InvalidateAllForUser is called whenever a purpose's token is
// superseded (e.g. a new reset request, or a password change) so old
// links stop working instead of remaining valid until their TTL.
func (r *EmailTokensRepo) InvalidateAllForUser(ctx context.Context, userID string, purpose models.EmailTokenPurpose) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE email_tokens SET consumed_at = now()
		WHERE user_id = $1 AND purpose = $2 AND consumed_at IS NULL
	`, userID, purpose)
	return err
}

func scanEmailToken(row pgx.Row) (*models.EmailToken, error) {
	var t models.EmailToken
	err := row.Scan(&t.ID, &t.UserID, &t.Purpose, &t.TokenHash, &t.ExpiresAt, &t.ConsumedAt, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}
