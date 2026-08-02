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

// Register creates a user and sends a verification email. It does not
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
	// Invalidate prior outstanding verify tokens so only the newest
	// link works — otherwise an old, forwarded email stays valid.
	if err := s.EmailTokens.InvalidateAllForUser(ctx, user.ID, models.EmailTokenVerify); err != nil {
		return err
	}

	rawToken, err := crypto.RandomToken(32)
	if err != nil {
		return err
	}
	if _, err := s.EmailTokens.Create(ctx, user.ID, models.EmailTokenVerify, crypto.HashToken(rawToken), time.Now().UTC().Add(s.Lifetimes.EmailTokenTTL)); err != nil {
		return err
	}

	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.AppBaseURL, rawToken)
	subject, body := mail.VerifyEmailMessage(verifyURL)
	return s.Mailer.Send(ctx, user.Email, subject, body)
}

// VerifyEmail redeems a verification token. Errors are not
// distinguished (expired vs. already-used vs. unknown) to avoid
// leaking which case applies to a guessed token.
func (s *Service) VerifyEmail(ctx context.Context, rawToken string) error {
	token, err := s.EmailTokens.ByTokenHash(ctx, crypto.HashToken(rawToken))
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return ErrInvalidToken
		}
		return err
	}
	if token.Purpose != models.EmailTokenVerify {
		return ErrInvalidToken
	}

	consumed, err := s.EmailTokens.Consume(ctx, token.ID)
	if err != nil {
		return err
	}
	if !consumed {
		return ErrInvalidToken
	}

	return s.Users.MarkEmailVerified(ctx, token.UserID)
}
