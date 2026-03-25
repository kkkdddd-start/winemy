//go:build windows

package m2_process

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
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
	pid := p.Pid
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$proc = Get-Process -Id %d -ErrorAction SilentlyContinue; if($proc -and $proc.Path) { (Get-WmiObject Win32_Process -Filter "ProcessId=%d" -ErrorAction SilentlyContinue).GetOwner().User }`, pid, pid))
	output, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("powershell", "-Command",
			fmt.Sprintf(`(Get-WmiObject Win32_Process -Filter "ProcessId=%d" -ErrorAction SilentlyContinue).GetOwner().User`, pid))
		output, err = cmd.Output()
		if err != nil {
			return ""
		}
	}
	user := strings.TrimSpace(string(output))
	if user == "" || user == "NULL" {
		return ""
	}
	return user
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

func (m *ProcessModule) DumpProcess(pid uint32, outputDir string) (*model.MemoryDumpDTO, error) {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil, fmt.Errorf("failed to create process: %w", err)
	}

	name, _ := p.Name()
	if name == "" {
		name = "unknown"
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := time.Now().Format("20060102150405")
	dumpFile := filepath.Join(outputDir, fmt.Sprintf("proc_%d_%s_%s.dmp", pid, name, timestamp))

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$procdump = Get-Command procdump -ErrorAction SilentlyContinue; if($procdump) { procdump -accepteula -ma %d "%s" } else { Add-Type -AssemblyName System.Diagnostics; $proc = Get-Process -Id %d; $dump = [System.Diagnostics.ProcessModule]::new(); $fs = [System.IO.File]::Create("%s"); $ptr = $proc.MainModule.BaseAddress; [void][System.Diagnostics.ProcessModule].GetMethod("GetModuleMemory", [System.Reflection.BindingFlags]36).Invoke($null, @($ptr, $fs)); $fs.Close() }`, pid, dumpFile, pid, dumpFile))

	output, err := cmd.CombinedOutput()
	if err != nil {
		_ = os.WriteFile(dumpFile+".txt", output, 0644)
	}

	if _, err := os.Stat(dumpFile); os.IsNotExist(err) {
		miniDumpCmd := exec.Command("powershell", "-Command",
			fmt.Sprintf(`$ErrorActionPreference='SilentlyContinue'; Add-Type @"
using System;
using System.Runtime.InteropServices;
using System.IO;
public class MiniDump {
    [DllImport("dbghelp.dll", SetLastError=true)]
    public static extern bool MiniDumpWriteDump(IntPtr hProcess, uint processId, IntPtr hFile, uint dumpType, IntPtr expParam, IntPtr userStreamParam, IntPtr callbackParam);
    public static uint DUMP_NORMAL = 0;
}
"@
$hProcess = (Get-Process -Id %d).Handle
$fs = [System.IO.File]::Create("%s")
MiniDump::MiniDumpWriteDump($hProcess, %d, $fs.Handle, MiniDump::DUMP_NORMAL, [IntPtr]::Zero, [IntPtr]::Zero, [IntPtr]::Zero)
$fs.Close()
if(Test-Path "%s") { Write-Output "SUCCESS" } else { Write-Output "FAILED" }`, pid, dumpFile, pid, dumpFile))
		miniDumpOutput, _ := miniDumpCmd.CombinedOutput()
		if strings.Contains(string(miniDumpOutput), "SUCCESS") {
			if _, err := os.Stat(dumpFile); err == nil {
				return m.createMemoryDumpDTO(pid, name, "MiniDump", dumpFile)
			}
		}
		return nil, fmt.Errorf("failed to dump process: %s", string(miniDumpOutput))
	}

	if _, err := os.Stat(dumpFile); err == nil {
		return m.createMemoryDumpDTO(pid, name, "MiniDump", dumpFile)
	}

	return nil, fmt.Errorf("dump file not created")
}

func (m *ProcessModule) createMemoryDumpDTO(pid uint32, name, dumpType, dumpFile string) (*model.MemoryDumpDTO, error) {
	info, err := os.Stat(dumpFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get dump file info: %w", err)
	}

	hash, err := computeFileHash(dumpFile)
	if err != nil {
		return nil, fmt.Errorf("failed to compute hash: %w", err)
	}

	return &model.MemoryDumpDTO{
		PID:         pid,
		ProcessName: name,
		DumpType:    dumpType,
		FilePath:    dumpFile,
		FileSize:    uint64(info.Size()),
		SHA256:      hash,
		CreatedAt:   time.Now(),
	}, nil
}

func computeFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
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
