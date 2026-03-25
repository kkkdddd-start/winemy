package m10_kernel

import (
	"context"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type KernelModule struct {
	ctx     context.Context
	storage registry.Storage
	drivers []model.DriverDTO
}

func New() *KernelModule {
	return &KernelModule{}
}

func (m *KernelModule) ID() int       { return 10 }
func (m *KernelModule) Name() string  { return "kernel" }
func (m *KernelModule) Priority() int { return 1 }

func (m *KernelModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	return nil
}

func (m *KernelModule) Collect(ctx context.Context) error {
	m.drivers = []model.DriverDTO{
		{
			Name:      "ntoskrnl.exe",
			Path:      "C:\\Windows\\System32\\ntoskrnl.exe",
			BaseAddr:  "0xFFFFF80000000000",
			Size:      15000000,
			IsSigned:  true,
			Signature: "Microsoft Windows",
			RiskLevel: model.RiskLow,
		},
		{
			Name:      "hal.dll",
			Path:      "C:\\Windows\\System32\\hal.dll",
			BaseAddr:  "0xFFFFF80000000000",
			Size:      350000,
			IsSigned:  true,
			Signature: "Microsoft Windows",
			RiskLevel: model.RiskLow,
		},
		{
			Name:      "UnknownDriver.sys",
			Path:      "C:\\Windows\\System32\\drivers\\UnknownDriver.sys",
			BaseAddr:  "0xFFFFF80000000000",
			Size:      50000,
			IsSigned:  false,
			Signature: "",
			RiskLevel: model.RiskCritical,
		},
		{
			Name:      "win32k.sys",
			Path:      "C:\\Windows\\System32\\win32k.sys",
			BaseAddr:  "0xFFFFF96000000000",
			Size:      3000000,
			IsSigned:  true,
			Signature: "Microsoft Windows",
			RiskLevel: model.RiskLow,
		},
	}

	return nil
}

func (m *KernelModule) Stop() error {
	return nil
}

func (m *KernelModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.drivers))
	for _, d := range m.drivers {
		result = append(result, map[string]interface{}{
			"name":       d.Name,
			"path":       d.Path,
			"base_addr":  d.BaseAddr,
			"size":       d.Size,
			"is_signed":  d.IsSigned,
			"signature":  d.Signature,
			"risk_level": d.RiskLevel,
		})
	}
	return result, nil
}
