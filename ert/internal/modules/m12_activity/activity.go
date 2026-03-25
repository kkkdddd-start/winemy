//go:build windows

package m12_activity

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type ActivityModule struct {
	ctx     context.Context
	storage registry.Storage
	recent  []RecentFileDTO
	usb     []USBDeviceDTO
	browser []BrowserHistoryDTO
}

type RecentFileDTO struct {
	Path       string          `json:"path"`
	Name       string          `json:"name"`
	TargetPath string          `json:"target_path"`
	Accessed   time.Time       `json:"accessed"`
	Created    time.Time       `json:"created"`
	RiskLevel  model.RiskLevel `json:"risk_level"`
}

type USBDeviceDTO struct {
	DeviceID   string          `json:"device_id"`
	Name       string          `json:"name"`
	LastInsert time.Time       `json:"last_insert"`
	RiskLevel  model.RiskLevel `json:"risk_level"`
}

type BrowserHistoryDTO struct {
	URL        string          `json:"url"`
	Title      string          `json:"title"`
	VisitCount int             `json:"visit_count"`
	VisitedAt  time.Time       `json:"visited_at"`
	RiskLevel  model.RiskLevel `json:"risk_level"`
}

func New() *ActivityModule {
	return &ActivityModule{}
}

func (m *ActivityModule) ID() int       { return 12 }
func (m *ActivityModule) Name() string  { return "activity" }
func (m *ActivityModule) Priority() int { return 2 }

func (m *ActivityModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	return nil
}

func (m *ActivityModule) Collect(ctx context.Context) error {
	m.recent = []RecentFileDTO{}
	m.usb = []USBDeviceDTO{}
	m.browser = []BrowserHistoryDTO{}

	m.collectRecentFiles()
	m.collectUSBDevices()
	m.collectBrowserHistory()

	return nil
}

func (m *ActivityModule) collectRecentFiles() {
	currentUser, err := user.Current()
	if err != nil {
		return
	}

	recentPath := filepath.Join(currentUser.HomeDir, "AppData", "Roaming", "Microsoft", "Windows", "Recent")

	entries, err := os.ReadDir(recentPath)
	if err != nil {
		return
	}

	count := 0
	maxRecent := 50
	for _, entry := range entries {
		if count >= maxRecent {
			break
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		name := entry.Name()
		recentFullPath := filepath.Join(recentPath, entry.Name())

		isLnk := strings.HasSuffix(strings.ToLower(name), ".lnk")
		targetPath := ""
		if isLnk {
			targetPath = m.parseLnkShortcut(recentFullPath)
			name = name[:len(name)-4]
		}

		riskLevel := model.RiskLow
		if isSuspiciousPath(recentFullPath) || (targetPath != "" && isSuspiciousPath(targetPath)) {
			riskLevel = model.RiskHigh
		}

		m.recent = append(m.recent, RecentFileDTO{
			Path:       recentFullPath,
			Name:       name,
			TargetPath: targetPath,
			Accessed:   info.ModTime(),
			Created:    m.getLnkCreationTime(recentFullPath),
			RiskLevel:  riskLevel,
		})
		count++
	}
}

func (m *ActivityModule) parseLnkShortcut(lnkPath string) string {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$shell = New-Object -ComObject WScript.Shell; $shortcut = $shell.CreateShortcut('%s'); Write-Output $shortcut.TargetPath`, lnkPath))
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func (m *ActivityModule) getLnkCreationTime(lnkPath string) time.Time {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`(Get-Item '%s' -Force -ErrorAction SilentlyContinue).CreationTime.ToString("yyyy-MM-ddTHH:mm:ssZ")`, lnkPath))
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

func (m *ActivityModule) collectUSBDevices() {
	m.usb = []USBDeviceDTO{}

	cmd := exec.Command("powershell", "-Command",
		`Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Enum\USBSTOR\*\*' -ErrorAction SilentlyContinue | Select-Object FriendlyName, DeviceDesc, Mfg, SerialNumber, Class | ConvertTo-Json`)
	output, err := cmd.Output()
	if err != nil {
		m.collectUSBDevicesFromRegistry()
		return
	}

	var devices []map[string]interface{}
	if err := json.Unmarshal(output, &devices); err != nil {
		var single map[string]interface{}
		if err := json.Unmarshal(output, &single); err == nil {
			devices = []map[string]interface{}{single}
		} else {
			m.collectUSBDevicesFromRegistry()
			return
		}
	}

	for _, dev := range devices {
		if dev["FriendlyName"] == nil {
			continue
		}
		deviceID := ""
		if dev["DeviceDesc"] != nil {
			deviceID = dev["DeviceDesc"].(string)
		}
		name := ""
		if dev["FriendlyName"] != nil {
			name = dev["FriendlyName"].(string)
		}

		m.usb = append(m.usb, USBDeviceDTO{
			DeviceID:   deviceID,
			Name:       name,
			LastInsert: time.Now(),
			RiskLevel:  model.RiskLow,
		})
	}
}

func (m *ActivityModule) collectUSBDevicesFromRegistry() {
	usbKeys := []string{
		`HKLM:\SYSTEM\CurrentControlSet\Enum\USBSTOR`,
		`HKLM:\SYSTEM\CurrentControlSet\Enum\USB`,
	}

	for _, key := range usbKeys {
		cmd := exec.Command("powershell", "-Command",
			fmt.Sprintf(`Get-ItemProperty -Path '%s' -ErrorAction SilentlyContinue | ForEach-Object { Get-ItemProperty $_.PSPath -ErrorAction SilentlyContinue }`, key))
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "FriendlyName") || strings.Contains(line, "DeviceDesc") {
				m.usb = append(m.usb, USBDeviceDTO{
					DeviceID:   strings.TrimSpace(line),
					Name:       "USB Device",
					LastInsert: time.Now(),
					RiskLevel:  model.RiskLow,
				})
			}
		}
	}
}

