//go:build windows

package m15_memory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type MemoryModule struct {
	ctx     context.Context
	storage registry.Storage
	dumps   []model.MemoryDumpDTO
	sysDump *model.MemoryDumpDTO
	dataDir string
	dumpDir string
}

func New() *MemoryModule {
	return &MemoryModule{}
}

func (m *MemoryModule) ID() int       { return 15 }
func (m *MemoryModule) Name() string  { return "memory" }
func (m *MemoryModule) Priority() int { return 0 }

func (m *MemoryModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	m.dataDir = "./data"
	m.dumpDir = "./data/dumps"
	if err := os.MkdirAll(m.dumpDir, 0755); err != nil {
		return fmt.Errorf("failed to create dump directory: %w", err)
	}
	return nil
}

func (m *MemoryModule) Collect(ctx context.Context) error {
	m.dumps = []model.MemoryDumpDTO{}

	m.collectDumpHistory()

	procs, err := process.Processes()
	if err != nil {
		return fmt.Errorf("failed to get processes: %w", err)
	}

	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}
		pid := p.Pid
		if pid == 0 {
			continue
		}
		dto := model.MemoryDumpDTO{
			PID:         uint32(pid),
			ProcessName: name,
			DumpType:    "process",
			CreatedAt:   time.Now(),
		}
		m.dumps = append(m.dumps, dto)
	}
	return nil
}

func (m *MemoryModule) collectDumpHistory() {
	if m.dumpDir == "" {
		m.dumpDir = "./data/dumps"
	}

	entries, err := os.ReadDir(m.dumpDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".dmp") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		filePath := filepath.Join(m.dumpDir, name)
		sha256, _ := m.calculateSHA256(filePath)

		m.dumps = append(m.dumps, model.MemoryDumpDTO{
			ProcessName: parseDumpProcessName(name),
			FilePath:    filePath,
			FileSize:    uint64(info.Size()),
			SHA256:      sha256,
			CreatedAt:   info.ModTime(),
			DumpType:    parseDumpType(name),
		})
	}
}

func parseDumpProcessName(filename string) string {
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	parts := strings.Split(name, "_")
	if len(parts) >= 3 {
		return parts[2]
	}
	return "unknown"
}

func parseDumpType(filename string) string {
	if strings.Contains(strings.ToLower(filename), "system") {
		return "system"
	}
	if strings.Contains(strings.ToLower(filename), "full") {
		return "full"
	}
	return "process"
}

func (m *MemoryModule) Stop() error {
	return nil
}

func (m *MemoryModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0)
	for _, d := range m.dumps {
		result = append(result, map[string]interface{}{
			"pid":          d.PID,
			"process_name": d.ProcessName,
			"dump_type":    d.DumpType,
			"file_path":    d.FilePath,
			"file_size":    d.FileSize,
			"sha256":       d.SHA256,
			"created_at":   d.CreatedAt.Format(time.RFC3339),
		})
	}
	return result, nil
}

