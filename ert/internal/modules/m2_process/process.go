package m2_process

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type ProcessModule struct {
	processes []model.ProcessDTO
}

func New() *ProcessModule {
	return &ProcessModule{}
}

func (m *ProcessModule) ID() int       { return 2 }
func (m *ProcessModule) Name() string  { return "process" }
func (m *ProcessModule) Priority() int { return 0 }

func (m *ProcessModule) Init(ctx context.Context, s registry.Storage) error {
	return nil
}

func (m *ProcessModule) Collect(ctx context.Context) error {
	procs, err := process.Processes()
	if err != nil {
		return err
	}

	m.processes = make([]model.ProcessDTO, 0, len(procs))
	for _, p := range procs {
		name, _ := p.Name()
		exe, _ := p.Exe()
		createTime, _ := p.CreateTime()
		memInfo, _ := p.MemoryInfo()
		cpuPercent, _ := p.CPUPercent()

		dto := model.ProcessDTO{
			PID:         uint32(p.Pid),
			Name:        name,
			Path:        exe,
			CommandLine: "",
			User:        "",
			CPU:         cpuPercent,
			Memory:      memInfo.RSS,
			StartTime:   time.Unix(int64(createTime)/1000, 0),
			RiskLevel:   model.RiskLow,
		}
		m.processes = append(m.processes, dto)
	}
	return nil
}

func (m *ProcessModule) Stop() error {
	return nil
}

func (m *ProcessModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.processes))
	for _, p := range m.processes {
		result = append(result, map[string]interface{}{
			"pid":          p.PID,
			"name":         p.Name,
			"path":         p.Path,
			"command_line": p.CommandLine,
			"user":         p.User,
			"cpu":          p.CPU,
			"memory":       p.Memory,
			"start_time":   p.StartTime.Format(time.RFC3339),
			"risk_level":   p.RiskLevel,
		})
	}
	return result, nil
}
