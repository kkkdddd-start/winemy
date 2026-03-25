package m15_memory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
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

	return dumpFile, fmt.Errorf("memory dump requires external tool (procdump.exe) - dump record created for: %s (%s)", name, exe)
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
