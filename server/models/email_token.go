package models

import "time"

type EmailTokenPurpose string

const (
	EmailTokenVerify EmailTokenPurpose = "verify"
	EmailTokenReset  EmailTokenPurpose = "reset"
)

type EmailToken struct {
	ID         string
	UserID     string
	Purpose    EmailTokenPurpose
	TokenHash  string
	ExpiresAt  time.Time
	ConsumedAt *time.Time
	CreatedAt  time.Time
}

func (t *EmailToken) Valid(now time.Time) bool {
	return t.ConsumedAt == nil && now.Before(t.ExpiresAt)
}
