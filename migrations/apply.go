package migrations

import (
	"embed"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

//go:embed *.sql
var files embed.FS

func Apply(db *gorm.DB) error {
	if db.Dialector.Name() != "sqlite" {
		return errors.New("当前发行只携带已验证的 SQLite 迁移，其他驱动不能直接启动")
	}
	if err := db.Exec(`create table if not exists schema_migration (
		version text primary key,
		applied_at text not null
	)`).Error; err != nil {
		return fmt.Errorf("创建迁移记录表: %w", err)
	}
	entries, err := files.ReadDir(".")
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		var count int64
		if err := db.Raw("select count(1) from schema_migration where version = ?", name).Scan(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		content, err := files.ReadFile(name)
		if err != nil {
			return err
		}
		err = db.Transaction(func(tx *gorm.DB) error {
			for _, statement := range statements(string(content)) {
				if err := tx.Exec(statement).Error; err != nil {
					return fmt.Errorf("%s: %w", name, err)
				}
			}
			return tx.Exec(
				"insert into schema_migration(version, applied_at) values (?, ?)",
				name,
				time.Now().UTC().Format(time.RFC3339),
			).Error
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func statements(script string) []string {
	parts := strings.Split(script, ";")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		var lines []string
		for _, line := range strings.Split(part, "\n") {
			trim := strings.TrimSpace(line)
			if trim == "" || strings.HasPrefix(trim, "--") {
				continue
			}
			lines = append(lines, line)
		}
		cleaned := strings.TrimSpace(strings.Join(lines, "\n"))
		if cleaned != "" {
			out = append(out, cleaned)
		}
	}
	return out
}
