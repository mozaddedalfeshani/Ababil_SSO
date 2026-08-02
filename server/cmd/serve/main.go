// Command serve runs the Ababil SSO API. It never applies migrations
// itself — run `migrate` first (see cmd/migrate). Boot order: config,
// database, Redis, services, router, then serve until a shutdown
// signal drains in-flight requests gracefully.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ababilx-sso/cache"
	"ababilx-sso/config"
	"ababilx-sso/db"
	"ababilx-sso/handlers"
	"ababilx-sso/middleware"
	"ababilx-sso/routes"
	"ababilx-sso/services/audit"
	"ababilx-sso/services/crypto"
	"ababilx-sso/services/identity"
	"ababilx-sso/services/keys"
	"ababilx-sso/services/mail"
	"ababilx-sso/services/oauth"
	"ababilx-sso/services/orgs"
	"ababilx-sso/services/privacy"
	"ababilx-sso/services/ratelimit"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	redisClient, err := cache.NewClient(ctx, cfg.RedisURL)
	if err != nil {
		slog.Error("redis connection failed", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	mailer, err := newMailer(cfg)
	if err != nil {
		slog.Error("mailer setup failed", "error", err)
		os.Exit(1)
	}

	cookies := middleware.CookiePolicy{Domain: cfg.CookieDomain, Secure: cfg.CookieSecure}

	identitySvc := &identity.Service{
		Users:            db.NewUsersRepo(pool),
		Sessions:         db.NewSessionsRepo(pool),
		EmailTokens:      db.NewEmailTokensRepo(pool),
		RecoveryCodes:    db.NewRecoveryCodesRepo(pool),
		Redis:            redisClient,
		Hasher:           crypto.NewPasswordHasher(crypto.DefaultArgon2Params(), crypto.NewHashSemaphore(16)),
		Mailer:           mailer,
		Audit:            audit.NewService(db.NewAuditRepo(pool)),
		IPHasher:         privacy.NewIPHasher(redisClient),
		Lockout:          ratelimit.NewLockout(redisClient),
		Issuer:           cfg.Issuer,
		AppName:          cfg.AppName,
		AppBaseURL:       cfg.AppBaseURL,
		KeyEncryptionKey: cfg.KeyEncryptionKey,
		Lifetimes:        identity.DefaultLifetimes(),
	}

	auditSvc := identitySvc.Audit
	signingKeys := keys.NewManager(pool, db.NewSigningKeysRepo(pool), cfg.KeyEncryptionKey)
	if err := signingKeys.EnsureActiveKey(ctx); err != nil {
		slog.Error("signing key setup failed", "error", err)
		os.Exit(1)
	}

	oauthSvc := &oauth.Service{
		Clients:          db.NewClientsRepo(pool),
		ClientSubjects:   db.NewClientSubjectsRepo(pool),
		Consents:         db.NewConsentsRepo(pool),
		Codes:            db.NewAuthorizationCodesRepo(pool),
		RefreshTokens:    db.NewRefreshTokensRepo(pool),
		Sessions:         identitySvc.Sessions,
		Users:            identitySvc.Users,
		Redis:            redisClient,
		Keys:             signingKeys,
		Audit:            auditSvc,
		IPHasher:         identitySvc.IPHasher,
		Issuer:           cfg.Issuer,
		KeyEncryptionKey: cfg.KeyEncryptionKey,
		Lifetimes:        oauth.DefaultLifetimes(),
	}

	orgsSvc := &orgs.Service{
		Orgs:    db.NewOrgsRepo(pool),
		Clients: oauthSvc.Clients,
	}

	engine, err := routes.NewEngine(cfg.TrustedProxyCIDRs, cfg.IsProduction())
	if err != nil {
		slog.Error("router setup failed", "error", err)
		os.Exit(1)
	}

	routes.Register(engine, routes.Deps{
		Health: &handlers.HealthHandler{DB: pool, Redis: redisClient},
		Auth: &handlers.AuthHandler{
			Identity:  identitySvc,
			RateLimit: ratelimit.NewLimiter(redisClient),
			IPHasher:  identitySvc.IPHasher,
			Redis:     redisClient,
			Cookies:   cookies,
		},
		Account: &handlers.AccountHandler{
			Identity: identitySvc,
			IPHasher: identitySvc.IPHasher,
			Cookies:  cookies,
		},
		OAuthMeta:       &handlers.OAuthMetaHandler{OAuth: oauthSvc, Issuer: cfg.Issuer},
		OAuthAuthorize:  &handlers.OAuthAuthorizeHandler{OAuth: oauthSvc, AppBaseURL: cfg.AppBaseURL},
		OAuthToken:      &handlers.OAuthTokenHandler{OAuth: oauthSvc, RateLimit: ratelimit.NewLimiter(redisClient)},
		OAuthUserInfo:   &handlers.OAuthUserInfoHandler{OAuth: oauthSvc},
		OAuthRevoke:     &handlers.OAuthRevokeHandler{OAuth: oauthSvc},
		OAuthIntrospect: &handlers.OAuthIntrospectHandler{OAuth: oauthSvc},
		OAuthLogout:     &handlers.OAuthLogoutHandler{OAuth: oauthSvc, AppBaseURL: cfg.AppBaseURL},
		AuthorizeConsent: &handlers.AuthorizeConsentHandler{
			Identity: identitySvc,
			OAuth:    oauthSvc,
			Cookies:  cookies,
		},
		Org:      &handlers.OrgHandler{Orgs: orgsSvc},
		Client:   &handlers.ClientHandler{Orgs: orgsSvc},
		Identity: identitySvc,
		Cookies:  cookies,
	})

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: engine,
	}

	go func() {
		slog.Info("listening", "addr", cfg.Addr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
}

// newMailer picks SMTP when configured, otherwise the console mailer —
// but never console in production, where a misconfigured SMTP_HOST
// must fail loudly at boot rather than silently logging password-reset
// links that nobody but the server operator ever sees.
func newMailer(cfg *config.Config) (mail.Mailer, error) {
	if cfg.SMTPConfigured() {
		return mail.NewSMTPMailer(mail.SMTPConfig{
			Host:     cfg.SMTP.Host,
			Port:     cfg.SMTP.Port,
			Username: cfg.SMTP.Username,
			Password: cfg.SMTP.Password,
			From:     cfg.SMTP.From,
		})
	}
	if cfg.IsProduction() {
		return nil, errors.New("SMTP_HOST is required in production (APP_ENV=production)")
	}
	slog.Warn("SMTP not configured — using console mailer; verify/reset links will be logged, not emailed")
	return mail.ConsoleMailer{}, nil
}