func (m *ActivityModule) collectBrowserHistory() {
	browserPaths := map[string][]string{
		"Chrome": {
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Google", "Chrome", "User Data", "Default", "History"),
		},
		"Edge": {
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Edge", "User Data", "Default", "History"),
		},
		"Firefox": {
			filepath.Join(os.Getenv("APPDATA"), "Mozilla", "Firefox", "Profiles"),
		},
	}

	for browserName, paths := range browserPaths {
		for _, historyPath := range paths {
			if browserName == "Firefox" {
				entries, err := os.ReadDir(historyPath)
				if err != nil {
					continue
				}
				for _, entry := range entries {
					if entry.IsDir() && strings.HasSuffix(entry.Name(), ".default") {
						historyPath = filepath.Join(historyPath, entry.Name(), "places.sqlite")
						break
					}
				}
			}

			if _, err := os.Stat(historyPath); os.IsNotExist(err) {
				continue
			}

			m.readBrowserHistory(browserName, historyPath)
		}
	}
}

func (m *ActivityModule) readBrowserHistory(browserName, dbPath string) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return
	}

	switch browserName {
	case "Chrome", "Edge":
		m.readChromeEdgeHistory(browserName, dbPath)
	case "Firefox":
		m.readFirefoxHistory(dbPath)
	}
}

