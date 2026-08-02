package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetSessionCookie is the only place the session cookie is written.
// HttpOnly + SameSite=Lax + Secure: Lax (not Strict) is required
// because the OAuth authorize redirect is a top-level GET that must
// carry the cookie for the "already logged in" fast path to work.
func SetSessionCookie(c *gin.Context, policy CookiePolicy, rawToken string, maxAgeSeconds int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(policy.SessionCookieName(), rawToken, maxAgeSeconds, "/", policy.Domain, policy.Secure, true)
}

func ClearSessionCookie(c *gin.Context, policy CookiePolicy) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(policy.SessionCookieName(), "", -1, "/", policy.Domain, policy.Secure, true)
}
