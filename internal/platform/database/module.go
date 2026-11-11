package database

import (
	"context"

	"go.uber.org/fx"
	"gorm.io/gorm"

	"poster.app/studio/internal/platform/config"
)

var Module = fx.Module(
	"database",
	fx.Provide(Provide),
)

func Provide(lc fx.Lifecycle, cfg *config.Config) (*gorm.DB, error) {
	if err := ensureDataDir(cfg); err != nil {
		return nil, err
	}
	db, err := Open(cfg)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDB.Close()
		},
	})
	return db, nil
}

func ensureDataDir(cfg *config.Config) error {
	return mkdir(cfg.App.DataDir)
}