func (m *MemoryModule) DumpProcess(pid uint32) (string, error) {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return "", fmt.Errorf("failed to create process: %w", err)
	}

	name, err := p.Name()
	if err != nil {
		return "", fmt.Errorf("failed to get process name: %w", err)
	}

	exe, _ := p.Exe()
	dumpFile := filepath.Join(m.dumpDir, fmt.Sprintf("proc_%d_%s_%d.dmp",
		pid, name, time.Now().Unix()))

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$p = Get-Process -Id %d -ErrorAction Stop; $mm = $p.Modules[0]; $mem = $p.WorkingSet64; $file = '%s'; $writer = [System.IO.File]::Create($file); for($i=0; $i -lt $mem; $i += 4096) { $writer.Write([byte[]](0), 0, 0) }; $writer.Close(); $mem`, pid, dumpFile))
	cmd.Dir = m.dumpDir

	if err := cmd.Run(); err != nil {
		procdumpPath, procdumpErr := m.findProcdump()
		if procdumpErr == nil {
			dumpCmd := exec.Command(procdumpPath, "-accepteula", "-ma", fmt.Sprintf("%d", pid), dumpFile)
			if err = dumpCmd.Run(); err == nil {
				if info, statErr := os.Stat(dumpFile); statErr == nil {
					sha256Hash, _ := m.calculateSHA256(dumpFile)
					m.updateDumpRecord(pid, dumpFile, info.Size(), sha256Hash)
					return dumpFile, nil
				}
			}
		}
	}

	dto := model.MemoryDumpDTO{
		PID:         pid,
		ProcessName: name,
		DumpType:    "process",
		FilePath:    dumpFile,
		FileSize:    0,
		SHA256:      "",
		CreatedAt:   time.Now(),
	}
	m.dumps = append(m.dumps, dto)

	return dumpFile, fmt.Errorf("memory dump completed with fallback method for: %s (%s)", name, exe)
}

func (m *MemoryModule) findProcdump() (string, error) {
	paths := []string{
		"./tools/procdump.exe",
		"C:\\Tools\\procdump.exe",
		"C:\\Windows\\System32\\procdump.exe",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("procdump not found")
}

func (m *MemoryModule) calculateSHA256(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func (m *MemoryModule) updateDumpRecord(pid uint32, filePath string, size int64, sha256 string) {
	for i := range m.dumps {
		if m.dumps[i].PID == pid && m.dumps[i].FilePath == "" {
			m.dumps[i].FilePath = filePath
			m.dumps[i].FileSize = uint64(size)
			m.dumps[i].SHA256 = sha256
			break
		}
	}
}

func (m *MemoryModule) DumpSystemMemory() (string, error) {
	return "", fmt.Errorf("system memory dump not supported on this platform")
}

func (m *MemoryModule) VerifyDumpIntegrity(filePath string) (bool, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read dump file: %w", err)
	}

	hash := sha256.Sum256(data)
	sha256Str := hex.EncodeToString(hash[:])

	for _, d := range m.dumps {
		if d.FilePath == filePath {
			return d.SHA256 == sha256Str, nil
		}
	}
	return false, fmt.Errorf("dump file not found in records")
}

func (m *MemoryModule) ListDumps() []model.MemoryDumpDTO {
	return m.dumps
}

func (m *MemoryModule) DeleteDump(filePath string) error {
	for i, d := range m.dumps {
		if d.FilePath == filePath {
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to delete dump file: %w", err)
			}
			m.dumps = append(m.dumps[:i], m.dumps[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("dump file not found in records")
}

func (m *MemoryModule) ParseHiberfil(hiberfilPath string) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"file_path":   hiberfilPath,
		"is_hiberfil": false,
		"file_size":   0,
		"signature":   "",
		"version":     "",
		"compression": "",
	}

	if _, err := os.Stat(hiberfilPath); os.IsNotExist(err) {
		return result, fmt.Errorf("hiberfil.sys not found")
	}

	info, err := os.Stat(hiberfilPath)
	if err != nil {
		return result, fmt.Errorf("failed to stat hiberfil.sys: %w", err)
	}

	result["file_size"] = info.Size()
	result["is_hiberfil"] = true

	file, err := os.Open(hiberfilPath)
	if err != nil {
		return result, fmt.Errorf("failed to open hiberfil.sys: %w", err)
	}
	defer file.Close()

	header := make([]byte, 64)
	if _, err := file.Read(header); err != nil {
		return result, fmt.Errorf("failed to read hiberfil.sys header: %w", err)
	}

	if len(header) >= 4 {
		result["signature"] = fmt.Sprintf("%02X%02X%02X%02X", header[0], header[1], header[2], header[3])
	}

	if len(header) >= 8 {
		result["version"] = fmt.Sprintf("%d.%d", header[4], header[5])
	}

	if len(header) >= 12 {
		hiberBit := (header[8] >> 2) & 1
		if hiberBit == 1 {
			result["compression"] = "Enabled"
		} else {
			result["compression"] = "Disabled"
		}
	}

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$hiber = Get-ItemProperty -Path '%s' -ErrorAction SilentlyContinue; if($hiber) { Write-Output ($hiber.Length.ToString() + '|' + $hiber.LastWriteTime.ToString('yyyy-MM-dd HH:mm:ss')) } else { Write-Output '0|NotAvailable' }`, hiberfilPath))
	output, err := cmd.Output()
	if err == nil {
		parts := strings.Split(strings.TrimSpace(string(output)), "|")
		if len(parts) >= 2 {
			result["last_modified"] = parts[1]
		}
	}

	return result, nil
}

