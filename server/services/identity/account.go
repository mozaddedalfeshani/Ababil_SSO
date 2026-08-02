package identity

import (
	"context"

	"ababilx-sso/models"
)

type ExportedAccount struct {
	User           *models.User
	ActiveSessions []*models.Session
	AuditEvents    []map[string]any
}

// ExportAccount is the user-facing data-export endpoint's data
// source — everything the account holds about itself, in one place,
// so the export can't quietly miss a table added later without this
// function being updated alongside it.
func (s *Service) ExportAccount(ctx context.Context, userID string) (*ExportedAccount, error) {
	user, err := s.Users.ByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	sessions, err := s.Sessions.ListActive(ctx, userID)
	if err != nil {
		return nil, err
	}
	auditEvents, err := s.Audit.ListForUser(ctx, userID, 500)
	if err != nil {
		return nil, err
	}

	return &ExportedAccount{User: user, ActiveSessions: sessions, AuditEvents: auditEvents}, nil
}

// DeleteAccount anonymizes the audit trail (kept for security
// investigation; see docs/architecture.md) then deletes the user row.
// Sessions, consents, refresh tokens, recovery codes, and email
// tokens cascade via ON DELETE CASCADE — no separate cleanup needed.
func (s *Service) DeleteAccount(ctx context.Context, userID string) error {
	if err := s.Audit.Anonymize(ctx, userID); err != nil {
		return err
	}
	return s.Users.Delete(ctx, userID)
}
