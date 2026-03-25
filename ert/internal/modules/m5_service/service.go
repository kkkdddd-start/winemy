//go:build windows

package m5_service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type ServiceModule struct {
	ctx      context.Context
	storage  registry.Storage
	services []model.ServiceDTO
}

func New() *ServiceModule {
	return &ServiceModule{}
}

func (m *ServiceModule) ID() int       { return 5 }
func (m *ServiceModule) Name() string  { return "service" }
func (m *ServiceModule) Priority() int { return 1 }

func (m *ServiceModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	return nil
}

func (m *ServiceModule) Collect(ctx context.Context) error {
	m.services = []model.ServiceDTO{}

	cmd := exec.Command("powershell", "-Command",
		`Get-Service | Select-Object Name, DisplayName, Status, StartType | ConvertTo-Json`)

	output, err := cmd.Output()
	if err != nil {
		m.services = append(m.services, model.ServiceDTO{
			Name:        "Error",
			DisplayName: "Failed to enumerate services",
			Status:      "Unknown",
			StartType:   "Unknown",
			Path:        err.Error(),
			RiskLevel:   model.RiskLow,
		})
		return nil
	}

	var services []struct {
		Name        string `json:"Name"`
		DisplayName string `json:"DisplayName"`
		Status      int    `json:"Status"`
		StartType   int    `json:"StartType"`
	}

	if err := json.Unmarshal(output, &services); err != nil {
		var single struct {
			Name        string `json:"Name"`
			DisplayName string `json:"DisplayName"`
			Status      int    `json:"Status"`
			StartType   int    `json:"StartType"`
		}
		if err2 := json.Unmarshal(output, &single); err2 == nil {
			services = []struct {
				Name        string `json:"Name"`
				DisplayName string `json:"DisplayName"`
				Status      int    `json:"Status"`
				StartType   int    `json:"StartType"`
			}{single}
		}
	}

	for _, s := range services {
		status := "Unknown"
		switch s.Status {
		case 1:
			status = "Stopped"
		case 2:
			status = "Start Pending"
		case 3:
			status = "Stop Pending"
		case 4:
			status = "Running"
		case 5:
			status = "Continue Pending"
		case 6:
			status = "Pause Pending"
		case 7:
			status = "Paused"
		}

		startType := "Unknown"
		switch s.StartType {
		case 0:
			startType = "Boot"
		case 1:
			startType = "System"
		case 2:
			startType = "Automatic"
		case 3:
			startType = "Manual"
		case 4:
			startType = "Disabled"
		}

		path, deps, desc := m.getServiceDetails(s.Name)
		riskLevel := model.RiskLow
		if isSuspiciousService(path, s.Name) {
			riskLevel = model.RiskHigh
		}

		displayName := s.DisplayName
		if displayName == "" {
			displayName = s.Name
		}

		m.services = append(m.services, model.ServiceDTO{
			Name:         s.Name,
			DisplayName:  displayName,
			Status:       status,
			StartType:    startType,
			Path:         path,
			Dependencies: deps,
			Description:  desc,
			RiskLevel:    riskLevel,
		})
	}

	return nil
}

