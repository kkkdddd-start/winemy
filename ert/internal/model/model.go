//go:build windows

package model

import (
	"time"
)

type RiskLevel int

const (
	RiskLow      RiskLevel = 0
	RiskMedium   RiskLevel = 1
	RiskHigh     RiskLevel = 2
	RiskCritical RiskLevel = 3
)

type SessionState int

const (
	SessionStateCreated    SessionState = 0
	SessionStateCollecting SessionState = 1
	SessionStatePaused     SessionState = 2
	SessionStateCompleted  SessionState = 3
	SessionStateFailed     SessionState = 4
	SessionStateRecovering SessionState = 5
)

type Session struct {
	ID        string       `json:"id"`
	Hostname  string       `json:"hostname"`
	Status    SessionState `json:"status"`
	StartedAt time.Time    `json:"started_at"`
	EndedAt   time.Time    `json:"ended_at"`
	UserName  string       `json:"user_name"`
	OSVersion string       `json:"os_version"`
	IsDomain  bool         `json:"is_domain"`
	Progress  float64      `json:"progress"`
	ErrorMsg  string       `json:"error_msg"`
}

type Progress struct {
	ModuleID   int       `json:"module_id"`
	Current    int       `json:"current"`
	Total      int       `json:"total"`
	Percentage float64   `json:"percentage"`
	Message    string    `json:"message"`
	ETA        time.Time `json:"eta"`
	StartedAt  time.Time `json:"started_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type SystemInfo struct {
	Hostname     string    `json:"hostname"`
	OSName       string    `json:"os_name"`
	OSVersion    string    `json:"os_version"`
	Architecture string    `json:"architecture"`
	BootTime     time.Time `json:"boot_time"`
	Uptime       uint64    `json:"uptime"`
	CurrentUser  string    `json:"current_user"`
	CPUCount     int       `json:"cpu_count"`
	MemoryTotal  uint64    `json:"memory_total"`
	DiskTotal    uint64    `json:"disk_total"`
	IsDomain     bool      `json:"is_domain"`
	DomainName   string    `json:"domain_name"`
}

type ProcessDTO struct {
	PID         uint32    `json:"pid"`
	PPID        uint32    `json:"ppid"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	CommandLine string    `json:"command_line"`
	User        string    `json:"user"`
	CPU         float64   `json:"cpu"`
	Memory      uint64    `json:"memory"`
	StartTime   time.Time `json:"start_time"`
	RiskLevel   RiskLevel `json:"risk_level"`
}

type ProcessTreeNode struct {
	PID      uint32             `json:"pid"`
	Name     string             `json:"name"`
	Children []*ProcessTreeNode `json:"children"`
}

type NetworkConnDTO struct {
	PID         uint32    `json:"pid"`
	ProcessName string    `json:"process_name"`
	Protocol    string    `json:"protocol"`
	LocalAddr   string    `json:"local_addr"`
	LocalPort   uint16    `json:"local_port"`
	RemoteAddr  string    `json:"remote_addr"`
	RemotePort  uint16    `json:"remote_port"`
	Country     string    `json:"country"`
	City        string    `json:"city"`
	State       string    `json:"state"`
	RiskLevel   RiskLevel `json:"risk_level"`
}

type RegistryKeyDTO struct {
	Path      string    `json:"path"`
	Name      string    `json:"name"`
	ValueType string    `json:"value_type"`
	Value     string    `json:"value"`
	Modified  time.Time `json:"modified"`
	RiskLevel RiskLevel `json:"risk_level"`
}

type ServiceDTO struct {
	Name         string    `json:"name"`
	DisplayName  string    `json:"display_name"`
	Status       string    `json:"status"`
	StartType    string    `json:"start_type"`
	Path         string    `json:"path"`
	Dependencies string    `json:"dependencies"`
	Description  string    `json:"description"`
	RiskLevel    RiskLevel `json:"risk_level"`
}

type ScheduledTaskDTO struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	State       string    `json:"state"`
	LastRunTime time.Time `json:"last_run_time"`
	NextRunTime time.Time `json:"next_run_time"`
	RiskLevel   RiskLevel `json:"risk_level"`
}

type EventLogDTO struct {
	EventID     int       `json:"event_id"`
	EventType   string    `json:"event_type"`
	Level       string    `json:"level"`
	Source      string    `json:"source"`
	Channel     string    `json:"channel"`
	Computer    string    `json:"computer"`
	TimeCreated time.Time `json:"time_created"`
	RawXML      string    `json:"raw_xml"`
	Message     string    `json:"message"`
}

