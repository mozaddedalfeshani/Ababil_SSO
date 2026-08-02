package keys

import (
	"context"
	"encoding/json"
	"fmt"

	josejwk "github.com/go-jose/go-jose/v4"
)

// JWKS returns every key that should currently be published — active,
// next (pre-announced before activation so RPs cache-warm), and
// recently-retired within its grace period.
func (m *Manager) JWKS(ctx context.Context) (*josejwk.JSONWebKeySet, error) {
	published, err := m.repo.Published(ctx)
	if err != nil {
		return nil, err
	}

	set := &josejwk.JSONWebKeySet{}
	for _, k := range published {
		var jwk josejwk.JSONWebKey
		if err := json.Unmarshal(k.PublicJWK, &jwk); err != nil {
			return nil, fmt.Errorf("unmarshal stored jwk %s: %w", k.KID, err)
		}
		set.Keys = append(set.Keys, jwk)
	}
	return set, nil
}
