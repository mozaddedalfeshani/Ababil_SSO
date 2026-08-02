// Command migrate applies pending schema migrations and exits. It is a
// separate binary from `serve` on purpose: a bad migration must fail a
// deliberate, observable step, not crash-loop the running fleet.
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

	if err := db.RunMigrations(ctx, pool); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}

	slog.Info("migrations applied")
}
