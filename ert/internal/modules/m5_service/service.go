package m5_service

import (
	"context"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type ServiceModule struct {
	ctx      context.Context
	storage  registry.Storage
	services []model.ServiceDTO
}

func New() *ServiceModule {
	return &ServiceModule{}
}

func (m *ServiceModule) ID() int       { return 5 }
func (m *ServiceModule) Name() string  { return "service" }
func (m *ServiceModule) Priority() int { return 1 }

func (m *ServiceModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	return nil
}

func (m *ServiceModule) Collect(ctx context.Context) error {
	m.services = []model.ServiceDTO{
		{
			Name:        "SampleService1",
			DisplayName: "Sample Service One",
			Status:      "Running",
			StartType:   "Automatic",
			Path:        "C:\\Windows\\System32\\sample1.exe",
			RiskLevel:   model.RiskLow,
		},
		{
			Name:        "SampleService2",
			DisplayName: "Sample Service Two",
			Status:      "Stopped",
			StartType:   "Manual",
			Path:        "C:\\Windows\\System32\\sample2.exe",
			RiskLevel:   model.RiskMedium,
		},
	}

	riskServices := []model.ServiceDTO{
		{
			Name:        "SuspiciousService",
			DisplayName: "Suspicious Windows Service",
			Status:      "Running",
			StartType:   "Automatic",
			Path:        "C:\\Temp\\malware.exe",
			RiskLevel:   model.RiskHigh,
		},
	}

	m.services = append(m.services, riskServices...)

	return nil
}

func (m *ServiceModule) Stop() error {
	return nil
}

func (m *ServiceModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.services))
	for _, s := range m.services {
		result = append(result, map[string]interface{}{
			"name":         s.Name,
			"display_name": s.DisplayName,
			"status":       s.Status,
			"start_type":   s.StartType,
			"path":         s.Path,
			"risk_level":   s.RiskLevel,
		})
	}
	return result, nil
}
