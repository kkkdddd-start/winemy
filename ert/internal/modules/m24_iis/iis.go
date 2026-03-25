//go:build windows

package m24_iis

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yourname/ert/internal/registry"
)

type IISModule struct {
	ctx        context.Context
	storage    registry.Storage
	iisLogs    []map[string]interface{}
	apacheLogs []map[string]interface{}
	nginxLogs  []map[string]interface{}
	sqlLogs    []map[string]interface{}
	tomcatLogs []map[string]interface{}
	logPaths   map[string]string
}

func New() *IISModule {
	return &IISModule{
		iisLogs:    []map[string]interface{}{},
		apacheLogs: []map[string]interface{}{},
		nginxLogs:  []map[string]interface{}{},
		sqlLogs:    []map[string]interface{}{},
		tomcatLogs: []map[string]interface{}{},
		logPaths:   map[string]string{},
	}
}

func (m *IISModule) ID() int       { return 24 }
func (m *IISModule) Name() string  { return "iis" }
func (m *IISModule) Priority() int { return 1 }

func (m *IISModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	m.detectLogPaths()
	return nil
}

func (m *IISModule) detectLogPaths() {
	m.logPaths = map[string]string{}

	cmd := exec.Command("powershell", "-Command",
		`$iisLogPaths = @(); try { Import-Module WebAdministration -ErrorAction Stop; Get-WebConfigurationProperty -Filter '//sites/site' -Name logFile.directory -PSPath 'IIS:\' -ErrorAction SilentlyContinue | ForEach-Object { $iisLogPaths += $_.Value } } catch { }; if($iisLogPaths.Count -eq 0) { $iisLogPaths += 'C:\inetpub\logs\LogFiles' }; $iisLogPaths | ConvertTo-Json`)
	output, err := cmd.Output()
	if err == nil {
		var paths []string
		if err := json.Unmarshal(output, &paths); err != nil {
			var single string
			if err := json.Unmarshal(output, &single); err == nil {
				paths = []string{single}
			}
		}
		for _, p := range paths {
			if p != "" {
				p = strings.Replace(p, "%SystemDrive%", "C:", 1)
				m.logPaths["iis_detected"] = p
			}
		}
	}

	if m.logPaths["iis_detected"] == "" {
		m.logPaths["iis"] = `C:\inetpub\logs\LogFiles`
	} else {
		m.logPaths["iis"] = m.logPaths["iis_detected"]
	}

	sqlLogPath := m.detectSQLServerLogPath()
	if sqlLogPath != "" {
		m.logPaths["sql_server"] = sqlLogPath
	} else {
		m.logPaths["sql_server"] = `C:\Program Files\Microsoft SQL Server\MSSQL\LOG`
	}

	apacheLogPaths := m.detectApacheLogPath()
	if len(apacheLogPaths) > 0 {
		m.logPaths["apache"] = apacheLogPaths[0]
		if len(apacheLogPaths) > 1 {
			m.logPaths["apache_xampp"] = apacheLogPaths[1]
		}
	} else {
		m.logPaths["apache"] = `C:\Apache24\logs`
		m.logPaths["apache_xampp"] = `C:\xampp\apache\logs`
	}

	nginxLogPath := m.detectNginxLogPath()
	if nginxLogPath != "" {
		m.logPaths["nginx"] = nginxLogPath
	} else {
		m.logPaths["nginx"] = `C:\nginx\logs`
	}

	m.logPaths["tomcat"] = `C:\Tomcat\logs`

	programFiles := os.Getenv("ProgramFiles")
	if programFiles != "" {
		m.logPaths["iis_w3svc"] = programFiles + `\IIS\logs`
	}

	programFilesX86 := os.Getenv("ProgramFiles(x86)")
	if programFilesX86 != "" {
		m.logPaths["iis_w3svc_x86"] = programFilesX86 + `\IIS\logs`
	}
}

