package m18_autostart

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type AutostartModule struct {
	ctx            context.Context
	storage        registry.Storage
	regKeys        []model.RegistryKeyDTO
	startupFiles   []model.RegistryKeyDTO
	scheduledTasks []model.ScheduledTaskDTO
	services       []model.ServiceDTO
	wmiItems       []map[string]interface{}
}

func New() *AutostartModule {
	return &AutostartModule{
		regKeys:        []model.RegistryKeyDTO{},
		startupFiles:   []model.RegistryKeyDTO{},
		scheduledTasks: []model.ScheduledTaskDTO{},
		services:       []model.ServiceDTO{},
		wmiItems:       []map[string]interface{}{},
	}
}

func (m *AutostartModule) ID() int       { return 18 }
func (m *AutostartModule) Name() string  { return "autostart" }
func (m *AutostartModule) Priority() int { return 1 }

func (m *AutostartModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	return nil
}

func (m *AutostartModule) Collect(ctx context.Context) error {
	if err := m.collectRegistryAutostart(); err != nil {
		return err
	}
	if err := m.collectStartupFolder(); err != nil {
		return err
	}
	if err := m.collectScheduledTasks(); err != nil {
		return err
	}
	if err := m.collectServicesAutostart(); err != nil {
		return err
	}
	if err := m.collectWMIAutostart(); err != nil {
		return err
	}
	return nil
}

func (m *AutostartModule) collectRegistryAutostart() error {
	m.regKeys = []model.RegistryKeyDTO{}

	autostartPaths := []string{
		`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`,
		`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnce`,
		`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`,
		`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnce`,
		`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnceEx`,
		`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnceEx`,
		`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Explorer\Run`,
		`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Explorer\Run`,
		`HKLM\SYSTEM\CurrentControlSet\Services`,
		`HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon`,
		`HKCU\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon`,
	}

	for _, path := range autostartPaths {
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
				valueType := "REG_" + strings.TrimSpace(subParts[0])
				value := strings.TrimSpace(subParts[1])

				nameEnd := strings.Index(strings.TrimSpace(line), "\t")
				name := ""
				if nameEnd > 0 {
					name = strings.TrimSpace(line[:nameEnd])
				}

				riskLevel := model.RiskLow
				suspicious := []string{"temp", "appdata", "download", "public", "http"}
				for _, s := range suspicious {
					if strings.Contains(strings.ToLower(value), s) {
						riskLevel = model.RiskMedium
						break
					}
				}

				dto := model.RegistryKeyDTO{
					Path:      path,
					Name:      name,
					ValueType: valueType,
					Value:     value,
					Modified:  time.Now(),
					RiskLevel: riskLevel,
				}
				m.regKeys = append(m.regKeys, dto)
			}
		}
	}
	return nil
}

func (m *AutostartModule) collectStartupFolder() error {
	m.startupFiles = []model.RegistryKeyDTO{}

	startupPaths := []string{
		`C:\ProgramData\Microsoft\Windows\Start Menu\Programs\Startup`,
		`C:\Users\Default\Microsoft\Windows\Start Menu\Programs\Startup`,
		os.Getenv("APPDATA") + `\\Microsoft\\Windows\\Start Menu\\Programs\\Startup`,
	}

	homeDir := os.Getenv("USERPROFILE")
	if homeDir != "" {
		startupPaths = append(startupPaths, filepath.Join(homeDir, "AppData", "Roaming", "Microsoft", "Windows", "Start Menu", "Programs", "Startup"))
	}

	for _, path := range startupPaths {
		entries, err := os.ReadDir(path)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			fi, _ := entry.Info()
			dto := model.RegistryKeyDTO{
				Path:      path,
				Name:      entry.Name(),
				ValueType: "File",
				Value:     filepath.Join(path, entry.Name()),
				Modified:  fi.ModTime(),
				RiskLevel: model.RiskLow,
			}
			m.startupFiles = append(m.startupFiles, dto)
		}
	}
	return nil
}

