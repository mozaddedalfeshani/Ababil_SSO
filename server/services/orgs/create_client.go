package orgs

import (
	"context"
	"fmt"

	"ababilx-sso/db"
	"ababilx-sso/models"
	"ababilx-sso/services/crypto"
)

type CreateClientParams struct {
	OrgID          string
	Name           string
	ClientType     models.ClientType
	RedirectURIs   []string
	PostLogoutURIs []string
	GrantTypes     []string
	AllowedScopes  []string
	SubjectType    models.SubjectType
	RequireConsent bool
	CreatedBy      string
}

// CreateClient generates the public client_id and, for confidential
// clients, a client_secret — returned in plaintext exactly once here;
// only its hash is stored (see services/crypto.HashToken).
func (s *Service) CreateClient(ctx context.Context, p CreateClientParams) (client *models.OAuthClient, plaintextSecret string, err error) {
	clientID, err := crypto.RandomToken(16)
	if err != nil {
		return nil, "", err
	}

	authMethod := models.AuthMethodNone
	var secretHMAC *string
	if p.ClientType == models.ClientTypeConfidential {
		plaintextSecret, err = crypto.RandomToken(32)
		if err != nil {
			return nil, "", err
		}
		hash := crypto.HashToken(plaintextSecret)
		secretHMAC = &hash
		authMethod = models.AuthMethodClientSecretBasic
	}

	subjectType := p.SubjectType
	if subjectType == "" {
		subjectType = models.SubjectTypePairwise
	}

	client, err = s.Clients.Create(ctx, db.CreateClientParams{
		OrgID:                   p.OrgID,
		ClientID:                clientID,
		ClientSecretHMAC:        secretHMAC,
		Name:                    p.Name,
		ClientType:              p.ClientType,
		TokenEndpointAuthMethod: authMethod,
		RedirectURIs:            p.RedirectURIs,
		PostLogoutRedirectURIs:  p.PostLogoutURIs,
		GrantTypes:              p.GrantTypes,
		AllowedScopes:           p.AllowedScopes,
		SubjectType:             subjectType,
		RequireConsent:          p.RequireConsent,
		CreatedBy:               p.CreatedBy,
	})
	if err != nil {
		return nil, "", fmt.Errorf("create client: %w", err)
	}
	return client, plaintextSecret, nil
}

// RotateSecret issues a new secret, invalidating the old one
// immediately — there is no overlap window, so callers must expect
// the previous secret to stop working the instant this returns.
func (s *Service) RotateSecret(ctx context.Context, clientID string) (plaintextSecret string, err error) {
	plaintextSecret, err = crypto.RandomToken(32)
	if err != nil {
		return "", err
	}
	if err := s.Clients.UpdateSecret(ctx, clientID, crypto.HashToken(plaintextSecret)); err != nil {
		return "", err
	}
	return plaintextSecret, nil
}
