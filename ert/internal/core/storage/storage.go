//go:build windows

package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

type Storage struct {
	db     *sql.DB
	dbPath string
	mu     sync.RWMutex
}

func New(dbPath string) (*Storage, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db dir: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	s := &Storage{
		db:     db,
		dbPath: dbPath,
	}

	if err := s.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return s, nil
}

func (s *Storage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		hostname TEXT NOT NULL,
		status INTEGER NOT NULL,
		started_at TEXT NOT NULL,
		ended_at TEXT,
		user_name TEXT,
		os_version TEXT,
		is_domain INTEGER DEFAULT 0,
		progress REAL DEFAULT 0,
		error_msg TEXT
	);

	CREATE TABLE IF NOT EXISTS processes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pid INTEGER NOT NULL,
		name TEXT NOT NULL,
		path TEXT,
		command_line TEXT,
		user_name TEXT,
		cpu_percent REAL,
		memory_bytes INTEGER,
		start_time TEXT,
		risk_level INTEGER DEFAULT 0,
		collected_at TEXT NOT NULL,
		session_id TEXT NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(id)
	);

	CREATE INDEX IF NOT EXISTS idx_processes_pid ON processes(pid);
	CREATE INDEX IF NOT EXISTS idx_processes_name ON processes(name);
	CREATE INDEX IF NOT EXISTS idx_processes_risk ON processes(risk_level);
	CREATE INDEX IF NOT EXISTS idx_processes_session ON processes(session_id);

	CREATE TABLE IF NOT EXISTS network_connections (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pid INTEGER,
		protocol TEXT,
		local_addr TEXT,
		local_port INTEGER,
		remote_addr TEXT,
		remote_port INTEGER,
		state TEXT,
		risk_level INTEGER DEFAULT 0,
		collected_at TEXT NOT NULL,
		session_id TEXT NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(id)
	);

	CREATE INDEX IF NOT EXISTS idx_network_session ON network_connections(session_id);
	CREATE INDEX IF NOT EXISTS idx_network_remote ON network_connections(remote_addr);

	CREATE TABLE IF NOT EXISTS registry_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT NOT NULL,
		name TEXT,
		value_type TEXT,
		value TEXT,
		modified_time TEXT,
		risk_level INTEGER DEFAULT 0,
		collected_at TEXT NOT NULL,
		session_id TEXT NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(id)
	);

	CREATE INDEX IF NOT EXISTS idx_registry_path ON registry_keys(path);
	CREATE INDEX IF NOT EXISTS idx_registry_session ON registry_keys(session_id);

	CREATE TABLE IF NOT EXISTS services (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		display_name TEXT,
		status TEXT,
		start_type TEXT,
		path TEXT,
		risk_level INTEGER DEFAULT 0,
		collected_at TEXT NOT NULL,
		session_id TEXT NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(id)
	);

	CREATE INDEX IF NOT EXISTS idx_services_name ON services(name);
	CREATE INDEX IF NOT EXISTS idx_services_session ON services(session_id);

	CREATE TABLE IF NOT EXISTS scheduled_tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		path TEXT,
		state TEXT,
		last_run_time TEXT,
		next_run_time TEXT,
		risk_level INTEGER DEFAULT 0,
		collected_at TEXT NOT NULL,
		session_id TEXT NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(id)
	);

	CREATE TABLE IF NOT EXISTS event_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_id INTEGER NOT NULL,
		event_type TEXT NOT NULL,
		level TEXT NOT NULL,
		source TEXT NOT NULL,
		channel TEXT NOT NULL,
		computer TEXT NOT NULL,
		time_created TEXT NOT NULL,
		raw_xml TEXT,
		session_id TEXT NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(id)
	);

	CREATE INDEX IF NOT EXISTS idx_event_logs_event_id ON event_logs(event_id);
	CREATE INDEX IF NOT EXISTS idx_event_logs_source ON event_logs(source);
	CREATE INDEX IF NOT EXISTS idx_event_logs_level ON event_logs(level);
	CREATE INDEX IF NOT EXISTS idx_event_logs_time ON event_logs(time_created);

	CREATE VIRTUAL TABLE IF NOT EXISTS logs_fts USING fts5(
		event_id,
		message,
		time_created,
		source,
		session_id,
		tokenize='porter unicode61'
	);

	CREATE TABLE IF NOT EXISTS accounts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		full_name TEXT,
		sid TEXT,
		domain TEXT,
		status TEXT,
		last_logon TEXT,
		risk_level INTEGER DEFAULT 0,
		collected_at TEXT NOT NULL,
		session_id TEXT NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(id)
	);

	CREATE TABLE IF NOT EXISTS drivers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		path TEXT,
		base_address TEXT,
		size INTEGER,
		is_signed INTEGER DEFAULT 0,
		signature TEXT,
		risk_level INTEGER DEFAULT 0,
		collected_at TEXT NOT NULL,
		session_id TEXT NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(id)
	);

	CREATE TABLE IF NOT EXISTS memory_dumps (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pid INTEGER,
		process_name TEXT,
		dump_type TEXT NOT NULL,
		file_path TEXT NOT NULL,
		file_size INTEGER,
		sha256 TEXT,
		created_at TEXT NOT NULL,
		session_id TEXT NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(id)
	);

	CREATE INDEX IF NOT EXISTS idx_memory_dumps_pid ON memory_dumps(pid);
	CREATE INDEX IF NOT EXISTS idx_memory_dumps_type ON memory_dumps(dump_type);

	CREATE TABLE IF NOT EXISTS audit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp TEXT NOT NULL,
		operator TEXT NOT NULL,
		action TEXT NOT NULL,
		target TEXT NOT NULL,
		result TEXT NOT NULL,
		details TEXT
	);

	CREATE TABLE IF NOT EXISTS checkpoints (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		task_id TEXT NOT NULL,
		state TEXT NOT NULL,
		version INTEGER NOT NULL,
		created_at TEXT NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(id)
	);

	CREATE TABLE IF NOT EXISTS codec_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		input TEXT NOT NULL,
		output TEXT NOT NULL,
		codec_type TEXT NOT NULL,
		operation TEXT NOT NULL,
		created_at TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_codec_history_created ON codec_history(created_at);
	CREATE INDEX IF NOT EXISTS idx_codec_history_type ON codec_history(codec_type);

	CREATE TABLE IF NOT EXISTS alert_events (
		id TEXT PRIMARY KEY,
		rule_id TEXT NOT NULL,
		rule_name TEXT NOT NULL,
		severity INTEGER NOT NULL,
		message TEXT NOT NULL,
		value REAL,
		threshold REAL,
		timestamp TEXT NOT NULL,
		module_id INTEGER NOT NULL,
		session_id TEXT NOT NULL,
		acknowledged INTEGER DEFAULT 0,
		acknowledged_by TEXT,
		acknowledged_at TEXT,
		FOREIGN KEY (session_id) REFERENCES sessions(id)
	);

	CREATE TABLE IF NOT EXISTS analysis_reports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		risk_score INTEGER,
		risk_level TEXT,
		report_path TEXT,
		created_at TEXT NOT NULL,
		FOREIGN KEY (session_id) REFERENCES sessions(id)
	);

	CREATE TABLE IF NOT EXISTS security_alerts (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		alert_type TEXT NOT NULL,
		severity TEXT NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		evidence TEXT,
		recommendation TEXT,
		time TEXT NOT NULL,
		acknowledged INTEGER DEFAULT 0,
		FOREIGN KEY (session_id) REFERENCES sessions(id)
	);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *Storage) Write(ctx context.Context, table string, data interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch v := data.(type) {
	case map[string]interface{}:
		columns := make([]string, 0, len(v))
		placeholders := make([]string, 0, len(v))
		values := make([]interface{}, 0, len(v))

		for k, val := range v {
			columns = append(columns, k)
			placeholders = append(placeholders, "?")
			values = append(values, val)
		}

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			table, joinColumns(columns), joinPlaceholders(len(columns)))

		_, err := s.db.ExecContext(ctx, query, values...)
		return err
	}

	return fmt.Errorf("unsupported data type: %T", data)
}

