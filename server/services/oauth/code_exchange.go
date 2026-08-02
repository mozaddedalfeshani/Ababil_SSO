package oauth

import (
	"context"
	"errors"

	"ababilx-sso/db"
	"ababilx-sso/models"
	"ababilx-sso/services/audit"
	"ababilx-sso/services/crypto"
)

type TokenResponse struct {
	AccessToken  string
	TokenType    string
	ExpiresIn    int
	Scope        string
	RefreshToken string
	IDToken      string
}

// ExchangeAuthorizationCode redeems a code for tokens. The replay path
// is the security-critical part: if the code was already consumed
// (checked from the row fetched before the atomic Consume call, so we
// know whether "not consumed" means replay vs. ordinary expiry), the
// entire refresh-token family tied to that authorization is revoked
// and a code_replay audit event is recorded — see
// docs/architecture.md.
func (s *Service) ExchangeAuthorizationCode(ctx context.Context, client *models.OAuthClient, rawCode, redirectURI, codeVerifier string) (*TokenResponse, error) {
	if !client.SupportsGrant("authorization_code") {
		return nil, ErrUnauthorizedClient
	}

	code, err := s.Codes.ByHash(ctx, crypto.HashToken(rawCode))
	if errors.Is(err, db.ErrNotFound) {
		return nil, ErrInvalidGrant
	}
	if err != nil {
		return nil, err
	}
	if code.ClientID != client.ID {
		return nil, ErrInvalidGrant
	}

	alreadyConsumed := code.ConsumedAt != nil

	consumed, err := s.Codes.Consume(ctx, code.ID)
	if err != nil {
		return nil, err
	}
	if !consumed {
		if alreadyConsumed {
			s.handleCodeReplay(ctx, client, code)
		}
		return nil, ErrInvalidGrant
	}

	if code.RedirectURI != redirectURI {
		return nil, ErrInvalidGrant
	}
	if !VerifyPKCE(code.CodeChallenge, codeVerifier) {
		return nil, ErrInvalidGrant
	}

	scopes := splitScope(code.Scope)
	return s.issueTokenSet(ctx, client, code.UserID, scopes, code.AuthTime, code.Nonce, code.SessionID, code.RefreshFamilyID, true)
}

// handleCodeReplay revokes every refresh token descended from this
// authorization and logs the event. Best-effort: a failure here must
// not mask the invalid_grant response the client already gets.
func (s *Service) handleCodeReplay(ctx context.Context, client *models.OAuthClient, code *models.AuthorizationCode) {
	_ = s.RefreshTokens.RevokeFamily(ctx, code.RefreshFamilyID, "code_replay")
	_ = s.Audit.Record(ctx, audit.Event{
		ActorUserID: code.UserID,
		Event:       audit.EventCodeReplay,
		Meta:        map[string]any{"client_id": client.ClientID, "family_id": code.RefreshFamilyID},
	})
}

func hasScope(scopes []string, target string) bool {
	for _, s := range scopes {
		if s == target {
			return true
		}
	}
	return false
}
