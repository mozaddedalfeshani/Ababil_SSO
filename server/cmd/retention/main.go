// Command retention runs one pass of scheduled data cleanup and
// exits. Intended to be invoked by cron / systemd timer / Kubernetes
// CronJob — not a long-running process, so a stuck run doesn't
// silently stop future ones the way an in-process ticker could.
//
// Currently: purges audit_logs rows older than AUDIT_RETENTION_DAYS
// (default 90) — see docs/architecture.md "Privacy".
package main

import (
	"context"
	"log/slog"
	"os"

	"ababilx-sso/config"
	"ababilx-sso/db"
)

func main() {
	ctx := context.Background()

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

	audit := db.NewAuditRepo(pool)
	purged, err := audit.Purge(ctx, cfg.AuditRetentionDays)
	if err != nil {
		slog.Error("audit purge failed", "error", err)
		os.Exit(1)
	}

	slog.Info("retention pass complete", "audit_rows_purged", purged, "retention_days", cfg.AuditRetentionDays)
}
