// Package oauth implements the authorization_code + PKCE, refresh, and
// client_credentials grants, plus userinfo/revoke/introspect/logout.
// This is the actual authorization boundary — handlers parse HTTP and
// delegate here; nothing about token issuance is decided in handlers.
package oauth

import (
	"time"

	"ababilx-sso/db"
	"ababilx-sso/services/audit"
	"ababilx-sso/services/keys"
	"ababilx-sso/services/privacy"

	"github.com/redis/go-redis/v9"
)

type Lifetimes struct {
	AuthRequestTTL  time.Duration
	CodeTTL         time.Duration
	AccessTokenTTL  time.Duration
	RefreshSliding  time.Duration
	RefreshAbsolute time.Duration
}

func DefaultLifetimes() Lifetimes {
	return Lifetimes{
		AuthRequestTTL:  10 * time.Minute,
		CodeTTL:         60 * time.Second,
		AccessTokenTTL:  10 * time.Minute,
		RefreshSliding:  30 * 24 * time.Hour,
		RefreshAbsolute: 90 * 24 * time.Hour,
	}
}

type Service struct {
	Clients        *db.ClientsRepo
	ClientSubjects *db.ClientSubjectsRepo
	Consents       *db.ConsentsRepo
	Codes          *db.AuthorizationCodesRepo
	RefreshTokens  *db.RefreshTokensRepo
	Sessions       *db.SessionsRepo
	Users          *db.UsersRepo
	Redis          *redis.Client
	Keys           *keys.Manager
	Audit          *audit.Service
	IPHasher       *privacy.IPHasher

	Issuer           string
	KeyEncryptionKey []byte
	Lifetimes        Lifetimes
}
