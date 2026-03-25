package m21_wmic

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type WmiModule struct {
	ctx            context.Context
	storage        registry.Storage
	commandHistory []map[string]interface{}
	suspiciousCmds []map[string]interface{}
}

func New() *WmiModule {
	return &WmiModule{
		commandHistory: []map[string]interface{}{},
		suspiciousCmds: []map[string]interface{}{},
	}
}

func (m *WmiModule) ID() int       { return 21 }
func (m *WmiModule) Name() string  { return "wmic" }
func (m *WmiModule) Priority() int { return 1 }

func (m *WmiModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	return nil
}

func (m *WmiModule) Collect(ctx context.Context) error {
	if err := m.collectWMICHistory(); err != nil {
		return err
	}
	m.detectSuspiciousCommands()
	return nil
}

func (m *WmiModule) collectWMICHistory() error {
	m.commandHistory = []map[string]interface{}{}

	historyPaths := []string{
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Explorer\RunMRU`,
		`HKCU\Software\Microsoft\Windows\PowerShell\PSReadline\ConsoleHost_History`,
	}

	for _, path := range historyPaths {
		cmd := exec.Command("reg", "query", path)
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "REG_") {
				parts := strings.SplitN(line, "REG_", 2)
				if len(parts) < 2 {
					continue
				}
				subParts := strings.SplitN(parts[1], "=", 2)
				if len(subParts) < 2 {
					continue
				}
				value := strings.TrimSpace(subParts[1])
				if value == "" || value == "(value not set)" {
					continue
				}

				cmdType := "registry"
				entry := map[string]interface{}{
					"type":       cmdType,
					"path":       path,
					"command":    value,
					"timestamp":  time.Now().Format(time.RFC3339),
					"risk_level": model.RiskLow,
				}

				if strings.HasPrefix(value, "wmic ") || strings.HasPrefix(value, "wmic") {
					entry["type"] = "wmic"
					cmdType = "wmic"
				}

				m.commandHistory = append(m.commandHistory, entry)
			}
		}
	}

	psHistoryPath := os.Getenv("APPDATA") + `\\Microsoft\\Windows\\PowerShell\\PSReadline\\ConsoleHost_history.txt`
	cmd := exec.Command("type", psHistoryPath)
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.HasPrefix(strings.ToLower(line), "wmic") {
				m.commandHistory = append(m.commandHistory, map[string]interface{}{
					"type":       "wmic",
					"source":     "powershell_history",
					"command":    line,
					"timestamp":  time.Now().Format(time.RFC3339),
					"risk_level": model.RiskLow,
				})
			}
		}
	}

	return nil
}

func (m *WmiModule) detectSuspiciousCommands() error {
	m.suspiciousCmds = []map[string]interface{}{}

	suspiciousPatterns := []struct {
		pattern     string
		description string
		severity    model.RiskLevel
	}{
		{"process call create", "Process creation via WMIC", model.RiskMedium},
		{"/node:", "Remote WMIC execution", model.RiskHigh},
		{"shadowcopy", "Volume shadow copy manipulation", model.RiskHigh},
		{"delete shadowcopy", "Shadow copy deletion", model.RiskCritical},
		{"firewall", "Firewall modification", model.RiskHigh},
		{"netsh", "Network configuration change", model.RiskMedium},
		{"useraccount", "User account manipulation", model.RiskHigh},
		{"group", "Group manipulation", model.RiskMedium},
		{"process call create \"cmd", "Command execution via WMIC", model.RiskCritical},
		{"call \"C:", "Suspicious command execution", model.RiskHigh},
		{"share", "Share manipulation", model.RiskMedium},
		{"service", "Service manipulation", model.RiskMedium},
		{"registry", "Registry manipulation", model.RiskMedium},
		{"alert", "Alert configuration", model.RiskLow},
		{"eventfilter", "WMI event filter creation", model.RiskHigh},
		{"consumer", "WMI consumer creation", model.RiskHigh},
		{"binding", "WMI binding", model.RiskMedium},
	}

	for _, entry := range m.commandHistory {
		cmd := entry["command"].(string)
		cmdLower := strings.ToLower(cmd)

		for _, pattern := range suspiciousPatterns {
			if strings.Contains(cmdLower, strings.ToLower(pattern.pattern)) {
				m.suspiciousCmds = append(m.suspiciousCmds, map[string]interface{}{
					"type":        "suspicious_wmic",
					"command":     cmd,
					"pattern":     pattern.pattern,
					"description": pattern.description,
					"source":      entry["source"],
					"risk_level":  pattern.severity,
					"detected_at": time.Now().Format(time.RFC3339),
				})
				break
			}
		}
	}

	return nil
}

func (m *WmiModule) BatchDetect(directory string) ([]map[string]interface{}, error) {
	results := []map[string]interface{}{}

	cmd := exec.Command("cmd", "/c", fmt.Sprintf("dir /s /b %s\\*.log %s\\*.txt 2>nul", directory, directory))
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	files := strings.Split(string(output), "\n")
	for _, file := range files {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}

		contentCmd := exec.Command("type", file)
		content, err := contentCmd.Output()
		if err != nil {
			continue
		}

		contentStr := string(content)
		for _, entry := range m.commandHistory {
			cmd := entry["command"].(string)
			if strings.Contains(contentStr, cmd) {
				results = append(results, map[string]interface{}{
					"file":       file,
					"matched":    cmd,
					"risk_level": entry["risk_level"],
				})
			}
		}
	}

	return results, nil
}

func (m *WmiModule) Stop() error {
	return nil
}

func (m *WmiModule) GetData() ([]map[string]interface{}, error) {
	result := []map[string]interface{}{}

	for _, h := range m.commandHistory {
		result = append(result, map[string]interface{}{
			"category":   "wmic_history",
			"type":       h["type"],
			"command":    h["command"],
			"source":     h["source"],
			"timestamp":  h["timestamp"],
			"risk_level": h["risk_level"],
		})
	}

	for _, s := range m.suspiciousCmds {
		result = append(result, map[string]interface{}{
			"category":    "suspicious_command",
			"command":     s["command"],
			"pattern":     s["pattern"],
			"description": s["description"],
			"source":      s["source"],
			"risk_level":  s["risk_level"],
			"detected_at": s["detected_at"],
		})
	}

	return result, nil
}

func (m *WmiModule) GetCommandHistory() []map[string]interface{} {
	return m.commandHistory
}

func (m *WmiModule) GetSuspiciousCommands() []map[string]interface{} {
	return m.suspiciousCmds
}
