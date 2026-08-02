package handlers

import (
	"errors"
	"net/http"
	"time"

	"ababilx-sso/middleware"
	"ababilx-sso/services/audit"
	"ababilx-sso/services/identity"
	"ababilx-sso/services/ratelimit"

	"github.com/gin-gonic/gin"
)

type verifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	ctx := c.Request.Context()

	var req verifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "token is required")
		return
	}

	if err := h.Identity.VerifyEmail(ctx, req.Token); err != nil {
		if errors.Is(err, identity.ErrInvalidToken) {
			respondError(c, http.StatusBadRequest, "invalid_token", "verification link is invalid or expired")
			return
		}
		respondInternalError(c, err)
		return
	}

	ipHash, ua := clientContext(ctx, c, h.IPHasher)
	_ = h.Identity.Audit.Record(ctx, identityAuditEvent("", audit.EventEmailVerified, ipHash, ua, nil))

	c.JSON(http.StatusOK, gin.H{"message": "email verified"})
}

// ResendVerification requires an authenticated session — you must
// already be logged in to ask for another link, which avoids using
// this endpoint as an email-enumeration or spam-relay oracle.
func (h *AuthHandler) ResendVerification(c *gin.Context) {
	ctx := c.Request.Context()

	if err := h.RateLimit.Allow(ctx, rlBucketResendVerify, middleware.UserID(c), 3, time.Hour, ratelimit.FailClosed); err != nil {
		respondError(c, http.StatusTooManyRequests, "rate_limited", "too many resend requests, try again later")
		return
	}

	if err := h.Identity.ResendVerification(ctx, middleware.UserID(c)); err != nil {
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "verification email sent"})
}
