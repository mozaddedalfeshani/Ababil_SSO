// Package routes only wires route groups to handlers. No handler logic
// or SQL belongs here.
package routes

import (
	"ababilx-sso/handlers"
	"ababilx-sso/middleware"
	"ababilx-sso/services/identity"

	"github.com/gin-gonic/gin"
)

type Deps struct {
	Health           *handlers.HealthHandler
	Auth             *handlers.AuthHandler
	Account          *handlers.AccountHandler
	OAuthMeta        *handlers.OAuthMetaHandler
	OAuthAuthorize   *handlers.OAuthAuthorizeHandler
	OAuthToken       *handlers.OAuthTokenHandler
	OAuthUserInfo    *handlers.OAuthUserInfoHandler
	OAuthRevoke      *handlers.OAuthRevokeHandler
	OAuthIntrospect  *handlers.OAuthIntrospectHandler
	OAuthLogout      *handlers.OAuthLogoutHandler
	AuthorizeConsent *handlers.AuthorizeConsentHandler
	Org              *handlers.OrgHandler
	Client           *handlers.ClientHandler
	Identity         *identity.Service
	Cookies          middleware.CookiePolicy
}

func Register(r *gin.Engine, deps Deps) {
	r.GET("/healthz", deps.Health.Healthz)
	r.GET("/readyz", deps.Health.Readyz)

	registerOAuthProtocolRoutes(r, deps)

	api := r.Group("/api")
	registerAuthRoutes(api, deps)
	registerAccountRoutes(api, deps)
	registerAuthorizeConsentRoutes(api, deps)
	registerOrgAndClientRoutes(api, deps)
}

// registerOAuthProtocolRoutes covers everything at the RFC-defined
// paths — no /api prefix, no session-cookie auth (client-authenticated
// or public per endpoint instead).
func registerOAuthProtocolRoutes(r *gin.Engine, deps Deps) {
	r.GET("/.well-known/openid-configuration", deps.OAuthMeta.Discovery)
	r.GET("/.well-known/oauth-authorization-server", deps.OAuthMeta.Discovery)
	r.GET("/oauth/jwks.json", deps.OAuthMeta.JWKS)
	r.GET("/oauth/authorize", deps.OAuthAuthorize.Authorize)
	r.POST("/oauth/token", deps.OAuthToken.Token)
	r.GET("/oauth/userinfo", deps.OAuthUserInfo.UserInfo)
	r.POST("/oauth/revoke", deps.OAuthRevoke.Revoke)
	r.POST("/oauth/introspect", deps.OAuthIntrospect.Introspect)
	r.GET("/oauth/logout", deps.OAuthLogout.Logout)
}

func registerAuthRoutes(api *gin.RouterGroup, deps Deps) {
	auth := api.Group("/auth")
	{
		auth.POST("/register", deps.Auth.Register)
		auth.POST("/login", deps.Auth.Login)
		auth.POST("/login/totp", deps.Auth.LoginTOTP)
		auth.POST("/verify-email", deps.Auth.VerifyEmail)
		auth.POST("/forgot-password", deps.Auth.ForgotPassword)
		auth.POST("/reset-password", deps.Auth.ResetPassword)
	}

	// Logout is intentionally CSRF-gated but not RequireSession-gated:
	// it must stay idempotent (200) for an already-expired or
	// already-cleared cookie, while still refusing a cross-site POST
	// that merely carries the cookie along.
	authCSRF := api.Group("/auth")
	authCSRF.Use(middleware.RequireCSRF(deps.Cookies))
	{
		authCSRF.POST("/logout", deps.Auth.Logout)
	}

	authed := api.Group("/auth")
	authed.Use(middleware.RequireSession(deps.Identity, deps.Cookies))
	{
		authed.POST("/resend-verification", deps.Auth.ResendVerification)
	}
}

func registerAccountRoutes(api *gin.RouterGroup, deps Deps) {
	authed := api.Group("")
	authed.Use(middleware.RequireSession(deps.Identity, deps.Cookies))
	{
		authed.GET("/me", deps.Account.Me)
		authed.GET("/me/export", deps.Account.ExportAccount)
		authed.GET("/me/sessions", deps.Account.ListSessions)
		authed.GET("/me/recovery-codes", deps.Account.RecoveryCodesStatus)
	}

	authedWrite := api.Group("")
	authedWrite.Use(middleware.RequireSession(deps.Identity, deps.Cookies))
	authedWrite.Use(middleware.RequireCSRF(deps.Cookies))
	{
		authedWrite.POST("/me/password", deps.Account.ChangePassword)
		authedWrite.POST("/me/totp/enroll", deps.Account.EnrollTOTP)
		authedWrite.POST("/me/totp/verify", deps.Account.ConfirmTOTP)
		authedWrite.POST("/me/totp/disable", deps.Account.DisableTOTP)
		authedWrite.DELETE("/me/sessions/:id", deps.Account.RevokeSession)
		authedWrite.DELETE("/me", deps.Account.DeleteAccount)
	}
}

// registerAuthorizeConsentRoutes are the Next-facing endpoints behind
// the /oauth/authorize and /oauth/logout redirects — the consent
// screen and logout confirmation read/act through these.
func registerAuthorizeConsentRoutes(api *gin.RouterGroup, deps Deps) {
	api.GET("/auth-request/:id", deps.AuthorizeConsent.AuthRequestInfo)

	authed := api.Group("")
	authed.Use(middleware.RequireSession(deps.Identity, deps.Cookies))
	authed.Use(middleware.RequireCSRF(deps.Cookies))
	{
		authed.POST("/authorize/consent", deps.AuthorizeConsent.Consent)
		authed.POST("/oauth/logout/confirm", deps.AuthorizeConsent.LogoutConfirm)
	}
}

func registerOrgAndClientRoutes(api *gin.RouterGroup, deps Deps) {
	reads := api.Group("")
	reads.Use(middleware.RequireSession(deps.Identity, deps.Cookies))
	{
		reads.GET("/orgs", deps.Org.List)
		reads.GET("/orgs/:id", deps.Org.Get)
		reads.GET("/orgs/:id/members", deps.Org.ListMembers)
		reads.GET("/orgs/:id/clients", deps.Client.List)
		reads.GET("/clients/:clientId", deps.Client.Get)
	}

	writes := api.Group("")
	writes.Use(middleware.RequireSession(deps.Identity, deps.Cookies))
	writes.Use(middleware.RequireCSRF(deps.Cookies))
	{
		writes.POST("/orgs", deps.Org.Create)
		writes.POST("/orgs/:id/members", deps.Org.AddMember)
		writes.DELETE("/orgs/:id/members/:userId", deps.Org.RemoveMember)

		writes.POST("/orgs/:id/clients", deps.Client.Create)
		writes.POST("/clients/:clientId/rotate-secret", deps.Client.RotateSecret)
		writes.DELETE("/clients/:clientId", deps.Client.Disable)
	}
}
