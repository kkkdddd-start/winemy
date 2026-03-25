//go:build windows

package m8_patch

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type PatchModule struct {
	ctx       context.Context
	storage   registry.Storage
	installed []HotfixDTO
	missing   []string
}

type HotfixDTO struct {
	HotfixID    string          `json:"hotfix_id"`
	Installed   time.Time       `json:"installed"`
	Description string          `json:"description"`
	RiskLevel   model.RiskLevel `json:"risk_level"`
}

func New() *PatchModule {
	return &PatchModule{}
}

func (m *PatchModule) ID() int       { return 8 }
func (m *PatchModule) Name() string  { return "patch" }
func (m *PatchModule) Priority() int { return 2 }

func (m *PatchModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	return nil
}

func (m *PatchModule) Collect(ctx context.Context) error {
	m.installed = []HotfixDTO{}
	m.missing = []string{}

	output, err := exec.Command("wmic", "qfe", "get", "HotFixID,InstalledOn,Description", "/format:csv").Output()
	if err != nil {
		m.installed = append(m.installed, HotfixDTO{
			HotfixID:    "Error",
			Installed:   time.Time{},
			Description: "Failed to query patches: " + err.Error(),
			RiskLevel:   model.RiskLow,
		})
		return nil
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 {
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")
		if len(fields) < 4 {
			continue
		}

		hotfixID := strings.TrimSpace(fields[1])
		installedOn := strings.TrimSpace(fields[2])
		description := strings.TrimSpace(fields[3])

		if hotfixID == "HotFixID" || hotfixID == "" {
			continue
		}

		installedTime := parseWMITime(installedOn)
		if installedTime.IsZero() {
			installedTime = time.Now()
		}

		riskLevel := model.RiskLow
		if isSecurityPatch(description) {
			riskLevel = model.RiskLow
		}

		m.installed = append(m.installed, HotfixDTO{
			HotfixID:    hotfixID,
			Installed:   installedTime,
			Description: description,
			RiskLevel:   riskLevel,
		})
	}

	return nil
}

func parseWMITime(dateStr string) time.Time {
	if dateStr == "" || dateStr == "NULL" {
		return time.Time{}
	}

	dateStr = strings.TrimSpace(dateStr)

	formats := []string{
		"1/2/2006",
		"01/02/2006",
		"2006-01-02",
		"1/2/2006 12:00:00 AM",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}
	return time.Time{}
}

func isSecurityPatch(description string) bool {
	descLower := strings.ToLower(description)
	securityKeywords := []string{
		"security",
		"security update",
		"security rollup",
		"cumulative update",
		"hotfix",
	}
	for _, keyword := range securityKeywords {
		if strings.Contains(descLower, keyword) {
			return true
		}
	}
	return false
}

func (m *PatchModule) Stop() error {
	return nil
}

func (m *PatchModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0)

	for _, p := range m.installed {
		installedStr := ""
		if !p.Installed.IsZero() {
			installedStr = p.Installed.Format(time.RFC3339)
		}
		result = append(result, map[string]interface{}{
			"type":        "installed",
			"hotfix_id":   p.HotfixID,
			"installed":   installedStr,
			"description": p.Description,
			"risk_level":  p.RiskLevel,
		})
	}

	for _, id := range m.missing {
		result = append(result, map[string]interface{}{
			"type":       "missing",
			"hotfix_id":  id,
			"risk_level": model.RiskHigh,
		})
	}

	return result, nil
}

func (m *PatchModule) DetectMissingPatches(requiredPatches []string) error {
	installedMap := make(map[string]bool)
	for _, p := range m.installed {
		installedMap[p.HotfixID] = true
	}

	m.missing = []string{}
	for _, required := range requiredPatches {
		if !installedMap[required] {
			m.missing = append(m.missing, required)
		}
	}
	return nil
}

func (m *PatchModule) DetectMissingPatchesFromMicrosoft() error {
	m.missing = []string{}

	cmd := exec.Command("powershell", "-Command",
		`$session = New-Object -ComObject Microsoft.Update.Session; $searcher = $session.CreateUpdateSearcher(); try { $history = $searcher.GetTotalHistoryCount(); $updates = $searcher.QueryHistory(0, $history) | Where-Object { $_.Operation -eq 1 -and $_.ResultCode -eq 2 }; $updates | ForEach-Object { $_.Title } } catch { }`)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to query update history: %w", err)
	}

	systemInfo, err := m.getSystemInfo()
	if err != nil {
		return fmt.Errorf("failed to get system info: %w", err)
	}

	cmd = exec.Command("powershell", "-Command",
		fmt.Sprintf(`$session = New-Object -ComObject Microsoft.Update.Session; $searcher = $session.CreateUpdateSearcher(); $searcher.Search("IsInstalled=0 and Type='Software'").Updates | Select-Object -ExpandProperty KBArticleIDs`, systemInfo))
	output, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to search for missing patches: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		kb := strings.TrimSpace(line)
		if kb != "" && !strings.HasPrefix(kb, "Exception") && !strings.HasPrefix(kb, "The") {
			if kb[:2] == "KB" || len(kb) >= 5 {
				if kb[:2] != "KB" {
					kb = "KB" + kb
				}
				m.missing = append(m.missing, kb)
			}
		}
	}
	return nil
}

func (m *PatchModule) getSystemInfo() (string, error) {
	cmd := exec.Command("powershell", "-Command",
		`(Get-WmiObject -Class Win32_OperatingSystem).OSArchitecture`)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	arch := strings.TrimSpace(string(output))
	if arch == "64-bit" {
		return "IsInstalled=0", nil
	}
	return "IsInstalled=0", nil
}
