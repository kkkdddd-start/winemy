//go:build windows

package memory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/yourname/ert/internal/model"
)

type MemoryDumper struct {
	outputDir string
	maxSize   uint64
}

type DumpType string

const (
	DumpProcess DumpType = "process"
	DumpFull    DumpType = "full"
	DumpKernel  DumpType = "kernel"
)

func New(outputDir string) *MemoryDumper {
	return &MemoryDumper{
		outputDir: outputDir,
		maxSize:   10 * 1024 * 1024 * 1024,
	}
}

func (m *MemoryDumper) DumpProcess(ctx context.Context, pid uint32) (*model.MemoryDumpDTO, error) {
	return m.dumpProcessByPID(ctx, pid)
}

func (m *MemoryDumper) DumpFull(ctx context.Context) (*model.MemoryDumpDTO, error) {
	return nil, fmt.Errorf("full system memory dump requires administrator privileges")
}

func (m *MemoryDumper) dumpProcessByPID(ctx context.Context, pid uint32) (*model.MemoryDumpDTO, error) {
	filename := fmt.Sprintf("process_%d_%s.dmp", pid, time.Now().Format("20060102150405"))
	filepath := filepath.Join(m.outputDir, filename)

	dto := &model.MemoryDumpDTO{
		PID:         pid,
		ProcessName: fmt.Sprintf("process_%d", pid),
		DumpType:    string(DumpProcess),
		FilePath:    filepath,
		CreatedAt:   time.Now(),
	}

	hash, err := m.calculateFileHash(filepath)
	if err == nil {
		dto.SHA256 = hash
	}

	return dto, nil
}

func (m *MemoryDumper) calculateFileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func (m *MemoryDumper) streamHash(ctx context.Context, path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	buf := make([]byte, 32*1024)

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			n, err := file.Read(buf)
			if n > 0 {
				hash.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func HybridHash(ctx context.Context, path string, chunkSize int64, maxRead int64) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return "", err
	}

	if fi.Size() <= maxRead {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		hash := sha256.Sum256(data)
		return hex.EncodeToString(hash[:]), nil
	}

	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%d%s", fi.Size(), fi.Name())))

	file.Seek(0, 0)
	if n, err := io.CopyN(hash, file, chunkSize); err == nil {
		_ = n
	}

	file.Seek(-chunkSize, 2)
	if n, err := io.CopyN(hash, file, chunkSize); err == nil {
		_ = n
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
