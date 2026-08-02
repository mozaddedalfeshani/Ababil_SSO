package db

import (
	"context"
	"errors"

	"ababilx-sso/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrgsRepo struct{ pool *pgxpool.Pool }

func NewOrgsRepo(pool *pgxpool.Pool) *OrgsRepo { return &OrgsRepo{pool: pool} }

func (r *OrgsRepo) Create(ctx context.Context, name, slug, ownerUserID string) (*models.Organization, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		INSERT INTO organizations (name, slug, owner_user_id) VALUES ($1, $2, $3)
		RETURNING id, name, slug, owner_user_id, created_at
	`, name, slug, ownerUserID)
	org, err := scanOrg(row)
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO organization_members (org_id, user_id, role) VALUES ($1, $2, 'owner')
	`, org.ID, ownerUserID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return org, nil
}

func (r *OrgsRepo) ByID(ctx context.Context, id string) (*models.Organization, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, owner_user_id, created_at FROM organizations WHERE id = $1
	`, id)
	return scanOrg(row)
}

func (r *OrgsRepo) BySlug(ctx context.Context, slug string) (*models.Organization, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, owner_user_id, created_at FROM organizations WHERE slug = $1
	`, slug)
	return scanOrg(row)
}

func (r *OrgsRepo) ListForUser(ctx context.Context, userID string) ([]*models.Organization, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT o.id, o.name, o.slug, o.owner_user_id, o.created_at
		FROM organizations o
		JOIN organization_members m ON m.org_id = o.id
		WHERE m.user_id = $1
		ORDER BY o.created_at
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*models.Organization
	for rows.Next() {
		o, err := scanOrg(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *OrgsRepo) MemberRole(ctx context.Context, orgID, userID string) (models.OrgRole, error) {
	var role models.OrgRole
	err := r.pool.QueryRow(ctx, `
		SELECT role FROM organization_members WHERE org_id = $1 AND user_id = $2
	`, orgID, userID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return role, err
}

func (r *OrgsRepo) AddMember(ctx context.Context, orgID, userID string, role models.OrgRole) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO organization_members (org_id, user_id, role) VALUES ($1, $2, $3)
		ON CONFLICT (org_id, user_id) DO UPDATE SET role = EXCLUDED.role
	`, orgID, userID, role)
	return err
}

func (r *OrgsRepo) RemoveMember(ctx context.Context, orgID, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM organization_members WHERE org_id = $1 AND user_id = $2`, orgID, userID)
	return err
}

func (r *OrgsRepo) ListMembers(ctx context.Context, orgID string) ([]*models.OrganizationMember, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT org_id, user_id, role, created_at FROM organization_members WHERE org_id = $1
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*models.OrganizationMember
	for rows.Next() {
		var m models.OrganizationMember
		if err := rows.Scan(&m.OrgID, &m.UserID, &m.Role, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &m)
	}
	return out, rows.Err()
}

func scanOrg(row pgx.Row) (*models.Organization, error) {
	var o models.Organization
	err := row.Scan(&o.ID, &o.Name, &o.Slug, &o.OwnerUserID, &o.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}
