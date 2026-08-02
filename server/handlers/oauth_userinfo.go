package handlers

import (
	"errors"
	"net/http"
	"strings"

	"ababilx-sso/services/oauth"

	"github.com/gin-gonic/gin"
)

type OAuthUserInfoHandler struct {
	OAuth *oauth.Service
}

func (h *OAuthUserInfoHandler) UserInfo(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		c.Header("WWW-Authenticate", `Bearer error="invalid_token"`)
		oauthError(c, http.StatusUnauthorized, "invalid_token", "missing bearer access token")
		return
	}
	accessToken := strings.TrimPrefix(auth, "Bearer ")

	claims, err := h.OAuth.UserInfo(c.Request.Context(), accessToken)
	if err != nil {
		status := http.StatusUnauthorized
		code := "invalid_token"
		if errors.Is(err, oauth.ErrInsufficientScope) {
			code = "insufficient_scope"
			status = http.StatusForbidden
		}
		c.Header("WWW-Authenticate", `Bearer error="`+code+`"`)
		oauthError(c, status, code, "")
		return
	}

	c.JSON(http.StatusOK, claims)
}
