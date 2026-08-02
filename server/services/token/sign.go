package token

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"

	josejwt "github.com/go-jose/go-jose/v4"
)

// SignAccessToken produces a JWT with `typ: at+jwt` (RFC 9068) so a
// resource server can reject an ID token presented as an access
// token — the two are otherwise structurally similar enough to
// confuse, and ID tokens are routinely held/logged by RPs.
func SignAccessToken(kid string, priv *ecdsa.PrivateKey, claims AccessTokenClaims) (string, error) {
	return signCompact(kid, priv, claims, "at+jwt")
}

// SignIDToken omits the `typ` override — OIDC ID tokens use the
// standard JWT type.
func SignIDToken(kid string, priv *ecdsa.PrivateKey, claims IDTokenClaims) (string, error) {
	return signCompact(kid, priv, claims, "")
}

func signCompact(kid string, priv *ecdsa.PrivateKey, claims any, typHeader string) (string, error) {
	opts := (&josejwt.SignerOptions{}).WithHeader("kid", kid)
	if typHeader != "" {
		opts = opts.WithType(josejwt.ContentType(typHeader))
	}

	signer, err := josejwt.NewSigner(josejwt.SigningKey{Algorithm: josejwt.ES256, Key: priv}, opts)
	if err != nil {
		return "", fmt.Errorf("create signer: %w", err)
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal claims: %w", err)
	}

	jws, err := signer.Sign(payload)
	if err != nil {
		return "", fmt.Errorf("sign: %w", err)
	}

	compact, err := jws.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("serialize: %w", err)
	}
	return compact, nil
}