func (m *ServiceModule) getServiceDetails(serviceName string) (path, dependencies, description string) {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$svc = Get-WmiObject Win32_Service -Filter "Name='%s'" -ErrorAction SilentlyContinue; if($svc) { Write-Output "$($svc.PathName)|$($svc.Dependencies)|$($svc.Description)" }`, serviceName))
	output, err := cmd.Output()
	if err != nil {
		return "", "", ""
	}

	parts := strings.Split(strings.TrimSpace(string(output)), "|")
	if len(parts) >= 1 {
		path = cleanServicePath(parts[0])
	}
	if len(parts) >= 2 {
		dependencies = parts[1]
	}
	if len(parts) >= 3 {
		description = strings.TrimSpace(parts[2])
	}
	return path, dependencies, description
}

func cleanServicePath(path string) string {
	if path == "" {
		return ""
	}
	path = strings.TrimSpace(path)
	path = strings.Trim(path, "\"")
	return path
}

func isSuspiciousService(path, name string) bool {
	pathLower := strings.ToLower(path)
	nameLower := strings.ToLower(name)

	suspiciousPaths := []string{
		"temp",
		"tmp",
		"appdata",
		"downloads",
		"desktop",
		"public",
	}

	suspiciousNames := []string{
		"malware",
		"trojan",
		"virus",
		"backdoor",
		"keylog",
		"spyware",
		"pupy",
		"meterpreter",
		"cobalt",
	}

	for _, sp := range suspiciousPaths {
		if strings.Contains(pathLower, sp) {
			return true
		}
	}

	for _, sn := range suspiciousNames {
		if strings.Contains(nameLower, sn) {
			return true
		}
	}

	r := regexp.MustCompile(`^[A-Z]:\\[^\\]+\.exe$`)
	if !r.MatchString(path) && path != "" {
		if strings.Contains(pathLower, ".exe") && !strings.HasPrefix(pathLower, "c:\\windows") && !strings.HasPrefix(pathLower, "c:\\program") {
			return true
		}
	}

	return false
}

func (m *ServiceModule) Stop() error {
	return nil
}

func (m *ServiceModule) StartService(serviceName string) error {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`Start-Service -Name '%s' -ErrorAction Stop`, serviceName))
	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to start service %s: %w", serviceName, err)
	}
	return nil
}

func (m *ServiceModule) StopService(serviceName string) error {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`Stop-Service -Name '%s' -Force -ErrorAction Stop`, serviceName))
	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to stop service %s: %w", serviceName, err)
	}
	return nil
}

func (m *ServiceModule) RestartService(serviceName string) error {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`Restart-Service -Name '%s' -Force -ErrorAction Stop`, serviceName))
	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to restart service %s: %w", serviceName, err)
	}
	return nil
}

func (m *ServiceModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.services))
	for _, s := range m.services {
		result = append(result, map[string]interface{}{
			"name":         s.Name,
			"display_name": s.DisplayName,
			"status":       s.Status,
			"start_type":   s.StartType,
			"path":         s.Path,
			"dependencies": s.Dependencies,
			"description":  s.Description,
			"risk_level":   s.RiskLevel,
		})
	}
	return result, nil
}

func (m *ServiceModule) Search(keyword string) []model.ServiceDTO {
	results := []model.ServiceDTO{}
	keywordLower := strings.ToLower(keyword)
	for _, s := range m.services {
		if strings.Contains(strings.ToLower(s.Name), keywordLower) ||
			strings.Contains(strings.ToLower(s.DisplayName), keywordLower) ||
			strings.Contains(strings.ToLower(s.Path), keywordLower) {
			results = append(results, s)
		}
	}
	return results
}

func (m *ServiceModule) DetectDisabledSecurityServices() []model.ServiceDTO {
	securityServices := []string{
		"WinDefend", "wscsvc", "SecurityHealthService", "WdNisSvc",
		"MsMpEng", "NisSrv", "SgrmBroker", "ThreatIntelligenceFiltering",
		"Windows Defender", "Windows Security Service", " Sense", "Defender",
	}
	disabled := []model.ServiceDTO{}
	for _, s := range m.services {
		if s.Status == "Stopped" || s.StartType == "Disabled" {
			for _, sec := range securityServices {
				if strings.Contains(strings.ToLower(s.Name), strings.ToLower(sec)) ||
					strings.Contains(strings.ToLower(s.DisplayName), strings.ToLower(sec)) {
					disabled = append(disabled, s)
					break
				}
			}
		}
	}
	return disabled
}

func (m *ServiceModule) ExportToJSON(filePath string) error {
	data, err := json.MarshalIndent(m.services, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal services: %w", err)
	}
	return os.WriteFile(filePath, data, 0644)
}

func (m *ServiceModule) ExportToCSV(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Name", "DisplayName", "Status", "StartType", "Path", "Dependencies", "Description", "RiskLevel"})
	for _, s := range m.services {
		writer.Write([]string{
			s.Name, s.DisplayName, s.Status, s.StartType, s.Path, s.Dependencies, s.Description, fmt.Sprintf("%v", s.RiskLevel),
		})
	}
	return nil
}