func (m *AutostartModule) collectScheduledTasks() error {
	m.scheduledTasks = []model.ScheduledTaskDTO{}

	cmd := exec.Command("schtasks", "/query", "/fo", "LIST", "/v")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(output), "\n")
	var currentTask model.ScheduledTaskDTO

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "TaskName:") {
			if currentTask.Name != "" {
				m.scheduledTasks = append(m.scheduledTasks, currentTask)
			}
			currentTask = model.ScheduledTaskDTO{
				Name: strings.TrimPrefix(line, "TaskName:"),
			}
		} else if strings.HasPrefix(line, "Status:") {
			currentTask.State = strings.TrimPrefix(line, "Status:")
		} else if strings.HasPrefix(line, "Next Run Time:") {
			nextRun := strings.TrimPrefix(line, "Next Run Time:")
			if nextRun != "N/A" {
				t, _ := time.Parse("1/2/2006 3:04:05 PM", strings.TrimSpace(nextRun))
				currentTask.NextRunTime = t
			}
		} else if strings.HasPrefix(line, "Last Run Time:") {
			lastRun := strings.TrimPrefix(line, "Last Run Time:")
			if lastRun != "N/A" && lastRun != "Never" {
				t, _ := time.Parse("1/2/2006 3:04:05 PM", strings.TrimSpace(lastRun))
				currentTask.LastRunTime = t
			}
		}
	}

	if currentTask.Name != "" {
		m.scheduledTasks = append(m.scheduledTasks, currentTask)
	}

	for i := range m.scheduledTasks {
		if strings.Contains(strings.ToLower(m.scheduledTasks[i].Name), "update") ||
			strings.Contains(strings.ToLower(m.scheduledTasks[i].Name), "security") {
			m.scheduledTasks[i].RiskLevel = model.RiskLow
		} else {
			m.scheduledTasks[i].RiskLevel = model.RiskMedium
		}
	}

	return nil
}

func (m *AutostartModule) collectServicesAutostart() error {
	m.services = []model.ServiceDTO{}

	cmd := exec.Command("sc", "query", "state=", "all")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(output), "\n")
	var currentSvc model.ServiceDTO

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SERVICE_NAME:") {
			if currentSvc.Name != "" {
				m.services = append(m.services, currentSvc)
			}
			currentSvc = model.ServiceDTO{
				Name: strings.TrimPrefix(line, "SERVICE_NAME:"),
			}
		} else if strings.HasPrefix(line, "DISPLAY_NAME:") {
			currentSvc.DisplayName = strings.TrimPrefix(line, "DISPLAY_NAME:")
		} else if strings.HasPrefix(line, "STATE") {
			stateStr := strings.TrimPrefix(line, "STATE")
			stateStr = strings.TrimSpace(strings.Split(stateStr, ":")[0])
			currentSvc.Status = stateStr
		} else if strings.HasPrefix(line, "START_TYPE:") {
			startStr := strings.TrimPrefix(line, "START_TYPE:")
			startStr = strings.TrimSpace(strings.Split(startStr, ":")[0])
			currentSvc.StartType = startStr
		}
	}

	if currentSvc.Name != "" {
		m.services = append(m.services, currentSvc)
	}

	return nil
}

func (m *AutostartModule) collectWMIAutostart() error {
	m.wmiItems = []map[string]interface{}{}

	cmd := exec.Command("wmic", "startup", "list", "brief")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "Caption") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			m.wmiItems = append(m.wmiItems, map[string]interface{}{
				"command":    parts[0],
				"location":   parts[1],
				"risk_level": model.RiskMedium,
			})
		}
	}

	return nil
}

func (m *AutostartModule) Stop() error {
	return nil
}

func (m *AutostartModule) GetData() ([]map[string]interface{}, error) {
	result := []map[string]interface{}{}

	for _, r := range m.regKeys {
		result = append(result, map[string]interface{}{
			"type":       "registry",
			"path":       r.Path,
			"name":       r.Name,
			"value_type": r.ValueType,
			"value":      r.Value,
			"modified":   r.Modified.Format(time.RFC3339),
			"risk_level": r.RiskLevel,
		})
	}

	for _, s := range m.startupFiles {
		result = append(result, map[string]interface{}{
			"type":       "startup_folder",
			"path":       s.Path,
			"name":       s.Name,
			"value":      s.Value,
			"modified":   s.Modified.Format(time.RFC3339),
			"risk_level": s.RiskLevel,
		})
	}

	for _, t := range m.scheduledTasks {
		result = append(result, map[string]interface{}{
			"type":       "scheduled_task",
			"name":       t.Name,
			"state":      t.State,
			"last_run":   t.LastRunTime.Format(time.RFC3339),
			"next_run":   t.NextRunTime.Format(time.RFC3339),
			"risk_level": t.RiskLevel,
		})
	}

	for _, s := range m.services {
		if s.StartType == "AUTO_START" {
			result = append(result, map[string]interface{}{
				"type":         "service",
				"name":         s.Name,
				"display_name": s.DisplayName,
				"state":        s.Status,
				"start_type":   s.StartType,
				"risk_level":   s.RiskLevel,
			})
		}
	}

	for _, w := range m.wmiItems {
		result = append(result, map[string]interface{}{
			"type":       "wmi",
			"command":    w["command"],
			"location":   w["location"],
			"risk_level": w["risk_level"],
		})
	}

	return result, nil
}

func (m *AutostartModule) GetRegistryKeys() []model.RegistryKeyDTO {
	return m.regKeys
}

func (m *AutostartModule) GetStartupFiles() []model.RegistryKeyDTO {
	return m.startupFiles
}

func (m *AutostartModule) GetScheduledTasks() []model.ScheduledTaskDTO {
	return m.scheduledTasks
}
