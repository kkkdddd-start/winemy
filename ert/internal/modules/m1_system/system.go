package m1_system

import (
	"context"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type SystemModule struct {
	info interface{}
}

func New() *SystemModule {
	return &SystemModule{}
}

func (m *SystemModule) ID() int       { return 1 }
func (m *SystemModule) Name() string  { return "system" }
func (m *SystemModule) Priority() int { return 0 }

func (m *SystemModule) Init(ctx context.Context, s registry.Storage) error {
	return nil
}

func (m *SystemModule) Collect(ctx context.Context) error {
	info, err := m.collectSystemInfo()
	if err != nil {
		return err
	}
	m.info = info
	return nil
}

func (m *SystemModule) Stop() error {
	return nil
}

func (m *SystemModule) GetInfo() (*model.SystemInfo, error) {
	if m.info == nil {
		return nil, ErrNotCollected
	}
	info, ok := m.info.(*model.SystemInfo)
	if !ok {
		return nil, ErrNotCollected
	}
	return info, nil
}

func (m *SystemModule) collectSystemInfo() (*model.SystemInfo, error) {
	info := &model.SystemInfo{}

	hostInfo, err := host.Info()
	if err == nil {
		info.Hostname = hostInfo.Hostname
		info.OSName = hostInfo.OS
		info.OSVersion = hostInfo.PlatformVersion
		info.Architecture = hostInfo.KernelArch
		info.BootTime = time.Unix(int64(hostInfo.BootTime), 0)
	}

	userInfo, err := host.Users()
	if err == nil && len(userInfo) > 0 {
		for _, u := range userInfo {
			if u.User != "" {
				info.CurrentUser = u.User
				break
			}
		}
	}

	memInfo, err := mem.VirtualMemory()
	if err == nil {
		info.MemoryTotal = memInfo.Total
	}

	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		info.CPUCount = len(cpuInfo)
	}

	diskInfo, err := disk.Partitions(false)
	if err == nil {
		var totalDisk uint64
		for _, d := range diskInfo {
			usage, err := disk.Usage(d.Mountpoint)
			if err == nil {
				totalDisk += usage.Total
			}
		}
		info.DiskTotal = totalDisk
	}

	ifaces, err := net.Interfaces()
	if err == nil {
		info.IsDomain = false
		for _, iface := range ifaces {
			for _, addr := range iface.Addrs {
				addrStr := addr.Addr
				if isPrivateIP(addrStr) {
					info.IsDomain = false
					break
				}
			}
		}
	}

	return info, nil
}

func isPrivateIP(ip string) bool {
	return strings.HasPrefix(ip, "192.168.") ||
		strings.HasPrefix(ip, "10.") ||
		strings.HasPrefix(ip, "172.")
}

var ErrNotCollected = &CollectError{Message: "system info not collected yet"}

type CollectError struct {
	Message string
}

func (e *CollectError) Error() string {
	return e.Message
}
