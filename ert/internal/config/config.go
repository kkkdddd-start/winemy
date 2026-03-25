package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App         AppConfig         `yaml:"app"`
	Server      ServerConfig      `yaml:"server"`
	Database    DatabaseConfig    `yaml:"database"`
	Storage     StorageConfig     `yaml:"storage"`
	Concurrency ConcurrencyConfig `yaml:"concurrency"`
	Timeout     TimeoutConfig     `yaml:"timeout"`
	Security    SecurityConfig    `yaml:"security"`
	Modules     ModulesConfig     `yaml:"modules"`
	UI          UIConfig          `yaml:"ui"`
	Log         LogConfig         `yaml:"log"`
}

type AppConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Debug   bool   `yaml:"debug"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	Main  DBConfig    `yaml:"main"`
	Cache CacheConfig `yaml:"cache"`
}

type DBConfig struct {
	Path         string `yaml:"path"`
	WALMode      bool   `yaml:"wal_mode"`
	BusyTimeout  int    `yaml:"busy_timeout"`
	MaxOpenConns int    `yaml:"max_open_conns"`
}

type CacheConfig struct {
	MaxCost    int `yaml:"max_cost"`
	BufferSize int `yaml:"buffer_size"`
	TTL        int `yaml:"ttl"`
}

type StorageConfig struct {
	DataDir    string `yaml:"data_dir"`
	DumpDir    string `yaml:"dump_dir"`
	ReportDir  string `yaml:"report_dir"`
	MaxStorage int    `yaml:"max_storage"`
}

type ConcurrencyConfig struct {
	HighPriorityWorkers   int         `yaml:"high_priority_workers"`
	MediumPriorityWorkers int         `yaml:"medium_priority_workers"`
	LowPriorityWorkers    int         `yaml:"low_priority_workers"`
	Aging                 AgingConfig `yaml:"aging"`
}

type AgingConfig struct {
	MaxWaitTime    int `yaml:"max_wait_time"`
	BoostThreshold int `yaml:"boost_threshold"`
	PriorityBoost  int `yaml:"priority_boost"`
	CheckInterval  int `yaml:"check_interval"`
}

type TimeoutConfig struct {
	Global string            `yaml:"global"`
	Module map[string]string `yaml:"module"`
}

type SecurityConfig struct {
	ReadOnly bool           `yaml:"readonly"`
	Response ResponseConfig `yaml:"response"`
	Audit    AuditConfig    `yaml:"audit"`
}

type ResponseConfig struct {
	RequireConfirmation bool     `yaml:"require_confirmation"`
	AllowKillCritical   bool     `yaml:"allow_kill_critical"`
	BackupBeforeAction  bool     `yaml:"backup_before_action"`
	CriticalProcesses   []string `yaml:"critical_processes"`
}

type AuditConfig struct {
	Enabled       bool `yaml:"enabled"`
	Encrypt       bool `yaml:"encrypt"`
	RetentionDays int  `yaml:"retention_days"`
}

type ModulesConfig struct {
	System   ModuleConfig  `yaml:"system"`
	Process  ModuleConfig  `yaml:"process"`
	Network  ModuleConfig  `yaml:"network"`
	Monitor  MonitorConfig `yaml:"monitor"`
	Logging  LoggingConfig `yaml:"logging"`
	Memory   MemoryConfig  `yaml:"memory"`
	Threat   ModuleConfig  `yaml:"threat"`
	Report   ModuleConfig  `yaml:"report"`
	Baseline ModuleConfig  `yaml:"baseline"`
	Codec    CodecConfig   `yaml:"codec"`
}

type ModuleConfig struct {
	RefreshInterval string `yaml:"refresh_interval"`
}

type MonitorConfig struct {
	Interval    string          `yaml:"interval"`
	HistorySize int             `yaml:"history_size"`
	Alerts      AlertThresholds `yaml:"alerts"`
}

type AlertThresholds struct {
	CPUThreshold  int `yaml:"cpu_threshold"`
	MemThreshold  int `yaml:"mem_threshold"`
	DiskThreshold int `yaml:"disk_threshold"`
}

type LoggingConfig struct {
	UseDefaultPaths  bool     `yaml:"use_default_paths"`
	CustomPaths      []string `yaml:"custom_paths"`
	Recursive        bool     `yaml:"recursive"`
	SupportedFormats []string `yaml:"supported_formats"`
	MaxFileSize      int64    `yaml:"max_file_size"`
	Charset          string   `yaml:"charset"`
}

type MemoryConfig struct {
	MaxDumpSize int64 `yaml:"max_dump_size"`
	BlockSize   int64 `yaml:"block_size"`
}

type CodecConfig struct {
	EnableHistory bool `yaml:"enable_history"`
	MaxHistory    int  `yaml:"max_history"`
	RetentionDays int  `yaml:"retention_days"`
}

type UIConfig struct {
	Theme       string          `yaml:"theme"`
	Language    string          `yaml:"language"`
	RefreshRate string          `yaml:"refresh_rate"`
	Table       TableConfig     `yaml:"table"`
	Chart       ChartConfig     `yaml:"chart"`
	Shortcuts   ShortcutsConfig `yaml:"shortcuts"`
}

type TableConfig struct {
	PageSize      int  `yaml:"page_size"`
	VirtualScroll bool `yaml:"virtual_scroll"`
}

type ChartConfig struct {
	Animation bool   `yaml:"animation"`
	Theme     string `yaml:"theme"`
}

type ShortcutsConfig struct {
	Enabled bool `yaml:"enabled"`
}

type LogConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

var (
	cfg  *Config
	once sync.Once
)

func Load(path string) (*Config, error) {
	var loadErr error
	once.Do(func() {
		data, err := os.ReadFile(path)
		if err != nil {
			loadErr = fmt.Errorf("failed to read config file: %w", err)
			return
		}

		cfg = &Config{}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			loadErr = fmt.Errorf("failed to parse config file: %w", err)
			return
		}
	})

	if loadErr != nil {
		return nil, loadErr
	}

	return cfg, nil
}

func Get() *Config {
	if cfg == nil {
		cfg = &Config{}
	}
	return cfg
}

func Save(path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
