package oauth

import (
	"context"

	"ababilx-sso/models"
	"ababilx-sso/services/crypto"
)

// Introspect implements RFC 7662, scoped to the caller: a client can
// only learn about tokens issued to itself. Any token belonging to a
// different client returns {"active": false} — the same response as
// an unknown token — so introspection can never be used to enumerate
// another client's token state. caller must already be an
// authenticated confidential client (enforced by the handler before
// calling this).
func (s *Service) Introspect(ctx context.Context, caller *models.OAuthClient, rawToken string) map[string]any {
	if claims, client, err := s.verifyAndResolveClient(ctx, rawToken); err == nil {
		if client.ID != caller.ID {
			return map[string]any{"active": false}
		}
		return map[string]any{
			"active":     true,
			"scope":      claims.Scope,
			"client_id":  claims.ClientID,
			"sub":        claims.Subject,
			"iss":        claims.Issuer,
			"exp":        claims.ExpiresAt,
			"iat":        claims.IssuedAt,
			"token_type": "Bearer",
		}
	}

	existing, err := s.RefreshTokens.ByTokenHash(ctx, crypto.HashToken(rawToken))
	if err != nil || existing.ClientID != caller.ID || !existing.Active(nowUTC()) {
		return map[string]any{"active": false}
	}
	return map[string]any{
		"active":    true,
		"scope":     existing.Scope,
		"client_id": caller.ClientID,
		"sub":       existing.UserID,
		"exp":       existing.ExpiresAt.Unix(),
	}
}
