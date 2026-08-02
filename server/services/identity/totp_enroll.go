package identity

import (
	"context"

	"ababilx-sso/services/totp"
)

type EnrollmentStart struct {
	Secret  string // base32, for manual entry
	OTPAuth string // otpauth:// URI, for QR rendering
}

// EnrollTOTP generates and stores a new secret in a disabled state —
// TOTPEnabledAt stays NULL until ConfirmEnrollment proves the user
// actually has a working authenticator, so a half-finished setup never
// silently gates future logins.
func (s *Service) EnrollTOTP(ctx context.Context, userID, userEmail string) (*EnrollmentStart, error) {
	user, err := s.Users.ByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.TOTPEnabled() {
		return nil, ErrTOTPAlreadyEnabled
	}

	secret, err := totp.GenerateSecret(s.AppName, userEmail)
	if err != nil {
		return nil, err
	}

	enc, err := s.encryptTOTPSecret(userID, secret.Base32)
	if err != nil {
		return nil, err
	}
	if err := s.Users.SetTOTPSecret(ctx, userID, enc); err != nil {
		return nil, err
	}

	return &EnrollmentStart{Secret: secret.Base32, OTPAuth: secret.OTPAuth}, nil
}

// ConfirmEnrollment verifies the first code from the new authenticator
// and, only on success, flips TOTPEnabledAt and issues recovery codes.
func (s *Service) ConfirmEnrollment(ctx context.Context, userID, code string) (recoveryCodes []string, err error) {
	user, err := s.Users.ByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.TOTPEnabled() {
		return nil, ErrTOTPAlreadyEnabled
	}

	secret, err := s.decryptTOTPSecret(user)
	if err != nil {
		return nil, err
	}

	step, ok, err := totp.Verify(code, secret, 0, nowUTC())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrTOTPInvalid
	}

	if _, err := s.Users.UpdateTOTPLastStep(ctx, userID, step); err != nil {
		return nil, err
	}
	if err := s.Users.EnableTOTP(ctx, userID); err != nil {
		return nil, err
	}

	plaintext, hashes, err := totp.GenerateRecoveryCodes()
	if err != nil {
		return nil, err
	}
	if err := s.RecoveryCodes.ReplaceAll(ctx, userID, hashes); err != nil {
		return nil, err
	}

	return plaintext, nil
}

// DisableTOTP requires the current password as re-authentication —
// removing a second factor is exactly the kind of action a hijacked
// session shouldn't be able to do on password alone... but since the
// session already passed 2FA to exist, requiring the password here
// specifically guards against a stolen/left-open browser session.
func (s *Service) DisableTOTP(ctx context.Context, userID, currentPassword string) error {
	user, err := s.Users.ByID(ctx, userID)
	if err != nil {
		return err
	}
	if !user.TOTPEnabled() {
		return ErrTOTPNotEnabled
	}

	ok, err := s.Hasher.Verify(ctx, currentPassword, user.PasswordHash)
	if err != nil {
		return err
	}
	if !ok {
		return ErrInvalidCredentials
	}

	return s.Users.DisableTOTP(ctx, userID)
}
