package handlers

import (
	"errors"
	"net/http"

	"ababilx-sso/middleware"
	"ababilx-sso/services/audit"
	"ababilx-sso/services/identity"

	"github.com/gin-gonic/gin"
)

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// ChangePassword revokes every session for the account (see
// identity.ChangePassword) then re-mints one for the request that
// just authenticated with the new password, so the caller isn't
// immediately logged out by their own password change.
func (h *AccountHandler) ChangePassword(c *gin.Context) {
	ctx := c.Request.Context()
	userID := middleware.UserID(c)

	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "current_password and new_password (min 8 chars) are required")
		return
	}

	// Capture the current session's AMR before it's revoked below, so
	// the re-minted session doesn't silently downgrade from
	// {"pwd","otp"} to {"pwd"} for a 2FA-enrolled account.
	amr := []string{"pwd"}
	if current, err := h.Identity.Sessions.ByID(ctx, middleware.SessionID(c)); err == nil {
		amr = current.AMR
	}

	if err := h.Identity.ChangePassword(ctx, userID, req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, identity.ErrInvalidCredentials) {
			respondError(c, http.StatusUnauthorized, "invalid_credentials", "current password is incorrect")
			return
		}
		respondInternalError(c, err)
		return
	}

	ipHash, ua := clientContext(ctx, c, h.IPHasher)
	rawToken, _, err := h.Identity.MintSession(ctx, userID, amr, ipHash, ua)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	middleware.SetSessionCookie(c, h.Cookies, rawToken, int(h.Identity.Lifetimes.SessionAbsolute.Seconds()))
	if err := middleware.IssueCSRFCookie(c, h.Cookies); err != nil {
		respondInternalError(c, err)
		return
	}

	_ = h.Identity.Audit.Record(ctx, identityAuditEvent(userID, audit.EventPasswordChanged, ipHash, ua, nil))
	c.JSON(http.StatusOK, gin.H{"message": "password changed"})
}
