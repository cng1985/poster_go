package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type locators struct {
	baseDir   string
	configDir string
	dataDir   string
	profile   string
	dataSet   bool
	version   bool
}

// Load 先确定定位参数，再加载 YAML，并用环境变量和命令行覆盖。
func Load(args []string) (*Config, error) {
	loc, err := parseLocators(args)
	if err != nil {
		return nil, err
	}
	if loc.version {
		return &Config{ShowVersion: true}, nil
	}
	base, err := filepath.Abs(loc.baseDir)
	if err != nil {
		return nil, fmt.Errorf("解析 base-dir: %w", err)
	}
	configDir := loc.configDir
	if !filepath.IsAbs(configDir) {
		configDir = filepath.Join(base, configDir)
	}
	cfg, sources, err := loadFiles(configDir, loc.profile)
	if err != nil {
		return nil, err
	}
	if err := applyEnv(cfg); err != nil {
		return nil, err
	}
	if loc.dataSet {
		cfg.App.DataDir = loc.dataDir
	}
	cfg.App.Profile = loc.profile
	cfg.App.BaseDir = base
	cfg.App.ConfigDir = configDir
	cfg.Sources = sources
	if err := resolvePaths(cfg); err != nil {
		return nil, err
	}
	if err := validate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// VersionRequested 通过预解析识别 --version，避免缺配置时无法查看版本。
func VersionRequested(args []string) bool {
	loc, err := parseLocators(args)
	return err == nil && loc.version
}

func parseLocators(args []string) (locators, error) {
	fs := flag.NewFlagSet("poster", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	baseDir := fs.String("base-dir", "", "应用根目录")
	configDir := fs.String("config-dir", "", "配置目录")
	dataDir := fs.String("data-dir", "", "数据目录")
	profile := fs.String("profile", "", "环境：dev 或 pro")
	version := fs.Bool("version", false, "打印版本")
	if err := fs.Parse(args); err != nil {
		return locators{}, err
	}
	loc := locators{version: *version}
	loc.baseDir = firstSet(*baseDir, os.Getenv("APP_BASE_DIR"), ".")
	loc.configDir = firstSet(*configDir, os.Getenv("APP_CONFIG_DIR"), "configs")
	if *dataDir != "" {
		loc.dataDir = *dataDir
		loc.dataSet = true
	} else if v, ok := os.LookupEnv("APP_DATA_DIR"); ok {
		if v == "" {
			return locators{}, errors.New("APP_DATA_DIR 已设置但为空")
		}
		loc.dataDir = v
		loc.dataSet = true
	}
	loc.profile = firstSet(*profile, os.Getenv("APP_PROFILE"), "")
	return loc, nil
}

func firstSet(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func loadFiles(configDir, profile string) (*Config, []string, error) {
	switch profile {
	case "", "dev", "pro":
	default:
		return nil, nil, fmt.Errorf("未知 profile %q，只允许 dev 或 pro", profile)
	}
	basePath := filepath.Join(configDir, "application.yaml")
	merged, err := readMap(basePath)
	if err != nil {
		return nil, nil, err
	}
	sources := []string{basePath}
	if profile != "" {
		profilePath := filepath.Join(configDir, "application-"+profile+".yaml")
		overlay, err := readMap(profilePath)
		if err != nil {
			return nil, nil, err
		}
		deepMerge(merged, overlay)
		sources = append(sources, profilePath)
	}
	if err := rejectLocatorOverride(merged, profile); err != nil {
		return nil, nil, err
	}
	raw, err := yaml.Marshal(merged)
	if err != nil {
		return nil, nil, err
	}
	decoder := yaml.NewDecoder(strings.NewReader(string(raw)))
	decoder.KnownFields(true)
	cfg := &Config{}
	if err := decoder.Decode(cfg); err != nil {
		return nil, nil, fmt.Errorf("解析配置: %w", err)
	}
	return cfg, sources, nil
}

func rejectLocatorOverride(merged map[string]any, profile string) error {
	app, _ := merged["app"].(map[string]any)
	if app == nil {
		return nil
	}
	for _, key := range []string{"base_dir", "config_dir"} {
		if _, ok := app[key]; ok {
			return fmt.Errorf("配置项 app.%s 不能写在 YAML 中，请使用启动参数或环境变量", key)
		}
	}
	if value, ok := app["profile"]; ok {
		text := strings.TrimSpace(fmt.Sprint(value))
		if text != "" && text != profile && !(text == "<nil>") {
			return fmt.Errorf("YAML 中的 app.profile=%v 与启动参数 %q 冲突", value, profile)
		}
	}
	delete(app, "base_dir")
	delete(app, "config_dir")
	delete(app, "profile")
	return nil
}

func readMap(path string) (map[string]any, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置 %s: %w", path, err)
	}
	var decoded map[string]any
	if err := yaml.Unmarshal(content, &decoded); err != nil {
		return nil, fmt.Errorf("解析 %s: %w", path, err)
	}
	if decoded == nil {
		decoded = map[string]any{}
	}
	return decoded, nil
}

func deepMerge(dst, src map[string]any) {
	for key, value := range src {
		child, ok := value.(map[string]any)
		if ok {
			if current, exists := dst[key].(map[string]any); exists {
				deepMerge(current, child)
				continue
			}
		}
		dst[key] = value
	}
}

func applyEnv(cfg *Config) error {
	setters := []struct {
		key string
		set func(string) error
	}{
		{"APP_NAME", func(v string) error { cfg.App.Name = v; return nil }},
		{"APP_SERVER_HOST", func(v string) error { cfg.Server.Host = v; return nil }},
		{"APP_SERVER_PORT", func(v string) error { return assignInt(v, &cfg.Server.Port) }},
		{"APP_SERVER_READ_HEADER_TIMEOUT", func(v string) error { cfg.Server.ReadHeaderTimeout = v; return nil }},
		{"APP_SERVER_IDLE_TIMEOUT", func(v string) error { cfg.Server.IdleTimeout = v; return nil }},
		{"APP_SERVER_MAX_HEADER_BYTES", func(v string) error { return assignInt(v, &cfg.Server.MaxHeaderBytes) }},
		{"APP_SERVER_SHUTDOWN_TIMEOUT", func(v string) error { cfg.Server.ShutdownTimeout = v; return nil }},
		{"APP_SERVER_REQUEST_TIMEOUT", func(v string) error { cfg.Server.RequestTimeout = v; return nil }},
		{"APP_SERVER_SSE_TIMEOUT", func(v string) error { cfg.Server.SSETimeout = v; return nil }},
		{"APP_SERVER_MAX_BODY_BYTES", func(v string) error { return assignInt64(v, &cfg.Server.MaxBodyBytes) }},
		{"APP_SERVER_TRUSTED_PROXIES", func(v string) error {
			var items []string
			if err := json.Unmarshal([]byte(v), &items); err != nil {
				return fmt.Errorf("APP_SERVER_TRUSTED_PROXIES 必须是 JSON 数组: %w", err)
			}
			cfg.Server.TrustedProxies = items
			return nil
		}},
		{"APP_DATABASE_ENABLED", func(v string) error { return assignBool(v, &cfg.Database.Enabled) }},
		{"APP_DATABASE_DRIVER", func(v string) error { cfg.Database.Driver = v; return nil }},
		{"APP_DATABASE_SQLITE_PATH", func(v string) error { cfg.Database.SQLite.Path = v; return nil }},
		{"APP_DATABASE_SQLITE_FOREIGN_KEYS", func(v string) error {
			var flag bool
			if err := assignBool(v, &flag); err != nil {
				return err
			}
			cfg.Database.SQLite.ForeignKeys = &flag
			return nil
		}},
		{"APP_DATABASE_SQLITE_BUSY_TIMEOUT", func(v string) error { cfg.Database.SQLite.BusyTimeout = v; return nil }},
		{"APP_DATABASE_SQLITE_JOURNAL_MODE", func(v string) error { cfg.Database.SQLite.JournalMode = v; return nil }},
		{"APP_DATABASE_DSN", func(v string) error {
			if v == "" {
				return errors.New("APP_DATABASE_DSN 已设置但为空")
			}
			cfg.Database.DSN = v
			return nil
		}},
		{"APP_DATABASE_MAX_OPEN_CONNS", func(v string) error { return assignInt(v, &cfg.Database.MaxOpenConns) }},
		{"APP_DATABASE_MAX_IDLE_CONNS", func(v string) error { return assignInt(v, &cfg.Database.MaxIdleConns) }},
		{"APP_DATABASE_CONN_MAX_LIFETIME", func(v string) error { cfg.Database.ConnMaxLifetime = v; return nil }},
		{"APP_DATABASE_BATCH_SIZE", func(v string) error { return assignInt(v, &cfg.Database.BatchSize) }},
		{"APP_AUTH_ENABLED", func(v string) error { return assignBool(v, &cfg.Auth.Enabled) }},
		{"APP_AUTH_SESSION_TTL", func(v string) error { cfg.Auth.SessionTTL = v; return nil }},
		{"APP_LOGGING_LEVEL", func(v string) error { cfg.Logging.Level = v; return nil }},
		{"APP_LOGGING_FORMAT", func(v string) error { cfg.Logging.Format = v; return nil }},
		{"APP_AI_API_KEY", func(v string) error {
			if v == "" {
				return errors.New("APP_AI_API_KEY 已设置但为空")
			}
			cfg.AI.APIKey = v
			return nil
		}},
		{"APP_AI_BASE_URL", func(v string) error { cfg.AI.BaseURL = v; return nil }},
		{"APP_AI_MODEL", func(v string) error { cfg.AI.Model = v; return nil }},
		{"APP_AI_TIMEOUT", func(v string) error { cfg.AI.Timeout = v; return nil }},
		{"APP_AI_FALLBACK_LOCAL", func(v string) error { return assignBool(v, &cfg.AI.FallbackLocal) }},
	}
	for _, item := range setters {
		value, ok := os.LookupEnv(item.key)
		if !ok {
			continue
		}
		if value == "" && item.key != "APP_DATABASE_DSN" && item.key != "APP_AI_API_KEY" && item.key != "APP_SERVER_TRUSTED_PROXIES" {
			return fmt.Errorf("%s 已设置但为空", item.key)
		}
		if err := item.set(value); err != nil {
			return err
		}
	}
	return nil
}

func assignInt(value string, dst *int) error {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("整数配置无效 %q", value)
	}
	*dst = parsed
	return nil
}

func assignInt64(value string, dst *int64) error {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("整数配置无效 %q", value)
	}
	*dst = parsed
	return nil
}

