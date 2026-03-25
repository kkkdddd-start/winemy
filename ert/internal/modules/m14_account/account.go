package m14_account

import (
	"context"
	"time"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type AccountModule struct {
	ctx      context.Context
	storage  registry.Storage
	accounts []model.AccountDTO
}

func New() *AccountModule {
	return &AccountModule{}
}

func (m *AccountModule) ID() int       { return 14 }
func (m *AccountModule) Name() string  { return "account" }
func (m *AccountModule) Priority() int { return 1 }

func (m *AccountModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	return nil
}

func (m *AccountModule) Collect(ctx context.Context) error {
	m.accounts = []model.AccountDTO{
		{
			Name:      "Administrator",
			FullName:  "Built-in account for administering the computer/domain",
			SID:       "S-1-5-21-1234567890-1234567890-1234567890-500",
			Domain:    "DESKTOP-ABC123",
			Status:    "Enabled",
			LastLogon: time.Now().AddDate(0, 0, -1),
			RiskLevel: model.RiskHigh,
		},
		{
			Name:      "Guest",
			FullName:  "Guest",
			SID:       "S-1-5-21-1234567890-1234567890-1234567890-501",
			Domain:    "DESKTOP-ABC123",
			Status:    "Disabled",
			LastLogon: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			RiskLevel: model.RiskMedium,
		},
		{
			Name:      "User1",
			FullName:  "User One",
			SID:       "S-1-5-21-1234567890-1234567890-1234567890-1001",
			Domain:    "DESKTOP-ABC123",
			Status:    "Enabled",
			LastLogon: time.Now().Add(-2 * time.Hour),
			RiskLevel: model.RiskLow,
		},
		{
			Name:      "SuspiciousUser",
			FullName:  "Suspicious Account",
			SID:       "S-1-5-21-9999999999-9999999999-9999999999-9999",
			Domain:    "UNKNOWN",
			Status:    "Enabled",
			LastLogon: time.Now().Add(-30 * time.Minute),
			RiskLevel: model.RiskCritical,
		},
	}

	return nil
}

func (m *AccountModule) Stop() error {
	return nil
}

func (m *AccountModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.accounts))
	for _, a := range m.accounts {
		result = append(result, map[string]interface{}{
			"name":       a.Name,
			"full_name":  a.FullName,
			"sid":        a.SID,
			"domain":     a.Domain,
			"status":     a.Status,
			"last_logon": a.LastLogon.Format(time.RFC3339),
			"risk_level": a.RiskLevel,
		})
	}
	return result, nil
}