func (s *Storage) WriteBatch(ctx context.Context, table string, dataList []interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(dataList) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, buildBatchInsertSQL(table, dataList[0]))
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, data := range dataList {
		if v, ok := data.(map[string]interface{}); ok {
			values := make([]interface{}, 0, len(v))
			for _, val := range v {
				values = append(values, val)
			}
			_, err := stmt.ExecContext(ctx, values...)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (s *Storage) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := []map[string]interface{}{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	return results, rows.Err()
}

func (s *Storage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Close()
}

func (s *Storage) IntegrityCheck() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow("PRAGMA integrity_check")
	var result string
	if err := row.Scan(&result); err != nil {
		return err
	}

	if result != "ok" {
		return fmt.Errorf("integrity check failed: %s", result)
	}
	return nil
}

func (s *Storage) Vacuum() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec("VACUUM")
	return err
}

func joinColumns(columns []string) string {
	result := ""
	for i, col := range columns {
		if i > 0 {
			result += ", "
		}
		result += col
	}
	return result
}

func joinPlaceholders(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		if i > 0 {
			result += ", "
		}
		result += "?"
	}
	return result
}

func buildBatchInsertSQL(table string, data interface{}) string {
	v, ok := data.(map[string]interface{})
	if !ok {
		return ""
	}

	columns := make([]string, 0, len(v))
	for k := range v {
		columns = append(columns, k)
	}

	placeholders := ""
	for i := range columns {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "?"
	}

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table, joinColumns(columns), placeholders)
}
