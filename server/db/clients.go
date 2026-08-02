package db

import (
	"context"
	"errors"

	"ababilx-sso/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ClientsRepo struct{ pool *pgxpool.Pool }

func NewClientsRepo(pool *pgxpool.Pool) *ClientsRepo { return &ClientsRepo{pool: pool} }

type CreateClientParams struct {
	OrgID                   string
	ClientID                string
	ClientSecretHMAC        *string
	Name                    string
	ClientType              models.ClientType
	TokenEndpointAuthMethod models.TokenEndpointAuthMethod
	RedirectURIs            []string
	PostLogoutRedirectURIs  []string
	GrantTypes              []string
	AllowedScopes           []string
	SubjectType             models.SubjectType
	RequireConsent          bool
	CreatedBy               string
}

func (r *ClientsRepo) Create(ctx context.Context, p CreateClientParams) (*models.OAuthClient, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO oauth_clients (org_id, client_id, client_secret_hmac, name, client_type,
			token_endpoint_auth_method, redirect_uris, post_logout_redirect_uris, grant_types,
			allowed_scopes, subject_type, require_consent, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, org_id, client_id, client_secret_hmac, name, logo_url, client_type,
			token_endpoint_auth_method, redirect_uris, post_logout_redirect_uris, grant_types,
			allowed_scopes, subject_type, sector_identifier, require_consent, created_by, disabled_at, created_at
	`, p.OrgID, p.ClientID, p.ClientSecretHMAC, p.Name, p.ClientType, p.TokenEndpointAuthMethod,
		nonNilStrings(p.RedirectURIs), nonNilStrings(p.PostLogoutRedirectURIs), nonNilStrings(p.GrantTypes),
		nonNilStrings(p.AllowedScopes), p.SubjectType, p.RequireConsent, p.CreatedBy)
	return scanClient(row)
}

func (r *ClientsRepo) ByClientID(ctx context.Context, clientID string) (*models.OAuthClient, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, org_id, client_id, client_secret_hmac, name, logo_url, client_type,
			token_endpoint_auth_method, redirect_uris, post_logout_redirect_uris, grant_types,
			allowed_scopes, subject_type, sector_identifier, require_consent, created_by, disabled_at, created_at
		FROM oauth_clients WHERE client_id = $1
	`, clientID)
	return scanClient(row)
}

func (r *ClientsRepo) ByID(ctx context.Context, id string) (*models.OAuthClient, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, org_id, client_id, client_secret_hmac, name, logo_url, client_type,
			token_endpoint_auth_method, redirect_uris, post_logout_redirect_uris, grant_types,
			allowed_scopes, subject_type, sector_identifier, require_consent, created_by, disabled_at, created_at
		FROM oauth_clients WHERE id = $1
	`, id)
	return scanClient(row)
}

func (r *ClientsRepo) ListForOrg(ctx context.Context, orgID string) ([]*models.OAuthClient, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, org_id, client_id, client_secret_hmac, name, logo_url, client_type,
			token_endpoint_auth_method, redirect_uris, post_logout_redirect_uris, grant_types,
			allowed_scopes, subject_type, sector_identifier, require_consent, created_by, disabled_at, created_at
		FROM oauth_clients WHERE org_id = $1 ORDER BY created_at DESC
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*models.OAuthClient
	for rows.Next() {
		cl, err := scanClient(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, cl)
	}
	return out, rows.Err()
}

func (r *ClientsRepo) UpdateSecret(ctx context.Context, id, secretHMAC string) error {
	_, err := r.pool.Exec(ctx, `UPDATE oauth_clients SET client_secret_hmac = $2 WHERE id = $1`, id, secretHMAC)
	return err
}

func (r *ClientsRepo) Update(ctx context.Context, id string, p CreateClientParams) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE oauth_clients SET name = $2, redirect_uris = $3, post_logout_redirect_uris = $4,
			grant_types = $5, allowed_scopes = $6, require_consent = $7
		WHERE id = $1
	`, id, p.Name, p.RedirectURIs, p.PostLogoutRedirectURIs, p.GrantTypes, p.AllowedScopes, p.RequireConsent)
	return err
}

func (r *ClientsRepo) Disable(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE oauth_clients SET disabled_at = now() WHERE id = $1`, id)
	return err
}

func (r *ClientsRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM oauth_clients WHERE id = $1`, id)
	return err
}

func scanClient(row pgx.Row) (*models.OAuthClient, error) {
	var c models.OAuthClient
	err := row.Scan(&c.ID, &c.OrgID, &c.ClientID, &c.ClientSecretHMAC, &c.Name, &c.LogoURL, &c.ClientType,
		&c.TokenEndpointAuthMethod, &c.RedirectURIs, &c.PostLogoutRedirectURIs, &c.GrantTypes,
		&c.AllowedScopes, &c.SubjectType, &c.SectorIdentifier, &c.RequireConsent, &c.CreatedBy,
		&c.DisabledAt, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}
