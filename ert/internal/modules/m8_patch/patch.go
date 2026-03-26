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
		`$ErrorActionPreference='SilentlyContinue'
$session = New-Object -ComObject Microsoft.Update.Session
$searcher = $session.CreateUpdateSearcher()
$updates = $searcher.Search("IsInstalled=0 and Type='Software'").Updates
foreach($update in $updates) {
    $kb = $update.KBArticleIDs | Where-Object { $_ -ne $null }
    if($kb) {
        Write-Output "KB$kb"
    } else {
        $title = $update.Title -replace '[^a-zA-Z0-9 ]', ''
        Write-Output "PATCH:$title"
    }
}`)

	output, err := cmd.Output()
	if err != nil {
		return m.detectMissingPatchesFallback()
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		kb := strings.TrimSpace(line)
		if kb == "" || strings.HasPrefix(kb, "Exception") || strings.HasPrefix(kb, "The") {
			continue
		}
		if len(kb) >= 3 {
			if !strings.HasPrefix(kb, "KB") && !strings.HasPrefix(kb, "PATCH:") {
				continue
			}
			m.missing = append(m.missing, kb)
		}
	}

	if len(m.missing) == 0 {
		return m.detectMissingPatchesFallback()
	}

	return nil
}

func (m *PatchModule) detectMissingPatchesFallback() error {
	cmd := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
$wu = New-Object -ComObject Microsoft.Update.AutoUpdate
$searcher = $wu.CreateUpdateSearcher()
try {
    $result = $searcher.Search("IsInstalled=0")
    $result.Updates | ForEach-Object {
        $_.KBArticleIDs | ForEach-Object { if($_ -ne $null) { Write-Output "KB$_" } }
    }
} catch { }`)

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to search for missing patches: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		kb := strings.TrimSpace(line)
		if kb == "" || strings.HasPrefix(kb, "Exception") || strings.HasPrefix(kb, "The") {
			continue
		}
		if len(kb) >= 3 && strings.HasPrefix(kb, "KB") {
			m.missing = append(m.missing, kb)
		}
	}

	return nil
}

func (m *PatchModule) Search(keyword string) []HotfixDTO {
	results := []HotfixDTO{}
	keywordLower := strings.ToLower(keyword)
	for _, p := range m.installed {
		if strings.Contains(strings.ToLower(p.HotfixID), keywordLower) ||
			strings.Contains(strings.ToLower(p.Description), keywordLower) {
			results = append(results, p)
		}
	}
	return results
}

func (m *PatchModule) GetCVEInfo(kbID string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$ErrorActionPreference='SilentlyContinue'
$searcher = New-Object -ComObject Microsoft.Update.Searcher
$kb = '%s' -replace 'KB', ''
$updates = $searcher.Search("UpdateID like '%%%s%%'").Updates
foreach($update in $updates) {
    $cves = @()
    foreach($article in $update.KBArticleIDs) {
        $cves += "KB$article"
    }
    if($cves.Count -eq 0) { $cves = @("N/A") }
    Write-Output ($update.Title + "|" + $cves[0])
}`, kbID, kbID))

	output, err := cmd.Output()
	if err != nil {
		return []map[string]interface{}{
			{
				"kb_id":       kbID,
				"title":       "Unknown",
				"cve_ids":     []string{},
				"description": "Unable to retrieve CVE information",
			},
		}, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 2 {
			results = append(results, map[string]interface{}{
				"kb_id":       kbID,
				"title":       parts[0],
				"cve_ids":     strings.Split(parts[1], ","),
				"description": parts[0],
			})
		}
	}

	if len(results) == 0 {
		results = append(results, map[string]interface{}{
			"kb_id":       kbID,
			"title":       "Not Found",
			"cve_ids":     []string{},
			"description": "KB update not found in Microsoft Update Catalog",
		})
	}
	return results, nil
}

