package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RecoveryCodesRepo struct{ pool *pgxpool.Pool }

func NewRecoveryCodesRepo(pool *pgxpool.Pool) *RecoveryCodesRepo {
	return &RecoveryCodesRepo{pool: pool}
}

// ReplaceAll deletes any existing codes and inserts a fresh batch —
// called on TOTP enrollment and on regeneration, never appended to.
func (r *RecoveryCodesRepo) ReplaceAll(ctx context.Context, userID string, codeHashes []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM user_recovery_codes WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for _, h := range codeHashes {
		if _, err := tx.Exec(ctx, `INSERT INTO user_recovery_codes (user_id, code_hash) VALUES ($1, $2)`, userID, h); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// Consume atomically redeems a matching unused code; returns whether
// one was found and consumed.
func (r *RecoveryCodesRepo) Consume(ctx context.Context, userID, codeHash string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE user_recovery_codes SET consumed_at = now()
		WHERE user_id = $1 AND code_hash = $2 AND consumed_at IS NULL
	`, userID, codeHash)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *RecoveryCodesRepo) CountRemaining(ctx context.Context, userID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT count(*) FROM user_recovery_codes WHERE user_id = $1 AND consumed_at IS NULL
	`, userID).Scan(&n)
	return n, err
}
