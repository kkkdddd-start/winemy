//go:build windows

package m9_software

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type SoftwareModule struct {
	ctx      context.Context
	storage  registry.Storage
	software []SoftwareDTO
}

type SoftwareDTO struct {
	Name      string          `json:"name"`
	Version   string          `json:"version"`
	Vendor    string          `json:"vendor"`
	Installed time.Time       `json:"installed"`
	RiskLevel model.RiskLevel `json:"risk_level"`
}

func New() *SoftwareModule {
	return &SoftwareModule{}
}

func (m *SoftwareModule) ID() int       { return 9 }
func (m *SoftwareModule) Name() string  { return "software" }
func (m *SoftwareModule) Priority() int { return 2 }

func (m *SoftwareModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	return nil
}

func (m *SoftwareModule) Collect(ctx context.Context) error {
	m.software = []SoftwareDTO{}

	cmd := exec.Command("powershell", "-Command",
		`Get-ItemProperty HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*, HKLM:\Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\* 2>$null | Where-Object { $_.DisplayName } | Select-Object DisplayName, DisplayVersion, Publisher, InstallDate | ConvertTo-Json`)

	output, err := cmd.Output()
	if err != nil {
		m.software = append(m.software, SoftwareDTO{
			Name:      "Error",
			Version:   "Failed to enumerate software",
			Vendor:    "",
			RiskLevel: model.RiskLow,
		})
		return nil
	}

	var software []struct {
		DisplayName    string `json:"DisplayName"`
		DisplayVersion string `json:"DisplayVersion"`
		Publisher      string `json:"Publisher"`
		InstallDate    string `json:"InstallDate"`
	}

	if err := json.Unmarshal(output, &software); err != nil {
		var single struct {
			DisplayName    string `json:"DisplayName"`
			DisplayVersion string `json:"DisplayVersion"`
			Publisher      string `json:"Publisher"`
			InstallDate    string `json:"InstallDate"`
		}
		if err2 := json.Unmarshal(output, &single); err2 == nil {
			software = []struct {
				DisplayName    string `json:"DisplayName"`
				DisplayVersion string `json:"DisplayVersion"`
				Publisher      string `json:"Publisher"`
				InstallDate    string `json:"InstallDate"`
			}{single}
		}
	}

	for _, s := range software {
		if s.DisplayName == "" {
			continue
		}

		installedTime := parseInstallDate(s.InstallDate)
		if installedTime.IsZero() {
			installedTime = time.Now()
		}

		riskLevel := model.RiskLow
		if isSuspiciousSoftware(s.DisplayName, s.Publisher) {
			riskLevel = model.RiskHigh
		}

		m.software = append(m.software, SoftwareDTO{
			Name:      s.DisplayName,
			Version:   s.DisplayVersion,
			Vendor:    s.Publisher,
			Installed: installedTime,
			RiskLevel: riskLevel,
		})
	}

	return nil
}

func parseInstallDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}

	dateStr = strings.TrimSpace(dateStr)
	formats := []string{
		"20060102",
		"01/02/2006",
		"1/2/2006",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}
	return time.Time{}
}

func isSuspiciousSoftware(name, vendor string) bool {
	nameLower := strings.ToLower(name)
	vendorLower := strings.ToLower(vendor)

	suspiciousPatterns := []string{
		"crack",
		"keygen",
		"patcher",
		" activator",
		"serial",
		"torrent",
		"download",
		"warez",
		"hack",
		"cracked",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(nameLower, pattern) || strings.Contains(vendorLower, pattern) {
			return true
		}
	}
	return false
}

func (m *SoftwareModule) Stop() error {
	return nil
}

func (m *SoftwareModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.software))
	for _, s := range m.software {
		installedStr := ""
		if !s.Installed.IsZero() {
			installedStr = s.Installed.Format(time.RFC3339)
		}
		result = append(result, map[string]interface{}{
			"name":       s.Name,
			"version":    s.Version,
			"vendor":     s.Vendor,
			"installed":  installedStr,
			"risk_level": s.RiskLevel,
		})
	}
	return result, nil
}

func (m *SoftwareModule) Search(keyword string) ([]SoftwareDTO, error) {
	results := []SoftwareDTO{}
	keywordLower := strings.ToLower(keyword)

	for _, s := range m.software {
		if strings.Contains(strings.ToLower(s.Name), keywordLower) ||
			strings.Contains(strings.ToLower(s.Version), keywordLower) ||
			strings.Contains(strings.ToLower(s.Vendor), keywordLower) {
			results = append(results, s)
		}
	}

	return results, nil
}

func (m *SoftwareModule) ExportJSON(filePath string) error {
	data := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"count":     len(m.software),
		"software":  m.software,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(filePath, jsonData, 0644)
}

func (m *SoftwareModule) ExportCSV(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Name", "Version", "Vendor", "Installed", "RiskLevel"})

	for _, s := range m.software {
		installedStr := ""
		if !s.Installed.IsZero() {
			installedStr = s.Installed.Format("2006-01-02")
		}
		writer.Write([]string{
			s.Name,
			s.Version,
			s.Vendor,
			installedStr,
			fmt.Sprintf("%v", s.RiskLevel),
		})
	}

	return nil
}
