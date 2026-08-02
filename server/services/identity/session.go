package identity

import (
	"context"
	"time"

	"ababilx-sso/db"
	"ababilx-sso/models"
	"ababilx-sso/services/crypto"
)

// MintSession issues a brand-new session (fresh random token, fresh
// row) rather than ever reusing an existing token — this is the
// session-fixation defense: login, TOTP completion, and password
// change each call this, so a cookie value that existed before
// authentication can never become the authenticated session.
func (s *Service) MintSession(ctx context.Context, userID string, amr []string, ipHash, userAgent *string) (rawToken string, session *models.Session, err error) {
	rawToken, err = crypto.RandomToken(32)
	if err != nil {
		return "", nil, err
	}

	now := time.Now().UTC()
	session, err = s.Sessions.Create(ctx, db.CreateSessionParams{
		UserID:            userID,
		TokenHash:         crypto.HashToken(rawToken),
		IPHash:            ipHash,
		UserAgent:         userAgent,
		AMR:               amr,
		AuthTime:          now,
		IdleExpiresAt:     now.Add(s.Lifetimes.SessionIdle),
		AbsoluteExpiresAt: now.Add(s.Lifetimes.SessionAbsolute),
	})
	if err != nil {
		return "", nil, err
	}
	return rawToken, session, nil
}

// ValidateSessionToken looks up an active session by its raw cookie
// value and slides the idle-expiry window forward. Returns
// db.ErrNotFound for anything invalid/expired/revoked — callers treat
// that uniformly as "not logged in", never distinguishing the reason
// (avoids leaking session-enumeration signal).
func (s *Service) ValidateSessionToken(ctx context.Context, rawToken string) (*models.Session, error) {
	session, err := s.Sessions.ByTokenHash(ctx, crypto.HashToken(rawToken))
	if err != nil {
		return nil, err
	}
	if !session.Active(time.Now().UTC()) {
		return nil, db.ErrNotFound
	}

	newIdle := time.Now().UTC().Add(s.Lifetimes.SessionIdle)
	_ = s.Sessions.ExtendIdle(ctx, session.ID, newIdle) // best-effort slide; failure doesn't invalidate the request
	return session, nil
}
