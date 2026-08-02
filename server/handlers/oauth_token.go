package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"ababilx-sso/services/oauth"
	"ababilx-sso/services/ratelimit"

	"github.com/gin-gonic/gin"
)

type OAuthTokenHandler struct {
	OAuth     *oauth.Service
	RateLimit *ratelimit.Limiter
}

// oauthError writes the RFC-shaped envelope every /oauth/* endpoint
// uses — deliberately different from handlers/errors.go's
// {"error":{"code","message"}} envelope, so a client can't confuse
// which API surface (protocol vs. dashboard) it's talking to.
func oauthError(c *gin.Context, status int, code, description string) {
	c.AbortWithStatusJSON(status, gin.H{"error": code, "error_description": description})
}

func (h *OAuthTokenHandler) Token(c *gin.Context) {
	ctx := c.Request.Context()

	if err := c.Request.ParseForm(); err != nil {
		oauthError(c, http.StatusBadRequest, "invalid_request", "malformed form body")
		return
	}
	form := c.Request.PostForm
	grantType := form.Get("grant_type")

	clientID, clientSecret := extractClientCredentials(c, form)
	if err := h.RateLimit.Allow(ctx, "oauth_token", clientID+":"+c.ClientIP(), 30, time.Minute, ratelimit.FailClosed); err != nil {
		oauthError(c, http.StatusTooManyRequests, "slow_down", "too many requests")
		return
	}

	client, err := h.OAuth.AuthenticateClient(ctx, clientID, clientSecret)
	if err != nil {
		oauthError(c, http.StatusUnauthorized, "invalid_client", "client authentication failed")
		return
	}

	var resp *oauth.TokenResponse
	switch grantType {
	case "authorization_code":
		resp, err = h.OAuth.ExchangeAuthorizationCode(ctx, client, form.Get("code"), form.Get("redirect_uri"), form.Get("code_verifier"))
	case "refresh_token":
		resp, err = h.OAuth.RefreshGrant(ctx, client, form.Get("refresh_token"), strings.Fields(form.Get("scope")))
	case "client_credentials":
		resp, err = h.OAuth.ClientCredentialsGrant(ctx, client, form.Get("scope"))
	default:
		oauthError(c, http.StatusBadRequest, "unsupported_grant_type", "grant_type must be authorization_code, refresh_token, or client_credentials")
		return
	}

	if err != nil {
		h.respondGrantError(c, err)
		return
	}

	body := gin.H{
		"access_token": resp.AccessToken,
		"token_type":   resp.TokenType,
		"expires_in":   resp.ExpiresIn,
		"scope":        resp.Scope,
	}
	if resp.RefreshToken != "" {
		body["refresh_token"] = resp.RefreshToken
	}
	if resp.IDToken != "" {
		body["id_token"] = resp.IDToken
	}
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.JSON(http.StatusOK, body)
}

func (h *OAuthTokenHandler) respondGrantError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, oauth.ErrInvalidGrant):
		oauthError(c, http.StatusBadRequest, "invalid_grant", "the grant is invalid, expired, or already used")
	case errors.Is(err, oauth.ErrInvalidScope):
		oauthError(c, http.StatusBadRequest, "invalid_scope", "requested scope exceeds what was originally granted")
	case errors.Is(err, oauth.ErrUnauthorizedClient):
		oauthError(c, http.StatusBadRequest, "unauthorized_client", "client is not authorized for this grant type")
	default:
		respondInternalError(c, err)
	}
}

// extractClientCredentials supports client_secret_basic (Authorization
// header) and client_secret_post (form body) — RFC 6749 §2.3.1. A
// public client sends only client_id, no secret.
func extractClientCredentials(c *gin.Context, form map[string][]string) (clientID, clientSecret string) {
	if user, pass, ok := c.Request.BasicAuth(); ok {
		return user, pass
	}
	get := func(k string) string {
		if v, ok := form[k]; ok && len(v) > 0 {
			return v[0]
		}
		return ""
	}
	return get("client_id"), get("client_secret")
}
