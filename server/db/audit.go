package db

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepo struct{ pool *pgxpool.Pool }

func NewAuditRepo(pool *pgxpool.Pool) *AuditRepo { return &AuditRepo{pool: pool} }

type WriteAuditParams struct {
	ActorUserID *string
	OrgID       *string
	ClientID    *string
	Event       string
	IPHash      *string
	UserAgent   *string
	Meta        map[string]any
}

func (r *AuditRepo) Write(ctx context.Context, p WriteAuditParams) error {
	meta := p.Meta
	if meta == nil {
		meta = map[string]any{}
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO audit_logs (actor_user_id, org_id, client_id, event, ip_hash, user_agent, meta)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, p.ActorUserID, p.OrgID, p.ClientID, p.Event, p.IPHash, p.UserAgent, metaJSON)
	return err
}

func (r *AuditRepo) ListForUser(ctx context.Context, userID string, limit int) ([]map[string]any, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT event, meta, created_at FROM audit_logs
		WHERE actor_user_id = $1 ORDER BY created_at DESC LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]any
	for rows.Next() {
		var event string
		var meta map[string]any
		var createdAt any
		if err := rows.Scan(&event, &meta, &createdAt); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"event": event, "meta": meta, "created_at": createdAt})
	}
	return out, rows.Err()
}

// AnonymizeForUser is called on account deletion: the trail stays
// (append-only, useful for security investigation) but stops naming
// the deleted account. ON DELETE SET NULL on actor_user_id already
// covers the FK; this only needs to run if the caller wants an
// explicit audit event recorded before the user row is removed.
func (r *AuditRepo) AnonymizeForUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE audit_logs SET meta = meta || '{"anonymized": true}'::jsonb
		WHERE actor_user_id = $1
	`, userID)
	return err
}
