package token

import (
	"encoding/json"
	"fmt"

	josejwk "github.com/go-jose/go-jose/v4"
	josejwt "github.com/go-jose/go-jose/v4/jwt"
)

// VerifyAccessToken checks the signature against the given JWKS
// (published keys, so a token signed just before rotation still
// verifies against a "retired-but-published" key) and returns the
// claims. It does not check expiry or scope — callers do that, since
// introspection and resource-server validation want to report *why*
// a token is invalid, not just that it is.
func VerifyAccessToken(jwks *josejwk.JSONWebKeySet, compact string) (*AccessTokenClaims, error) {
	parsed, err := josejwt.ParseSigned(compact, []josejwk.SignatureAlgorithm{josejwk.ES256})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	if len(parsed.Headers) == 0 {
		return nil, fmt.Errorf("token has no headers")
	}

	kid := parsed.Headers[0].KeyID
	matches := jwks.Key(kid)
	if len(matches) == 0 {
		return nil, fmt.Errorf("unknown signing key: %s", kid)
	}

	var raw json.RawMessage
	if err := parsed.Claims(matches[0].Key, &raw); err != nil {
		return nil, fmt.Errorf("verify signature: %w", err)
	}

	var claims AccessTokenClaims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return nil, fmt.Errorf("unmarshal claims: %w", err)
	}
	return &claims, nil
}
