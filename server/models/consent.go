package models

import "time"

type Consent struct {
	ID        string
	UserID    string
	ClientID  string
	Scopes    []string
	GrantedAt time.Time
	RevokedAt *time.Time
}

// Satisfies reports whether this consent already covers every
// requested scope — if so, the authorize flow can skip re-prompting
// the user.
func (c *Consent) Satisfies(requested []string) bool {
	if c.RevokedAt != nil {
		return false
	}
	granted := make(map[string]bool, len(c.Scopes))
	for _, s := range c.Scopes {
		granted[s] = true
	}
	for _, s := range requested {
		if !granted[s] {
			return false
		}
	}
	return true
}
