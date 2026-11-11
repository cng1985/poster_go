package config

import "time"

// Config 是启动后的有效配置。定位参数已解析为绝对路径。
type Config struct {
	ShowVersion bool           `yaml:"-"`
	App         AppConfig      `yaml:"app"`
	Server      ServerConfig   `yaml:"server"`
	Database    DatabaseConfig `yaml:"database"`
	Auth        AuthConfig     `yaml:"auth"`
	Logging     LoggingConfig  `yaml:"logging"`
	AI          AIConfig       `yaml:"ai"`
	Upload      UploadConfig   `yaml:"upload"`
	Runtime     Runtime        `yaml:"-"`
	Sources     []string       `yaml:"-"`
}

type AppConfig struct {
	Name      string `yaml:"name"`
	Profile   string `yaml:"profile"`
	BaseDir   string `yaml:"base_dir"`
	ConfigDir string `yaml:"config_dir"`
	DataDir   string `yaml:"data_dir"`
}

type ServerConfig struct {
	Host              string   `yaml:"host"`
	Port              int      `yaml:"port"`
	ReadHeaderTimeout string   `yaml:"read_header_timeout"`
	IdleTimeout       string   `yaml:"idle_timeout"`
	MaxHeaderBytes    int      `yaml:"max_header_bytes"`
	ShutdownTimeout   string   `yaml:"shutdown_timeout"`
	RequestTimeout    string   `yaml:"request_timeout"`
	SSETimeout        string   `yaml:"sse_timeout"`
	MaxBodyBytes      int64    `yaml:"max_body_bytes"`
	TrustedProxies    []string `yaml:"trusted_proxies"`
}

type SQLiteConfig struct {
	Path        string `yaml:"path"`
	ForeignKeys *bool  `yaml:"foreign_keys"`
	BusyTimeout string `yaml:"busy_timeout"`
	JournalMode string `yaml:"journal_mode"`
}

type DatabaseConfig struct {
	Enabled         bool         `yaml:"enabled"`
	Driver          string       `yaml:"driver"`
	DSN             string       `yaml:"dsn"`
	MaxOpenConns    int          `yaml:"max_open_conns"`
	MaxIdleConns    int          `yaml:"max_idle_conns"`
	ConnMaxLifetime string       `yaml:"conn_max_lifetime"`
	BatchSize       int          `yaml:"batch_size"`
	SQLite          SQLiteConfig `yaml:"sqlite"`
}

type AuthConfig struct {
	Enabled     bool   `yaml:"enabled"`
	SessionTTL  string `yaml:"session_ttl"`
	LoginMax    int    `yaml:"login_max_attempts"`
	LoginWindow string `yaml:"login_window"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type AIConfig struct {
	Enabled       bool   `yaml:"enabled"`
	Provider      string `yaml:"provider"`
	BaseURL       string `yaml:"base_url"`
	Model         string `yaml:"model"`
	Timeout       string `yaml:"timeout"`
	FallbackLocal bool   `yaml:"fallback_local"`
	MaxPerMinute  int    `yaml:"max_per_minute"`
	APIKey        string `yaml:"-"`
}

type UploadConfig struct {
	MaxBytes int64  `yaml:"max_bytes"`
	Dir      string `yaml:"dir"`
}

// Runtime 保存解析后的时长和路径，避免业务层重复解析。
type Runtime struct {
	ReadHeaderTimeout time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	RequestTimeout    time.Duration
	SSETimeout        time.Duration
	SessionTTL        time.Duration
	LoginWindow       time.Duration
	ConnMaxLifetime   time.Duration
	SQLiteBusy        time.Duration
	AITimeout         time.Duration
	SQLitePath        string
	UploadDir         string
	ForeignKeys       bool
}

func (c Config) SafeLines() []any {
	keySet := c.AI.APIKey != ""
	dsnSet := c.Database.DSN != ""
	return []any{
		"name", c.App.Name,
		"profile", c.App.Profile,
		"base_dir", c.App.BaseDir,
		"config_dir", c.App.ConfigDir,
		"data_dir", c.App.DataDir,
		"sqlite_path", c.Runtime.SQLitePath,
		"driver", c.Database.Driver,
		"dsn_set", dsnSet,
		"ai_key_set", keySet,
		"ai_fallback_local", c.AI.FallbackLocal,
		"sources", c.Sources,
	}
}
