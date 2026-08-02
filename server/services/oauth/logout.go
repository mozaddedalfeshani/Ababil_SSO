package oauth

import (
	"context"
	"errors"
	"net/url"

	"ababilx-sso/db"
	"ababilx-sso/models"

	josejwk "github.com/go-jose/go-jose/v4"
	josejwt "github.com/go-jose/go-jose/v4/jwt"
)

var ErrInvalidLogoutRequest = errors.New("invalid logout request")

// ResolveLogoutClient validates an RP-initiated logout request per
// OIDC RP-Initiated Logout 1.0: id_token_hint identifies which client
// is asking (its signature is checked against our own JWKS — we
// issued it, so this also rules out a forged hint naming an arbitrary
// client), and post_logout_redirect_uri must be registered for that
// client. Both checks matter: without them this endpoint would be an
// open redirect plus a way to end an arbitrary user's session via a
// bare GET (CSRF) — see docs/architecture.md "Logout hardening".
// Handlers additionally require interactive confirmation before
// calling CompleteLogout.
func (s *Service) ResolveLogoutClient(ctx context.Context, idTokenHint, postLogoutRedirectURI string) (*models.OAuthClient, error) {
	if idTokenHint == "" {
		return nil, ErrInvalidLogoutRequest
	}

	jwks, err := s.Keys.JWKS(ctx)
	if err != nil {
		return nil, err
	}
	parsed, err := josejwt.ParseSigned(idTokenHint, []josejwk.SignatureAlgorithm{josejwk.ES256})
	if err != nil || len(parsed.Headers) == 0 {
		return nil, ErrInvalidLogoutRequest
	}
	matches := jwks.Key(parsed.Headers[0].KeyID)
	if len(matches) == 0 {
		return nil, ErrInvalidLogoutRequest
	}

	var claims struct {
		Audience string `json:"aud"`
	}
	if err := parsed.Claims(matches[0].Key, &claims); err != nil {
		return nil, ErrInvalidLogoutRequest
	}

	client, err := s.Clients.ByClientID(ctx, claims.Audience)
	if errors.Is(err, db.ErrNotFound) {
		return nil, ErrInvalidLogoutRequest
	}
	if err != nil {
		return nil, err
	}

	if postLogoutRedirectURI != "" && !client.AllowsPostLogoutRedirectURI(postLogoutRedirectURI) {
		return nil, ErrInvalidLogoutRequest
	}
	return client, nil
}

// CompleteLogout revokes the session and every session-bound refresh
// token tied to it (offline_access tokens are untouched by design —
// see docs/architecture.md "Token/session binding").
func (s *Service) CompleteLogout(ctx context.Context, sessionID string) error {
	if err := s.RefreshTokens.RevokeSessionBound(ctx, sessionID, "logout"); err != nil {
		return err
	}
	return s.Sessions.Revoke(ctx, sessionID)
}

func BuildLogoutRedirect(postLogoutRedirectURI, state string) (string, error) {
	if postLogoutRedirectURI == "" {
		return "", nil
	}
	u, err := url.Parse(postLogoutRedirectURI)
	if err != nil {
		return "", err
	}
	if state != "" {
		q := u.Query()
		q.Set("state", state)
		u.RawQuery = q.Encode()
	}
	return u.String(), nil
}
