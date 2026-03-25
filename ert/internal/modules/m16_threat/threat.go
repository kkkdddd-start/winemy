//go:build windows

package m16_threat

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type ThreatModule struct {
	ctx           context.Context
	storage       registry.Storage
	malProcesses  []model.ProcessDTO
	suspiciousNet []model.NetworkConnDTO
	sensitiveActs []map[string]interface{}
	threatRules   []model.AlertRule
}

func New() *ThreatModule {
	return &ThreatModule{
		malProcesses:  []model.ProcessDTO{},
		suspiciousNet: []model.NetworkConnDTO{},
		sensitiveActs: []map[string]interface{}{},
		threatRules:   []model.AlertRule{},
	}
}

func (m *ThreatModule) ID() int       { return 16 }
func (m *ThreatModule) Name() string  { return "threat" }
func (m *ThreatModule) Priority() int { return 0 }

func (m *ThreatModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	m.loadThreatRules()
	return nil
}

func (m *ThreatModule) loadThreatRules() {
	m.threatRules = []model.AlertRule{
		{ID: "rule1", Name: "Suspicious Process", ModuleID: 16, Severity: model.RiskHigh, Enabled: true},
		{ID: "rule2", Name: "Known Malware Hash", ModuleID: 16, Severity: model.RiskCritical, Enabled: true},
		{ID: "rule3", Name: "Suspicious Network", ModuleID: 16, Severity: model.RiskMedium, Enabled: true},
		{ID: "rule4", Name: "Sensitive Behavior", ModuleID: 16, Severity: model.RiskMedium, Enabled: true},
	}
}

func (m *ThreatModule) Collect(ctx context.Context) error {
	if err := m.detectMaliciousProcesses(); err != nil {
		return err
	}
	if err := m.detectSuspiciousNetwork(); err != nil {
		return err
	}
	m.detectSensitiveBehaviors()
	return nil
}

func (m *ThreatModule) detectMaliciousProcesses() error {
	procs, err := process.Processes()
	if err != nil {
		return err
	}

	m.malProcesses = []model.ProcessDTO{}
	suspiciousNames := []string{
		"mimikatz", "pwdump", "procdump", "lsass", "cmd",
		"powershell", "nc", "netcat", "psexec", "wce",
		"gsecdump", "fgdump", "hashdump", "metasploit",
	}

	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}

		nameLower := strings.ToLower(name)
		for _, sus := range suspiciousNames {
			if strings.Contains(nameLower, sus) {
				cpu, _ := p.CPUPercent()
				mem, _ := p.MemoryInfo()
				createTime, _ := p.CreateTime()
				ppid, _ := p.Ppid()

				dto := model.ProcessDTO{
					PID:         uint32(p.Pid),
					Name:        name,
					Path:        "",
					CommandLine: "",
					User:        "",
					CPU:         cpu,
					Memory:      mem.RSS,
					StartTime:   time.Unix(int64(createTime)/1000, 0),
					RiskLevel:   model.RiskHigh,
				}
				_ = ppid
				m.malProcesses = append(m.malProcesses, dto)
				break
			}
		}
	}
	return nil
}

func (m *ThreatModule) detectSuspiciousNetwork() error {
	conns, err := net.Connections("tcp")
	if err != nil {
		return err
	}

	m.suspiciousNet = []model.NetworkConnDTO{}
	suspiciousPorts := map[uint32]bool{
		4444:  true,
		5555:  true,
		6666:  true,
		7777:  true,
		31337: true,
	}

	for _, c := range conns {
		if suspiciousPorts[c.Raddr.Port] {
			proto := "tcp"
			if c.Type == 2 {
				proto = "udp"
			}
			dto := model.NetworkConnDTO{
				PID:        uint32(c.Pid),
				Protocol:   proto,
				LocalAddr:  c.Laddr.IP,
				LocalPort:  uint16(c.Laddr.Port),
				RemoteAddr: c.Raddr.IP,
				RemotePort: uint16(c.Raddr.Port),
				State:      c.Status,
				RiskLevel:  model.RiskHigh,
			}
			m.suspiciousNet = append(m.suspiciousNet, dto)
		}

		if c.Status == "listening" && c.Raddr.Port == 0 {
			continue
		}

		remoteStr := c.Raddr.IP
		if remoteStr == "127.0.0.1" || remoteStr == "0.0.0.0" {
			continue
		}

		suspiciousIPs := []string{"192.0.2.", "198.51.100.", "203.0.113."}
		for _, sip := range suspiciousIPs {
			if strings.HasPrefix(remoteStr, sip) {
				proto := "tcp"
				if c.Type == 2 {
					proto = "udp"
				}
				dto := model.NetworkConnDTO{
					PID:        uint32(c.Pid),
					Protocol:   proto,
					LocalAddr:  c.Laddr.IP,
					LocalPort:  uint16(c.Laddr.Port),
					RemoteAddr: c.Raddr.IP,
					RemotePort: uint16(c.Raddr.Port),
					State:      c.Status,
					RiskLevel:  model.RiskMedium,
				}
				m.suspiciousNet = append(m.suspiciousNet, dto)
				break
			}
		}
	}
	return nil
}

