package handlers

import (
	"net/http"

	"ababilx-sso/middleware"
	"ababilx-sso/services/oauth"

	"github.com/gin-gonic/gin"
)

type logoutConfirmRequest struct {
	IDTokenHint           string `json:"id_token_hint"`
	PostLogoutRedirectURI string `json:"post_logout_redirect_uri"`
	State                 string `json:"state"`
}

// LogoutConfirm is the interactive-confirmation step RP-initiated
// logout requires — mounted behind RequireSession + RequireCSRF.
// Re-validates the client/redirect binding from scratch rather than
// trusting anything Next passed through from the GET step.
func (h *AuthorizeConsentHandler) LogoutConfirm(c *gin.Context) {
	ctx := c.Request.Context()
	sessionID := middleware.SessionID(c)

	var body logoutConfirmRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "malformed request")
		return
	}

	if _, err := h.OAuth.ResolveLogoutClient(ctx, body.IDTokenHint, body.PostLogoutRedirectURI); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "invalid logout request")
		return
	}

	if err := h.OAuth.CompleteLogout(ctx, sessionID); err != nil {
		respondInternalError(c, err)
		return
	}
	middleware.ClearSessionCookie(c, h.Cookies)

	redirectTo, err := oauth.BuildLogoutRedirect(body.PostLogoutRedirectURI, body.State)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"redirect_to": redirectTo})
}
