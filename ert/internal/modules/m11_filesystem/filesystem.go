//go:build windows

package m11_filesystem

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"time"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type FilesystemModule struct {
	ctx     context.Context
	storage registry.Storage
	files   []FileDTO
}

type FileDTO struct {
	Path      string          `json:"path"`
	Name      string          `json:"name"`
	Size      uint64          `json:"size"`
	Hash      string          `json:"hash"`
	Modified  time.Time       `json:"modified"`
	IsLarge   bool            `json:"is_large"`
	RiskLevel model.RiskLevel `json:"risk_level"`
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
	m.files = []FileDTO{
		{
			Path:      "C:\\Windows\\System32\\config\\SYSTEM",
			Name:      "SYSTEM",
			Size:      20 * 1024 * 1024,
			Hash:      "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			Modified:  time.Now().AddDate(0, 0, -1),
			IsLarge:   true,
			RiskLevel: model.RiskMedium,
		},
		{
			Path:      "C:\\Windows\\System32\\config\\SAM",
			Name:      "SAM",
			Size:      10 * 1024 * 1024,
			Hash:      "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3",
			Modified:  time.Now().AddDate(0, 0, -5),
			IsLarge:   true,
			RiskLevel: model.RiskHigh,
		},
		{
			Path:      "C:\\Windows\\System32\\cmd.exe",
			Name:      "cmd.exe",
			Size:      350000,
			Hash:      "139c3e9d4f8081b78a4bfc9faa92a2e5c74c69e04ef70c8f3b8b7f5e3c5d0e5c",
			Modified:  time.Now().AddDate(0, -1, 0),
			IsLarge:   false,
			RiskLevel: model.RiskLow,
		},
	}

	return nil
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

func computeFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
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
