package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/gatie-io/gatie-server/internal/repository/postgres"
)

func runTokenCleanup(ctx context.Context, repo *postgres.AuthRepository) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := repo.DeleteExpiredRefreshTokens(ctx); err != nil {
				slog.Error("failed to purge expired refresh tokens", "error", err)
			}
		}
	}
}
