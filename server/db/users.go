package db

import (
	"context"
	"errors"

	"ababilx-sso/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type UsersRepo struct{ pool *pgxpool.Pool }

func NewUsersRepo(pool *pgxpool.Pool) *UsersRepo { return &UsersRepo{pool: pool} }

func (r *UsersRepo) Create(ctx context.Context, normalizedEmail, passwordHash string) (*models.User, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, email_verified_at, password_hash, totp_secret_enc,
			totp_last_step, totp_enabled_at, status, created_at, last_login_at
	`, normalizedEmail, passwordHash)
	return scanUser(row)
}

func (r *UsersRepo) ByEmail(ctx context.Context, normalizedEmail string) (*models.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, email_verified_at, password_hash, totp_secret_enc,
			totp_last_step, totp_enabled_at, status, created_at, last_login_at
		FROM users WHERE lower(email) = lower($1)
	`, normalizedEmail)
	return scanUser(row)
}

func (r *UsersRepo) ByID(ctx context.Context, id string) (*models.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, email_verified_at, password_hash, totp_secret_enc,
			totp_last_step, totp_enabled_at, status, created_at, last_login_at
		FROM users WHERE id = $1
	`, id)
	return scanUser(row)
}

func (r *UsersRepo) MarkEmailVerified(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET email_verified_at = now() WHERE id = $1`, userID)
	return err
}

func (r *UsersRepo) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET password_hash = $2 WHERE id = $1`, userID, passwordHash)
	return err
}

func (r *UsersRepo) UpdateLastLogin(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET last_login_at = now() WHERE id = $1`, userID)
	return err
}

func (r *UsersRepo) SetTOTPSecret(ctx context.Context, userID string, encSecret []byte) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET totp_secret_enc = $2, totp_enabled_at = NULL, totp_last_step = NULL
		WHERE id = $1
	`, userID, encSecret)
	return err
}

func (r *UsersRepo) EnableTOTP(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET totp_enabled_at = now() WHERE id = $1`, userID)
	return err
}

func (r *UsersRepo) DisableTOTP(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET totp_secret_enc = NULL, totp_enabled_at = NULL, totp_last_step = NULL
		WHERE id = $1
	`, userID)
	return err
}

// UpdateTOTPLastStep is the replay guard: it only advances forward,
// enforced by the WHERE clause, so a code for a time step already
// consumed can never be recorded as newly-used from an older step.
func (r *UsersRepo) UpdateTOTPLastStep(ctx context.Context, userID string, step int64) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE users SET totp_last_step = $2
		WHERE id = $1 AND (totp_last_step IS NULL OR totp_last_step < $2)
	`, userID, step)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *UsersRepo) Delete(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	return err
}

func scanUser(row pgx.Row) (*models.User, error) {
	var u models.User
	err := row.Scan(&u.ID, &u.Email, &u.EmailVerifiedAt, &u.PasswordHash, &u.TOTPSecretEnc,
		&u.TOTPLastStep, &u.TOTPEnabledAt, &u.Status, &u.CreatedAt, &u.LastLoginAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
