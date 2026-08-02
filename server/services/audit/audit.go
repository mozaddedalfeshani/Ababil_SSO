// Package audit records security-relevant events. It never logs
// tokens, codes, secrets, or raw emails — only hashes and IDs.
package audit

import (
	"context"

	"ababilx-sso/db"
)

type Service struct {
	repo *db.AuditRepo
}

func NewService(repo *db.AuditRepo) *Service { return &Service{repo: repo} }

type Event struct {
	ActorUserID string // empty if not attributable to a user
	Event       string
	IPHash      string
	UserAgent   string
	Meta        map[string]any
}

// Record best-effort logs an event. Audit writes never block or fail
// the request they describe — a missed audit row is far cheaper than
// a broken login — but the call is still made synchronously so
// callers can see write errors in their own logs during development.
func (s *Service) Record(ctx context.Context, e Event) error {
	var actorUserID, ipHash, userAgent *string
	if e.ActorUserID != "" {
		actorUserID = &e.ActorUserID
	}
	if e.IPHash != "" {
		ipHash = &e.IPHash
	}
	if e.UserAgent != "" {
		userAgent = &e.UserAgent
	}

	return s.repo.Write(ctx, db.WriteAuditParams{
		ActorUserID: actorUserID,
		Event:       e.Event,
		IPHash:      ipHash,
		UserAgent:   userAgent,
		Meta:        e.Meta,
	})
}

// Anonymize strips the user linkage from that user's audit rows on
// account deletion — the trail survives (append-only, useful for
// security investigation) but stops naming the deleted account.
func (s *Service) Anonymize(ctx context.Context, userID string) error {
	return s.repo.AnonymizeForUser(ctx, userID)
}

// ListForUser powers the /api/me/export endpoint.
func (s *Service) ListForUser(ctx context.Context, userID string, limit int) ([]map[string]any, error) {
	return s.repo.ListForUser(ctx, userID, limit)
}

// Event name constants keep call sites from typo-diverging on strings
// that later get grepped/alerted on.
const (
	EventRegistered           = "user_registered"
	EventEmailVerified        = "email_verified"
	EventLoginSuccess         = "login_success"
	EventLoginFailed          = "login_failed"
	EventLoginLockedOut       = "login_locked_out"
	EventTOTPEnabled          = "totp_enabled"
	EventTOTPDisabled         = "totp_disabled"
	EventTOTPFailed           = "totp_failed"
	EventTOTPReplayRejected   = "totp_replay_rejected"
	EventRecoveryCodeUsed     = "recovery_code_used"
	EventPasswordChanged      = "password_changed"
	EventPasswordResetRequest = "password_reset_requested"
	EventPasswordResetDone    = "password_reset_completed"
	EventSessionRevoked       = "session_revoked"
	EventAccountDeleted       = "account_deleted"
	EventCodeReplay           = "code_replay"
	EventRefreshReuseDetected = "refresh_reuse_detected"
)
