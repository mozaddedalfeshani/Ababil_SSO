package middleware

import (
	"net/http"

	"ababilx-sso/db"
	"ababilx-sso/services/identity"

	"github.com/gin-gonic/gin"
)

const contextUserIDKey = "sso_user_id"
const contextSessionIDKey = "sso_session_id"

// RequireSession validates the session cookie against Postgres on
// every request — Next's proxy.ts may do an optimistic redirect for
// UX, but this is the actual authorization boundary. No session state
// is trusted from anywhere except this lookup.
func RequireSession(svc *identity.Service, policy CookiePolicy) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, err := c.Cookie(policy.SessionCookieName())
		if err != nil || raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "unauthenticated", "message": "no active session"},
			})
			return
		}

		session, err := svc.ValidateSessionToken(c.Request.Context(), raw)
		if err != nil {
			status := http.StatusInternalServerError
			if err == db.ErrNotFound {
				status = http.StatusUnauthorized
			}
			c.AbortWithStatusJSON(status, gin.H{
				"error": gin.H{"code": "unauthenticated", "message": "session invalid or expired"},
			})
			return
		}

		c.Set(contextUserIDKey, session.UserID)
		c.Set(contextSessionIDKey, session.ID)
		c.Next()
	}
}

func UserID(c *gin.Context) string {
	v, _ := c.Get(contextUserIDKey)
	s, _ := v.(string)
	return s
}

func SessionID(c *gin.Context) string {
	v, _ := c.Get(contextSessionIDKey)
	s, _ := v.(string)
	return s
}
