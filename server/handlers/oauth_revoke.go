package handlers

import (
	"net/http"

	"ababilx-sso/services/oauth"

	"github.com/gin-gonic/gin"
)

type OAuthRevokeHandler struct {
	OAuth *oauth.Service
}

// Revoke always responds 200 regardless of whether the token existed
// or belonged to this client — RFC 7009's own anti-enumeration
// requirement. See oauth.Service.Revoke.
func (h *OAuthRevokeHandler) Revoke(c *gin.Context) {
	ctx := c.Request.Context()
	if err := c.Request.ParseForm(); err != nil {
		oauthError(c, http.StatusBadRequest, "invalid_request", "malformed form body")
		return
	}
	form := c.Request.PostForm

	clientID, clientSecret := extractClientCredentials(c, form)
	client, err := h.OAuth.AuthenticateClient(ctx, clientID, clientSecret)
	if err != nil {
		oauthError(c, http.StatusUnauthorized, "invalid_client", "client authentication failed")
		return
	}

	h.OAuth.Revoke(ctx, client, form.Get("token"))
	c.Status(http.StatusOK)
}
