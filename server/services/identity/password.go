package identity

import (
	"context"
	"errors"

	"ababilx-sso/db"
	"ababilx-sso/models"
	"ababilx-sso/services/mail"
)

// ChangePassword requires the current password and revokes every
// other session — a password change is a signal the old credential
// may have leaked, so every session it could have created should die,
// keeping only the one making this request alive at the caller's
// discretion (handler re-mints if it wants to keep the current one).
func (s *Service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	user, err := s.Users.ByID(ctx, userID)
	if err != nil {
		return err
	}

	ok, err := s.Hasher.Verify(ctx, currentPassword, user.PasswordHash)
	if err != nil {
		return err
	}
	if !ok {
		return ErrInvalidCredentials
	}

	newHash, err := s.Hasher.Hash(ctx, newPassword)
	if err != nil {
		return err
	}
	if err := s.Users.UpdatePassword(ctx, userID, newHash); err != nil {
		return err
	}

	return s.Sessions.RevokeAllForUser(ctx, userID)
}

// RequestPasswordReset always returns nil on a well-formed request
// regardless of whether the email exists — existence must not be
// observable from this endpoint's response.
func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	normalized := NormalizeEmail(email)
	user, err := s.Users.ByEmail(ctx, normalized)
	if errors.Is(err, db.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	otp, err := s.issueEmailOTP(ctx, user.ID, models.EmailTokenReset, s.Lifetimes.ResetTokenTTL)
	if err != nil {
		return err
	}
	subject, body := mail.ResetPasswordMessage(otp)
	return s.Mailer.Send(ctx, user.Email, subject, body)
}

// ResetPassword redeems a reset OTP, sets the new password, and —
// same reasoning as ChangePassword — revokes every existing session
// and every outstanding reset token for the account.
func (s *Service) ResetPassword(ctx context.Context, email, otp, newPassword string) error {
	userID, err := s.consumeEmailOTP(ctx, email, otp, models.EmailTokenReset)
	if err != nil {
		return err
	}

	newHash, err := s.Hasher.Hash(ctx, newPassword)
	if err != nil {
		return err
	}
	if err := s.Users.UpdatePassword(ctx, userID, newHash); err != nil {
		return err
	}
	if err := s.EmailTokens.InvalidateAllForUser(ctx, userID, models.EmailTokenReset); err != nil {
		return err
	}

	return s.Sessions.RevokeAllForUser(ctx, userID)
}