func (m *ActivityModule) readChromeEdgeHistory(browserName, dbPath string) {
	tempDb := dbPath + ".tmp"
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$ErrorActionPreference='SilentlyContinue'
$history = '%s'
if(Test-Path $history) {
    Copy-Item $history '%s' -Force
    $conn = New-Object System.Data.SQLite.SQLiteConnection
    $conn.ConnectionString = 'Data Source=%s;Version=3;'
    $conn.Open()
    $cmd = $conn.CreateCommand()
    $cmd.CommandText = 'SELECT url, title, visit_count, last_visit_time FROM urls ORDER BY last_visit_time DESC LIMIT 100'
    $adapter = New-Object System.Data.SQLite.SQLiteDataAdapter $cmd
    $dataset = New-Object System.Data.DataSet
    [void]$adapter.Fill($dataset)
    $conn.Close()
    Remove-Item '%s' -Force -ErrorAction SilentlyContinue
    foreach($row in $dataset.Tables[0].Rows) {
        $timestamp = [long]$row.last_visit_time
        $visiteTime = (Get-Date '1601-01-01').AddSeconds($timestamp / 10000000).ToString('yyyy-MM-ddTHH:mm:ssZ')
        Write-Output ($row.url + '|' + $row.title + '|' + $row.visit_count + '|' + $visiteTime)
    }
}`, dbPath, tempDb, tempDb, tempDb))

	output, err := cmd.Output()
	if err != nil {
		m.readBrowserHistoryFallback(browserName, dbPath)
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 1 {
			url := parts[0]
			title := ""
			visitCount := 0
			visitedAt := time.Now()

			if len(parts) > 1 {
				title = parts[1]
			}
			if len(parts) > 2 {
				if count, err := strconv.Atoi(parts[2]); err == nil {
					visitCount = count
				}
			}
			if len(parts) > 3 {
				if t, err := time.Parse("2006-01-02T15:04:05Z", parts[3]); err == nil {
					visitedAt = t
				}
			}

			m.browser = append(m.browser, BrowserHistoryDTO{
				URL:        url,
				Title:      title,
				VisitCount: visitCount,
				VisitedAt:  visitedAt,
				RiskLevel:  model.RiskLow,
			})
		}
	}
}

func (m *ActivityModule) readFirefoxHistory(dbPath string) {
	tempDb := dbPath + ".tmp"
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$ErrorActionPreference='SilentlyContinue'
$places = '%s'
if(Test-Path $places) {
    Copy-Item $places '%s' -Force
    $conn = New-Object System.Data.SQLite.SQLiteConnection
    $conn.ConnectionString = 'Data Source=%s;Version=3;'
    $conn.Open()
    $cmd = $conn.CreateCommand()
    $cmd.CommandText = 'SELECT url, title, visit_count FROM moz_places ORDER BY last_visit_date DESC LIMIT 100'
    $adapter = New-Object System.Data.SQLite.SQLiteDataAdapter $cmd
    $dataset = New-Object System.Data.DataSet
    [void]$adapter.Fill($dataset)
    $conn.Close()
    Remove-Item '%s' -Force -ErrorAction SilentlyContinue
    $dataset.Tables[0] | ForEach-Object { Write-Output ($_.url + '|' + $_.title + '|' + $_.visit_count) }
}`, dbPath, tempDb, tempDb, tempDb))

	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 1 {
			url := parts[0]
			title := ""
			if len(parts) > 1 {
				title = parts[1]
			}

			m.browser = append(m.browser, BrowserHistoryDTO{
				URL:       url,
				Title:     title,
				VisitedAt: time.Now(),
				RiskLevel: model.RiskLow,
			})
		}
	}
}

