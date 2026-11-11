package database

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	sqlite "github.com/glebarez/sqlite"
	mysqldriver "gorm.io/driver/mysql"
	pgdriver "gorm.io/driver/postgres"
	sqlserverdriver "gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"poster.app/studio/internal/platform/config"
)

func Open(cfg *config.Config) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}
	var dialector gorm.Dialector
	switch cfg.Database.Driver {
	case "sqlite":
		if err := os.MkdirAll(filepath.Dir(cfg.Runtime.SQLitePath), 0o750); err != nil {
			return nil, fmt.Errorf("创建数据目录: %w", err)
		}
		dialector = sqlite.Open(sqliteDSN(cfg))
	case "mysql":
		dialector = mysqldriver.Open(cfg.Database.DSN)
	case "postgres", "pgsql":
		dialector = pgdriver.Open(cfg.Database.DSN)
	case "sqlserver":
		dialector = sqlserverdriver.Open(cfg.Database.DSN)
	default:
		return nil, fmt.Errorf("未知数据库驱动 %q", cfg.Database.Driver)
	}
	db, err := gorm.Open(dialector, gormCfg)
	if err != nil {
		if cfg.Database.Driver == "sqlite" {
			return nil, fmt.Errorf("连接 SQLite 失败: %w", err)
		}
		return nil, errors.New("连接网络数据库失败")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Runtime.ConnMaxLifetime)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		if cfg.Database.Driver == "sqlite" {
			return nil, fmt.Errorf("数据库不可用: %w", err)
		}
		return nil, errors.New("数据库不可用")
	}
	if cfg.Database.Driver == "sqlite" {
		if err := verifySQLite(db, cfg); err != nil {
			_ = sqlDB.Close()
			return nil, err
		}
	}
	return db, nil
}

func sqliteDSN(cfg *config.Config) string {
	fk := "0"
	if cfg.Runtime.ForeignKeys {
		fk = "1"
	}
	return fmt.Sprintf(
		"file:%s?_pragma=foreign_keys(%s)&_pragma=busy_timeout(%d)&_pragma=journal_mode(%s)",
		cfg.Runtime.SQLitePath,
		fk,
		cfg.Runtime.SQLiteBusy.Milliseconds(),
		cfg.Database.SQLite.JournalMode,
	)
}

func verifySQLite(db *gorm.DB, cfg *config.Config) error {
	var fk int
	if err := db.Raw("pragma foreign_keys").Scan(&fk).Error; err != nil {
		return fmt.Errorf("读取 foreign_keys: %w", err)
	}
	if cfg.Runtime.ForeignKeys && fk != 1 {
		return errors.New("SQLite 外键没有在当前连接生效")
	}
	var journal string
	if err := db.Raw("pragma journal_mode").Scan(&journal).Error; err != nil {
		return fmt.Errorf("读取 journal_mode: %w", err)
	}
	if !strings.EqualFold(journal, cfg.Database.SQLite.JournalMode) {
		return fmt.Errorf("SQLite journal_mode=%s，期望 %s", journal, cfg.Database.SQLite.JournalMode)
	}
	var busy int
	if err := db.Raw("pragma busy_timeout").Scan(&busy).Error; err != nil {
		return fmt.Errorf("读取 busy_timeout: %w", err)
	}
	if busy != int(cfg.Runtime.SQLiteBusy.Milliseconds()) {
		return fmt.Errorf("SQLite busy_timeout=%d，期望 %d", busy, cfg.Runtime.SQLiteBusy.Milliseconds())
	}
	return nil
}

func OpenSQLiteFile(path string) (*gorm.DB, error) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Driver:       "sqlite",
			MaxOpenConns: 1,
			MaxIdleConns: 1,
			SQLite:       config.SQLiteConfig{JournalMode: "DELETE"},
		},
		Runtime: config.Runtime{
			SQLitePath:      path,
			SQLiteBusy:      5 * time.Second,
			ForeignKeys:     true,
			ConnMaxLifetime: 30 * time.Minute,
		},
	}
	return Open(cfg)
}