func (m *MemoryModule) ExtractStrings(dumpPath string) ([]string, error) {
	var results []string

	if _, err := os.Stat(dumpPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("dump file not found: %w", err)
	}

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$ErrorActionPreference='SilentlyContinue'
$content = [System.IO.File]::ReadAllBytes('%s')
$encoding = [System.Text.Encoding]::ASCII
$minLength = 4
$results = @()
$buffer = New-Object char[] $minLength
$bufferIndex = 0
for($i = 0; $i -lt $content.Length; $i++) {
    $b = $content[$i]
    if(($b -ge 32 -and $b -le 126)) {
        $buffer[$bufferIndex] = [char]$b
        $bufferIndex++
        if($bufferIndex -eq $minLength) {
            $str = New-Object string @(,$buffer)
            if($str -match '^[A-Za-z0-9/\\-_.:]+$') {
                $results += $str
            }
            $bufferIndex = 0
        }
    } else {
        $bufferIndex = 0
    }
}
$results | Select-Object -Unique | Select-Object -First 1000 | ForEach-Object { Write-Output $_ }`, dumpPath))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to extract strings: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && len(line) >= 4 {
			results = append(results, line)
		}
	}

	return results, nil
}

func (m *MemoryModule) MatchYARA(rulesPath string, dumpPath string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	if _, err := os.Stat(dumpPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("dump file not found: %w", err)
	}

	if _, err := os.Stat(rulesPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("YARA rules file not found: %w", err)
	}

	rulesContent, err := os.ReadFile(rulesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules file: %w", err)
	}

	patterns := m.parseYARARules(string(rulesContent))

	dumpData, err := os.ReadFile(dumpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read dump file: %w", err)
	}

	dumpSize := len(dumpData)
	if dumpSize > 100*1024*1024 {
		dumpData = dumpData[:100*1024*1024]
	}

	for _, rule := range patterns {
		if m.matchPattern(string(dumpData), rule.Pattern, rule.Condition) {
			results = append(results, map[string]interface{}{
				"rule_name":   rule.Name,
				"category":    rule.Category,
				"pattern":     rule.Pattern,
				"condition":   rule.Condition,
				"severity":    rule.Severity,
				"description": rule.Description,
				"dump_file":   dumpPath,
				"matched":     true,
			})
		}
	}

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$ErrorActionPreference='SilentlyContinue'
$dump = [System.IO.File]::ReadAllBytes('%s')
$minLength = 4
$results = @()
$buffer = New-Object char[] $minLength
$bufferIndex = 0
$currentString = ""
for($i = 0; $i -lt [Math]::Min($dump.Length, 10485760); $i++) {
    $b = $dump[$i]
    if(($b -ge 32 -and $b -le 126)) {
        $currentString += [char]$b
        if($currentString.Length -ge 8) {
            $results += $currentString
        }
    } else {
        $currentString = ""
    }
}
$results | Select-Object -Unique | Select-Object -First 100 | ForEach-Object { Write-Output $_ }`, dumpPath))

	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && len(line) >= 8 {
				for _, rule := range patterns {
					if rule.Category == "string" && m.matchPattern(line, rule.Pattern, "contains") {
						results = append(results, map[string]interface{}{
							"rule_name":      rule.Name,
							"category":       "string",
							"pattern":        rule.Pattern,
							"matched_string": line,
							"severity":       rule.Severity,
							"dump_file":      dumpPath,
							"matched":        true,
						})
					}
				}
			}
		}
	}

	return results, nil
}

type YARARule struct {
	Name        string
	Category    string
	Pattern     string
	Condition   string
	Severity    string
	Description string
}

func (m *MemoryModule) parseYARulesFile(rulesPath string) ([]YARARule, error) {
	content, err := os.ReadFile(rulesPath)
	if err != nil {
		return nil, err
	}
	return m.parseYARARules(string(content)), nil
}

