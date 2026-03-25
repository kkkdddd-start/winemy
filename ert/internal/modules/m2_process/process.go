//go:build windows

package m2_process

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type ProcessModule struct {
	ctx       context.Context
	storage   registry.Storage
	processes []model.ProcessDTO
	tree      map[uint32]*model.ProcessTreeNode
}

func New() *ProcessModule {
	return &ProcessModule{
		tree: make(map[uint32]*model.ProcessTreeNode),
	}
}

func (m *ProcessModule) ID() int       { return 2 }
func (m *ProcessModule) Name() string  { return "process" }
func (m *ProcessModule) Priority() int { return 0 }

func (m *ProcessModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	return nil
}

func (m *ProcessModule) Collect(ctx context.Context) error {
	procs, err := process.Processes()
	if err != nil {
		return err
	}

	m.processes = make([]model.ProcessDTO, 0, len(procs))
	m.tree = make(map[uint32]*model.ProcessTreeNode)

	for _, p := range procs {
		pid := uint32(p.Pid)

		name, _ := p.Name()
		exe, _ := p.Exe()
		cmdline, _ := p.CmdlineSlice()
		createTime, _ := p.CreateTime()
		memInfo, _ := p.MemoryInfo()
		cpuPercent, _ := p.CPUPercent()
		ppid, _ := p.Ppid()

		dto := model.ProcessDTO{
			PID:         pid,
			PPID:        uint32(ppid),
			Name:        name,
			Path:        exe,
			CommandLine: strings.Join(cmdline, " "),
			User:        getProcessUser(p),
			CPU:         cpuPercent,
			Memory:      memInfo.RSS,
			StartTime:   time.Unix(int64(createTime)/1000, 0),
			RiskLevel:   assessRiskLevel(name, exe, cmdline),
		}
		m.processes = append(m.processes, dto)

		m.tree[pid] = &model.ProcessTreeNode{
			PID:      pid,
			Name:     name,
			Children: []*model.ProcessTreeNode{},
		}
	}

	m.buildProcessTree()
	return nil
}

func getProcessUser(p *process.Process) string {
	return ""
}

func assessRiskLevel(name, path string, cmdline []string) model.RiskLevel {
	nameLower := strings.ToLower(name)
	pathLower := strings.ToLower(path)
	fullCmd := strings.ToLower(strings.Join(cmdline, " "))

	suspiciousNames := []string{
		"mimikatz", "pwdump", "procdump", "lsass", "lsass.exe",
		"gsecdump", "fgdump", "hashdump",
		"metasploit", "meterpreter", "cobalt", "covenant",
		"empire", "powerup", "powerview",
	}

	suspiciousCmds := []string{
		"mimikatz", "pwdump", "procdump", "lsass",
		"powershell -enc", "powershell -encodedcommand",
		"IEX ", "Invoke-Expression",
		" downloadString", " downloadfile",
		"net user ", "net localgroup administrators",
		"reg add HKLM", "schtasks /create",
		"wmic process call create", "vssadmin delete",
		"bcdedit", "wbadmin delete",
	}

	for _, sus := range suspiciousNames {
		if strings.Contains(nameLower, sus) || strings.Contains(pathLower, sus) {
			return model.RiskHigh
		}
	}

	for _, sus := range suspiciousCmds {
		if strings.Contains(fullCmd, sus) {
			return model.RiskHigh
		}
	}

	if path == "" && name != "" {
		return model.RiskMedium
	}

	if strings.Contains(pathLower, "temp") || strings.Contains(pathLower, "tmp") {
		return model.RiskMedium
	}

	return model.RiskLow
}

func (m *ProcessModule) buildProcessTree() {
	for pid, node := range m.tree {
		ppid := m.processes[pid].PPID
		if ppid != 0 {
			if parent, ok := m.tree[ppid]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}
}

func (m *ProcessModule) GetProcessTree() []*model.ProcessTreeNode {
	var roots []*model.ProcessTreeNode
	for _, node := range m.tree {
		proc := m.getProcessByPID(node.PID)
		if proc != nil && (proc.PPID == 0 || proc.PPID == 4) {
			roots = append(roots, node)
		}
	}
	return roots
}

func (m *ProcessModule) getProcessByPID(pid uint32) *model.ProcessDTO {
	for _, p := range m.processes {
		if p.PID == pid {
			return &p
		}
	}
	return nil
}

func (m *ProcessModule) Search(keyword string) []model.ProcessDTO {
	results := []model.ProcessDTO{}
	keywordLower := strings.ToLower(keyword)
	for _, p := range m.processes {
		if strings.Contains(strings.ToLower(p.Name), keywordLower) ||
			strings.Contains(strings.ToLower(p.Path), keywordLower) ||
			strings.Contains(strings.ToLower(p.CommandLine), keywordLower) {
			results = append(results, p)
		}
	}
	return results
}

func (m *ProcessModule) KillProcess(pid uint32) error {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return fmt.Errorf("failed to create process: %w", err)
	}

	name, _ := p.Name()
	criticalProcs := map[string]bool{
		"system": true, "csrss.exe": true, "smss.exe": true,
		"wininit.exe": true, "services.exe": true, "lsass.exe": true,
		"winlogon.exe": true, "dwm.exe": true,
	}

	if criticalProcs[strings.ToLower(name)] {
		return fmt.Errorf("cannot kill critical system process: %s", name)
	}

	if err := p.Kill(); err != nil {
		return fmt.Errorf("failed to kill process: %w", err)
	}

	return nil
}

func (m *ProcessModule) DumpProcess(pid uint32, outputDir string) (string, error) {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return "", fmt.Errorf("failed to create process: %w", err)
	}

	name, _ := p.Name()

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	dumpFile := filepath.Join(outputDir, fmt.Sprintf("proc_%d_%s_%s.dmp",
		pid, name, time.Now().Format("20060102150405")))

	return dumpFile, nil
}

func (m *ProcessModule) Stop() error {
	return nil
}

func (m *ProcessModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.processes))
	for _, p := range m.processes {
		result = append(result, map[string]interface{}{
			"pid":          p.PID,
			"ppid":         p.PPID,
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

func (m *ProcessModule) GetProcess(pid uint32) *model.ProcessDTO {
	return m.getProcessByPID(pid)
}
