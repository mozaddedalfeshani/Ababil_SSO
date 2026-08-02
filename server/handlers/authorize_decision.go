package handlers

import (
	"net/http"

	"ababilx-sso/middleware"

	"github.com/gin-gonic/gin"
)

type consentDecisionRequest struct {
	RequestID string `json:"req_id" binding:"required"`
	Approved  bool   `json:"approved"`
}

// Consent is mounted behind RequireSession + RequireCSRF (see
// routes.go) — this is the actual authorization decision, re-validated
// server-side regardless of what the consent screen displayed.
// Requesting only the scopes originally parked in the auth request
// (never trusting a scope list from the request body) means there is
// no path for a tampered client-side request to grant more than what
// was already validated against the client's allow-list at
// /oauth/authorize time.
func (h *AuthorizeConsentHandler) Consent(c *gin.Context) {
	ctx := c.Request.Context()
	userID := middleware.UserID(c)
	sessionID := middleware.SessionID(c)

	var body consentDecisionRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "req_id is required")
		return
	}

	req, err := h.OAuth.GetAuthRequest(ctx, body.RequestID)
	if err != nil {
		respondError(c, http.StatusNotFound, "not_found", "authorization request not found or expired")
		return
	}

	if !body.Approved {
		c.JSON(http.StatusOK, gin.H{"redirect_to": h.OAuth.DenyAuthorization(ctx, req)})
		return
	}

	user, err := h.Identity.Users.ByID(ctx, userID)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	if !user.EmailVerified() {
		respondError(c, http.StatusForbidden, "email_not_verified", "verify your email before authorizing an application")
		return
	}

	session, err := h.Identity.Sessions.ByID(ctx, sessionID)
	if err != nil {
		respondInternalError(c, err)
		return
	}

	redirectTo, err := h.OAuth.CompleteAuthorization(ctx, req, userID, sessionID, session.AuthTime, req.Scope)
	if err != nil {
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"redirect_to": redirectTo})
}
