//go:build windows

package m12_activity

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
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
	Path      string          `json:"path"`
	Name      string          `json:"name"`
	Accessed  time.Time       `json:"accessed"`
	RiskLevel model.RiskLevel `json:"risk_level"`
}

type USBDeviceDTO struct {
	DeviceID   string          `json:"device_id"`
	Name       string          `json:"name"`
	LastInsert time.Time       `json:"last_insert"`
	RiskLevel  model.RiskLevel `json:"risk_level"`
}

type BrowserHistoryDTO struct {
	URL       string          `json:"url"`
	Title     string          `json:"title"`
	VisitedAt time.Time       `json:"visited_at"`
	RiskLevel model.RiskLevel `json:"risk_level"`
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
		if strings.HasSuffix(strings.ToLower(name), ".lnk") {
			name = name[:len(name)-4]
		}

		recentFullPath := filepath.Join(recentPath, entry.Name())

		riskLevel := model.RiskLow
		if isSuspiciousPath(recentFullPath) {
			riskLevel = model.RiskHigh
		}

		m.recent = append(m.recent, RecentFileDTO{
			Path:      recentFullPath,
			Name:      name,
			Accessed:  info.ModTime(),
			RiskLevel: riskLevel,
		})
		count++
	}
}

func (m *ActivityModule) collectUSBDevices() {
	m.usb = []USBDeviceDTO{}
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
	if browserName == "Chrome" || browserName == "Edge" {
		m.browser = append(m.browser, BrowserHistoryDTO{
			URL:       fmt.Sprintf("Browser: %s, History DB: %s", browserName, dbPath),
			Title:     "Browser history requires database reading (SQLite)",
			VisitedAt: time.Time{},
			RiskLevel: model.RiskLow,
		})
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
