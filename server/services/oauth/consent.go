package oauth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"ababilx-sso/db"
	"ababilx-sso/services/crypto"
)

// NeedsConsent reports whether the user must see the consent screen:
// first-party clients that opt out of it, or a returning user whose
// existing (unrevoked) consent already covers every requested scope,
// skip the prompt — anything else, including an explicit
// prompt=consent from the RP, shows it.
func (s *Service) NeedsConsent(ctx context.Context, req *AuthRequest, userID string) (bool, error) {
	client, err := s.Clients.ByID(ctx, req.ClientID)
	if err != nil {
		return false, err
	}
	if !client.RequireConsent {
		return false, nil
	}
	if req.Prompt == "consent" {
		return true, nil
	}

	existing, err := s.Consents.Active(ctx, userID, req.ClientID)
	if errors.Is(err, db.ErrNotFound) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return !existing.Satisfies(req.Scope), nil
}

// CompleteAuthorization records consent, mints a single-use
// authorization code, and returns the redirect URL back to the RP.
// grantedScopes may be a subset of req.Scope if the user unchecked
// something on the consent screen — never a superset (see
// models.OAuthClient.AllowsScopes, already enforced when the request
// was parked).
func (s *Service) CompleteAuthorization(ctx context.Context, req *AuthRequest, userID, sessionID string, authTime time.Time, grantedScopes []string) (redirectURL string, err error) {
	if _, err := s.Consents.Grant(ctx, userID, req.ClientID, grantedScopes); err != nil {
		return "", fmt.Errorf("record consent: %w", err)
	}

	rawCode, err := crypto.RandomToken(32)
	if err != nil {
		return "", err
	}

	var nonce *string
	if req.Nonce != "" {
		nonce = &req.Nonce
	}

	if _, err := s.Codes.Create(ctx, db.CreateCodeParams{
		CodeHash:      crypto.HashToken(rawCode),
		ClientID:      req.ClientID,
		UserID:        userID,
		SessionID:     sessionID,
		RedirectURI:   req.RedirectURI,
		Scope:         joinScope(grantedScopes),
		Nonce:         nonce,
		CodeChallenge: req.CodeChallenge,
		AuthTime:      authTime,
		ExpiresAt:     time.Now().UTC().Add(s.Lifetimes.CodeTTL),
	}); err != nil {
		return "", fmt.Errorf("create authorization code: %w", err)
	}

	s.deleteAuthRequest(ctx, req.ID)

	u, err := url.Parse(req.RedirectURI)
	if err != nil {
		return "", fmt.Errorf("parse redirect_uri: %w", err)
	}
	q := u.Query()
	q.Set("code", rawCode)
	if req.State != "" {
		q.Set("state", req.State)
	}
	q.Set("iss", s.Issuer) // RFC 9207 — prevents IdP mix-up attacks
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// DenyAuthorization builds the access_denied redirect back to the RP
// and clears the parked request — used when the user declines consent
// on the screen.
func (s *Service) DenyAuthorization(ctx context.Context, req *AuthRequest) string {
	s.deleteAuthRequest(ctx, req.ID)

	u, err := url.Parse(req.RedirectURI)
	if err != nil {
		return req.RedirectURI
	}
	q := u.Query()
	q.Set("error", "access_denied")
	q.Set("error_description", "the user declined the authorization request")
	if req.State != "" {
		q.Set("state", req.State)
	}
	q.Set("iss", s.Issuer)
	u.RawQuery = q.Encode()
	return u.String()
}

func joinScope(scopes []string) string {
	out := ""
	for i, s := range scopes {
		if i > 0 {
			out += " "
		}
		out += s
	}
	return out
}
