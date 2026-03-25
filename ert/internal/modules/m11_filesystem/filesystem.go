//go:build windows

package m11_filesystem

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type FilesystemModule struct {
	ctx     context.Context
	storage registry.Storage
	files   []model.FileDTO
}

func New() *FilesystemModule {
	return &FilesystemModule{}
}

func (m *FilesystemModule) ID() int       { return 11 }
func (m *FilesystemModule) Name() string  { return "filesystem" }
func (m *FilesystemModule) Priority() int { return 2 }

func (m *FilesystemModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	return nil
}

func (m *FilesystemModule) Collect(ctx context.Context) error {
	m.files = []model.FileDTO{}

	scanPaths := []string{
		`C:\Windows\System32`,
		`C:\Program Files`,
		`C:\Program Files (x86)`,
		`C:\Users`,
	}

	for _, path := range scanPaths {
		files, err := m.scanDirectory(path, 3)
		if err == nil {
			m.files = append(m.files, files...)
		}
	}

	return nil
}

func (m *FilesystemModule) scanDirectory(root string, maxDepth int) ([]model.FileDTO, error) {
	var results []model.FileDTO
	return results, filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return filepath.SkipDir
		}

		if info.IsDir() {
			if strings.Contains(strings.ToLower(path), "temp") ||
				strings.Contains(strings.ToLower(path), "cache") {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))

		file := model.FileDTO{
			Path:      path,
			Name:      info.Name(),
			Size:      uint64(info.Size()),
			Modified:  info.ModTime(),
			RiskLevel: model.RiskLow,
		}

		if ext == ".exe" || ext == ".dll" || ext == ".sys" {
			hash, err := m.computeFileHash(path)
			if err == nil {
				file.Hash = hash
				file.RiskLevel = assessFileRisk(path, hash)
			}
		}

		if info.Size() > 100*1024*1024*1024 {
			file.IsLarge = true
		}

		results = append(results, file)
		return nil
	})
}

func (m *FilesystemModule) computeFileHash(path string) (string, error) {
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

func assessFileRisk(path, hash string) model.RiskLevel {
	nameLower := strings.ToLower(path)

	suspiciousExts := []string{".tmp", ".vbs", ".js", ".jse", ".vbe", ".ws", ".wsh", ".scr", ".pif", ".msi", ".msp"}

	ext := strings.ToLower(filepath.Ext(nameLower))
	for _, sus := range suspiciousExts {
		if ext == sus {
			return model.RiskMedium
		}
	}

	if strings.Contains(nameLower, "temp") || strings.Contains(nameLower, "tmp") {
		return model.RiskMedium
	}

	if strings.Contains(nameLower, "downloads") {
		return model.RiskMedium
	}

	systemDirs := []string{`\windows\system32`, `\windows\syswow64`}
	for _, sys := range systemDirs {
		if strings.Contains(nameLower, sys) {
			return model.RiskLow
		}
	}

	return model.RiskLow
}

func (m *FilesystemModule) ScanPath(path string, recursive bool) ([]model.FileDTO, error) {
	if recursive {
		return m.scanDirectory(path, -1)
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var results []model.FileDTO
	for _, f := range files {
		info, _ := f.Info()
		results = append(results, model.FileDTO{
			Path:      filepath.Join(path, f.Name()),
			Name:      f.Name(),
			Size:      uint64(info.Size()),
			Modified:  info.ModTime(),
			RiskLevel: model.RiskLow,
		})
	}
	return results, nil
}

func (m *FilesystemModule) Search(keyword string) ([]model.FileDTO, error) {
	results := []model.FileDTO{}
	keywordLower := strings.ToLower(keyword)
	for _, f := range m.files {
		if strings.Contains(strings.ToLower(f.Name), keywordLower) ||
			strings.Contains(strings.ToLower(f.Path), keywordLower) {
			results = append(results, f)
		}
	}
	return results, nil
}

func (m *FilesystemModule) GetFileHash(path string) (string, error) {
	return m.computeFileHash(path)
}

func (m *FilesystemModule) Stop() error {
	return nil
}

func (m *FilesystemModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.files))
	for _, f := range m.files {
		result = append(result, map[string]interface{}{
			"path":       f.Path,
			"name":       f.Name,
			"size":       f.Size,
			"hash":       f.Hash,
			"modified":   f.Modified.Format(time.RFC3339),
			"is_large":   f.IsLarge,
			"risk_level": f.RiskLevel,
		})
	}
	return result, nil
}
