package db

import (
	"context"
	"errors"
	"time"

	"ababilx-sso/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionsRepo struct{ pool *pgxpool.Pool }

func NewSessionsRepo(pool *pgxpool.Pool) *SessionsRepo { return &SessionsRepo{pool: pool} }

type CreateSessionParams struct {
	UserID            string
	TokenHash         string
	IPHash            *string
	UserAgent         *string
	AMR               []string
	AuthTime          time.Time
	IdleExpiresAt     time.Time
	AbsoluteExpiresAt time.Time
}

func (r *SessionsRepo) Create(ctx context.Context, p CreateSessionParams) (*models.Session, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO sessions (user_id, token_hash, ip_hash, user_agent, amr, auth_time, idle_expires_at, absolute_expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, token_hash, ip_hash, user_agent, amr, auth_time, idle_expires_at, absolute_expires_at, revoked_at, created_at
	`, p.UserID, p.TokenHash, p.IPHash, p.UserAgent, nonNilStrings(p.AMR), p.AuthTime, p.IdleExpiresAt, p.AbsoluteExpiresAt)
	return scanSession(row)
}

func (r *SessionsRepo) ByTokenHash(ctx context.Context, tokenHash string) (*models.Session, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, ip_hash, user_agent, amr, auth_time, idle_expires_at, absolute_expires_at, revoked_at, created_at
		FROM sessions WHERE token_hash = $1
	`, tokenHash)
	return scanSession(row)
}

func (r *SessionsRepo) ByID(ctx context.Context, id string) (*models.Session, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, ip_hash, user_agent, amr, auth_time, idle_expires_at, absolute_expires_at, revoked_at, created_at
		FROM sessions WHERE id = $1
	`, id)
	return scanSession(row)
}

func (r *SessionsRepo) ListActive(ctx context.Context, userID string) ([]*models.Session, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, token_hash, ip_hash, user_agent, amr, auth_time, idle_expires_at, absolute_expires_at, revoked_at, created_at
		FROM sessions WHERE user_id = $1 AND revoked_at IS NULL AND absolute_expires_at > now()
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*models.Session
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// ExtendIdle slides the idle-timeout window forward on activity,
// bounded by the absolute expiry which never moves.
func (r *SessionsRepo) ExtendIdle(ctx context.Context, id string, newIdleExpiresAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE sessions SET idle_expires_at = $2
		WHERE id = $1 AND revoked_at IS NULL AND $2 < absolute_expires_at
	`, id, newIdleExpiresAt)
	return err
}

func (r *SessionsRepo) Revoke(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE sessions SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL`, id)
	return err
}

// RevokeAllForUser is used on password reset/change: every existing
// session dies, since a password change is a signal the prior
// credential may have been compromised.
func (r *SessionsRepo) RevokeAllForUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	return err
}

func scanSession(row pgx.Row) (*models.Session, error) {
	var s models.Session
	err := row.Scan(&s.ID, &s.UserID, &s.TokenHash, &s.IPHash, &s.UserAgent, &s.AMR,
		&s.AuthTime, &s.IdleExpiresAt, &s.AbsoluteExpiresAt, &s.RevokedAt, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}