func assignBool(value string, dst *bool) error {
	switch value {
	case "true":
		*dst = true
	case "false":
		*dst = false
	default:
		return fmt.Errorf("布尔配置只接受 true 或 false，收到 %q", value)
	}
	return nil
}

func resolvePaths(cfg *Config) error {
	dataDir := cfg.App.DataDir
	if dataDir == "" {
		dataDir = "./data"
	}
	if !filepath.IsAbs(dataDir) {
		dataDir = filepath.Join(cfg.App.BaseDir, dataDir)
	}
	abs, err := filepath.Abs(dataDir)
	if err != nil {
		return err
	}
	cfg.App.DataDir = abs
	sqlitePath := cfg.Database.SQLite.Path
	if sqlitePath == "" {
		sqlitePath = "app.db"
	}
	if !filepath.IsAbs(sqlitePath) {
		sqlitePath = filepath.Join(abs, sqlitePath)
	}
	cfg.Runtime.SQLitePath = sqlitePath
	uploadDir := cfg.Upload.Dir
	if uploadDir == "" {
		uploadDir = "uploads"
	}
	if !filepath.IsAbs(uploadDir) {
		uploadDir = filepath.Join(abs, uploadDir)
	}
	cfg.Runtime.UploadDir = uploadDir
	return nil
}

