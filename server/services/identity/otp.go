package identity

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"ababilx-sso/cache"
	"ababilx-sso/db"
	"ababilx-sso/models"
	"ababilx-sso/services/crypto"

	"github.com/redis/go-redis/v9"
)

const (
	emailOTPLength   = 6
	maxEmailOTPFails = 5
)

var digitsOnly = regexp.MustCompile(`^\d{6}$`)

// hashEmailOTP binds the code to the user so two accounts can share the
// same 6-digit value without colliding on email_tokens.token_hash, and
// so presenting an OTP without the matching account is useless.
func hashEmailOTP(userID, otp string) string {
	return crypto.HashToken(userID + ":" + otp)
}

func normalizeOTP(raw string) (string, bool) {
	otp := strings.TrimSpace(raw)
	if !digitsOnly.MatchString(otp) {
		return "", false
	}
	return otp, true
}

func (s *Service) issueEmailOTP(ctx context.Context, userID string, purpose models.EmailTokenPurpose, ttl time.Duration) (otp string, err error) {
	if err := s.EmailTokens.InvalidateAllForUser(ctx, userID, purpose); err != nil {
		return "", err
	}
	otp, err = crypto.RandomDigits(emailOTPLength)
	if err != nil {
		return "", err
	}
	if _, err := s.EmailTokens.Create(ctx, userID, purpose, hashEmailOTP(userID, otp), time.Now().UTC().Add(ttl)); err != nil {
		return "", err
	}
	_ = s.clearOTPFailures(ctx, purpose, userID)
	return otp, nil
}

// consumeEmailOTP looks up the outstanding OTP for email+purpose,
// enforces the 5-attempt cap, and consumes on success. Wrong/missing
// codes all surface as ErrInvalidToken.
func (s *Service) consumeEmailOTP(ctx context.Context, email, rawOTP string, purpose models.EmailTokenPurpose) (userID string, err error) {
	otp, ok := normalizeOTP(rawOTP)
	if !ok {
		return "", ErrInvalidToken
	}

	normalized := NormalizeEmail(email)
	user, err := s.Users.ByEmail(ctx, normalized)
	if errors.Is(err, db.ErrNotFound) {
		return "", ErrInvalidToken
	}
	if err != nil {
		return "", err
	}

	fails, err := s.otpFailureCount(ctx, purpose, user.ID)
	if err != nil {
		return "", err
	}
	if fails >= maxEmailOTPFails {
		return "", ErrInvalidToken
	}

	token, err := s.EmailTokens.ActiveByUserPurpose(ctx, user.ID, purpose)
	if errors.Is(err, db.ErrNotFound) {
		_, _ = s.recordOTPFailure(ctx, purpose, user.ID)
		return "", ErrInvalidToken
	}
	if err != nil {
		return "", err
	}

	if !crypto.EqualTokenHash(token.TokenHash, hashEmailOTP(user.ID, otp)) {
		exhausted, failErr := s.recordOTPFailure(ctx, purpose, user.ID)
		if failErr != nil {
			return "", failErr
		}
		if exhausted {
			_ = s.EmailTokens.InvalidateAllForUser(ctx, user.ID, purpose)
		}
		return "", ErrInvalidToken
	}

	consumed, err := s.EmailTokens.Consume(ctx, token.ID)
	if err != nil {
		return "", err
	}
	if !consumed {
		return "", ErrInvalidToken
	}
	_ = s.clearOTPFailures(ctx, purpose, user.ID)
	return user.ID, nil
}

func (s *Service) otpFailureCount(ctx context.Context, purpose models.EmailTokenPurpose, userID string) (int64, error) {
	n, err := s.Redis.Get(ctx, cache.OTPFailKey(string(purpose), userID)).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	return n, err
}

func (s *Service) recordOTPFailure(ctx context.Context, purpose models.EmailTokenPurpose, userID string) (exhausted bool, err error) {
	key := cache.OTPFailKey(string(purpose), userID)
	count, err := s.Redis.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		ttl := s.Lifetimes.EmailTokenTTL
		if s.Lifetimes.ResetTokenTTL > ttl {
			ttl = s.Lifetimes.ResetTokenTTL
		}
		s.Redis.Expire(ctx, key, ttl)
	}
	return count >= maxEmailOTPFails, nil
}

func (s *Service) clearOTPFailures(ctx context.Context, purpose models.EmailTokenPurpose, userID string) error {
	return s.Redis.Del(ctx, cache.OTPFailKey(string(purpose), userID)).Err()
}