func (m *ActivityModule) readBrowserHistoryFallback(browserName, dbPath string) {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$ErrorActionPreference='SilentlyContinue'
$db = '%s'
$tempDb = $db + '.bak'
Copy-Item $db $tempDb -Force
$lines = @()
try {
    $conn = New-Object System.Data.SQLite.SQLiteConnection
    $conn.ConnectionString = 'Data Source=' + $tempDb + ';Version=3;'
    $conn.Open()
    $reader = $conn.CreateCommand()
    $reader.CommandText = 'SELECT url, title FROM urls LIMIT 50'
    $result = $reader.ExecuteReader()
    while($result.Read()) {
        $lines += $result['url'] + '|' + $result['title']
    }
    $result.Close()
    $conn.Close()
} catch { }
Remove-Item $tempDb -Force -ErrorAction SilentlyContinue
$lines`, dbPath))

	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 1 {
			url := parts[0]
			title := ""
			if len(parts) > 1 {
				title = parts[1]
			}

			m.browser = append(m.browser, BrowserHistoryDTO{
				URL:       url,
				Title:     title,
				VisitedAt: time.Now(),
				RiskLevel: model.RiskLow,
			})
		}
	}
}

func isSuspiciousPath(path string) bool {
	pathLower := strings.ToLower(path)

	suspiciousPatterns := []string{
		"temp",
		"tmp",
		"downloads",
		"desktop",
		"public",
		"suspicious",
		"malware",
		"trojan",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(pathLower, pattern) {
			return true
		}
	}
	return false
}

func (m *ActivityModule) Stop() error {
	return nil
}

func (m *ActivityModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0)

	for _, r := range m.recent {
		accessedStr := ""
		if !r.Accessed.IsZero() {
			accessedStr = r.Accessed.Format(time.RFC3339)
		}
		result = append(result, map[string]interface{}{
			"type":       "recent_file",
			"path":       r.Path,
			"name":       r.Name,
			"accessed":   accessedStr,
			"risk_level": r.RiskLevel,
		})
	}

	for _, u := range m.usb {
		lastInsertStr := ""
		if !u.LastInsert.IsZero() {
			lastInsertStr = u.LastInsert.Format(time.RFC3339)
		}
		result = append(result, map[string]interface{}{
			"type":        "usb",
			"device_id":   u.DeviceID,
			"name":        u.Name,
			"last_insert": lastInsertStr,
			"risk_level":  u.RiskLevel,
		})
	}

	for _, b := range m.browser {
		visitedStr := ""
		if !b.VisitedAt.IsZero() {
			visitedStr = b.VisitedAt.Format(time.RFC3339)
		}
		result = append(result, map[string]interface{}{
			"type":       "browser",
			"url":        b.URL,
			"title":      b.Title,
			"visited_at": visitedStr,
			"risk_level": b.RiskLevel,
		})
	}

	return result, nil
}

func (m *ActivityModule) CollectRDPHistory() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	cmd := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
Get-ItemProperty -Path 'HKCU:\Software\Microsoft\Terminal Server Client' -Name '*' | ForEach-Object {
    $server = $_.Default
    if($server) {
        Write-Output "Server:$server"
    }
}`)

	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Server:") {
				results = append(results, map[string]interface{}{
					"type":       "rdp_history",
					"server":     strings.TrimPrefix(line, "Server:"),
					"risk_level": model.RiskLow,
				})
			}
		}
	}

	cmd2 := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
$rdpFiles = Get-ChildItem -Path "$env:USERPROFILE\Documents" -Filter "*.rdp" -ErrorAction SilentlyContinue
foreach($f in $rdpFiles) {
    Write-Output ("RDPFile:" + $f.FullName)
}`)

	output2, err := cmd2.Output()
	if err == nil {
		lines := strings.Split(string(output2), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "RDPFile:") {
				results = append(results, map[string]interface{}{
					"type":       "rdp_file",
					"path":       strings.TrimPrefix(line, "RDPFile:"),
					"risk_level": model.RiskMedium,
				})
			}
		}
	}

	cmd3 := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
wevtutil qe Security /c:20 /f:text /rd:true | Select-String -Pattern "4624.*LogonType.*10" | Select-Object -First 10`)

	output3, err := cmd3.Output()
	if err == nil {
		lines := strings.Split(string(output3), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				results = append(results, map[string]interface{}{
					"type":       "rdp_login",
					"event":      strings.TrimSpace(line),
					"risk_level": model.RiskMedium,
				})
			}
		}
	}

	return results, nil
}

func (m *ActivityModule) CollectShareHistory() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	cmd := exec.Command("net", "share")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "Share name") && !strings.HasPrefix(line, "-") {
				fields := strings.Fields(line)
				if len(fields) >= 3 {
					results = append(results, map[string]interface{}{
						"type":       "share",
						"name":       fields[0],
						"path":       fields[2],
						"risk_level": model.RiskLow,
					})
				}
			}
		}
	}

	cmd2 := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