func (m *ThreatModule) detectSensitiveBehaviors() {
	m.sensitiveActs = []map[string]interface{}{}

	sensitivePatterns := []struct {
		name        string
		pattern     string
		description string
		severity    model.RiskLevel
	}{
		{"PowerShell Encoded Command", "enc", "PowerShell encoded command detected", model.RiskHigh},
		{"Registry Modification", "reg add", "Registry modification detected", model.RiskMedium},
		{"Service Creation", "sc create", "Windows service creation detected", model.RiskMedium},
		{"WMI Execution", "wmic ", "WMI command execution detected", model.RiskMedium},
		{"Scheduled Task", "schtasks", "Scheduled task creation detected", model.RiskMedium},
	}

	procs, _ := process.Processes()
	for _, p := range procs {
		cmdline, err := p.CmdlineSlice()
		if err != nil || len(cmdline) == 0 {
			continue
		}
		fullCmd := strings.ToLower(strings.Join(cmdline, " "))

		for _, pattern := range sensitivePatterns {
			if strings.Contains(fullCmd, pattern.pattern) {
				name, _ := p.Name()
				m.sensitiveActs = append(m.sensitiveActs, map[string]interface{}{
					"pid":         p.Pid,
					"name":        name,
					"pattern":     pattern.name,
					"description": pattern.description,
					"command":     strings.Join(cmdline, " "),
					"risk_level":  pattern.severity,
					"timestamp":   time.Now().Format(time.RFC3339),
				})
			}
		}
	}
}

func (m *ThreatModule) Stop() error {
	return nil
}

func (m *ThreatModule) GetData() ([]map[string]interface{}, error) {
	result := []map[string]interface{}{}

	for _, p := range m.malProcesses {
		result = append(result, map[string]interface{}{
			"type":       "malicious_process",
			"pid":        p.PID,
			"name":       p.Name,
			"path":       p.Path,
			"cpu":        p.CPU,
			"memory":     p.Memory,
			"start_time": p.StartTime.Format(time.RFC3339),
			"risk_level": p.RiskLevel,
		})
	}

	for _, n := range m.suspiciousNet {
		result = append(result, map[string]interface{}{
			"type":        "suspicious_network",
			"pid":         n.PID,
			"protocol":    n.Protocol,
			"local_addr":  n.LocalAddr,
			"local_port":  n.LocalPort,
			"remote_addr": n.RemoteAddr,
			"remote_port": n.RemotePort,
			"state":       n.State,
			"risk_level":  n.RiskLevel,
		})
	}

	for _, a := range m.sensitiveActs {
		result = append(result, map[string]interface{}{
			"type":        "sensitive_behavior",
			"pid":         a["pid"],
			"name":        a["name"],
			"pattern":     a["pattern"],
			"description": a["description"],
			"command":     a["command"],
			"risk_level":  a["risk_level"],
			"timestamp":   a["timestamp"],
		})
	}

	return result, nil
}

func (m *ThreatModule) GetThreatRules() []model.AlertRule {
	return m.threatRules
}

func (m *ThreatModule) GetMaliciousProcesses() []model.ProcessDTO {
	return m.malProcesses
}

func (m *ThreatModule) GetSuspiciousNetwork() []model.NetworkConnDTO {
	return m.suspiciousNet
}