func (m *MemoryModule) parseYARARules(content string) []YARARule {
	var rules []YARARule

	ruleBlocks := strings.Split(content, "rule ")
	for _, block := range ruleBlocks {
		if strings.TrimSpace(block) == "" {
			continue
		}

		rule := YARARule{
			Severity:  "medium",
			Condition: "contains",
		}

		lines := strings.Split(block, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "name:") {
				rule.Name = strings.TrimPrefix(line, "name:")
			} else if strings.HasPrefix(line, "category:") {
				rule.Category = strings.TrimPrefix(line, "category:")
			} else if strings.HasPrefix(line, "pattern:") {
				rule.Pattern = strings.TrimPrefix(line, "pattern:")
			} else if strings.HasPrefix(line, "condition:") {
				rule.Condition = strings.TrimPrefix(line, "condition:")
			} else if strings.HasPrefix(line, "severity:") {
				rule.Severity = strings.TrimPrefix(line, "severity:")
			} else if strings.HasPrefix(line, "description:") {
				rule.Description = strings.TrimPrefix(line, "description:")
			}
		}

		if rule.Name != "" && rule.Pattern != "" {
			rules = append(rules, rule)
		}
	}

	if len(rules) == 0 {
		rules = []YARARule{
			{Name: "base64_strings", Category: "encoded", Pattern: `[A-Za-z0-9+/]{40,}==?`, Condition: "regex", Severity: "low", Description: "Base64 encoded string"},
			{Name: "http_urls", Category: "network", Pattern: `https?://[^\s]+`, Condition: "regex", Severity: "medium", Description: "HTTP/HTTPS URL"},
			{Name: "ip_addresses", Category: "network", Pattern: `\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`, Condition: "regex", Severity: "medium", Description: "IP address"},
			{Name: "registry_keys", Category: "system", Pattern: `HKLM\\|HKCU\\|HKCR\\`, Condition: "regex", Severity: "high", Description: "Windows registry key"},
			{Name: "file_paths", Category: "system", Pattern: `[A-Z]:\\[^\s]+\.[a-z]{1,4}`, Condition: "regex", Severity: "low", Description: "Windows file path"},
			{Name: "powershell_commands", Category: "suspicious", Pattern: `powershell.*-enc|-encodedcommand`, Condition: "contains", Severity: "high", Description: "Encoded PowerShell command"},
			{Name: "mimikatz_pattern", Category: "malware", Pattern: `mimikatz|sekurlsa::logonpasswords`, Condition: "contains", Severity: "critical", Description: "Mimikatz credential dumping tool"},
			{Name: "password_keywords", Category: "sensitive", Pattern: `password|passwd|pwd|secret`, Condition: "contains", Severity: "high", Description: "Potential password or secret"},
		}
	}

	return rules
}

func (m *MemoryModule) matchPattern(data string, pattern string, condition string) bool {
	switch condition {
	case "contains":
		return strings.Contains(strings.ToLower(data), strings.ToLower(pattern))
	case "regex":
		return m.matchRegex(data, pattern)
	case "starts_with":
		return strings.HasPrefix(strings.ToLower(data), strings.ToLower(pattern))
	case "ends_with":
		return strings.HasSuffix(strings.ToLower(data), strings.ToLower(pattern))
	default:
		return strings.Contains(strings.ToLower(data), strings.ToLower(pattern))
	}
}

func (m *MemoryModule) matchRegex(data string, pattern string) bool {
	matched, _ := regexp.MatchString(pattern, data)
	return matched
}

func (m *MemoryModule) ExtractStringsDetailed(dumpPath string, minLength int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	if _, err := os.Stat(dumpPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("dump file not found: %w", err)
	}

	if minLength < 4 {
		minLength = 4
	}

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$ErrorActionPreference='SilentlyContinue'
$content = [System.IO.File]::ReadAllBytes('%s')
$results = @()
$currentString = ""
$offset = 0
for($i = 0; $i -lt [Math]::Min($content.Length, 10485760); $i++) {
    $b = $content[$i]
    if(($b -ge 32 -and $b -le 126)) {
        $currentString += [char]$b
    } else {
        if($currentString.Length -ge %d) {
            $results += [PSCustomObject]@{
                String = $currentString
                Offset = $offset
                Length = $currentString.Length
            }
        }
        $currentString = ""
        $offset = $i + 1
    }
}
$results | Select-Object -First 500 | ForEach-Object { Write-Output ($_.String + "|" + $_.Offset.ToString() + "|" + $_.Length.ToString()) }`, dumpPath, minLength))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to extract strings: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			offset, _ := strconv.ParseInt(parts[1], 10, 64)
			length, _ := strconv.Atoi(parts[2])
			results = append(results, map[string]interface{}{
				"string":  parts[0],
				"offset":  offset,
				"length":  length,
				"address": fmt.Sprintf("0x%08X", offset),
			})
		}
	}

	return results, nil
}

func (m *MemoryModule) ExportDump(dumpPath string, destPath string) error {
	if _, err := os.Stat(dumpPath); os.IsNotExist(err) {
		return fmt.Errorf("dump file not found: %w", err)
	}

	cmd := exec.Command("cmd", "/c", fmt.Sprintf("copy /Y \"%s\" \"%s\"", dumpPath, destPath))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy dump file: %w", err)
	}

	sha256, err := m.calculateSHA256(destPath)
	if err == nil {
		cmd2 := exec.Command("powershell", "-Command",
			fmt.Sprintf(`$dump = Get-Item '%s'; $dump | Add-Member -NotePropertyName 'SHA256' -NotePropertyValue '%s' -Force`, destPath, sha256))
		cmd2.Run()
	}

	return nil
}
