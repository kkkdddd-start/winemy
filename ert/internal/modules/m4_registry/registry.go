package m4_registry

import (
	"context"
	"time"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type RegistryModule struct {
	ctx     context.Context
	storage registry.Storage
	keys    []model.RegistryKeyDTO
}

func New() *RegistryModule {
	return &RegistryModule{}
}

func (m *RegistryModule) ID() int       { return 4 }
func (m *RegistryModule) Name() string  { return "registry" }
func (m *RegistryModule) Priority() int { return 1 }

func (m *RegistryModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	return nil
}

func (m *RegistryModule) Collect(ctx context.Context) error {
	m.keys = make([]model.RegistryKeyDTO, 0)

	criticalKeys := []struct {
		Path string
		Name string
		Risk model.RiskLevel
	}{
		{`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, "Autorun", model.RiskHigh},
		{`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnce`, "RunOnce", model.RiskHigh},
		{`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, "UserAutorun", model.RiskMedium},
		{`HKLM\SYSTEM\CurrentControlSet\Services`, "Services", model.RiskMedium},
		{`HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon`, "Winlogon", model.RiskHigh},
		{`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`, "Uninstall", model.RiskLow},
	}

	for _, ck := range criticalKeys {
		m.keys = append(m.keys, model.RegistryKeyDTO{
			Path:      ck.Path,
			Name:      ck.Name,
			ValueType: "key",
			Value:     "",
			Modified:  time.Now(),
			RiskLevel: ck.Risk,
		})
	}

	autostartKeys := []string{
		`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`,
		`HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Run`,
		`HKLM\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Run`,
	}

	for _, key := range autostartKeys {
		m.keys = append(m.keys, model.RegistryKeyDTO{
			Path:      key,
			Name:      "AutostartLocation",
			ValueType: "key",
			Value:     "",
			Modified:  time.Now(),
			RiskLevel: model.RiskMedium,
		})
	}

	return nil
}

func (m *RegistryModule) Stop() error {
	return nil
}

func (m *RegistryModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.keys))
	for _, k := range m.keys {
		result = append(result, map[string]interface{}{
			"path":       k.Path,
			"name":       k.Name,
			"value_type": k.ValueType,
			"value":      k.Value,
			"modified":   k.Modified.Format(time.RFC3339),
			"risk_level": k.RiskLevel,
		})
	}
	return result, nil
}
