package db

import (
	"context"
	"errors"

	"ababilx-sso/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConsentsRepo struct{ pool *pgxpool.Pool }

func NewConsentsRepo(pool *pgxpool.Pool) *ConsentsRepo { return &ConsentsRepo{pool: pool} }

func (r *ConsentsRepo) Active(ctx context.Context, userID, clientID string) (*models.Consent, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, client_id, scopes, granted_at, revoked_at
		FROM consents WHERE user_id = $1 AND client_id = $2 AND revoked_at IS NULL
	`, userID, clientID)
	return scanConsent(row)
}

// Grant replaces any existing active consent with a fresh one covering
// exactly the given scopes (not a union) — re-consenting with fewer
// scopes actually narrows what's granted, matching user intent from
// the consent screen rather than silently accumulating permissions.
func (r *ConsentsRepo) Grant(ctx context.Context, userID, clientID string, scopes []string) (*models.Consent, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		UPDATE consents SET revoked_at = now() WHERE user_id = $1 AND client_id = $2 AND revoked_at IS NULL
	`, userID, clientID); err != nil {
		return nil, err
	}

	row := tx.QueryRow(ctx, `
		INSERT INTO consents (user_id, client_id, scopes) VALUES ($1, $2, $3)
		RETURNING id, user_id, client_id, scopes, granted_at, revoked_at
	`, userID, clientID, scopes)
	consent, err := scanConsent(row)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return consent, nil
}

func (r *ConsentsRepo) Revoke(ctx context.Context, userID, clientID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE consents SET revoked_at = now() WHERE user_id = $1 AND client_id = $2 AND revoked_at IS NULL
	`, userID, clientID)
	return err
}

func (r *ConsentsRepo) ListActiveForUser(ctx context.Context, userID string) ([]*models.Consent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, client_id, scopes, granted_at, revoked_at
		FROM consents WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*models.Consent
	for rows.Next() {
		c, err := scanConsent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func scanConsent(row pgx.Row) (*models.Consent, error) {
	var c models.Consent
	err := row.Scan(&c.ID, &c.UserID, &c.ClientID, &c.Scopes, &c.GrantedAt, &c.RevokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}
