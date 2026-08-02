package keys

import (
	"crypto/ecdsa"
	"crypto/x509"
	"fmt"
)

func marshalECPrivateKey(priv *ecdsa.PrivateKey) ([]byte, error) {
	der, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("marshal ec private key: %w", err)
	}
	return der, nil
}

func parseECPrivateKey(der []byte) (*ecdsa.PrivateKey, error) {
	priv, err := x509.ParseECPrivateKey(der)
	if err != nil {
		return nil, fmt.Errorf("parse ec private key: %w", err)
	}
	return priv, nil
}
