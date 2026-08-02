// Package identity orchestrates registration, login, TOTP, and
// password management. It is the only layer that touches multiple
// repos/services for a single user-facing action — handlers stay thin
// and call into here; this package never imports Gin.
package identity

import (
	"time"

	"ababilx-sso/db"
	"ababilx-sso/services/audit"
	"ababilx-sso/services/crypto"
	"ababilx-sso/services/mail"
	"ababilx-sso/services/privacy"
	"ababilx-sso/services/ratelimit"

	"github.com/redis/go-redis/v9"
)

type Lifetimes struct {
	SessionIdle     time.Duration
	SessionAbsolute time.Duration
	EmailTokenTTL   time.Duration
	ResetTokenTTL   time.Duration
	MFAPendingTTL   time.Duration
}

func DefaultLifetimes() Lifetimes {
	return Lifetimes{
		SessionIdle:     14 * 24 * time.Hour,
		SessionAbsolute: 30 * 24 * time.Hour,
		EmailTokenTTL:   24 * time.Hour,
		ResetTokenTTL:   1 * time.Hour,
		MFAPendingTTL:   5 * time.Minute,
	}
}

type Service struct {
	Users            *db.UsersRepo
	Sessions         *db.SessionsRepo
	EmailTokens      *db.EmailTokensRepo
	RecoveryCodes    *db.RecoveryCodesRepo
	Redis            *redis.Client
	Hasher           *crypto.PasswordHasher
	Mailer           mail.Mailer
	Audit            *audit.Service
	IPHasher         *privacy.IPHasher
	Lockout          *ratelimit.Lockout
	Issuer           string // OIDC issuer (protocol `iss`, e.g. https://auth.example.com)
	AppName          string // human-readable name shown in authenticator apps, e.g. "Ababil SSO"
	AppBaseURL       string // for building verify/reset links
	KeyEncryptionKey []byte // seals TOTP secrets at rest (AES-256-GCM)
	Lifetimes        Lifetimes
}
