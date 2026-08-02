package identity

import (
	"context"
	"errors"

	"ababilx-sso/cache"
	"ababilx-sso/db"
	"ababilx-sso/models"
	"ababilx-sso/services/crypto"
)

// LoginResult tells the handler what to do next: either the login is
// complete (RawSessionToken set) or a second factor is pending
// (MFAPendingID set) and the handler must not set a session cookie.
type LoginResult struct {
	Complete        bool
	RawSessionToken string
	Session         *models.Session
	MFAPendingID    string
}

// LoginPassword verifies credentials and, if the account has TOTP
// enabled, stops short of minting a session — see docs/architecture.md
// "Session / MFA state machine" for why password success alone must
// never be a full session.
func (s *Service) LoginPassword(ctx context.Context, email, password string, ipHash, userAgent *string) (*LoginResult, error) {
	normalized := NormalizeEmail(email)
	emailHash := crypto.HashToken(normalized)

	locked, err := s.Lockout.Locked(ctx, emailHash)
	if err != nil {
		return nil, err
	}
	if locked {
		return nil, ErrAccountLocked
	}

	user, err := s.Users.ByEmail(ctx, normalized)
	if errors.Is(err, db.ErrNotFound) {
		// Dummy-hash verify so timing/shape matches the wrong-password
		// case exactly — no user-enumeration signal either way.
		_, _ = s.Hasher.Verify(ctx, password, crypto.DummyHash)
		_ = s.Lockout.RecordFailure(ctx, emailHash)
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	ok, err := s.Hasher.Verify(ctx, password, user.PasswordHash)
	if err != nil {
		return nil, err
	}
	if !ok {
		_ = s.Lockout.RecordFailure(ctx, emailHash)
		return nil, ErrInvalidCredentials
	}

	_ = s.Lockout.Clear(ctx, emailHash)
	_ = s.Users.UpdateLastLogin(ctx, user.ID)

	if user.TOTPEnabled() {
		pendingID, err := crypto.RandomToken(24)
		if err != nil {
			return nil, err
		}
		payload := user.ID
		if err := s.Redis.Set(ctx, cache.MFAPendingKey(pendingID), payload, s.Lifetimes.MFAPendingTTL).Err(); err != nil {
			return nil, err
		}
		return &LoginResult{Complete: false, MFAPendingID: pendingID}, nil
	}

	rawToken, session, err := s.MintSession(ctx, user.ID, []string{"pwd"}, ipHash, userAgent)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Complete: true, RawSessionToken: rawToken, Session: session}, nil
}

// mfaPendingUserID resolves a pending-MFA challenge to its user ID.
// Kept unexported: only login_totp.go's completion methods should
// consume this key, per the single-purpose design of MFAPendingKey.
func (s *Service) mfaPendingUserID(ctx context.Context, pendingID string) (string, error) {
	userID, err := s.Redis.Get(ctx, cache.MFAPendingKey(pendingID)).Result()
	if err != nil {
		return "", ErrMFAPendingNotFound
	}
	return userID, nil
}

func (s *Service) clearMFAPending(ctx context.Context, pendingID string) {
	s.Redis.Del(ctx, cache.MFAPendingKey(pendingID))
}
