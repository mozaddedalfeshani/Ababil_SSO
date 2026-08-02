// Package totp implements TOTP enrollment and verification with a
// persistent replay guard — the step counter lives in Postgres
// (users.totp_last_step), not just checked against a ±1 window, so a
// phished or shoulder-surfed code cannot be reused even within its
// nominal validity period.
package totp

import (
	"fmt"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const stepSeconds = 30

type Secret struct {
	Base32  string
	OTPAuth string // otpauth:// URI for QR code rendering
}

func GenerateSecret(issuer, accountEmail string) (*Secret, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountEmail,
	})
	if err != nil {
		return nil, fmt.Errorf("generate totp key: %w", err)
	}
	return &Secret{Base32: key.Secret(), OTPAuth: key.URL()}, nil
}

// Verify checks code against secret at the current time step, and
// additionally requires step > lastConsumedStep — the actual replay
// guard. The library's built-in ±1 window is intentionally NOT used
// here (via ValidateCustom with Skew 0) because tolerating adjacent
// steps widens the replay window this guard exists to close; the
// user experience cost is a resend, which is cheap.
func Verify(code, secretBase32 string, lastConsumedStep int64, now time.Time) (step int64, ok bool, err error) {
	step = now.Unix() / stepSeconds

	valid, err := totp.ValidateCustom(code, secretBase32, now, totp.ValidateOpts{
		Period:    stepSeconds,
		Skew:      0,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return step, false, fmt.Errorf("validate totp: %w", err)
	}
	if !valid {
		return step, false, nil
	}
	if step <= lastConsumedStep {
		return step, false, nil
	}
	return step, true, nil
}
