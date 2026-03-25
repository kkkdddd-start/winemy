package m6_schedule

import (
	"context"
	"time"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type ScheduleModule struct {
	ctx     context.Context
	storage registry.Storage
	tasks   []model.ScheduledTaskDTO
}

func New() *ScheduleModule {
	return &ScheduleModule{}
}

func (m *ScheduleModule) ID() int       { return 6 }
func (m *ScheduleModule) Name() string  { return "schedule" }
func (m *ScheduleModule) Priority() int { return 1 }

func (m *ScheduleModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	return nil
}

func (m *ScheduleModule) Collect(ctx context.Context) error {
	m.tasks = []model.ScheduledTaskDTO{
		{
			Name:        "Task1",
			Path:        "\\Microsoft\\Windows\\Task1",
			State:       "Ready",
			LastRunTime: time.Now().Add(-24 * time.Hour),
			NextRunTime: time.Now().Add(24 * time.Hour),
			RiskLevel:   model.RiskLow,
		},
		{
			Name:        "SuspiciousTask",
			Path:        "\\Temp\\SuspiciousTask",
			State:       "Running",
			LastRunTime: time.Now().Add(-1 * time.Hour),
			NextRunTime: time.Now().Add(1 * time.Hour),
			RiskLevel:   model.RiskHigh,
		},
		{
			Name:        "DailyMaintenance",
			Path:        "\\Microsoft\\Windows\\Maintenance",
			State:       "Ready",
			LastRunTime: time.Now().Add(-12 * time.Hour),
			NextRunTime: time.Now().Add(12 * time.Hour),
			RiskLevel:   model.RiskLow,
		},
	}

	return nil
}

func (m *ScheduleModule) Stop() error {
	return nil
}

func (m *ScheduleModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.tasks))
	for _, t := range m.tasks {
		result = append(result, map[string]interface{}{
			"name":          t.Name,
			"path":          t.Path,
			"state":         t.State,
			"last_run_time": t.LastRunTime.Format(time.RFC3339),
			"next_run_time": t.NextRunTime.Format(time.RFC3339),
			"risk_level":    t.RiskLevel,
		})
	}
	return result, nil
}
