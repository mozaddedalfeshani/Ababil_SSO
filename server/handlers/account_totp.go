package handlers

import (
	"errors"
	"net/http"

	"ababilx-sso/middleware"
	"ababilx-sso/services/audit"
	"ababilx-sso/services/identity"

	"github.com/gin-gonic/gin"
)

func (h *AccountHandler) EnrollTOTP(c *gin.Context) {
	ctx := c.Request.Context()
	userID := middleware.UserID(c)

	user, err := h.Identity.Users.ByID(ctx, userID)
	if err != nil {
		respondInternalError(c, err)
		return
	}

	enrollment, err := h.Identity.EnrollTOTP(ctx, userID, user.Email)
	if err != nil {
		if errors.Is(err, identity.ErrTOTPAlreadyEnabled) {
			respondError(c, http.StatusConflict, "totp_already_enabled", "two-factor authentication is already enabled")
			return
		}
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"secret": enrollment.Secret, "otpauth_url": enrollment.OTPAuth})
}

type confirmTOTPRequest struct {
	Code string `json:"code" binding:"required"`
}

// ConfirmTOTP proves the authenticator app was actually set up
// correctly before enabling 2FA — see identity.ConfirmEnrollment.
// Recovery codes are returned exactly once here; they cannot be
// re-fetched afterward.
func (h *AccountHandler) ConfirmTOTP(c *gin.Context) {
	ctx := c.Request.Context()
	userID := middleware.UserID(c)

	var req confirmTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "code is required")
		return
	}

	recoveryCodes, err := h.Identity.ConfirmEnrollment(ctx, userID, req.Code)
	if err != nil {
		if errors.Is(err, identity.ErrTOTPInvalid) {
			respondError(c, http.StatusBadRequest, "invalid_code", "invalid code")
			return
		}
		if errors.Is(err, identity.ErrTOTPAlreadyEnabled) {
			respondError(c, http.StatusConflict, "totp_already_enabled", "two-factor authentication is already enabled")
			return
		}
		respondInternalError(c, err)
		return
	}

	ipHash, ua := clientContext(ctx, c, h.IPHasher)
	_ = h.Identity.Audit.Record(ctx, identityAuditEvent(userID, audit.EventTOTPEnabled, ipHash, ua, nil))

	c.JSON(http.StatusOK, gin.H{"recovery_codes": recoveryCodes})
}

type disableTOTPRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
}

func (h *AccountHandler) DisableTOTP(c *gin.Context) {
	ctx := c.Request.Context()
	userID := middleware.UserID(c)

	var req disableTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "current_password is required")
		return
	}

	if err := h.Identity.DisableTOTP(ctx, userID, req.CurrentPassword); err != nil {
		if errors.Is(err, identity.ErrInvalidCredentials) {
			respondError(c, http.StatusUnauthorized, "invalid_credentials", "current password is incorrect")
			return
		}
		if errors.Is(err, identity.ErrTOTPNotEnabled) {
			respondError(c, http.StatusConflict, "totp_not_enabled", "two-factor authentication is not enabled")
			return
		}
		respondInternalError(c, err)
		return
	}

	ipHash, ua := clientContext(ctx, c, h.IPHasher)
	_ = h.Identity.Audit.Record(ctx, identityAuditEvent(userID, audit.EventTOTPDisabled, ipHash, ua, nil))

	c.JSON(http.StatusOK, gin.H{"message": "two-factor authentication disabled"})
}
