// Package keys manages ES256 signing-key generation, sealing, rotation,
// and JWKS assembly. See docs/architecture.md "Key management" for why
// generation happens under the same advisory lock as migrations, and
// why retired keys stay published for a grace period.
package keys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"

	"ababilx-sso/services/crypto"

	josejwk "github.com/go-jose/go-jose/v4"
)

const Algorithm = "ES256"

// generateKeyPair creates a new P-256 key and returns everything
// needed to persist it: a random kid, the sealed private key, and the
// public JWK ready to embed in a JWKS document.
func generateKeyPair(kek []byte) (kid string, sealedPrivate []byte, publicJWKJSON []byte, err error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", nil, nil, fmt.Errorf("generate ec key: %w", err)
	}

	kid, err = crypto.RandomToken(16)
	if err != nil {
		return "", nil, nil, fmt.Errorf("generate kid: %w", err)
	}

	privDER, err := marshalECPrivateKey(priv)
	if err != nil {
		return "", nil, nil, err
	}
	sealedPrivate, err = crypto.Seal(kek, privDER, []byte(kid))
	if err != nil {
		return "", nil, nil, fmt.Errorf("seal private key: %w", err)
	}

	publicJWK := josejwk.JSONWebKey{
		Key:       priv.Public(),
		KeyID:     kid,
		Algorithm: Algorithm,
		Use:       "sig",
	}
	publicJWKJSON, err = json.Marshal(publicJWK)
	if err != nil {
		return "", nil, nil, fmt.Errorf("marshal public jwk: %w", err)
	}

	return kid, sealedPrivate, publicJWKJSON, nil
}

func unsealPrivateKey(sealed []byte, kid string, kek []byte) (*ecdsa.PrivateKey, error) {
	der, err := crypto.Open(kek, sealed, []byte(kid))
	if err != nil {
		return nil, fmt.Errorf("open private key: %w", err)
	}
	return parseECPrivateKey(der)
}
