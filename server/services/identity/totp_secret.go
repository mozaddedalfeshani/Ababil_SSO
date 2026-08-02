package identity

import (
	"fmt"

	"ababilx-sso/models"
	"ababilx-sso/services/crypto"
)

// TOTP secrets are sealed with AES-256-GCM under KEY_ENCRYPTION_KEY,
// AAD-bound to the user ID so a ciphertext copied to a different row
// (e.g. via a DB restore mistake) fails to decrypt rather than
// silently unlocking someone else's account.

func (s *Service) encryptTOTPSecret(userID, secretBase32 string) ([]byte, error) {
	ct, err := crypto.Seal(s.KeyEncryptionKey, []byte(secretBase32), []byte(userID))
	if err != nil {
		return nil, fmt.Errorf("seal totp secret: %w", err)
	}
	return ct, nil
}

func (s *Service) decryptTOTPSecret(user *models.User) (string, error) {
	if user.TOTPSecretEnc == nil {
		return "", ErrTOTPNotEnabled
	}
	pt, err := crypto.Open(s.KeyEncryptionKey, user.TOTPSecretEnc, []byte(user.ID))
	if err != nil {
		return "", fmt.Errorf("open totp secret: %w", err)
	}
	return string(pt), nil
}
