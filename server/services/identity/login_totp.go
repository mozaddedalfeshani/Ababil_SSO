package identity

import (
	"context"

	"ababilx-sso/services/crypto"
	"ababilx-sso/services/totp"
)

// CompleteTOTPLogin consumes a pending-MFA challenge with a TOTP code
// and, only on success, mints the real session. This is the only
// place mfa_pending:<id> is ever read — see docs/architecture.md.
func (s *Service) CompleteTOTPLogin(ctx context.Context, pendingID, code string, ipHash, userAgent *string) (*LoginResult, error) {
	userID, err := s.mfaPendingUserID(ctx, pendingID)
	if err != nil {
		return nil, err
	}

	user, err := s.Users.ByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !user.TOTPEnabled() {
		return nil, ErrTOTPNotEnabled
	}

	secret, err := s.decryptTOTPSecret(user)
	if err != nil {
		return nil, err
	}

	var lastStep int64
	if user.TOTPLastStep != nil {
		lastStep = *user.TOTPLastStep
	}

	step, ok, err := totp.Verify(code, secret, lastStep, nowUTC())
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrTOTPInvalid
	}

	advanced, err := s.Users.UpdateTOTPLastStep(ctx, user.ID, step)
	if err != nil {
		return nil, err
	}
	if !advanced {
		// Someone else redeemed this step first — treat as replay.
		return nil, ErrTOTPInvalid
	}

	s.clearMFAPending(ctx, pendingID)

	rawToken, session, err := s.MintSession(ctx, user.ID, []string{"pwd", "otp"}, ipHash, userAgent)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Complete: true, RawSessionToken: rawToken, Session: session}, nil
}

// CompleteRecoveryCodeLogin is the TOTP-lost-device fallback: consumes
// one single-use recovery code instead of a TOTP code.
func (s *Service) CompleteRecoveryCodeLogin(ctx context.Context, pendingID, code string, ipHash, userAgent *string) (*LoginResult, error) {
	userID, err := s.mfaPendingUserID(ctx, pendingID)
	if err != nil {
		return nil, err
	}

	consumed, err := s.RecoveryCodes.Consume(ctx, userID, crypto.HashToken(code))
	if err != nil {
		return nil, err
	}
	if !consumed {
		return nil, ErrTOTPInvalid
	}

	s.clearMFAPending(ctx, pendingID)

	rawToken, session, err := s.MintSession(ctx, userID, []string{"pwd", "recovery_code"}, ipHash, userAgent)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Complete: true, RawSessionToken: rawToken, Session: session}, nil
}