func (m *PatchModule) GetPatchSource(kbID string) (string, error) {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$ErrorActionPreference='SilentlyContinue'
$session = New-Object -ComObject Microsoft.Update.Session
$searcher = $session.CreateUpdateSearcher()
$kb = '%s' -replace 'KB', ''
try {
    $updates = $searcher.Search("UpdateID like '%%%s%%'").Updates
    foreach($update in $updates) {
        if($update.KBArticleIDs -contains $kb -or $update.Title -like "*$kb*") {
            Write-Output $update.SupportUrl
            break
        }
    }
} catch { }
if(-not $updates) {
    Write-Output "https://catalog.update.microsoft.com/"
}`, kbID, kbID))

	output, err := cmd.Output()
	if err != nil {
		return "https://catalog.update.microsoft.com/", nil
	}
	result := strings.TrimSpace(string(output))
	if result == "" {
		return "https://catalog.update.microsoft.com/", nil
	}
	return result, nil
}

func (m *PatchModule) GetPatchSize(kbID string) (map[string]interface{}, error) {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$ErrorActionPreference='SilentlyContinue'
$session = New-Object -ComObject Microsoft.Update.Session
$searcher = $session.CreateUpdateSearcher()
$kb = '%s' -replace 'KB', ''
try {
    $updates = $searcher.Search("UpdateID like '%%%s%%'").Updates
    foreach($update in $updates) {
        if($update.KBArticleIDs -contains $kb -or $update.Title -like "*$kb*") {
            $sizeKB = [math]::Round($update.MaxDownloadSize / 1024, 2)
            $sizeMB = [math]::Round($update.MaxDownloadSize / 1024 / 1024, 2)
            Write-Output "$sizeKB|$sizeMB"
            break
        }
    }
} catch { }
if(-not $updates) {
    Write-Output "0|0"
}`, kbID, kbID))

	output, err := cmd.Output()
	if err != nil {
		return map[string]interface{}{
			"kb_id":     kbID,
			"size_kb":   0,
			"size_mb":   0,
			"size_byte": 0,
		}, nil
	}

	parts := strings.Split(strings.TrimSpace(string(output)), "|")
	sizeKB := float64(0)
	sizeMB := float64(0)
	if len(parts) >= 2 {
		fmt.Sscanf(parts[0], "%f", &sizeKB)
		fmt.Sscanf(parts[1], "%f", &sizeMB)
	}

	return map[string]interface{}{
		"kb_id":     kbID,
		"size_kb":   sizeKB,
		"size_mb":   sizeMB,
		"size_byte": int64(sizeMB * 1024 * 1024),
	}, nil
}

func (m *PatchModule) DetectRollback() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	cmd := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Services\EventLog\Security' -Name File | Select-Object -ExpandProperty File`)

	output, err := cmd.Output()
	if err == nil {
		eventLogPath := strings.TrimSpace(string(output))
		if eventLogPath != "" {
			cmd2 := exec.Command("powershell", "-Command",
				`wevtutil qe Security /c:50 /f:text /rd:true | Select-String -Pattern "Rollback|1202"`)
			output2, err2 := cmd2.Output()
			if err2 == nil {
				lines := strings.Split(string(output2), "\n")
				for _, line := range lines {
					if strings.Contains(strings.ToLower(line), "rollback") || strings.Contains(line, "1202") {
						results = append(results, map[string]interface{}{
							"event":      strings.TrimSpace(line),
							"type":       "rollback_detected",
							"risk_level": model.RiskHigh,
						})
					}
				}
			}
		}
	}

	cmd3 := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
$hotfixes = Get-HotFix | Where-Object { $_.Description -eq 'Hotfix' -or $_.Description -eq 'Update' }
foreach($hf in $hotfixes) {
    $installedOn = $hf.InstalledOn
    if($installedOn -and ($installedOn -gt (Get-Date).AddDays(-7))) {
        Write-Output ("Recent:" + $hf.HotFixID + ":" + $installedOn.ToString('yyyy-MM-dd'))
    }
}`)

	output3, err := cmd3.Output()
	if err == nil {
		lines := strings.Split(string(output3), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Recent:") {
				parts := strings.Split(strings.TrimPrefix(line, "Recent:"), ":")
				if len(parts) >= 3 {
					results = append(results, map[string]interface{}{
						"hotfix_id":  parts[1],
						"type":       "recent_update",
						"installed":  parts[2],
						"risk_level": model.RiskMedium,
					})
				}
			}
		}
	}

	if len(results) == 0 {
		results = append(results, map[string]interface{}{
			"type":       "no_rollback_detected",
			"risk_level": model.RiskLow,
		})
	}

	return results, nil
}