Get-ItemProperty -Path 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Network\Lanman\*' | ForEach-Object {
    if($_.Pipe) {
        Write-Output ("NamedPipe:" + $_.Pipe)
    }
}`)

	output2, err := cmd2.Output()
	if err == nil {
		lines := strings.Split(string(output2), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "NamedPipe:") {
				results = append(results, map[string]interface{}{
					"type":       "named_pipe",
					"pipe":       strings.TrimPrefix(line, "NamedPipe:"),
					"risk_level": model.RiskMedium,
				})
			}
		}
	}

	return results, nil
}

func (m *ActivityModule) CollectPrintHistory() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	cmd := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
Get-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Explorer\RunMru' | Select-Object -ExpandProperty MRUList | ForEach-Object {
    $value = Get-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Explorer\RunMru' -Name $_ -ErrorAction SilentlyContinue
    if($value.$_ -match 'print| spool') {
        Write-Output ("PrintCommand:" + $value.$_)
    }
}`)

	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "PrintCommand:") {
				results = append(results, map[string]interface{}{
					"type":       "print_command",
					"command":    strings.TrimPrefix(line, "PrintCommand:"),
					"risk_level": model.RiskLow,
				})
			}
		}
	}

	cmd2 := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
Get-WinEvent -LogName 'Microsoft-Windows-PrintService/Operational' -MaxEvents 50 -ErrorAction SilentlyContinue | ForEach-Object {
    if($_.Message -match 'document|printed') {
        Write-Output ("PrintEvent:" + $_.TimeCreated.ToString('yyyy-MM-dd HH:mm:ss') + "|" + $_.Message.Substring(0, [Math]::Min(100, $_.Message.Length)))
    }
}`)

	output2, err := cmd2.Output()
	if err == nil {
		lines := strings.Split(string(output2), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "PrintEvent:") {
				parts := strings.Split(strings.TrimPrefix(line, "PrintEvent:"), "|")
				if len(parts) >= 2 {
					results = append(results, map[string]interface{}{
						"type":       "print_event",
						"timestamp":  parts[0],
						"message":    parts[1],
						"risk_level": model.RiskLow,
					})
				}
			}
		}
	}

	return results, nil
}

func (m *ActivityModule) Search(keyword string) ([]map[string]interface{}, error) {
	results := []map[string]interface{}{}
	keywordLower := strings.ToLower(keyword)

	for _, r := range m.recent {
		if strings.Contains(strings.ToLower(r.Name), keywordLower) ||
			strings.Contains(strings.ToLower(r.Path), keywordLower) {
			results = append(results, map[string]interface{}{
				"type":   "recent_file",
				"name":   r.Name,
				"path":   r.Path,
				"source": "recent",
			})
		}
	}

	for _, u := range m.usb {
		if strings.Contains(strings.ToLower(u.Name), keywordLower) ||
			strings.Contains(strings.ToLower(u.DeviceID), keywordLower) {
			results = append(results, map[string]interface{}{
				"type":   "usb_device",
				"name":   u.Name,
				"source": "usb",
			})
		}
	}

	for _, b := range m.browser {
		if strings.Contains(strings.ToLower(b.URL), keywordLower) ||
			strings.Contains(strings.ToLower(b.Title), keywordLower) {
			results = append(results, map[string]interface{}{
				"type":   "browser_history",
				"url":    b.URL,
				"title":  b.Title,
				"source": "browser",
			})
		}
	}

	return results, nil
}

func (m *ActivityModule) ExportJSON(filePath string) error {
	data := map[string]interface{}{
		"timestamp":     time.Now().Format(time.RFC3339),
		"recent_files":  m.recent,
		"usb_devices":   m.usb,
		"browser_items": m.browser,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(filePath, jsonData, 0644)
}

func (m *ActivityModule) ExportCSV(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Type", "Name", "Value", "Timestamp", "RiskLevel"})

	for _, r := range m.recent {
		writer.Write([]string{
			"recent_file",
			r.Name,
			r.Path,
			r.Accessed.Format(time.RFC3339),
			fmt.Sprintf("%v", r.RiskLevel),
		})
	}

	for _, u := range m.usb {
		writer.Write([]string{
			"usb",
			u.Name,
			u.DeviceID,
			u.LastInsert.Format(time.RFC3339),
			fmt.Sprintf("%v", u.RiskLevel),
		})
	}

	for _, b := range m.browser {
		writer.Write([]string{
			"browser",
			b.Title,
			b.URL,
			b.VisitedAt.Format(time.RFC3339),
			fmt.Sprintf("%v", b.RiskLevel),
		})
	}

	return nil
}
