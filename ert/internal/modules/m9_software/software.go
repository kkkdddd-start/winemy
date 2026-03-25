package m9_software

import (
	"context"
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
	m.software = []SoftwareDTO{
		{
			Name:      "Microsoft Windows",
			Version:   "10.0.19045",
			Vendor:    "Microsoft Corporation",
			Installed: time.Now().AddDate(-1, 0, 0),
			RiskLevel: model.RiskLow,
		},
		{
			Name:      "Google Chrome",
			Version:   "120.0.6099.130",
			Vendor:    "Google LLC",
			Installed: time.Now().AddDate(0, -3, 0),
			RiskLevel: model.RiskLow,
		},
		{
			Name:      "Suspicious Software",
			Version:   "1.0.0",
			Vendor:    "Unknown",
			Installed: time.Now().AddDate(0, 0, -5),
			RiskLevel: model.RiskHigh,
		},
	}

	return nil
}

func (m *SoftwareModule) Stop() error {
	return nil
}

func (m *SoftwareModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.software))
	for _, s := range m.software {
		result = append(result, map[string]interface{}{
			"name":       s.Name,
			"version":    s.Version,
			"vendor":     s.Vendor,
			"installed":  s.Installed.Format(time.RFC3339),
			"risk_level": s.RiskLevel,
		})
	}
	return result, nil
}
