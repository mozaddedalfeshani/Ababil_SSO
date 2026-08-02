package handlers

import (
	"net/http"

	"ababilx-sso/middleware"
	"ababilx-sso/services/audit"

	"github.com/gin-gonic/gin"
)

// Logout revokes the current session row (not just the cookie) so a
// copy of the cookie taken before logout can't still be used.
func (h *AuthHandler) Logout(c *gin.Context) {
	ctx := c.Request.Context()

	if raw, err := c.Cookie(h.Cookies.SessionCookieName()); err == nil && raw != "" {
		if session, err := h.Identity.ValidateSessionToken(ctx, raw); err == nil {
			_ = h.Identity.Sessions.Revoke(ctx, session.ID)
			ipHash, ua := clientContext(ctx, c, h.IPHasher)
			_ = h.Identity.Audit.Record(ctx, identityAuditEvent(session.UserID, audit.EventSessionRevoked, ipHash, ua, map[string]any{"reason": "logout"}))
		}
	}

	middleware.ClearSessionCookie(c, h.Cookies)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
