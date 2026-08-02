package models

import "time"

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
)

type User struct {
	ID              string
	Email           string
	EmailVerifiedAt *time.Time
	PasswordHash    string
	TOTPSecretEnc   []byte
	TOTPLastStep    *int64
	TOTPEnabledAt   *time.Time
	Status          UserStatus
	CreatedAt       time.Time
	LastLoginAt     *time.Time
}

func (u *User) EmailVerified() bool { return u.EmailVerifiedAt != nil }
func (u *User) TOTPEnabled() bool   { return u.TOTPEnabledAt != nil }
