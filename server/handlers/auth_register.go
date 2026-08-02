package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"ababilx-sso/services/audit"
	"ababilx-sso/services/identity"
	"ababilx-sso/services/ratelimit"

	"github.com/gin-gonic/gin"
)

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	ctx := c.Request.Context()

	if err := h.RateLimit.Allow(ctx, rlBucketRegister, c.ClientIP(), 10, time.Hour, ratelimit.FailClosed); err != nil {
		respondError(c, http.StatusTooManyRequests, "rate_limited", "too many registration attempts, try again later")
		return
	}

	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "email and password (min 8 chars) are required")
		return
	}

	user, err := h.Identity.Register(ctx, req.Email, req.Password)
	if errors.Is(err, identity.ErrEmailTaken) {
		// Same response as success — registering with a taken email
		// must not reveal that the account exists.
		c.JSON(http.StatusAccepted, gin.H{"message": "if that email is available, check your inbox to verify it"})
		return
	}
	if err != nil && user == nil {
		respondInternalError(c, err)
		return
	}
	if err != nil {
		// Account created; verification email failed to send. Logged,
		// not surfaced — the response must not distinguish this case
		// from the ordinary success path, and the user can request a
		// resend once they notice nothing arrived.
		slog.Warn("verification_email_send_failed", "error", err)
	}

	ipHash, ua := clientContext(ctx, c, h.IPHasher)
	_ = h.Identity.Audit.Record(ctx, identityAuditEvent(user.ID, audit.EventRegistered, ipHash, ua, nil))

	c.JSON(http.StatusAccepted, gin.H{"message": "if that email is available, check your inbox to verify it"})
}
