//go:build windows

package m6_schedule

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"os/exec"
	"strings"
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
	m.tasks = []model.ScheduledTaskDTO{}

	output, err := exec.Command("schtasks", "/query", "/fo", "CSV", "/v").Output()
	if err != nil {
		m.tasks = append(m.tasks, model.ScheduledTaskDTO{
			Name:        "Error",
			Path:        fmt.Sprintf("Failed to query scheduled tasks: %v", err),
			State:       "Unknown",
			LastRunTime: time.Time{},
			NextRunTime: time.Time{},
			RiskLevel:   model.RiskLow,
		})
		return nil
	}

	reader := csv.NewReader(bytes.NewReader(output))
	records, err := reader.ReadAll()
	if err != nil {
		m.tasks = append(m.tasks, model.ScheduledTaskDTO{
			Name:        "Error",
			Path:        fmt.Sprintf("Failed to parse CSV: %v", err),
			State:       "Unknown",
			LastRunTime: time.Time{},
			NextRunTime: time.Time{},
			RiskLevel:   model.RiskLow,
		})
		return nil
	}

	for i, record := range records {
		if i == 0 {
			continue
		}
		if len(record) < 11 {
			continue
		}

		taskName := strings.TrimSpace(record[1])
		nextRun := strings.TrimSpace(record[2])
		status := strings.TrimSpace(record[3])
		lastRun := strings.TrimSpace(record[5])
		taskToRun := strings.TrimSpace(record[8])

		if taskName == "" {
			continue
		}

		lastRunTime := parseTime(lastRun)
		nextRunTime := parseTime(nextRun)

		riskLevel := model.RiskLow
		if isSuspiciousTask(taskName, taskToRun) {
			riskLevel = model.RiskHigh
		}

		m.tasks = append(m.tasks, model.ScheduledTaskDTO{
			Name:        taskName,
			Path:        taskToRun,
			State:       status,
			LastRunTime: lastRunTime,
			NextRunTime: nextRunTime,
			RiskLevel:   riskLevel,
		})
	}

	return nil
}

func parseTime(timeStr string) time.Time {
	timeStr = strings.TrimSpace(timeStr)
	if timeStr == "N/A" || timeStr == "Never" || timeStr == "" {
		return time.Time{}
	}

	formats := []string{
		"1/2/2006 3:04:05 PM",
		"1/2/2006 15:04:05",
		"2006-01-02 15:04:05",
		"01/02/2006 15:04:05",
		"01/02/2006 3:04:05 PM",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t
		}
	}
	return time.Time{}
}

func isSuspiciousTask(name, path string) bool {
	nameLower := strings.ToLower(name)
	pathLower := strings.ToLower(path)

	suspiciousPatterns := []string{
		"temp",
		"tmp",
		"appdata",
		"downloads",
		"desktop",
		"public",
		"powershell",
		"cmd.exe",
		"wscript",
		"cscript",
		"mshta",
		"certutil",
		"bitsadmin",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(nameLower, pattern) || strings.Contains(pathLower, pattern) {
			return true
		}
	}
	return false
}

func (m *ScheduleModule) Stop() error {
	return nil
}

func (m *ScheduleModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.tasks))
	for _, t := range m.tasks {
		lastRunStr := ""
		if !t.LastRunTime.IsZero() {
			lastRunStr = t.LastRunTime.Format(time.RFC3339)
		}
		nextRunStr := ""
		if !t.NextRunTime.IsZero() {
			nextRunStr = t.NextRunTime.Format(time.RFC3339)
		}
		result = append(result, map[string]interface{}{
			"name":          t.Name,
			"path":          t.Path,
			"state":         t.State,
			"last_run_time": lastRunStr,
			"next_run_time": nextRunStr,
			"risk_level":    t.RiskLevel,
		})
	}
	return result, nil
}