func validate(cfg *Config) error {
	if strings.TrimSpace(cfg.App.Name) == "" {
		return errors.New("app.name 不能为空")
	}
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return errors.New("server.port 必须在 1 到 65535 之间")
	}
	if strings.TrimSpace(cfg.Server.Host) == "" {
		return errors.New("server.host 不能为空")
	}
	var err error
	if cfg.Runtime.ReadHeaderTimeout, err = parsePositive(cfg.Server.ReadHeaderTimeout); err != nil {
		return fmt.Errorf("server.read_header_timeout: %w", err)
	}
	if cfg.Runtime.IdleTimeout, err = parsePositive(cfg.Server.IdleTimeout); err != nil {
		return fmt.Errorf("server.idle_timeout: %w", err)
	}
	if cfg.Runtime.ShutdownTimeout, err = parsePositive(cfg.Server.ShutdownTimeout); err != nil {
		return fmt.Errorf("server.shutdown_timeout: %w", err)
	}
	if cfg.Runtime.RequestTimeout, err = parsePositive(cfg.Server.RequestTimeout); err != nil {
		return fmt.Errorf("server.request_timeout: %w", err)
	}
	if cfg.Runtime.SSETimeout, err = parsePositive(cfg.Server.SSETimeout); err != nil {
		return fmt.Errorf("server.sse_timeout: %w", err)
	}
	if cfg.Server.MaxHeaderBytes < 1024 {
		return errors.New("server.max_header_bytes 过小")
	}
	if cfg.Server.MaxBodyBytes < 1024 {
		return errors.New("server.max_body_bytes 过小")
	}
	if !cfg.Database.Enabled {
		return errors.New("海报工坊需要数据库，database.enabled 不能为 false")
	}
	switch cfg.Database.Driver {
	case "sqlite", "mysql", "postgres", "pgsql", "sqlserver":
	default:
		return fmt.Errorf("未知 database.driver %q", cfg.Database.Driver)
	}
	if cfg.Database.Driver != "sqlite" && strings.TrimSpace(cfg.Database.DSN) == "" {
		return errors.New("网络数据库必须通过 APP_DATABASE_DSN 提供连接串")
	}
	if cfg.Database.MaxOpenConns < 1 || cfg.Database.MaxIdleConns < 1 || cfg.Database.MaxIdleConns > cfg.Database.MaxOpenConns {
		return errors.New("数据库连接池参数无效")
	}
	if cfg.Database.BatchSize < 1 {
		return errors.New("database.batch_size 必须为正数")
	}
	if cfg.Runtime.ConnMaxLifetime, err = parsePositive(cfg.Database.ConnMaxLifetime); err != nil {
		return fmt.Errorf("database.conn_max_lifetime: %w", err)
	}
	if cfg.Runtime.SQLiteBusy, err = parsePositive(cfg.Database.SQLite.BusyTimeout); err != nil {
		return fmt.Errorf("database.sqlite.busy_timeout: %w", err)
	}
	switch strings.ToUpper(cfg.Database.SQLite.JournalMode) {
	case "DELETE", "WAL", "TRUNCATE", "PERSIST", "MEMORY", "OFF":
	default:
		return fmt.Errorf("不支持的 journal_mode %q", cfg.Database.SQLite.JournalMode)
	}
	cfg.Database.SQLite.JournalMode = strings.ToUpper(cfg.Database.SQLite.JournalMode)
	cfg.Runtime.ForeignKeys = true
	if cfg.Database.SQLite.ForeignKeys != nil {
		cfg.Runtime.ForeignKeys = *cfg.Database.SQLite.ForeignKeys
	}
	if !cfg.Auth.Enabled {
		return errors.New("海报工坊需要认证，auth.enabled 不能为 false")
	}
	if cfg.Runtime.SessionTTL, err = parsePositive(cfg.Auth.SessionTTL); err != nil {
		return fmt.Errorf("auth.session_ttl: %w", err)
	}
	if cfg.Auth.LoginWindow == "" {
		cfg.Auth.LoginWindow = "10m"
	}
	if cfg.Runtime.LoginWindow, err = parsePositive(cfg.Auth.LoginWindow); err != nil {
		return fmt.Errorf("auth.login_window: %w", err)
	}
	if cfg.Auth.LoginMax <= 0 {
		cfg.Auth.LoginMax = 8
	}
	switch cfg.Logging.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("未知 logging.level %q", cfg.Logging.Level)
	}
	switch cfg.Logging.Format {
	case "json", "text":
	default:
		return fmt.Errorf("未知 logging.format %q", cfg.Logging.Format)
	}
	if cfg.AI.Enabled {
		if cfg.Runtime.AITimeout, err = parsePositive(cfg.AI.Timeout); err != nil {
			return fmt.Errorf("ai.timeout: %w", err)
		}
		if cfg.AI.MaxPerMinute <= 0 {
			cfg.AI.MaxPerMinute = 10
		}
		if strings.TrimSpace(cfg.AI.BaseURL) == "" || strings.TrimSpace(cfg.AI.Model) == "" {
			return errors.New("启用 AI 时必须配置 base_url 和 model")
		}
		if cfg.Runtime.AITimeout >= cfg.Runtime.RequestTimeout {
			return errors.New("ai.timeout 必须小于 server.request_timeout")
		}
	}
	if cfg.Upload.MaxBytes < 1024 {
		return errors.New("upload.max_bytes 过小")
	}
	if cfg.Upload.MaxBytes > 32*1024*1024 {
		return errors.New("upload.max_bytes 不能超过 32MiB")
	}
	return nil
}

func parsePositive(value string) (time.Duration, error) {
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("时长无效 %q", value)
	}
	if parsed <= 0 {
		return 0, errors.New("时长必须为正数")
	}
	return parsed, nil
}
