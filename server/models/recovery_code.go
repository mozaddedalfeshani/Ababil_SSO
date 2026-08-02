package models

import "time"

type RecoveryCode struct {
	ID         string
	UserID     string
	CodeHash   string
	ConsumedAt *time.Time
	CreatedAt  time.Time
}
