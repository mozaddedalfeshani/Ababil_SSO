package oauth

import (
	"context"
	"errors"

	"ababilx-sso/db"
	"ababilx-sso/models"
	"ababilx-sso/services/crypto"
)

var ErrClientAuthFailed = errors.New("client authentication failed")

// AuthenticateClient resolves a client by its public client_id and,
// for confidential clients, verifies the presented secret. A public
// client (client_type=public, auth_method=none) presents no secret at
// all — PKCE is what authenticates the authorization-code exchange in
// that case, per OAuth 2.1.
func (s *Service) AuthenticateClient(ctx context.Context, clientID, clientSecret string) (*models.OAuthClient, error) {
	client, err := s.Clients.ByClientID(ctx, clientID)
	if errors.Is(err, db.ErrNotFound) {
		return nil, ErrClientAuthFailed
	}
	if err != nil {
		return nil, err
	}
	if client.Disabled() {
		return nil, ErrClientAuthFailed
	}

	if !client.IsConfidential() {
		return client, nil
	}

	if client.ClientSecretHMAC == nil || clientSecret == "" {
		return nil, ErrClientAuthFailed
	}
	if !crypto.EqualTokenHash(crypto.HashToken(clientSecret), *client.ClientSecretHMAC) {
		return nil, ErrClientAuthFailed
	}
	return client, nil
}
