package m8_patch

import (
	"context"
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
	m.installed = []HotfixDTO{
		{
			HotfixID:    "KB5000001",
			Installed:   time.Now().AddDate(0, -1, 0),
			Description: "Security Update",
			RiskLevel:   model.RiskLow,
		},
		{
			HotfixID:    "KB5000002",
			Installed:   time.Now().AddDate(0, 0, -15),
			Description: "Security Update",
			RiskLevel:   model.RiskLow,
		},
		{
			HotfixID:    "KB5000003",
			Installed:   time.Now().AddDate(0, -2, 0),
			Description: "Critical Update",
			RiskLevel:   model.RiskLow,
		},
	}

	m.missing = []string{
		"KB5034441",
		"KB5034203",
	}

	return nil
}

func (m *PatchModule) Stop() error {
	return nil
}

func (m *PatchModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0)

	for _, p := range m.installed {
		result = append(result, map[string]interface{}{
			"type":        "installed",
			"hotfix_id":   p.HotfixID,
			"installed":   p.Installed.Format(time.RFC3339),
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
