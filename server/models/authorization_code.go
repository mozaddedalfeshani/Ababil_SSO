package models

import "time"

type AuthorizationCode struct {
	ID              string
	CodeHash        string
	ClientID        string
	UserID          string
	SessionID       string
	RedirectURI     string
	Scope           string
	Nonce           *string
	CodeChallenge   string
	AuthTime        time.Time
	RefreshFamilyID string
	ExpiresAt       time.Time
	ConsumedAt      *time.Time
	CreatedAt       time.Time
}

func (a *AuthorizationCode) Expired(now time.Time) bool { return now.After(a.ExpiresAt) }
