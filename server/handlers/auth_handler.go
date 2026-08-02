package handlers

import (
	"ababilx-sso/middleware"
	"ababilx-sso/services/identity"
	"ababilx-sso/services/privacy"
	"ababilx-sso/services/ratelimit"

	"github.com/redis/go-redis/v9"
)

// AuthHandler covers registration, login (both factors), logout, and
// email/password recovery. All business logic lives in
// services/identity — this struct only parses requests, applies rate
// limits, and shapes responses.
type AuthHandler struct {
	Identity  *identity.Service
	RateLimit *ratelimit.Limiter
	IPHasher  *privacy.IPHasher
	Redis     *redis.Client
	Cookies   middleware.CookiePolicy
}

const (
	rlBucketLogin          = "login"
	rlBucketRegister       = "register"
	rlBucketForgotPassword = "forgot_password"
	rlBucketTOTPVerify     = "totp_verify"
	rlBucketResendVerify   = "resend_verify"
)
