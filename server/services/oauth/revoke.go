package oauth

import (
	"context"

	"ababilx-sso/models"
	"ababilx-sso/services/crypto"
)

// Revoke implements RFC 7009. It always succeeds from the caller's
// perspective (no error for an unknown/already-revoked/foreign token)
// — that's the spec's own anti-enumeration requirement, not a bug: a
// client probing which tokens are valid must learn nothing from this
// endpoint's response. Access tokens are stateless JWTs with no
// server-side record, so "revoking" one is a no-op beyond its natural
// (short) expiry; only refresh tokens are actually mutated.
func (s *Service) Revoke(ctx context.Context, client *models.OAuthClient, rawToken string) {
	existing, err := s.RefreshTokens.ByTokenHash(ctx, crypto.HashToken(rawToken))
	if err != nil || existing.ClientID != client.ID || existing.RevokedAt != nil {
		return
	}
	_ = s.RefreshTokens.RevokeFamily(ctx, existing.FamilyID, "client_revoked")
}
