package oauth

import (
	"context"
	"errors"

	"ababilx-sso/db"
	"ababilx-sso/models"
	"ababilx-sso/services/audit"
	"ababilx-sso/services/crypto"
)

// RefreshGrant redeems a refresh token, always rotating: the
// presented token is consumed and a brand-new one is issued in its
// place, linked via family_id. A token presented with ConsumedAt
// already set is the actual reuse signal — someone is replaying a
// token that was already rotated away, meaning the token leaked and
// both the original and current holder now have "valid-looking"
// copies. That's treated as compromise: the entire family is revoked
// and a refresh_reuse_detected audit event is recorded. A concurrent
// redemption race (two requests hit Rotate at the same instant) is
// caught the same way as a fallback. requestedScopes, if non-empty,
// may only narrow the original grant.
func (s *Service) RefreshGrant(ctx context.Context, client *models.OAuthClient, rawToken string, requestedScopes []string) (*TokenResponse, error) {
	if !client.SupportsGrant("refresh_token") {
		return nil, ErrUnauthorizedClient
	}

	existing, err := s.RefreshTokens.ByTokenHash(ctx, crypto.HashToken(rawToken))
	if errors.Is(err, db.ErrNotFound) {
		return nil, ErrInvalidGrant
	}
	if err != nil {
		return nil, err
	}
	if existing.ClientID != client.ID {
		return nil, ErrInvalidGrant
	}

	if existing.ConsumedAt != nil {
		s.handleRefreshReuse(ctx, client, existing)
		return nil, ErrInvalidGrant
	}
	if existing.RevokedAt != nil || !nowUTC().Before(existing.ExpiresAt) {
		// Already revoked (e.g. a prior reuse event on this family, or
		// an explicit "revoke this app") or simply expired — refuse,
		// but this alone isn't reuse evidence.
		return nil, ErrInvalidGrant
	}

	grantedScopes := splitScope(existing.Scope)
	scopes := grantedScopes
	if len(requestedScopes) > 0 {
		if !scopeSubsetOf(requestedScopes, grantedScopes) {
			return nil, ErrInvalidScope
		}
		scopes = requestedScopes
	}

	sessionID := ""
	if existing.SessionID != nil {
		sessionID = *existing.SessionID
	}

	tokens, err := s.issueTokenSet(ctx, client, existing.UserID, scopes, nowUTC(), nil, sessionID, existing.FamilyID, true)
	if err != nil {
		return nil, err
	}

	// The new token's hash was just written by issueTokenSet; look it
	// up to get its ID for Rotate's linkage. Cheaper than threading
	// the ID back out of issueTokenSet for this one caller.
	newToken, err := s.RefreshTokens.ByTokenHash(ctx, crypto.HashToken(tokens.RefreshToken))
	if err != nil {
		return nil, err
	}

	rotated, err := s.RefreshTokens.Rotate(ctx, existing.ID, newToken.ID)
	if err != nil {
		return nil, err
	}
	if !rotated {
		// Lost a race with a concurrent redemption of the same token —
		// treat identically to reuse: kill the family, including the
		// token we just minted.
		s.handleRefreshReuse(ctx, client, existing)
		return nil, ErrInvalidGrant
	}

	return tokens, nil
}

func (s *Service) handleRefreshReuse(ctx context.Context, client *models.OAuthClient, reused *models.RefreshToken) {
	_ = s.RefreshTokens.RevokeFamily(ctx, reused.FamilyID, "refresh_reuse_detected")
	_ = s.Audit.Record(ctx, audit.Event{
		ActorUserID: reused.UserID,
		Event:       audit.EventRefreshReuseDetected,
		Meta:        map[string]any{"client_id": client.ClientID, "family_id": reused.FamilyID},
	})
}

func scopeSubsetOf(requested, granted []string) bool {
	grantedSet := make(map[string]bool, len(granted))
	for _, s := range granted {
		grantedSet[s] = true
	}
	for _, s := range requested {
		if !grantedSet[s] {
			return false
		}
	}
	return true
}
