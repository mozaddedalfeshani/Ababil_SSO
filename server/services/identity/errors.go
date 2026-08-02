package identity

import "errors"

var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountLocked      = errors.New("account temporarily locked")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrTOTPRequired       = errors.New("totp code required")
	ErrTOTPInvalid        = errors.New("invalid totp code")
	ErrTOTPAlreadyEnabled = errors.New("totp already enabled")
	ErrTOTPNotEnabled     = errors.New("totp not enabled")
	ErrMFAPendingNotFound = errors.New("mfa challenge not found or expired")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrRateLimited        = errors.New("too many requests")
)