func (m *IISModule) detectSQLServerLogPath() string {
	cmd := exec.Command("powershell", "-Command",
		`$ErrorActionPreference = 'SilentlyContinue'; $sqlPaths = @(); Get-ChildItem 'HKLM:\SOFTWARE\Microsoft\Microsoft SQL Server' -Recurse | ForEach-Object { $key = $_; try { $instDir = (Get-ItemProperty -Path $key.PSPath -Name InstallDir -ErrorAction SilentlyContinue).InstallDir; if($instDir) { $logPath = $instDir + '\Log'; if(Test-Path $logPath) { $sqlPaths += $logPath } } } catch { } }; Get-ChildItem 'HKLM:\SOFTWARE\Wow6432Node\Microsoft\Microsoft SQL Server' -Recurse | ForEach-Object { $key = $_; try { $instDir = (Get-ItemProperty -Path $key.PSPath -Name InstallDir -ErrorAction SilentlyContinue).InstallDir; if($instDir) { $logPath = $instDir + '\Log'; if(Test-Path $logPath) { $sqlPaths += $logPath } } } catch { } }; $sqlPaths | Select-Object -First 1`)
	output, err := cmd.Output()
	if err == nil {
		path := strings.TrimSpace(string(output))
		if path != "" && !strings.Contains(path, "Microsoft") {
			return path
		}
	}
	return ""
}

func (m *IISModule) detectApacheLogPath() []string {
	paths := []string{}

	programFiles := os.Getenv("ProgramFiles")
	if programFiles != "" {
		apachePath := programFiles + "\\Apache"
		if _, err := os.Stat(apachePath); err == nil {
			entries, _ := os.ReadDir(apachePath)
			for _, entry := range entries {
				if entry.IsDir() && strings.HasPrefix(entry.Name(), "Apache") {
					logPath := apachePath + "\\" + entry.Name() + "\\logs"
					if _, err := os.Stat(logPath); err == nil {
						paths = append(paths, logPath)
					}
				}
			}
		}
	}

	xamppPath := `C:\xampp\apache\logs`
	if _, err := os.Stat(xamppPath); err == nil {
		paths = append(paths, xamppPath)
	}

	return paths
}

func (m *IISModule) detectNginxLogPath() string {
	cmd := exec.Command("powershell", "-Command",
		`$ErrorActionPreference = 'SilentlyContinue'; $nginxPath = (Get-Command nginx -ErrorAction SilentlyContinue).Source; if($nginxPath) { $nginxDir = Split-Path $nginxPath; $logPath = $nginxDir + '\logs'; if(Test-Path $logPath) { Write-Output $logPath } else { Write-Output $nginxDir } } else { $commonPaths = @('C:\nginx\logs', 'C:\Program Files\nginx\logs', 'C:\Program Files (x86)\nginx\logs'); foreach($p in $commonPaths) { if(Test-Path $p) { Write-Output $p; break } } }`)
	output, err := cmd.Output()
	if err == nil {
		path := strings.TrimSpace(string(output))
		if path != "" && !strings.Contains(path, "nginx") {
			if strings.HasSuffix(path, `\logs`) || strings.HasSuffix(path, `\\logs`) {
				return path
			}
			return path + `\logs`
		}
	}
	return ""
}

func (m *IISModule) Collect(ctx context.Context) error {
	if err := m.collectIISLogs(); err != nil {
		return err
	}
	if err := m.collectApacheLogs(); err != nil {
		return err
	}
	if err := m.collectNginxLogs(); err != nil {
		return err
	}
	if err := m.collectSQLLogs(); err != nil {
		return err
	}
	if err := m.collectTomcatLogs(); err != nil {
		return err
	}
	return nil
}

func (m *IISModule) collectIISLogs() error {
	m.iisLogs = []map[string]interface{}{}

	logDir := m.logPaths["iis"]
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		return nil
	}

	w3svcPattern := regexp.MustCompile(`(?i)u_ex\d{8}\.log`)
	sftpPattern := regexp.MustCompile(`(?i)s_ex\d{8}\.log`)

	err := filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		if w3svcPattern.MatchString(info.Name()) || sftpPattern.MatchString(info.Name()) {
			logs, err := m.parseIISLogFile(path)
			if err == nil {
				m.iisLogs = append(m.iisLogs, logs...)
			}
		}
		return nil
	})

	return err
}

func (m *IISModule) parseIISLogFile(filePath string) ([]map[string]interface{}, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var headers []string
	var logs []map[string]interface{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			if strings.HasPrefix(line, "#Fields:") {
				headers = strings.Split(strings.TrimPrefix(line, "#Fields:"), " ")
			}
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < len(headers) {
			continue
		}

		entry := map[string]interface{}{
			"source": filepath.Base(filePath),
			"raw":    line,
		}

		for i, h := range headers {
			if i < len(fields) {
				entry[h] = fields[i]
			}
		}

		entry["risk_level"] = m.assessIISLogRisk(entry)

		logs = append(logs, entry)
	}

	return logs, nil
}

