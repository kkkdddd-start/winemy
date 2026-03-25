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
		{`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnceEx`, model.RiskHigh},
		{`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, model.RiskMedium},
		{`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnce`, model.RiskMedium},
		{`HKLM\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Run`, model.RiskMedium},
		{`HKLM\SYSTEM\CurrentControlSet\Services`, model.RiskMedium},
		{`HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon`, model.RiskHigh},
		{`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\RunMRU`, model.RiskMedium},
		{`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\FileExts`, model.RiskMedium},
		{`HKLM\SOFTWARE\Classes\*\shell\open\command`, model.RiskHigh},
		{`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\App Paths`, model.RiskMedium},
		{`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Explorer\Run`, model.RiskHigh},
		{`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Explorer\Run`, model.RiskHigh},
		{`HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Image File Execution Options`, model.RiskMedium},
		{`HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon\Userinit`, model.RiskHigh},
		{`HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon\Shell`, model.RiskHigh},
		{`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\UFH\SFU`, model.RiskMedium},
		{`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\UFH\SFU`, model.RiskMedium},
	}

	for _, ap := range autostartPaths {
		values := m.readRegistryValues(ap.Path)
		for _, v := range values {
			v.RiskLevel = ap.Risk
			m.keys = append(m.keys, v)
		}
	}

	m.keys = append(m.keys, m.collectServices()...)

	m.collectAppInitDLLs()
	m.collectKnownDLLs()
	m.collectLSAProtection()

	return nil
}

func (m *RegistryModule) readRegistryValues(keyPath string) []model.RegistryKeyDTO {
	powershellPath := keyPath
	powershellPath = strings.Replace(powershellPath, "HKLM\\", "HKLM:", -1)
	powershellPath = strings.Replace(powershellPath, "HKCU\\", "HKCU:", -1)

	keyModified := m.getRegistryKeyModifiedTime(powershellPath)

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$ErrorActionPreference='SilentlyContinue'; Get-ItemProperty -Path '%s' | ForEach-Object { $_.PSObject.Properties | Where-Object { $_.Name -notmatch '^PS' } | ForEach-Object { Add-Member -InputObject $_ -MemberType NoteProperty -Name ValueType -Value ($_.TypeNameOfValue -replace 'System\\.', '') -PassThru } } | Select-Object PSChildName, Value, ValueType | ConvertTo-Json -Compress`, powershellPath))

	output, err := cmd.Output()
	if err != nil {
		return m.readRegistryValuesFallback(keyPath)
	}

	data := strings.TrimSpace(string(output))
	if data == "" || data == "null" || data == "[]" {
		return nil
	}

	var items []map[string]interface{}
	if err := json.Unmarshal(output, &items); err != nil {
		var single map[string]interface{}
		if err := json.Unmarshal(output, &single); err == nil {
			items = []map[string]interface{}{single}
		} else {
			return m.readRegistryValuesFallback(keyPath)
		}
	}

	var result []model.RegistryKeyDTO
	for _, item := range items {
		name := ""
		value := ""
		valueType := "REG_SZ"

		if n, ok := item["PSChildName"].(string); ok {
			name = n
		}
		if v, ok := item["Value"]; ok {
			value = fmt.Sprintf("%v", v)
		}
		if vt, ok := item["ValueType"].(string); ok {
			valueType = m.normalizeValueType(vt)
		}

		if name == "" {
			continue
		}

		result = append(result, model.RegistryKeyDTO{
			Path:      keyPath,
			Name:      name,
			ValueType: valueType,
			Value:     value,
			Modified:  keyModified,
			RiskLevel: model.RiskLow,
		})
	}

	return result
}

