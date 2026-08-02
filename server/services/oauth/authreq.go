package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"ababilx-sso/cache"
	"ababilx-sso/db"
	"ababilx-sso/services/crypto"

	"github.com/redis/go-redis/v9"
)

// AuthRequest is the pending authorization request parked in Redis
// between the /oauth/authorize redirect and the consent decision.
// It is not durable — losing it only loses an in-flight login, never
// account or token state (see docs/architecture.md).
type AuthRequest struct {
	ID             string   `json:"id"`
	ClientID       string   `json:"client_id"` // internal UUID, not the public client_id string
	PublicClientID string   `json:"public_client_id"`
	RedirectURI    string   `json:"redirect_uri"`
	Scope          []string `json:"scope"`
	State          string   `json:"state"`
	Nonce          string   `json:"nonce"`
	CodeChallenge  string   `json:"code_challenge"`
	Prompt         string   `json:"prompt"`
}

type StartAuthorizeParams struct {
	ClientID            string
	RedirectURI         string
	ResponseType        string
	Scope               string
	State               string
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
	Prompt              string
}

// StartAuthorize validates an /oauth/authorize request and, if valid,
// parks it in Redis for the consent screen to resolve. Returns a
// *ClientError for anything that must render as a page (bad client_id
// or redirect_uri — see errors.go), and a *RedirectError for anything
// that should bounce back to the RP with an `error` parameter.
func (s *Service) StartAuthorize(ctx context.Context, p StartAuthorizeParams) (id string, err error) {
	client, err := s.Clients.ByClientID(ctx, p.ClientID)
	if errors.Is(err, db.ErrNotFound) || (err == nil && client.Disabled()) {
		return "", NewClientError("unknown or disabled client")
	}
	if err != nil {
		return "", err
	}

	if !client.AllowsRedirectURI(p.RedirectURI) {
		return "", NewClientError("redirect_uri is not registered for this client")
	}

	if p.ResponseType != "code" {
		return "", NewRedirectError("unsupported_response_type", "only response_type=code is supported")
	}

	if !ValidCodeChallengeMethod(p.CodeChallengeMethod) || p.CodeChallenge == "" {
		return "", NewRedirectError("invalid_request", "code_challenge with method S256 is required")
	}

	scopes := splitScope(p.Scope)
	if len(scopes) == 0 || !client.AllowsScopes(scopes) {
		return "", NewRedirectError("invalid_scope", "requested scope is not permitted for this client")
	}

	reqID, err := crypto.RandomToken(24)
	if err != nil {
		return "", err
	}

	authReq := AuthRequest{
		ID:             reqID,
		ClientID:       client.ID,
		PublicClientID: client.ClientID,
		RedirectURI:    p.RedirectURI,
		Scope:          scopes,
		State:          p.State,
		Nonce:          p.Nonce,
		CodeChallenge:  p.CodeChallenge,
		Prompt:         p.Prompt,
	}
	payload, err := json.Marshal(authReq)
	if err != nil {
		return "", err
	}

	if err := s.Redis.Set(ctx, cache.AuthRequestKey(reqID), payload, s.Lifetimes.AuthRequestTTL).Err(); err != nil {
		return "", err
	}
	return reqID, nil
}

var ErrAuthRequestNotFound = errors.New("authorization request not found or expired")

func (s *Service) GetAuthRequest(ctx context.Context, id string) (*AuthRequest, error) {
	raw, err := s.Redis.Get(ctx, cache.AuthRequestKey(id)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, ErrAuthRequestNotFound
	}
	if err != nil {
		return nil, err
	}

	var req AuthRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func (s *Service) deleteAuthRequest(ctx context.Context, id string) {
	s.Redis.Del(ctx, cache.AuthRequestKey(id))
}

func splitScope(scope string) []string {
	fields := strings.Fields(scope)
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}
