package handlers

import (
	"net/http"

	"ababilx-sso/services/oauth"

	"github.com/gin-gonic/gin"
)

type OAuthMetaHandler struct {
	OAuth  *oauth.Service
	Issuer string
}

// Discovery serves both .well-known/openid-configuration and
// .well-known/oauth-authorization-server — the same document covers
// both discovery specs for the core-only surface this deployment
// supports (see docs/architecture.md "Out of scope for v1").
func (h *OAuthMetaHandler) Discovery(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"issuer":                                h.Issuer,
		"authorization_endpoint":                h.Issuer + "/oauth/authorize",
		"token_endpoint":                        h.Issuer + "/oauth/token",
		"userinfo_endpoint":                     h.Issuer + "/oauth/userinfo",
		"jwks_uri":                              h.Issuer + "/oauth/jwks.json",
		"revocation_endpoint":                   h.Issuer + "/oauth/revoke",
		"introspection_endpoint":                h.Issuer + "/oauth/introspect",
		"end_session_endpoint":                  h.Issuer + "/oauth/logout",
		"response_types_supported":              []string{"code"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token", "client_credentials"},
		"code_challenge_methods_supported":      []string{"S256"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_basic", "client_secret_post", "none"},
		"subject_types_supported":               []string{"pairwise", "public"},
		"id_token_signing_alg_values_supported": []string{"ES256"},
		"scopes_supported":                      []string{"openid", "profile", "email", "offline_access"},
		"claims_supported":                      []string{"sub", "email", "email_verified", "iss", "aud", "exp", "iat", "auth_time"},
	})
}

func (h *OAuthMetaHandler) JWKS(c *gin.Context) {
	set, err := h.OAuth.Keys.JWKS(c.Request.Context())
	if err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, set)
}
