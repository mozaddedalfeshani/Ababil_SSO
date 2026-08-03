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

type verifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required"`
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	ctx := c.Request.Context()

	var req verifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "email and otp are required")
		return
	}

	emailKey := crypto.HashToken(identity.NormalizeEmail(req.Email))
	if err := h.RateLimit.Allow(ctx, "verify_email_otp", emailKey, 20, time.Hour, ratelimit.FailClosed); err != nil {
		respondError(c, http.StatusTooManyRequests, "rate_limited", "too many attempts, try again later")
		return
	}

	if err := h.Identity.VerifyEmail(ctx, req.Email, req.OTP); err != nil {
		if errors.Is(err, identity.ErrInvalidToken) {
			respondError(c, http.StatusBadRequest, "invalid_token", "verification code is invalid or expired")
			return
		}
		respondInternalError(c, err)
		return
	}

	ipHash, ua := clientContext(ctx, c, h.IPHasher)
	_ = h.Identity.Audit.Record(ctx, identityAuditEvent("", audit.EventEmailVerified, ipHash, ua, nil))

	c.JSON(http.StatusOK, gin.H{"message": "email verified"})
}

type resendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResendVerification is public (email in body). Always returns the
// same message — see identity.RequestVerificationResend. Rate-limited
// to 3/hour per normalized email to avoid spam-relay abuse without
// requiring a session (register success screen has no cookie yet).
func (h *AuthHandler) ResendVerification(c *gin.Context) {
	ctx := c.Request.Context()

	var req resendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "email is required")
		return
	}

	emailKey := crypto.HashToken(identity.NormalizeEmail(req.Email))
	if err := h.RateLimit.Allow(ctx, rlBucketResendVerify, emailKey, 3, time.Hour, ratelimit.FailClosed); err != nil {
		respondError(c, http.StatusTooManyRequests, "rate_limited", "too many resend requests, try again later")
		return
	}

	if err := h.Identity.RequestVerificationResend(ctx, req.Email); err != nil {
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "if that email is available, a verification code has been sent"})
}
