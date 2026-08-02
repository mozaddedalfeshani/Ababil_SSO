package handlers

import (
	"ababilx-sso/middleware"
	"ababilx-sso/services/identity"
	"ababilx-sso/services/privacy"
)

// AccountHandler covers the authenticated /api/me/* surface: profile,
// password, TOTP, sessions, and data export/delete. Every route this
// handler serves is mounted behind middleware.RequireSession — see
// routes.go — so middleware.UserID(c) is always populated.
type AccountHandler struct {
	Identity *identity.Service
	IPHasher *privacy.IPHasher
	Cookies  middleware.CookiePolicy
}
