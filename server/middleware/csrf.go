package middleware

import (
	"net/http"

	"ababilx-sso/services/crypto"

	"github.com/gin-gonic/gin"
)

// RequireCSRF implements double-submit: the CSRF cookie is readable
// JS-side (not HttpOnly) so the frontend can echo its value back in a
// header; a cross-site request can trigger the cookie's presence but
// cannot read it to also set the header, which is what defeats CSRF
// here. This only guards state-changing requests — SameSite=Lax on
// the session cookie already blocks it being sent cross-site on a
// plain top-level GET navigation, but OAuth redirects need Lax rather
// than Strict, so this is the second layer for POST/PATCH/DELETE.
func RequireCSRF(policy CookiePolicy) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookieVal, err := c.Cookie(policy.CSRFCookieName())
		if err != nil || cookieVal == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{"code": "csrf_missing", "message": "missing csrf cookie"},
			})
			return
		}

		headerVal := c.GetHeader(CSRFHeaderName)
		if headerVal == "" || !crypto.EqualTokenHash(cookieVal, headerVal) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{"code": "csrf_mismatch", "message": "csrf token mismatch"},
			})
			return
		}

		c.Next()
	}
}

// IssueCSRFCookie sets a fresh CSRF token if one isn't already
// present. Called on any response that might precede a form submit
// (login page load, session start) so the frontend always has a
// current token to echo back.
func IssueCSRFCookie(c *gin.Context, policy CookiePolicy) error {
	if existing, err := c.Cookie(policy.CSRFCookieName()); err == nil && existing != "" {
		return nil
	}
	token, err := crypto.RandomToken(32)
	if err != nil {
		return err
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(policy.CSRFCookieName(), token, 0, "/", policy.Domain, policy.Secure, false)
	return nil
}