func (m *ThreatModule) DetectRemoteThread() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	cmd := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
Get-WinEvent -FilterHashtable @{LogName='Security';ID=8} -MaxEvents 100 -ErrorAction SilentlyContinue | ForEach-Object {
    $xml = [xml]$_.ToXml()
    $eventData = $xml.Event.EventData.Data
    $targetPID = ($eventData | Where-Object { $_.Name -eq 'TargetProcessId' }).'#text'
    $sourcePID = ($eventData | Where-Object { $_.Name -eq 'SourceProcessId' }).'#text'
    $newThread = ($eventData | Where-Object { $_.Name -eq 'NewThreadId' }).'#text'
    if($targetPID -and $sourcePID) {
        $sourceProc = Get-Process -Id $sourcePID -ErrorAction SilentlyContinue
        $targetProc = Get-Process -Id $targetPID -ErrorAction SilentlyContinue
        Write-Output ($sourceProc.ProcessName + '|' + $targetProc.ProcessName + '|' + $sourcePID + '|' + $targetPID + '|' + $_.TimeCreated.ToString('yyyy-MM-dd HH:mm:ss'))
    }
}`)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to detect remote thread: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			results = append(results, map[string]interface{}{
				"source_process": parts[0],
				"target_process": parts[1],
				"source_pid":     parts[2],
				"target_pid":     parts[3],
				"timestamp":      parts[4],
				"type":           "remote_thread",
				"risk_level":     model.RiskHigh,
			})
		}
	}

	cmd2 := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
$procs = Get-Process
foreach($p in $procs) {
    $threads = $p.Threads
    foreach($t in $threads) {
        if($t.WaitReason -eq 'Suspended' -and $t.ThreadState -eq 'Wait') {
            Write-Output ($p.ProcessName + '|' + $p.Id.ToString() + '|Suspended')
        }
    }
}`)

	output2, err := cmd2.Output()
	if err == nil {
		lines2 := strings.Split(string(output2), "\n")
		for _, line := range lines2 {
			parts := strings.Split(strings.TrimSpace(line), "|")
			if len(parts) >= 3 && parts[2] == "Suspended" {
				results = append(results, map[string]interface{}{
					"source_process": parts[0],
					"target_pid":     parts[1],
					"type":           "suspicious_thread",
					"risk_level":     model.RiskMedium,
				})
			}
		}
	}

	return results, nil
}

func (m *ThreatModule) LoadThreatDB(dbPath string) error {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("threat database file not found: %w", err)
	}

	content, err := os.ReadFile(dbPath)
	if err != nil {
		return fmt.Errorf("failed to read threat database: %w", err)
	}

	var threatDB struct {
		Rules []struct {
			ID       string   `json:"id"`
			Name     string   `json:"name"`
			Patterns []string `json:"patterns"`
			Severity string   `json:"severity"`
		} `json:"rules"`
	}

	if err := json.Unmarshal(content, &threatDB); err != nil {
		return fmt.Errorf("failed to parse threat database: %w", err)
	}

	for _, rule := range threatDB.Rules {
		m.threatRules = append(m.threatRules, model.AlertRule{
			ID:      rule.ID,
			Name:    rule.Name,
			Enabled: true,
		})
	}

	return nil
}

func (m *ThreatModule) MatchYARA(rulesPath string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	if _, err := os.Stat(rulesPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("YARA rules file not found: %w", err)
	}

	psRulePath := rulesPath + ".ps1"
	scriptContent := fmt.Sprintf(`
$ErrorActionPreference='SilentlyContinue'
$rules = Get-Content '%s' -Raw
$processes = Get-Process | Select-Object Name, Path, Id

$yaraMatches = @()

if($rules -match 'rule\s+(\w+)\s*\{') {
    $matches = [regex]::Matches($rules, 'rule\s+(\w+)\s*\{[^}]*strings[^}]*\}')
    foreach($m in $matches) {
        $ruleName = $m.Groups[1].Value
        $yaraMatches += @{
            RuleName = $ruleName
            Matched = $false
            ProcessName = ''
            ProcessID = 0
        }
    }
}

foreach($p in $processes) {
    if($p.Path) {
        $content = Get-Content $p.Path -Raw -ErrorAction SilentlyContinue
        if($content -and $rules) {
            if($rules -match $p.Name -or $content -match 'mimikatz|password|credential') {
                foreach($ym in $yaraMatches) {
                    if($ym.RuleName -match $p.Name) {
                        $ym.Matched = $true
                        $ym.ProcessName = $p.Name
                        $ym.ProcessID = $p.Id
                    }
                }
            }
        }
    }
}

$yaraMatches | Where-Object { $_.Matched } | ForEach-Object { Write-Output ($_.RuleName + '|' + $_.ProcessName + '|' + $_.ProcessID.ToString()) }
`, rulesPath)

	if err := os.WriteFile(psRulePath, []byte(scriptContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write temporary script: %w", err)
	}
	defer os.Remove(psRulePath)

	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", psRulePath)
	output, err := cmd.Output()
	if err != nil {
		return results, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			results = append(results, map[string]interface{}{
				"rule_name":    parts[0],
				"process_name": parts[1],
				"process_id":   parts[2],
				"type":         "yara_match",
			})
		}
	}

	return results, nil
}