type FileDTO struct {
	Path      string    `json:"path"`
	Name      string    `json:"name"`
	Size      uint64    `json:"size"`
	MD5       string    `json:"md5"`
	SHA1      string    `json:"sha1"`
	SHA256    string    `json:"sha256"`
	Modified  time.Time `json:"modified"`
	Created   time.Time `json:"created"`
	IsLarge   bool      `json:"is_large"`
	IsHidden  bool      `json:"is_hidden"`
	IsSystem  bool      `json:"is_system"`
	IsSigned  bool      `json:"is_signed"`
	Signature string    `json:"signature"`
	HasADS    bool      `json:"has_ads"`
	RiskLevel RiskLevel `json:"risk_level"`
}

type AccountDTO struct {
	Name      string    `json:"name"`
	FullName  string    `json:"full_name"`
	SID       string    `json:"sid"`
	Domain    string    `json:"domain"`
	Status    string    `json:"status"`
	LastLogon time.Time `json:"last_logon"`
	RiskLevel RiskLevel `json:"risk_level"`
}

type DriverDTO struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	BaseAddr  string    `json:"base_address"`
	Size      uint64    `json:"size"`
	IsSigned  bool      `json:"is_signed"`
	Signature string    `json:"signature"`
	RiskLevel RiskLevel `json:"risk_level"`
}

type MemoryDumpDTO struct {
	PID         uint32    `json:"pid"`
	ProcessName string    `json:"process_name"`
	DumpType    string    `json:"dump_type"`
	FilePath    string    `json:"file_path"`
	FileSize    uint64    `json:"file_size"`
	SHA256      string    `json:"sha256"`
	CreatedAt   time.Time `json:"created_at"`
}

type AlertRule struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ModuleID  int       `json:"module_id"`
	Condition string    `json:"condition"`
	Metric    string    `json:"metric"`
	Threshold float64   `json:"threshold"`
	Duration  int       `json:"duration"`
	Severity  RiskLevel `json:"severity"`
	Enabled   bool      `json:"enabled"`
	Actions   []string  `json:"actions"`
}

type AlertEvent struct {
	ID        string    `json:"id"`
	RuleID    string    `json:"rule_id"`
	RuleName  string    `json:"rule_name"`
	Severity  RiskLevel `json:"severity"`
	Message   string    `json:"message"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
	Timestamp time.Time `json:"timestamp"`
	ModuleID  int       `json:"module_id"`
}

type CompareResult struct {
	Session1         string           `json:"session1"`
	Session2         string           `json:"session2"`
	AddedProcesses   []ProcessDTO     `json:"added_processes"`
	RemovedProcesses []ProcessDTO     `json:"removed_processes"`
	AddedNetwork     []NetworkConnDTO `json:"added_network"`
	RemovedNetwork   []NetworkConnDTO `json:"removed_network"`
	AddedRegistry    []RegistryKeyDTO `json:"added_registry"`
	RemovedRegistry  []RegistryKeyDTO `json:"removed_registry"`
	AddedServices    []ServiceDTO     `json:"added_services"`
	RemovedServices  []ServiceDTO     `json:"removed_services"`
}

type TimelineEvent struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	ModuleID    int       `json:"module_id"`
	EventType   string    `json:"event_type"`
	Severity    RiskLevel `json:"severity"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Source      string    `json:"source"`
	Target      string    `json:"target"`
	Details     string    `json:"details"`
	SessionID   string    `json:"session_id"`
}

type SessionSummary struct {
	SessionID     string    `json:"session_id"`
	Hostname      string    `json:"hostname"`
	CollectedAt   time.Time `json:"collected_at"`
	ProcessCount  int       `json:"process_count"`
	NetworkCount  int       `json:"network_count"`
	RegistryCount int       `json:"registry_count"`
	ServiceCount  int       `json:"service_count"`
	RiskScore     int       `json:"risk_score"`
	RiskLevel     RiskLevel `json:"risk_level"`
}

type CompareSummary struct {
	TotalAdded    int `json:"total_added"`
	TotalRemoved  int `json:"total_removed"`
	TotalModified int `json:"total_modified"`
	RiskIncreased int `json:"risk_increased"`
	RiskDecreased int `json:"risk_decreased"`
}

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
