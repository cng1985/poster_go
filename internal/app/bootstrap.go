package app

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"poster.app/studio/internal/modules/studio/template"
	"poster.app/studio/internal/platform/config"
	"poster.app/studio/internal/platform/server"
	"poster.app/studio/migrations"
)

func Bootstrap(db *gorm.DB, cfg *config.Config, logger *slog.Logger, state *server.State) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := migrations.Apply(db.WithContext(ctx)); err != nil {
		return err
	}
	if err := template.Seed(ctx, db); err != nil {
		return err
	}
	state.MarkReady()
	logger.Info("bootstrap complete", cfg.SafeLines()...)
	return nil
}
