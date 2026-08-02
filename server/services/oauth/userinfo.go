package oauth

import (
	"context"
	"errors"
	"time"

	"ababilx-sso/db"
	"ababilx-sso/models"
	"ababilx-sso/services/token"
)

var (
	ErrInvalidToken      = errors.New("invalid_token")
	ErrInsufficientScope = errors.New("insufficient_scope")
)

// UserInfo implements the OIDC userinfo endpoint: verify the access
// token, require the `openid` scope, and reverse the token's subject
// (pairwise or public — see models.SubjectType) back to a user row.
func (s *Service) UserInfo(ctx context.Context, accessToken string) (map[string]any, error) {
	claims, client, err := s.verifyAndResolveClient(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	scopes := splitScope(claims.Scope)
	if !hasScope(scopes, "openid") {
		return nil, ErrInsufficientScope
	}

	userID, err := s.resolveUserID(ctx, client, claims.Subject)
	if err != nil {
		return nil, err
	}
	user, err := s.Users.ByID(ctx, userID)
	if errors.Is(err, db.ErrNotFound) {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}

	out := map[string]any{"sub": claims.Subject}
	if hasScope(scopes, "email") {
		out["email"] = user.Email
		out["email_verified"] = user.EmailVerified()
	}
	return out, nil
}

func (s *Service) verifyAndResolveClient(ctx context.Context, accessToken string) (*token.AccessTokenClaims, *models.OAuthClient, error) {
	jwks, err := s.Keys.JWKS(ctx)
	if err != nil {
		return nil, nil, err
	}
	claims, err := token.VerifyAccessToken(jwks, accessToken)
	if err != nil {
		return nil, nil, ErrInvalidToken
	}
	if time.Now().UTC().Unix() >= claims.ExpiresAt {
		return nil, nil, ErrInvalidToken
	}

	client, err := s.Clients.ByClientID(ctx, claims.ClientID)
	if errors.Is(err, db.ErrNotFound) {
		return nil, nil, ErrInvalidToken
	}
	if err != nil {
		return nil, nil, err
	}
	return claims, client, nil
}

func (s *Service) resolveUserID(ctx context.Context, client *models.OAuthClient, subject string) (string, error) {
	if client.SubjectType == models.SubjectTypePublic {
		return subject, nil
	}
	userID, err := s.ClientSubjects.ByPairwiseSub(ctx, client.ID, subject)
	if errors.Is(err, db.ErrNotFound) {
		return "", ErrInvalidToken
	}
	return userID, err
}
