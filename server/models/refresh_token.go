package models

import "time"

type RefreshToken struct {
	ID           string
	TokenHash    string
	FamilyID     string
	ClientID     string
	UserID       string
	SessionID    *string
	Scope        string
	SessionBound bool // false = offline_access; survives session logout
	ExpiresAt    time.Time
	ConsumedAt   *time.Time
	RotatedTo    *string
	RevokedAt    *time.Time
	RevokeReason *string
	CreatedAt    time.Time
}

func (r *RefreshToken) Active(now time.Time) bool {
	return r.RevokedAt == nil && r.ConsumedAt == nil && now.Before(r.ExpiresAt)
}
