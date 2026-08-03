package identity

import (
	"context"
	"errors"
	"fmt"

	"ababilx-sso/db"
	"ababilx-sso/models"
	"ababilx-sso/services/mail"
)

// Register creates a user and sends a verification OTP. It does not
// mint a session — the caller (handler) decides whether to log the
// user in immediately (allowed: email verification gates OAuth
// consent and org/client ownership, not login itself).
func (s *Service) Register(ctx context.Context, email, password string) (*models.User, error) {
	normalized := NormalizeEmail(email)

	if _, err := s.Users.ByEmail(ctx, normalized); err == nil {
		return nil, ErrEmailTaken
	} else if !errors.Is(err, db.ErrNotFound) {
		return nil, err
	}

	hash, err := s.Hasher.Hash(ctx, password)
	if err != nil {
		return nil, err
	}

	user, err := s.Users.Create(ctx, normalized, hash)
	if err != nil {
		return nil, err
	}

	if err := s.sendVerificationEmail(ctx, user); err != nil {
		// Registration still succeeded; the user can request a resend.
		return user, fmt.Errorf("account created but verification email failed to send: %w", err)
	}
	return user, nil
}

// RequestVerificationResend always returns nil on a well-formed
// request regardless of whether the email exists or is already
// verified — existence/verification must not be observable from the
// response (same posture as RequestPasswordReset).
func (s *Service) RequestVerificationResend(ctx context.Context, email string) error {
	normalized := NormalizeEmail(email)
	user, err := s.Users.ByEmail(ctx, normalized)
	if errors.Is(err, db.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if user.EmailVerified() {
		return nil
	}
	return s.sendVerificationEmail(ctx, user)
}

// ResendVerification sends another OTP for an already-authenticated
// user. Prefer RequestVerificationResend for unauthenticated flows.
func (s *Service) ResendVerification(ctx context.Context, userID string) error {
	user, err := s.Users.ByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.EmailVerified() {
		return nil
	}
	return s.sendVerificationEmail(ctx, user)
}

func (s *Service) sendVerificationEmail(ctx context.Context, user *models.User) error {
	otp, err := s.issueEmailOTP(ctx, user.ID, models.EmailTokenVerify, s.Lifetimes.EmailTokenTTL)
	if err != nil {
		return err
	}
	subject, body := mail.VerifyEmailMessage(otp)
	return s.Mailer.Send(ctx, user.Email, subject, body)
}

// VerifyEmail redeems a verification OTP. Errors are not distinguished
// (wrong vs expired vs unknown email) to avoid leaking which case
// applies.
func (s *Service) VerifyEmail(ctx context.Context, email, otp string) error {
	userID, err := s.consumeEmailOTP(ctx, email, otp, models.EmailTokenVerify)
	if err != nil {
		return err
	}
	return s.Users.MarkEmailVerified(ctx, userID)
}
