package handlers

import (
	"errors"
	"net/http"
	"time"

	"ababilx-sso/services/audit"
	"ababilx-sso/services/crypto"
	"ababilx-sso/services/identity"
	"ababilx-sso/services/ratelimit"

	"github.com/gin-gonic/gin"
)

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ForgotPassword always responds the same way regardless of whether
// the email exists — see identity.RequestPasswordReset.
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	ctx := c.Request.Context()

	if err := h.RateLimit.Allow(ctx, rlBucketForgotPassword, c.ClientIP(), 5, time.Hour, ratelimit.FailClosed); err != nil {
		respondError(c, http.StatusTooManyRequests, "rate_limited", "too many requests, try again later")
		return
	}

	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "email is required")
		return
	}

	if err := h.Identity.RequestPasswordReset(ctx, req.Email); err != nil {
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "if that email is registered, a reset code has been sent"})
}

type resetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	OTP         string `json:"otp" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	ctx := c.Request.Context()

	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "email, otp, and new_password (min 8 chars) are required")
		return
	}

	emailKey := crypto.HashToken(identity.NormalizeEmail(req.Email))
	if err := h.RateLimit.Allow(ctx, "reset_password_otp", emailKey, 20, time.Hour, ratelimit.FailClosed); err != nil {
		respondError(c, http.StatusTooManyRequests, "rate_limited", "too many attempts, try again later")
		return
	}

	if err := h.Identity.ResetPassword(ctx, req.Email, req.OTP, req.NewPassword); err != nil {
		if errors.Is(err, identity.ErrInvalidToken) {
			respondError(c, http.StatusBadRequest, "invalid_token", "reset code is invalid or expired")
			return
		}
		respondInternalError(c, err)
		return
	}

	ipHash, ua := clientContext(ctx, c, h.IPHasher)
	_ = h.Identity.Audit.Record(ctx, identityAuditEvent("", audit.EventPasswordResetDone, ipHash, ua, nil))

	c.JSON(http.StatusOK, gin.H{"message": "password reset — all sessions have been signed out"})
}
