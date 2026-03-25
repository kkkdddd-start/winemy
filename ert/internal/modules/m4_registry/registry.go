//go:build windows

package m4_registry

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type RegistryModule struct {
	ctx     context.Context
	storage registry.Storage
	keys    []model.RegistryKeyDTO
}

func New() *RegistryModule {
	return &RegistryModule{}
}

func (m *RegistryModule) ID() int       { return 4 }
func (m *RegistryModule) Name() string  { return "registry" }
func (m *RegistryModule) Priority() int { return 1 }

func (m *RegistryModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	return nil
}

func (m *RegistryModule) Collect(ctx context.Context) error {
	m.keys = []model.RegistryKeyDTO{}

	autostartPaths := []struct {
		Path string
		Risk model.RiskLevel
	}{
		{`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, model.RiskHigh},
		{`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnce`, model.RiskHigh},
		{`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, model.RiskMedium},
		{`HKLM\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Run`, model.RiskMedium},
		{`HKLM\SYSTEM\CurrentControlSet\Services`, model.RiskMedium},
		{`HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon`, model.RiskHigh},
		{`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\RunMRU`, model.RiskMedium},
	}

	for _, ap := range autostartPaths {
		values := m.readRegistryValues(ap.Path)
		for _, v := range values {
			v.RiskLevel = ap.Risk
			m.keys = append(m.keys, v)
		}
	}

	m.keys = append(m.keys, m.collectServices()...)

	return nil
}

func (m *RegistryModule) readRegistryValues(keyPath string) []model.RegistryKeyDTO {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`Get-ItemProperty -Path '%s' -ErrorAction SilentlyContinue | ConvertTo-Json -Compress`, keyPath))

	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var result []model.RegistryKeyDTO
	data := strings.TrimSpace(string(output))
	if data == "" || data == "null" {
		return nil
	}

	var props map[string]interface{}
	if err := json.Unmarshal(output, &props); err != nil {
		return m.parseMultiValue(keyPath, output)
	}

	for name, value := range props {
		if name == "PSPath" || name == "PSParentPath" || name == "PSChildName" || name == "PSDrive" || name == "PSProvider" {
			continue
		}

		valueStr := fmt.Sprintf("%v", value)
		valueType := "REG_SZ"
		if value == nil {
			valueType = "REG_NONE"
		}

		result = append(result, model.RegistryKeyDTO{
			Path:      keyPath,
			Name:      name,
			ValueType: valueType,
			Value:     valueStr,
			Modified:  time.Now(),
			RiskLevel: model.RiskLow,
		})
	}

	return result
}

func (m *RegistryModule) parseMultiValue(keyPath string, output []byte) []model.RegistryKeyDTO {
	var items []map[string]interface{}
	if err := json.Unmarshal(output, &items); err == nil {
		var result []model.RegistryKeyDTO
		for _, item := range items {
			for name, value := range item {
				if name == "PSPath" || strings.HasPrefix(name, "PS") {
					continue
				}
				valueStr := fmt.Sprintf("%v", value)
				result = append(result, model.RegistryKeyDTO{
					Path:      keyPath,
					Name:      name,
					ValueType: "REG_SZ",
					Value:     valueStr,
					Modified:  time.Now(),
					RiskLevel: model.RiskLow,
				})
			}
		}
		return result
	}
	return nil
}

func (m *RegistryModule) collectServices() []model.RegistryKeyDTO {
	cmd := exec.Command("powershell", "-Command",
		`Get-ItemProperty 'HKLM:\SYSTEM\CurrentControlSet\Services\*' -ErrorAction SilentlyContinue | Select Name, DisplayName, Start, Type, ImagePath | ConvertTo-Json -Compress`)

	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var services []struct {
		Name        string `json:"Name"`
		DisplayName string `json:"DisplayName"`
		Start       int    `json:"Start"`
		Type        int    `json:"Type"`
		ImagePath   string `json:"ImagePath"`
	}

	if err := json.Unmarshal(output, &services); err != nil {
		return nil
	}

	var result []model.RegistryKeyDTO
	for _, svc := range services {
		path := `HKLM\SYSTEM\CurrentControlSet\Services\` + svc.Name
		risk := m.assessServiceRisk(svc.Name, svc.ImagePath)

		result = append(result, model.RegistryKeyDTO{
			Path:      path,
			Name:      svc.DisplayName,
			ValueType: "SERVICE",
			Value:     svc.ImagePath,
			Modified:  time.Now(),
			RiskLevel: risk,
		})
	}

	return result
}

func (m *RegistryModule) assessServiceRisk(name, path string) model.RiskLevel {
	nameLower := strings.ToLower(name)
	pathLower := strings.ToLower(path)

	suspiciousNames := []string{
		"mimikatz", "pwdump", "procdump", "lsass",
		"metasploit", "meterpreter", "cobalt", "empire",
		"ransomware", "trojan", "backdoor",
	}

	suspiciousPaths := []string{
		"temp", "tmp", "appdata", "downloads",
		"public", "UNC ", "\\\\",
	}

	for _, sus := range suspiciousNames {
		if strings.Contains(nameLower, sus) {
			return model.RiskHigh
		}
	}

	for _, sus := range suspiciousPaths {
		if strings.Contains(pathLower, sus) {
			return model.RiskMedium
		}
	}

	if path == "" {
		return model.RiskMedium
	}

	return model.RiskLow
}

func (m *RegistryModule) Search(keyword string) []model.RegistryKeyDTO {
	results := []model.RegistryKeyDTO{}
	keywordLower := strings.ToLower(keyword)
	for _, k := range m.keys {
		if strings.Contains(strings.ToLower(k.Name), keywordLower) ||
			strings.Contains(strings.ToLower(k.Path), keywordLower) ||
			strings.Contains(strings.ToLower(k.Value), keywordLower) {
			results = append(results, k)
		}
	}
	return results
}

func (m *RegistryModule) Stop() error {
	return nil
}

func (m *RegistryModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.keys))
	for _, k := range m.keys {
		result = append(result, map[string]interface{}{
			"path":       k.Path,
			"name":       k.Name,
			"value_type": k.ValueType,
			"value":      k.Value,
			"modified":   k.Modified.Format(time.RFC3339),
			"risk_level": k.RiskLevel,
		})
	}
	return result, nil
}