func (m *RegistryModule) getRegistryKeyModifiedTime(powershellPath string) time.Time {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$item = Get-Item -Path '%s' -ErrorAction SilentlyContinue; if($item) { $item.LastWriteTime.ToString("yyyy-MM-ddTHH:mm:ssZ") }`, powershellPath))
	output, err := cmd.Output()
	if err != nil {
		return time.Now()
	}
	t, err := time.Parse("2006-01-02T15:04:05Z", strings.TrimSpace(string(output)))
	if err != nil {
		return time.Now()
	}
	return t
}

func (m *RegistryModule) normalizeValueType(typeName string) string {
	switch typeName {
	case "String", "REG_SZ":
		return "REG_SZ"
	case "DWord", "DWORD", "Int32", "Int64", "UInt32":
		return "REG_DWORD"
	case "Binary", "REG_BINARY":
		return "REG_BINARY"
	case "MultiString", "REG_MULTI_SZ":
		return "REG_MULTI_SZ"
	case "QWord", "QWORD":
		return "REG_QWORD"
	case "None", "REG_NONE":
		return "REG_NONE"
	case "ExpandString", "REG_EXPAND_SZ":
		return "REG_EXPAND_SZ"
	default:
		if strings.Contains(typeName, "String") {
			return "REG_SZ"
		}
		if strings.Contains(typeName, "DWord") || strings.Contains(typeName, "Int") {
			return "REG_DWORD"
		}
		return typeName
	}
}

func (m *RegistryModule) readRegistryValuesFallback(keyPath string) []model.RegistryKeyDTO {
	powershellPath := keyPath
	powershellPath = strings.Replace(powershellPath, "HKLM\\", "HKLM:", -1)
	powershellPath = strings.Replace(powershellPath, "HKCU\\", "HKCU:", -1)

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`Get-ItemProperty -Path '%s' -ErrorAction SilentlyContinue | ConvertTo-Json -Compress`, powershellPath))

	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	data := strings.TrimSpace(string(output))
	if data == "" || data == "null" {
		return nil
	}

	var props map[string]interface{}
	if err := json.Unmarshal(output, &props); err != nil {
		return nil
	}

	var result []model.RegistryKeyDTO
	for name, value := range props {
		if strings.HasPrefix(name, "PS") {
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

func (m *RegistryModule) collectAppInitDLLs() {
	appInitPaths := []string{
		`HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Windows`,
		`HKLM\SOFTWARE\WOW6432Node\Microsoft\Windows NT\CurrentVersion\Windows`,
	}

	for _, path := range appInitPaths {
		cmd := exec.Command("powershell", "-Command",
			fmt.Sprintf(`Get-ItemProperty -Path '%s' -ErrorAction SilentlyContinue | Select-Object AppInit_DLLs, LoadAppInit_DLLs | ConvertTo-Json`, path))
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		var result map[string]interface{}
		if err := json.Unmarshal(output, &result); err != nil {
			continue
		}

		if appInit, ok := result["AppInit_DLLs"].(string); ok && appInit != "" {
			m.keys = append(m.keys, model.RegistryKeyDTO{
				Path:      path,
				Name:      "AppInit_DLLs",
				ValueType: "REG_SZ",
				Value:     appInit,
				Modified:  time.Now(),
				RiskLevel: model.RiskHigh,
			})
		}

		if loadAppInit, ok := result["LoadAppInit_DLLs"].(float64); ok && loadAppInit == 1 {
			m.keys = append(m.keys, model.RegistryKeyDTO{
				Path:      path,
				Name:      "LoadAppInit_DLLs",
				ValueType: "REG_DWORD",
				Value:     "1 (Enabled)",
				Modified:  time.Now(),
				RiskLevel: model.RiskHigh,
			})
		}
	}
}

func (m *RegistryModule) collectKnownDLLs() {
	cmd := exec.Command("powershell", "-Command",
		`Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\KnownDLLs' -ErrorAction SilentlyContinue | ConvertTo-Json`)
	output, err := cmd.Output()
	if err != nil {
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return
	}

	for name, value := range result {
		if strings.HasPrefix(name, "PS") {
			continue
		}
		valueStr := fmt.Sprintf("%v", value)
		if valueStr == "" || valueStr == "0" {
			continue
		}
		m.keys = append(m.keys, model.RegistryKeyDTO{
			Path:      `HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\KnownDLLs`,
			Name:      name,
			ValueType: "REG_SZ",
			Value:     valueStr,
			Modified:  time.Now(),
			RiskLevel: model.RiskMedium,
		})
	}
}

func (m *RegistryModule) collectLSAProtection() {
	cmd := exec.Command("powershell", "-Command",
		`Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Control\Lsa' -ErrorAction SilentlyContinue | Select-Object RunAsPPL | ConvertTo-Json`)
	output, err := cmd.Output()
	if err != nil {
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return
	}

	if runAsPPL, ok := result["RunAsPPL"]; ok {
		valueStr := fmt.Sprintf("%v", runAsPPL)
		m.keys = append(m.keys, model.RegistryKeyDTO{
			Path:      `HKLM\SYSTEM\CurrentControlSet\Control\Lsa`,
			Name:      "RunAsPPL",
			ValueType: "REG_DWORD",
			Value:     valueStr,
			Modified:  time.Now(),
			RiskLevel: model.RiskLow,
		})
	}
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
