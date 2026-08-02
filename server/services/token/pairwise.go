package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// DerivePairwiseSub computes a stable-but-uncorrelatable subject
// identifier for a (client, user) pair, per OIDC Core §8.1. HMAC
// keyed by the server's KEY_ENCRYPTION_KEY so the value can't be
// recomputed by anyone outside this server, and the sector identifier
// input (client ID, not client name/URL) means renaming a client
// doesn't change the sub value users are already correlated under.
func DerivePairwiseSub(kek []byte, sectorIdentifier, userID string) string {
	mac := hmac.New(sha256.New, kek)
	mac.Write([]byte(sectorIdentifier))
	mac.Write([]byte{0})
	mac.Write([]byte(userID))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
