package handlers

import (
	"net/http"

	"ababilx-sso/middleware"
	"ababilx-sso/services/audit"

	"github.com/gin-gonic/gin"
)

func (h *AccountHandler) ExportAccount(c *gin.Context) {
	ctx := c.Request.Context()

	export, err := h.Identity.ExportAccount(ctx, middleware.UserID(c))
	if err != nil {
		respondInternalError(c, err)
		return
	}

	// Sessions are re-shaped rather than marshaled as-is: models.Session
	// exports TokenHash, which has no business leaving the server even
	// hashed — an export endpoint should return account data, not
	// internal auth-token bookkeeping.
	sessions := make([]gin.H, 0, len(export.ActiveSessions))
	for _, s := range export.ActiveSessions {
		sessions = append(sessions, gin.H{
			"id":         s.ID,
			"user_agent": s.UserAgent,
			"amr":        s.AMR,
			"auth_time":  s.AuthTime,
			"created_at": s.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":             export.User.ID,
			"email":          export.User.Email,
			"email_verified": export.User.EmailVerified(),
			"totp_enabled":   export.User.TOTPEnabled(),
			"created_at":     export.User.CreatedAt,
			"last_login_at":  export.User.LastLoginAt,
		},
		"active_sessions": sessions,
		"audit_events":    export.AuditEvents,
	})
}

type deleteAccountRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
}

// DeleteAccount requires the current password as re-authentication —
// account deletion is irreversible, so a hijacked/left-open session
// alone must not be sufficient to trigger it.
func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	ctx := c.Request.Context()
	userID := middleware.UserID(c)

	var req deleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "current_password is required")
		return
	}

	user, err := h.Identity.Users.ByID(ctx, userID)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	ok, err := h.Identity.Hasher.Verify(ctx, req.CurrentPassword, user.PasswordHash)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	if !ok {
		respondError(c, http.StatusUnauthorized, "invalid_credentials", "current password is incorrect")
		return
	}

	ipHash, ua := clientContext(ctx, c, h.IPHasher)
	_ = h.Identity.Audit.Record(ctx, identityAuditEvent(userID, audit.EventAccountDeleted, ipHash, ua, nil))

	if err := h.Identity.DeleteAccount(ctx, userID); err != nil {
		respondInternalError(c, err)
		return
	}

	middleware.ClearSessionCookie(c, h.Cookies)
	c.JSON(http.StatusOK, gin.H{"message": "account deleted"})
}
