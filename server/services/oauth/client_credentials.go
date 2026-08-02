package oauth

import (
	"context"
	"fmt"
	"time"

	"ababilx-sso/models"
	"ababilx-sso/services/crypto"
	"ababilx-sso/services/token"
)

// ClientCredentialsGrant issues a machine-to-machine access token with
// no user identity: `sub` equals the client_id, never a pairwise or
// public user subject. Confidential clients only — a public client
// (no secret) authenticating as itself would let anyone who obtains
// the client_id mint tokens, since there's nothing to verify. No
// refresh token: a client that can re-authenticate with its own
// secret just requests a new access token directly.
func (s *Service) ClientCredentialsGrant(ctx context.Context, client *models.OAuthClient, requestedScope string) (*TokenResponse, error) {
	if !client.SupportsGrant("client_credentials") {
		return nil, ErrUnauthorizedClient
	}
	if !client.IsConfidential() {
		return nil, ErrUnauthorizedClient
	}

	scopes := splitScope(requestedScope)
	if len(scopes) == 0 {
		scopes = client.AllowedScopes
	}
	if !client.AllowsScopes(scopes) {
		return nil, ErrInvalidScope
	}
	if hasScope(scopes, "openid") || hasScope(scopes, "offline_access") {
		// Machine tokens never carry an identity or a refresh token.
		return nil, ErrInvalidScope
	}

	kid, priv, err := s.Keys.ActiveSigningKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("load signing key: %w", err)
	}

	now := time.Now().UTC()
	jti, err := crypto.RandomToken(16)
	if err != nil {
		return nil, err
	}

	accessToken, err := token.SignAccessToken(kid, priv, token.AccessTokenClaims{
		Issuer:    s.Issuer,
		Subject:   client.ClientID,
		Audience:  client.ClientID,
		ClientID:  client.ClientID,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(s.Lifetimes.AccessTokenTTL).Unix(),
		Scope:     joinScope(scopes),
		JTI:       jti,
	})
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	return &TokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.Lifetimes.AccessTokenTTL.Seconds()),
		Scope:       joinScope(scopes),
	}, nil
}
