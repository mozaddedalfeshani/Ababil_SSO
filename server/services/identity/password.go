package identity

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ababilx-sso/db"
	"ababilx-sso/models"
	"ababilx-sso/services/crypto"
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

	if err := s.EmailTokens.InvalidateAllForUser(ctx, user.ID, models.EmailTokenReset); err != nil {
		return err
	}

	rawToken, err := crypto.RandomToken(32)
	if err != nil {
		return err
	}
	if _, err := s.EmailTokens.Create(ctx, user.ID, models.EmailTokenReset, crypto.HashToken(rawToken), time.Now().UTC().Add(s.Lifetimes.ResetTokenTTL)); err != nil {
		return err
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.AppBaseURL, rawToken)
	subject, body := mail.ResetPasswordMessage(resetURL)
	return s.Mailer.Send(ctx, user.Email, subject, body)
}

// ResetPassword redeems a reset token, sets the new password, and —
// same reasoning as ChangePassword — revokes every existing session
// and every outstanding reset token for the account.
func (s *Service) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	token, err := s.EmailTokens.ByTokenHash(ctx, crypto.HashToken(rawToken))
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return ErrInvalidToken
		}
		return err
	}
	if token.Purpose != models.EmailTokenReset {
		return ErrInvalidToken
	}

	consumed, err := s.EmailTokens.Consume(ctx, token.ID)
	if err != nil {
		return err
	}
	if !consumed {
		return ErrInvalidToken
	}

	newHash, err := s.Hasher.Hash(ctx, newPassword)
	if err != nil {
		return err
	}
	if err := s.Users.UpdatePassword(ctx, token.UserID, newHash); err != nil {
		return err
	}
	if err := s.EmailTokens.InvalidateAllForUser(ctx, token.UserID, models.EmailTokenReset); err != nil {
		return err
	}

	return s.Sessions.RevokeAllForUser(ctx, token.UserID)
}
