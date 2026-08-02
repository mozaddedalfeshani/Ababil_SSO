package totp

import (
	"fmt"

	"ababilx-sso/services/crypto"
)

const recoveryCodeCount = 10

// GenerateRecoveryCodes returns plaintext codes (shown to the user
// exactly once) and their hashes (what gets stored). Format
// "xxxx-xxxx" — 8 hex chars split for readability, ~32 bits of entropy
// per code which is adequate since each code is single-use and the
// account is already behind a password + this second factor.
func GenerateRecoveryCodes() (plaintext []string, hashes []string, err error) {
	plaintext = make([]string, recoveryCodeCount)
	hashes = make([]string, recoveryCodeCount)

	for i := 0; i < recoveryCodeCount; i++ {
		raw, err := crypto.RandomHex(4)
		if err != nil {
			return nil, nil, fmt.Errorf("generate recovery code: %w", err)
		}
		code := raw[:4] + "-" + raw[4:]
		plaintext[i] = code
		hashes[i] = crypto.HashToken(code)
	}
	return plaintext, hashes, nil
}
