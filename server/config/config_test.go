package config

import (
	"encoding/base64"
	"testing"
)

func setRequiredEnv(t *testing.T) {
	t.Helper()
	kek := base64.StdEncoding.EncodeToString(make([]byte, 32))
	env := map[string]string{
		"DATABASE_URL":       "postgres://localhost/test",
		"REDIS_URL":          "redis://localhost:6379/0",
		"SSO_ISSUER":         "https://auth.example.com",
		"KEY_ENCRYPTION_KEY": kek,
	}
	for k, v := range env {
		t.Setenv(k, v)
	}
}

func TestLoadFailsFastOnMissingSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("SSO_ISSUER", "https://auth.example.com")
	// Explicitly cleared, not just "never set" — a developer running
	// tests in a shell that already sourced .env would otherwise
	// inherit a real KEY_ENCRYPTION_KEY and silently pass this test
	// for the wrong reason.
	t.Setenv("KEY_ENCRYPTION_KEY", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected Load to fail without KEY_ENCRYPTION_KEY")
	}
}

func TestLoadRejectsWrongLengthKey(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("KEY_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(make([]byte, 16))) // wrong length

	if _, err := Load(); err == nil {
		t.Fatal("expected Load to reject a non-32-byte key")
	}
}

func TestLoadSucceedsWithDefaults(t *testing.T) {
	setRequiredEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Env != "development" {
		t.Errorf("expected default env=development, got %q", cfg.Env)
	}
	if cfg.CookieSecure {
		t.Error("expected CookieSecure to default false outside production")
	}
	if cfg.SMTPConfigured() {
		t.Error("expected SMTP to be unconfigured by default")
	}
}

func TestLoadProductionDefaultsCookieSecureTrue(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("APP_ENV", "production")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.CookieSecure {
		t.Error("expected CookieSecure to default true in production")
	}
}
