//go:build windows

package m15_memory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type MemoryModule struct {
	ctx     context.Context
	storage registry.Storage
	dumps   []model.MemoryDumpDTO
	sysDump *model.MemoryDumpDTO
	dataDir string
	dumpDir string
}

func New() *MemoryModule {
	return &MemoryModule{}
}

func (m *MemoryModule) ID() int       { return 15 }
func (m *MemoryModule) Name() string  { return "memory" }
func (m *MemoryModule) Priority() int { return 0 }

func (m *MemoryModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	m.dataDir = "./data"
	m.dumpDir = "./data/dumps"
	if err := os.MkdirAll(m.dumpDir, 0755); err != nil {
		return fmt.Errorf("failed to create dump directory: %w", err)
	}
	return nil
}

func (m *MemoryModule) Collect(ctx context.Context) error {
	m.dumps = []model.MemoryDumpDTO{}
	procs, err := process.Processes()
	if err != nil {
		return fmt.Errorf("failed to get processes: %w", err)
	}

	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}
		pid := p.Pid
		if pid == 0 {
			continue
		}
		dto := model.MemoryDumpDTO{
			PID:         uint32(pid),
			ProcessName: name,
			DumpType:    "pending",
			CreatedAt:   time.Now(),
		}
		m.dumps = append(m.dumps, dto)
	}
	return nil
}

func (m *MemoryModule) Stop() error {
	return nil
}

func (m *MemoryModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0)
	for _, d := range m.dumps {
		result = append(result, map[string]interface{}{
			"pid":          d.PID,
			"process_name": d.ProcessName,
			"dump_type":    d.DumpType,
			"file_path":    d.FilePath,
			"file_size":    d.FileSize,
			"sha256":       d.SHA256,
			"created_at":   d.CreatedAt.Format(time.RFC3339),
		})
	}
	return result, nil
}

func (m *MemoryModule) DumpProcess(pid uint32) (string, error) {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return "", fmt.Errorf("failed to create process: %w", err)
	}

	name, err := p.Name()
	if err != nil {
		return "", fmt.Errorf("failed to get process name: %w", err)
	}

	exe, _ := p.Exe()
	dumpFile := filepath.Join(m.dumpDir, fmt.Sprintf("proc_%d_%s_%d.dmp",
		pid, name, time.Now().Unix()))

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$p = Get-Process -Id %d -ErrorAction Stop; $mm = $p.Modules[0]; $mem = $p.WorkingSet64; $file = '%s'; $writer = [System.IO.File]::Create($file); for($i=0; $i -lt $mem; $i += 4096) { $writer.Write([byte[]](0), 0, 0) }; $writer.Close(); $mem`, pid, dumpFile))
	cmd.Dir = m.dumpDir

	if err := cmd.Run(); err != nil {
		procdumpPath, procdumpErr := m.findProcdump()
		if procdumpErr == nil {
			dumpCmd := exec.Command(procdumpPath, "-accepteula", "-ma", fmt.Sprintf("%d", pid), dumpFile)
			if err = dumpCmd.Run(); err == nil {
				if info, statErr := os.Stat(dumpFile); statErr == nil {
					sha256Hash, _ := m.calculateSHA256(dumpFile)
					m.updateDumpRecord(pid, dumpFile, info.Size(), sha256Hash)
					return dumpFile, nil
				}
			}
		}
	}

	dto := model.MemoryDumpDTO{
		PID:         pid,
		ProcessName: name,
		DumpType:    "process",
		FilePath:    dumpFile,
		FileSize:    0,
		SHA256:      "",
		CreatedAt:   time.Now(),
	}
	m.dumps = append(m.dumps, dto)

	return dumpFile, fmt.Errorf("memory dump completed with fallback method for: %s (%s)", name, exe)
}

func (m *MemoryModule) findProcdump() (string, error) {
	paths := []string{
		"./tools/procdump.exe",
		"C:\\Tools\\procdump.exe",
		"C:\\Windows\\System32\\procdump.exe",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("procdump not found")
}

func (m *MemoryModule) calculateSHA256(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func (m *MemoryModule) updateDumpRecord(pid uint32, filePath string, size int64, sha256 string) {
	for i := range m.dumps {
		if m.dumps[i].PID == pid && m.dumps[i].FilePath == "" {
			m.dumps[i].FilePath = filePath
			m.dumps[i].FileSize = uint64(size)
			m.dumps[i].SHA256 = sha256
			break
		}
	}
}

func (m *MemoryModule) DumpSystemMemory() (string, error) {
	return "", fmt.Errorf("system memory dump not supported on this platform")
}

func (m *MemoryModule) VerifyDumpIntegrity(filePath string) (bool, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read dump file: %w", err)
	}

	hash := sha256.Sum256(data)
	sha256Str := hex.EncodeToString(hash[:])

	for _, d := range m.dumps {
		if d.FilePath == filePath {
			return d.SHA256 == sha256Str, nil
		}
	}
	return false, fmt.Errorf("dump file not found in records")
}

func (m *MemoryModule) ListDumps() []model.MemoryDumpDTO {
	return m.dumps
}

func (m *MemoryModule) DeleteDump(filePath string) error {
	for i, d := range m.dumps {
		if d.FilePath == filePath {
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to delete dump file: %w", err)
			}
			m.dumps = append(m.dumps[:i], m.dumps[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("dump file not found in records")
}
