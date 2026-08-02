// Package orgs implements organization and OAuth client management —
// the developer-facing half of the platform (as opposed to
// services/oauth, which is the protocol runtime). Authorization checks
// (who may manage a client) live here, not in handlers.
package orgs

import (
	"context"
	"errors"

	"ababilx-sso/db"
	"ababilx-sso/models"
)

var (
	ErrForbidden = errors.New("insufficient permissions for this organization")
)

type Service struct {
	Orgs    *db.OrgsRepo
	Clients *db.ClientsRepo
}

// RequireRole errors unless the user is a member with at least the
// given role (owner > admin > member). Every mutating org/client
// action goes through this — membership alone is not authorization.
func (s *Service) RequireRole(ctx context.Context, orgID, userID string, minRole models.OrgRole) error {
	role, err := s.Orgs.MemberRole(ctx, orgID, userID)
	if errors.Is(err, db.ErrNotFound) {
		return ErrForbidden
	}
	if err != nil {
		return err
	}
	if !roleAtLeast(role, minRole) {
		return ErrForbidden
	}
	return nil
}

func roleAtLeast(have, want models.OrgRole) bool {
	rank := map[models.OrgRole]int{
		models.OrgRoleMember: 1,
		models.OrgRoleAdmin:  2,
		models.OrgRoleOwner:  3,
	}
	return rank[have] >= rank[want]
}