func (m *IISModule) assessIISLogRisk(entry map[string]interface{}) int {
	if sc, ok := entry["sc-status"].(string); ok {
		if sc == "401" || sc == "403" || sc == "500" {
			return 2
		}
	}

	if uri, ok := entry["cs-uri-stem"].(string); ok {
		suspicious := []string{".asp", ".aspx", ".jsp", ".php", ".cmd", ".exe", ".bat", ".ps1"}
		for _, s := range suspicious {
			if strings.Contains(strings.ToLower(uri), s) {
				return 2
			}
		}
	}

	if csUriQuery, ok := entry["cs-uri-query"].(string); ok {
		suspicious := []string{"union", "select", "exec", "xp_", "sp_"}
		for _, s := range suspicious {
			if strings.Contains(strings.ToLower(csUriQuery), s) {
				return 3
			}
		}
	}

	return 0
}

func (m *IISModule) collectApacheLogs() error {
	m.apacheLogs = []map[string]interface{}{}

	apachePaths := []string{m.logPaths["apache"], m.logPaths["apache_xampp"]}

	for _, logDir := range apachePaths {
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}

			if strings.HasSuffix(info.Name(), ".log") {
				logs, err := m.parseApacheLogFile(path)
				if err == nil {
					m.apacheLogs = append(m.apacheLogs, logs...)
				}
			}
			return nil
		})

		if err == nil {
			break
		}
	}

	return nil
}

func (m *IISModule) parseApacheLogFile(filePath string) ([]map[string]interface{}, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var logs []map[string]interface{}

	accessLogPattern := regexp.MustCompile(`^(\S+) (\S+) (\S+) \[([^\]]+)\] "([^"]*)" (\d+) (\d+) "([^"]*)" "([^"]*)"`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		matches := accessLogPattern.FindStringSubmatch(line)
		if len(matches) > 0 {
			statusCode := 0
			fmt.Sscanf(matches[6], "%d", &statusCode)

			logs = append(logs, map[string]interface{}{
				"source":     filepath.Base(filePath),
				"ip":         matches[1],
				"user":       matches[3],
				"timestamp":  matches[4],
				"request":    matches[5],
				"status":     statusCode,
				"size":       matches[7],
				"referer":    matches[8],
				"user_agent": matches[9],
				"raw":        line,
				"risk_level": m.assessApacheLogRisk(statusCode, matches[5]),
			})
		} else {
			logs = append(logs, map[string]interface{}{
				"source":     filepath.Base(filePath),
				"raw":        line,
				"risk_level": 0,
			})
		}
	}

	return logs, nil
}

func (m *IISModule) assessApacheLogRisk(statusCode int, request string) int {
	if statusCode >= 400 {
		return 2
	}

	suspicious := []string{"../", "etc/passwd", "union select", "eval(", "base64_"}
	for _, s := range suspicious {
		if strings.Contains(strings.ToLower(request), s) {
			return 3
		}
	}

	return 0
}

func (m *IISModule) collectNginxLogs() error {
	m.nginxLogs = []map[string]interface{}{}

	logDir := m.logPaths["nginx"]
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		return nil
	}

	err := filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(info.Name(), ".log") {
			logs, err := m.parseNginxLogFile(path)
			if err == nil {
				m.nginxLogs = append(m.nginxLogs, logs...)
			}
		}
		return nil
	})

	return err
}

func (m *IISModule) parseNginxLogFile(filePath string) ([]map[string]interface{}, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var logs []map[string]interface{}

	nginxPattern := regexp.MustCompile(`^(\S+) - (\S+) \[([^\]]+)\] "([^"]*)" (\d+) (\d+) "([^"]*)" "([^"]*)"`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := nginxPattern.FindStringSubmatch(line)
		if len(matches) > 0 {
			statusCode := 0
			fmt.Sscanf(matches[5], "%d", &statusCode)

			logs = append(logs, map[string]interface{}{
				"source":     filepath.Base(filePath),
				"ip":         matches[1],
				"user":       matches[2],
				"timestamp":  matches[3],
				"request":    matches[4],
				"status":     statusCode,
				"size":       matches[6],
				"referer":    matches[7],
				"user_agent": matches[8],
				"raw":        line,
				"risk_level": m.assessApacheLogRisk(statusCode, matches[4]),
			})
		}
	}

	return logs, nil
}

