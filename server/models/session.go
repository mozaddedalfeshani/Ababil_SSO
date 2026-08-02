package models

import "time"

type Session struct {
	ID                string
	UserID            string
	TokenHash         string
	IPHash            *string
	UserAgent         *string
	AMR               []string // authentication methods reference, e.g. {"pwd","otp"}
	AuthTime          time.Time
	IdleExpiresAt     time.Time
	AbsoluteExpiresAt time.Time
	RevokedAt         *time.Time
	CreatedAt         time.Time
}

func (s *Session) Active(now time.Time) bool {
	if s.RevokedAt != nil {
		return false
	}
	return now.Before(s.IdleExpiresAt) && now.Before(s.AbsoluteExpiresAt)
}

func (s *Session) HasAMR(method string) bool {
	for _, m := range s.AMR {
		if m == method {
			return true
		}
	}
	return false
}
