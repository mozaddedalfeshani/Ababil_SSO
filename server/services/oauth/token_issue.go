package oauth

import (
	"context"
	"fmt"
	"time"

	"ababilx-sso/db"
	"ababilx-sso/models"
	"ababilx-sso/services/crypto"
	"ababilx-sso/services/token"
)

// issueTokenSet mints access/ID/refresh tokens for a grant. Shared by
// the authorization_code and refresh_token grants — client_credentials
// has its own path (no user, no ID token, no refresh) in
// client_credentials.go.
//
// sessionID may be empty (client_credentials has none). familyID
// threads a refresh token's lineage back to the authorization that
// started it, so RevokeFamily on reuse detection kills every
// descendant, not just the immediately reused token.
func (s *Service) issueTokenSet(ctx context.Context, client *models.OAuthClient, userID string, scopes []string, authTime time.Time, nonce *string, sessionID, familyID string, mintRefresh bool) (*TokenResponse, error) {
	kid, priv, err := s.Keys.ActiveSigningKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("load signing key: %w", err)
	}

	sub, err := s.subjectFor(ctx, client, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	jti, err := crypto.RandomToken(16)
	if err != nil {
		return nil, err
	}

	accessToken, err := token.SignAccessToken(kid, priv, token.AccessTokenClaims{
		Issuer:    s.Issuer,
		Subject:   sub,
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

	resp := &TokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.Lifetimes.AccessTokenTTL.Seconds()),
		Scope:       joinScope(scopes),
	}

	if hasScope(scopes, "openid") {
		idClaims := token.IDTokenClaims{
			Issuer:    s.Issuer,
			Subject:   sub,
			Audience:  client.ClientID,
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(s.Lifetimes.AccessTokenTTL).Unix(),
			AuthTime:  authTime.Unix(),
		}
		if nonce != nil {
			idClaims.Nonce = *nonce
		}
		if sessionID != "" {
			if session, err := s.Sessions.ByID(ctx, sessionID); err == nil {
				idClaims.AMR = session.AMR
			}
		}
		if hasScope(scopes, "email") {
			if user, err := s.Users.ByID(ctx, userID); err == nil {
				idClaims.Email = user.Email
				verified := user.EmailVerified()
				idClaims.EmailVerified = &verified
			}
		}

		idToken, err := token.SignIDToken(kid, priv, idClaims)
		if err != nil {
			return nil, fmt.Errorf("sign id token: %w", err)
		}
		resp.IDToken = idToken
	}

	if mintRefresh && client.SupportsGrant("refresh_token") {
		rawRefresh, err := crypto.RandomToken(32)
		if err != nil {
			return nil, err
		}

		var sessionIDPtr *string
		if sessionID != "" {
			sessionIDPtr = &sessionID
		}

		if _, err := s.RefreshTokens.Create(ctx, db.CreateRefreshTokenParams{
			TokenHash:    crypto.HashToken(rawRefresh),
			FamilyID:     familyID,
			ClientID:     client.ID,
			UserID:       userID,
			SessionID:    sessionIDPtr,
			Scope:        joinScope(scopes),
			SessionBound: !hasScope(scopes, "offline_access"),
			ExpiresAt:    now.Add(s.Lifetimes.RefreshAbsolute),
		}); err != nil {
			return nil, fmt.Errorf("create refresh token: %w", err)
		}
		resp.RefreshToken = rawRefresh
	}

	return resp, nil
}

// subjectFor returns the pairwise subject for the client's sector
// unless the client explicitly opted into public subjects — see
// docs/architecture.md and models.SubjectType.
func (s *Service) subjectFor(ctx context.Context, client *models.OAuthClient, userID string) (string, error) {
	if client.SubjectType == models.SubjectTypePublic {
		return userID, nil
	}

	sector := client.ClientID
	if client.SectorIdentifier != nil && *client.SectorIdentifier != "" {
		sector = *client.SectorIdentifier
	}
	pairwise := token.DerivePairwiseSub(s.KeyEncryptionKey, sector, userID)

	stored, err := s.ClientSubjects.GetOrCreate(ctx, client.ID, userID, pairwise)
	if err != nil {
		return "", fmt.Errorf("resolve pairwise subject: %w", err)
	}
	return stored, nil
}