func (m *ThreatModule) DeduplicateAlerts() []model.AlertEvent {
	seen := make(map[string]bool)
	var unique []model.AlertEvent

	allAlerts := make([]model.AlertEvent, 0)
	for _, p := range m.malProcesses {
		allAlerts = append(allAlerts, model.AlertEvent{
			ID:       fmt.Sprintf("proc_%d_%s", p.PID, p.Name),
			RuleID:   "malicious_process",
			RuleName: "Malicious Process Detected",
			Severity: p.RiskLevel,
			Message:  fmt.Sprintf("Suspicious process: %s (PID: %d)", p.Name, p.PID),
			ModuleID: 16,
		})
	}
	for _, n := range m.suspiciousNet {
		allAlerts = append(allAlerts, model.AlertEvent{
			ID:       fmt.Sprintf("net_%d_%s_%d", n.PID, n.RemoteAddr, n.RemotePort),
			RuleID:   "suspicious_network",
			RuleName: "Suspicious Network Connection",
			Severity: n.RiskLevel,
			Message:  fmt.Sprintf("Suspicious connection from %s:%d to %s:%d", n.LocalAddr, n.LocalPort, n.RemoteAddr, n.RemotePort),
			ModuleID: 16,
		})
	}
	for _, a := range m.sensitiveActs {
		if name, ok := a["name"].(string); ok {
			allAlerts = append(allAlerts, model.AlertEvent{
				ID:       fmt.Sprintf("beh_%s_%v", name, a["pid"]),
				RuleID:   "sensitive_behavior",
				RuleName: "Sensitive Behavior Detected",
				Severity: model.RiskMedium,
				Message:  fmt.Sprintf("Sensitive behavior: %v", a["description"]),
				ModuleID: 16,
			})
		}
	}

	for _, alert := range allAlerts {
		key := fmt.Sprintf("%s_%s_%v", alert.RuleID, alert.RuleName, alert.Message)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, alert)
		}
	}

	return unique
}

func (m *ThreatModule) Search(keyword string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	keywordLower := strings.ToLower(keyword)

	for _, p := range m.malProcesses {
		if strings.Contains(strings.ToLower(p.Name), keywordLower) ||
			strings.Contains(strings.ToLower(p.Path), keywordLower) {
			results = append(results, map[string]interface{}{
				"type":       "malicious_process",
				"name":       p.Name,
				"pid":        p.PID,
				"path":       p.Path,
				"risk_level": p.RiskLevel,
			})
		}
	}

	for _, n := range m.suspiciousNet {
		if strings.Contains(strings.ToLower(n.RemoteAddr), keywordLower) ||
			strings.Contains(strings.ToLower(n.ProcessName), keywordLower) {
			results = append(results, map[string]interface{}{
				"type":        "suspicious_network",
				"process":     n.ProcessName,
				"remote_addr": n.RemoteAddr,
				"remote_port": n.RemotePort,
				"risk_level":  n.RiskLevel,
			})
		}
	}

	for _, a := range m.sensitiveActs {
		if desc, ok := a["description"].(string); ok {
			if strings.Contains(strings.ToLower(desc), keywordLower) {
				results = append(results, map[string]interface{}{
					"type":        "sensitive_behavior",
					"description": desc,
					"risk_level":  a["risk_level"],
				})
			}
		}
	}

	return results, nil
}
