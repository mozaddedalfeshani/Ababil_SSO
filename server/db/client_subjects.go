package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ClientSubjectsRepo struct{ pool *pgxpool.Pool }

func NewClientSubjectsRepo(pool *pgxpool.Pool) *ClientSubjectsRepo {
	return &ClientSubjectsRepo{pool: pool}
}

// GetOrCreate is the core of the pairwise-subject-identifier privacy
// control: the first call for a (client, user) pair mints a random
// value and every subsequent call returns that same value, so the
// identifier is stable per-client but uncorrelatable across clients.
// ON CONFLICT makes this safe under concurrent first-authorize races.
func (r *ClientSubjectsRepo) GetOrCreate(ctx context.Context, clientID, userID, newPairwiseSub string) (string, error) {
	var sub string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO client_subjects (client_id, user_id, pairwise_sub)
		VALUES ($1, $2, $3)
		ON CONFLICT (client_id, user_id) DO UPDATE SET client_id = EXCLUDED.client_id
		RETURNING pairwise_sub
	`, clientID, userID, newPairwiseSub).Scan(&sub)
	return sub, err
}

// ByPairwiseSub reverses the mapping — used by userinfo/introspect
// when a request only carries the pairwise sub from a token.
func (r *ClientSubjectsRepo) ByPairwiseSub(ctx context.Context, clientID, pairwiseSub string) (string, error) {
	var userID string
	err := r.pool.QueryRow(ctx, `
		SELECT user_id FROM client_subjects WHERE client_id = $1 AND pairwise_sub = $2
	`, clientID, pairwiseSub).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return userID, err
}
