package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOverridesFalseAndEnv(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "configs")
	if err := os.MkdirAll(base, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "application.yaml"), []byte(sampleYAML), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "application-dev.yaml"), []byte("logging:\n  level: debug\n  format: text\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	t.Setenv("APP_PROFILE", "dev")
	t.Setenv("APP_SERVER_PORT", "9090")
	t.Setenv("APP_DATABASE_SQLITE_FOREIGN_KEYS", "false")
	t.Setenv("APP_AI_FALLBACK_LOCAL", "false")
	cfg, err := Load([]string{"--base-dir", dir, "--config-dir", "configs"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Port != 9090 {
		t.Fatalf("port = %d", cfg.Server.Port)
	}
	if cfg.Runtime.ForeignKeys {
		t.Fatal("foreign_keys false was ignored")
	}
	if cfg.AI.FallbackLocal {
		t.Fatal("fallback false was ignored")
	}
	if cfg.Logging.Level != "debug" || cfg.Logging.Format != "text" {
		t.Fatalf("profile logging = %s %s", cfg.Logging.Level, cfg.Logging.Format)
	}
	if cfg.App.Profile != "dev" {
		t.Fatal(cfg.App.Profile)
	}
}

func TestUnknownProfile(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "configs"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "configs", "application.yaml"), []byte(sampleYAML), 0o640); err != nil {
		t.Fatal(err)
	}
	if _, err := Load([]string{"--base-dir", dir, "--profile", "staging"}); err == nil {
		t.Fatal("expected unknown profile error")
	}
}

const sampleYAML = `
app:
  name: poster
  data_dir: ./data
server:
  host: 127.0.0.1
  port: 8080
  read_header_timeout: 5s
  idle_timeout: 60s
  max_header_bytes: 1048576
  shutdown_timeout: 30s
  request_timeout: 90s
  sse_timeout: 15m
  max_body_bytes: 2097152
  trusted_proxies: []
database:
  enabled: true
  driver: sqlite
  max_open_conns: 1
  max_idle_conns: 1
  conn_max_lifetime: 30m
  batch_size: 100
  sqlite:
    path: app.db
    foreign_keys: true
    busy_timeout: 5s
    journal_mode: DELETE
auth:
  enabled: true
  session_ttl: 8h
  login_max_attempts: 8
  login_window: 10m
logging:
  level: info
  format: json
ai:
  enabled: true
  provider: openai
  base_url: https://api.openai.com/v1
  model: gpt-4o-mini
  timeout: 45s
  fallback_local: true
  max_per_minute: 10
upload:
  max_bytes: 8388608
  dir: uploads
`
