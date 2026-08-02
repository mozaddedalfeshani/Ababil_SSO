package handlers

import (
	"net/http"

	"ababilx-sso/middleware"
	"ababilx-sso/services/identity"
	"ababilx-sso/services/oauth"

	"github.com/gin-gonic/gin"
)

type AuthorizeConsentHandler struct {
	Identity *identity.Service
	OAuth    *oauth.Service
	Cookies  middleware.CookiePolicy
}

// currentSession re-validates the session cookie the same way
// middleware.RequireSession does. This handler can't just use that
// middleware because AuthRequestInfo must also serve anonymous
// visitors — they need to know what they're about to log in for
// before authenticating.
func (h *AuthorizeConsentHandler) currentSession(c *gin.Context) *identity.LoginResult {
	raw, err := c.Cookie(h.Cookies.SessionCookieName())
	if err != nil || raw == "" {
		return nil
	}
	session, err := h.Identity.ValidateSessionToken(c.Request.Context(), raw)
	if err != nil {
		return nil
	}
	return &identity.LoginResult{Complete: true, Session: session}
}

func (h *AuthorizeConsentHandler) AuthRequestInfo(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	req, err := h.OAuth.GetAuthRequest(ctx, id)
	if err != nil {
		respondError(c, http.StatusNotFound, "not_found", "authorization request not found or expired")
		return
	}

	client, err := h.OAuth.Clients.ByID(ctx, req.ClientID)
	if err != nil {
		respondInternalError(c, err)
		return
	}

	resp := gin.H{
		"client": gin.H{
			"name":     client.Name,
			"logo_url": client.LogoURL,
		},
		"scopes":           req.Scope,
		"requires_login":   true,
		"requires_consent": true,
		"email_unverified": false,
	}

	if login := h.currentSession(c); login != nil {
		resp["requires_login"] = false

		needsConsent, err := h.OAuth.NeedsConsent(ctx, req, login.Session.UserID)
		if err != nil {
			respondInternalError(c, err)
			return
		}
		resp["requires_consent"] = needsConsent

		user, err := h.Identity.Users.ByID(ctx, login.Session.UserID)
		if err == nil && !user.EmailVerified() {
			resp["email_unverified"] = true
		}
	}

	c.JSON(http.StatusOK, resp)
}