func (m *IISModule) collectSQLLogs() error {
	m.sqlLogs = []map[string]interface{}{}

	logDir := m.logPaths["sql_server"]
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		cmd := exec.Command("powershell", "-Command", "Get-WinEvent -ListLog *SQL* | Select-Object -First 1 -ExpandProperty LogName")
		output, err := cmd.Output()
		if err == nil {
			logName := strings.TrimSpace(string(output))
			if logName != "" {
				return m.collectSQLLogsFromEventLog(logName)
			}
		}
		return nil
	}

	return filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(info.Name(), ".log") || strings.HasSuffix(info.Name(), ".trc") {
			logs, err := m.parseSQLLogFile(path)
			if err == nil {
				m.sqlLogs = append(m.sqlLogs, logs...)
			}
		}
		return nil
	})
}

func (m *IISModule) collectSQLLogsFromEventLog(logName string) error {
	cmd := exec.Command("powershell", "-Command", fmt.Sprintf(`Get-WinEvent -LogName "%s" -MaxEvents 100 | Select-Object TimeCreated, Id, LevelDisplayName, Message | ConvertTo-Json`, logName))
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var events []map[string]interface{}
	if err := json.Unmarshal(output, &events); err != nil {
		var single map[string]interface{}
		if err := json.Unmarshal(output, &single); err == nil {
			events = []map[string]interface{}{single}
		}
	}

	for _, e := range events {
		m.sqlLogs = append(m.sqlLogs, map[string]interface{}{
			"source":     logName,
			"timestamp":  e["TimeCreated"],
			"event_id":   e["Id"],
			"level":      e["LevelDisplayName"],
			"message":    e["Message"],
			"risk_level": 1,
		})
	}

	return nil
}

func (m *IISModule) parseSQLLogFile(filePath string) ([]map[string]interface{}, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var logs []map[string]interface{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "---") {
			continue
		}

		risk := 0
		suspicious := []string{"error", "fail", "denied", "permission", "login failed"}
		for _, s := range suspicious {
			if strings.Contains(strings.ToLower(line), s) {
				risk = 2
				break
			}
		}

		logs = append(logs, map[string]interface{}{
			"source":     filepath.Base(filePath),
			"message":    line,
			"risk_level": risk,
		})
	}

	return logs, nil
}

func (m *IISModule) collectTomcatLogs() error {
	m.tomcatLogs = []map[string]interface{}{}

	logDir := m.logPaths["tomcat"]
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		return nil
	}

	err := filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(info.Name(), ".log") || strings.HasSuffix(info.Name(), ".txt") {
			logs, err := m.parseTomcatLogFile(path)
			if err == nil {
				m.tomcatLogs = append(m.tomcatLogs, logs...)
			}
		}
		return nil
	})

	return err
}

func (m *IISModule) parseTomcatLogFile(filePath string) ([]map[string]interface{}, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var logs []map[string]interface{}

	tomcatPattern := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\.(\d+)\s+(\w+)\s+\[([^\]]+)\]\s+([^\s]+)\s+(\d+):\s+(.*)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := tomcatPattern.FindStringSubmatch(line)
		if len(matches) > 0 {
			level := matches[3]
			risk := 0
			if level == "SEVERE" || level == "ERROR" {
				risk = 3
			} else if level == "WARNING" || level == "WARN" {
				risk = 2
			}

			logs = append(logs, map[string]interface{}{
				"source":     filepath.Base(filePath),
				"timestamp":  matches[1],
				"level":      level,
				"thread":     matches[4],
				"class":      matches[5],
				"line":       matches[6],
				"message":    matches[7],
				"raw":        line,
				"risk_level": risk,
			})
		} else {
			logs = append(logs, map[string]interface{}{
				"source":     filepath.Base(filePath),
				"message":    line,
				"risk_level": 0,
			})
		}
	}

	return logs, nil
}

func (m *IISModule) Stop() error {
	return nil
}

func (m *IISModule) GetData() ([]map[string]interface{}, error) {
	result := []map[string]interface{}{}

	for _, l := range m.iisLogs {
		l["type"] = "iis"
		result = append(result, l)
	}

	for _, l := range m.apacheLogs {
		l["type"] = "apache"
		result = append(result, l)
	}

	for _, l := range m.nginxLogs {
		l["type"] = "nginx"
		result = append(result, l)
	}

	for _, l := range m.sqlLogs {
		l["type"] = "sql_server"
		result = append(result, l)
	}

	for _, l := range m.tomcatLogs {
		l["type"] = "tomcat"
		result = append(result, l)
	}

	return result, nil
}

func (m *IISModule) GetIISLogs() []map[string]interface{} {
	return m.iisLogs
}

func (m *IISModule) GetLogPaths() map[string]string {
	return m.logPaths
}
