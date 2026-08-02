package handlers

import (
	"net/http"

	"ababilx-sso/services/oauth"

	"github.com/gin-gonic/gin"
)

type OAuthIntrospectHandler struct {
	OAuth *oauth.Service
}

// Introspect requires client authentication — RFC 7662 §2.1 makes this
// the caller's own credential, which is also what scopes the response
// to the caller's own tokens (see oauth.Service.Introspect). Public
// clients (no secret) cannot call this endpoint at all.
func (h *OAuthIntrospectHandler) Introspect(c *gin.Context) {
	ctx := c.Request.Context()
	if err := c.Request.ParseForm(); err != nil {
		oauthError(c, http.StatusBadRequest, "invalid_request", "malformed form body")
		return
	}
	form := c.Request.PostForm

	clientID, clientSecret := extractClientCredentials(c, form)
	client, err := h.OAuth.AuthenticateClient(ctx, clientID, clientSecret)
	if err != nil || !client.IsConfidential() {
		oauthError(c, http.StatusUnauthorized, "invalid_client", "client authentication failed")
		return
	}

	result := h.OAuth.Introspect(ctx, client, form.Get("token"))
	c.JSON(http.StatusOK, result)
}
