package testkit

import (
	"path/filepath"
	"testing"
	"time"

	"gorm.io/gorm"

	"poster.app/studio/internal/platform/config"
	"poster.app/studio/internal/platform/database"
	"poster.app/studio/migrations"
)

func Config(t *testing.T) *config.Config {
	t.Helper()
	dir := t.TempDir()
	return &config.Config{
		App: config.AppConfig{Name: "poster", DataDir: dir},
		Server: config.ServerConfig{
			Host:           "127.0.0.1",
			Port:           8080,
			MaxHeaderBytes: 1 << 20,
			MaxBodyBytes:   2 << 20,
		},
		Database: config.DatabaseConfig{
			Enabled:      true,
			Driver:       "sqlite",
			MaxOpenConns: 1,
			MaxIdleConns: 1,
			BatchSize:    100,
			SQLite:       config.SQLiteConfig{JournalMode: "DELETE"},
		},
		Auth:    config.AuthConfig{Enabled: true, LoginMax: 8},
		Logging: config.LoggingConfig{Level: "error", Format: "text"},
		AI: config.AIConfig{
			Enabled: true, Provider: "openai", BaseURL: "https://example.test/v1",
			Model: "test", FallbackLocal: true, MaxPerMinute: 10,
		},
		Upload: config.UploadConfig{MaxBytes: 2 << 20, Dir: "uploads"},
		Runtime: config.Runtime{
			RequestTimeout:    5 * time.Second,
			ReadHeaderTimeout: time.Second,
			IdleTimeout:       time.Second,
			ShutdownTimeout:   time.Second,
			SessionTTL:        time.Hour,
			LoginWindow:       time.Minute,
			ConnMaxLifetime:   time.Minute,
			SQLiteBusy:        5 * time.Second,
			AITimeout:         time.Second,
			ForeignKeys:       true,
			SQLitePath:        filepath.Join(dir, "app.db"),
			UploadDir:         filepath.Join(dir, "uploads"),
		},
	}
}

func DB(t *testing.T) (*gorm.DB, *config.Config) {
	t.Helper()
	cfg := Config(t)
	db, err := database.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
	if err := migrations.Apply(db); err != nil {
		t.Fatal(err)
	}
	return db, cfg
}
