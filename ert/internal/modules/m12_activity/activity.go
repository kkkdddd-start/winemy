package m12_activity

import (
	"context"
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
	m.recent = []RecentFileDTO{
		{
			Path:      "C:\\Users\\Admin\\Documents\\important.docx",
			Name:      "important.docx",
			Accessed:  time.Now().Add(-1 * time.Hour),
			RiskLevel: model.RiskLow,
		},
		{
			Path:      "C:\\Temp\\suspicious.bat",
			Name:      "suspicious.bat",
			Accessed:  time.Now().Add(-30 * time.Minute),
			RiskLevel: model.RiskHigh,
		},
		{
			Path:      "C:\\Windows\\System32\\config\\SYSTEM",
			Name:      "SYSTEM",
			Accessed:  time.Now().Add(-2 * time.Hour),
			RiskLevel: model.RiskMedium,
		},
	}

	m.usb = []USBDeviceDTO{
		{
			DeviceID:   "USB\\VID_0000&PID_0001",
			Name:       "Sandisk USB",
			LastInsert: time.Now().AddDate(0, 0, -3),
			RiskLevel:  model.RiskMedium,
		},
		{
			DeviceID:   "USB\\VID_0000&PID_0002",
			Name:       "Unknown USB Device",
			LastInsert: time.Now().Add(-1 * time.Hour),
			RiskLevel:  model.RiskHigh,
		},
	}

	m.browser = []BrowserHistoryDTO{
		{
			URL:       "https://google.com",
			Title:     "Google",
			VisitedAt: time.Now().Add(-10 * time.Minute),
			RiskLevel: model.RiskLow,
		},
		{
			URL:       "https://suspicious-site.com",
			Title:     "Suspicious Site",
			VisitedAt: time.Now().Add(-1 * time.Hour),
			RiskLevel: model.RiskHigh,
		},
	}

	return nil
}

func (m *ActivityModule) Stop() error {
	return nil
}

func (m *ActivityModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0)

	for _, r := range m.recent {
		result = append(result, map[string]interface{}{
			"type":       "recent_file",
			"path":       r.Path,
			"name":       r.Name,
			"accessed":   r.Accessed.Format(time.RFC3339),
			"risk_level": r.RiskLevel,
		})
	}

	for _, u := range m.usb {
		result = append(result, map[string]interface{}{
			"type":        "usb",
			"device_id":   u.DeviceID,
			"name":        u.Name,
			"last_insert": u.LastInsert.Format(time.RFC3339),
			"risk_level":  u.RiskLevel,
		})
	}

	for _, b := range m.browser {
		result = append(result, map[string]interface{}{
			"type":       "browser",
			"url":        b.URL,
			"title":      b.Title,
			"visited_at": b.VisitedAt.Format(time.RFC3339),
			"risk_level": b.RiskLevel,
		})
	}

	return result, nil
}
