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

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login is the password step. On success it either sets the session
// cookie (no 2FA configured) or returns an mfa_pending id the client
// must complete via /api/auth/login/totp — see
// docs/architecture.md "Session / MFA state machine" for why password
// success never mints a session by itself when TOTP is enabled.
func (h *AuthHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	if err := h.RateLimit.Allow(ctx, rlBucketLogin, c.ClientIP(), 20, time.Minute, ratelimit.FailClosed); err != nil {
		respondError(c, http.StatusTooManyRequests, "rate_limited", "too many attempts, try again later")
		return
	}

	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "email and password are required")
		return
	}

	ipHash, ua := clientContext(ctx, c, h.IPHasher)

	result, err := h.Identity.LoginPassword(ctx, req.Email, req.Password, ipHash, ua)
	if err != nil {
		h.respondLoginError(c, err)
		return
	}

	if !result.Complete {
		c.JSON(http.StatusOK, gin.H{"mfa_required": true, "mfa_pending_id": result.MFAPendingID})
		return
	}

	h.completeLoginResponse(c, result, audit.EventLoginSuccess)
}

func (h *AuthHandler) respondLoginError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, identity.ErrAccountLocked):
		respondError(c, http.StatusTooManyRequests, "account_locked", "too many failed attempts; try again later")
	case errors.Is(err, identity.ErrInvalidCredentials):
		respondError(c, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
	default:
		respondInternalError(c, err)
	}
}

// completeLoginResponse sets the session + CSRF cookies and returns
// the profile. Shared by the password-only path and the TOTP
// completion path in auth_login_totp.go.
func (h *AuthHandler) completeLoginResponse(c *gin.Context, result *identity.LoginResult, auditEvent string) {
	maxAge := int(h.Identity.Lifetimes.SessionAbsolute.Seconds())
	middleware.SetSessionCookie(c, h.Cookies, result.RawSessionToken, maxAge)
	if err := middleware.IssueCSRFCookie(c, h.Cookies); err != nil {
		respondInternalError(c, err)
		return
	}

	ipHash, ua := clientContext(c.Request.Context(), c, h.IPHasher)
	_ = h.Identity.Audit.Record(c.Request.Context(), identityAuditEvent(result.Session.UserID, auditEvent, ipHash, ua, nil))

	c.JSON(http.StatusOK, gin.H{"user_id": result.Session.UserID})
}
