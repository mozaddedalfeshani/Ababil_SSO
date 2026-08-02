package handlers

import (
	"net/http"

	"ababilx-sso/middleware"
	"ababilx-sso/services/audit"

	"github.com/gin-gonic/gin"
)

func (h *AccountHandler) ListSessions(c *gin.Context) {
	ctx := c.Request.Context()
	sessions, err := h.Identity.Sessions.ListActive(ctx, middleware.UserID(c))
	if err != nil {
		respondInternalError(c, err)
		return
	}

	currentID := middleware.SessionID(c)
	out := make([]gin.H, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, gin.H{
			"id":         s.ID,
			"user_agent": s.UserAgent,
			"amr":        s.AMR,
			"auth_time":  s.AuthTime,
			"created_at": s.CreatedAt,
			"is_current": s.ID == currentID,
		})
	}
	c.JSON(http.StatusOK, gin.H{"sessions": out})
}

// RevokeSession lets a user sign out a specific device/browser
// remotely. It only ever revokes a session that belongs to the
// authenticated user — ownership is checked before revocation so one
// account can't revoke another's session by guessing an ID.
func (h *AccountHandler) RevokeSession(c *gin.Context) {
	ctx := c.Request.Context()
	userID := middleware.UserID(c)
	targetID := c.Param("id")

	session, err := h.Identity.Sessions.ByID(ctx, targetID)
	if err != nil || session.UserID != userID {
		respondError(c, http.StatusNotFound, "not_found", "session not found")
		return
	}

	if err := h.Identity.Sessions.Revoke(ctx, targetID); err != nil {
		respondInternalError(c, err)
		return
	}

	ipHash, ua := clientContext(ctx, c, h.IPHasher)
	_ = h.Identity.Audit.Record(ctx, identityAuditEvent(userID, audit.EventSessionRevoked, ipHash, ua, map[string]any{"session_id": targetID, "reason": "user_revoked"}))

	// Revoking the current session should also clear its cookie.
	if targetID == middleware.SessionID(c) {
		middleware.ClearSessionCookie(c, h.Cookies)
	}

	c.JSON(http.StatusOK, gin.H{"message": "session revoked"})
}
