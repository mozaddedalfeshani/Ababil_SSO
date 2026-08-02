package models

import "time"

type AuditLog struct {
	ID          string
	ActorUserID *string
	OrgID       *string
	ClientID    *string
	Event       string
	IPHash      *string
	UserAgent   *string
	Meta        map[string]any
	CreatedAt   time.Time
}
