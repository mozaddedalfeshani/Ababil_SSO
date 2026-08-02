package handlers

import (
	"errors"
	"net/http"
	"time"

	"ababilx-sso/services/audit"
	"ababilx-sso/services/identity"
	"ababilx-sso/services/ratelimit"

	"github.com/gin-gonic/gin"
)

type loginTOTPRequest struct {
	MFAPendingID string `json:"mfa_pending_id" binding:"required"`
	Code         string `json:"code" binding:"required"`
	RecoveryCode bool   `json:"recovery_code"`
}

func (h *AuthHandler) LoginTOTP(c *gin.Context) {
	ctx := c.Request.Context()

	if err := h.RateLimit.Allow(ctx, rlBucketTOTPVerify, c.ClientIP(), 10, time.Minute, ratelimit.FailClosed); err != nil {
		respondError(c, http.StatusTooManyRequests, "rate_limited", "too many attempts, try again later")
		return
	}

	var req loginTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "mfa_pending_id and code are required")
		return
	}

	ipHash, ua := clientContext(ctx, c, h.IPHasher)

	var result *identity.LoginResult
	var err error
	if req.RecoveryCode {
		result, err = h.Identity.CompleteRecoveryCodeLogin(ctx, req.MFAPendingID, req.Code, ipHash, ua)
	} else {
		result, err = h.Identity.CompleteTOTPLogin(ctx, req.MFAPendingID, req.Code, ipHash, ua)
	}

	if err != nil {
		switch {
		case errors.Is(err, identity.ErrMFAPendingNotFound):
			respondError(c, http.StatusUnauthorized, "mfa_expired", "verification session expired, please log in again")
		case errors.Is(err, identity.ErrTOTPInvalid):
			respondError(c, http.StatusUnauthorized, "invalid_code", "invalid code")
		default:
			respondInternalError(c, err)
		}
		return
	}

	event := audit.EventLoginSuccess
	if req.RecoveryCode {
		event = audit.EventRecoveryCodeUsed
	}
	h.completeLoginResponse(c, result, event)
}
