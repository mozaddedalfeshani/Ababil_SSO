// Package config loads and validates process configuration from the
// environment. Loading fails fast: a missing required secret must stop
// boot, never fall back silently.
package config

import (
	"encoding/base64"
	"fmt"
)

type Config struct {
	Env  string // "development" | "production"
	Addr string // HTTP listen address, e.g. ":7897"

	DatabaseURL string
	RedisURL    string

	// Issuer is the OIDC issuer URL (e.g. https://auth.example.com).
	// Used verbatim in discovery, JWKS `iss`, and RFC 9207 `iss` on
	// authorization responses.
	Issuer string

	// AppName is the human-readable name shown in authenticator apps
	// (TOTP QR issuer label) — deliberately not the issuer URL, which
	// renders unreadably inside an authenticator's account label.
	AppName string

	// KeyEncryptionKey seals signing-key private halves and TOTP
	// secrets at rest (AES-256-GCM). Losing it is unrecoverable.
	KeyEncryptionKey []byte

	// TrustedProxyCIDRs are the only hops allowed to set
	// X-Forwarded-For; anything else is attacker-controlled.
	TrustedProxyCIDRs []string

	CookieDomain string // empty for __Host- cookies (recommended)
	CookieSecure bool   // false only for http://localhost dev

	// AppBaseURL is where the Next.js UI lives — verify-email and
	// reset-password links point here, not at the Go API.
	AppBaseURL string

	SMTP SMTPConfig

	// AuditRetentionDays bounds how long audit_logs rows are kept —
	// see docs/architecture.md "Privacy". Purged by cmd/retention, run
	// on a schedule (cron / systemd timer) outside the main process.
	AuditRetentionDays int
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func (c *Config) SMTPConfigured() bool { return c.SMTP.Host != "" }

func Load() (*Config, error) {
	env := optionalEnv("APP_ENV", "development")

	addr := optionalEnv("HTTP_ADDR", ":7897")

	dbURL, err := requireEnv("DATABASE_URL")
	if err != nil {
		return nil, err
	}

	redisURL, err := requireEnv("REDIS_URL")
	if err != nil {
		return nil, err
	}

	issuer, err := requireEnv("SSO_ISSUER")
	if err != nil {
		return nil, err
	}

	kekB64, err := requireEnv("KEY_ENCRYPTION_KEY")
	if err != nil {
		return nil, err
	}
	kek, err := base64.StdEncoding.DecodeString(kekB64)
	if err != nil {
		return nil, fmt.Errorf("KEY_ENCRYPTION_KEY must be base64: %w", err)
	}
	if len(kek) != 32 {
		return nil, fmt.Errorf("KEY_ENCRYPTION_KEY must decode to 32 bytes, got %d", len(kek))
	}

	trustedProxies := splitCSV(optionalEnv("TRUSTED_PROXY_CIDRS", "127.0.0.1/32,::1/128"))
	cookieDomain := optionalEnv("COOKIE_DOMAIN", "")
	cookieSecure, err := optionalEnvBool("COOKIE_SECURE", env == "production")
	if err != nil {
		return nil, err
	}

	appName := optionalEnv("APP_NAME", "Ababil SSO")
	appBaseURL := optionalEnv("NEXT_PUBLIC_APP_URL", issuer)

	smtpPort, err := optionalEnvInt("SMTP_PORT", 587)
	if err != nil {
		return nil, err
	}
	smtp := SMTPConfig{
		Host:     optionalEnv("SMTP_HOST", ""),
		Port:     smtpPort,
		Username: optionalEnv("SMTP_USERNAME", ""),
		Password: optionalEnv("SMTP_PASSWORD", ""),
		From:     optionalEnv("SMTP_FROM", "noreply@localhost"),
	}

	auditRetentionDays, err := optionalEnvInt("AUDIT_RETENTION_DAYS", 90)
	if err != nil {
		return nil, err
	}

	return &Config{
		Env:                env,
		Addr:               addr,
		DatabaseURL:        dbURL,
		RedisURL:           redisURL,
		Issuer:             issuer,
		AppName:            appName,
		KeyEncryptionKey:   kek,
		TrustedProxyCIDRs:  trustedProxies,
		CookieDomain:       cookieDomain,
		CookieSecure:       cookieSecure,
		AppBaseURL:         appBaseURL,
		SMTP:               smtp,
		AuditRetentionDays: auditRetentionDays,
	}, nil
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}
