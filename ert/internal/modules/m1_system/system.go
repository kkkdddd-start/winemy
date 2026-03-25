//go:build windows

package m1_system

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
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
	info    interface{}
	metrics *RealtimeMetrics
}

type RealtimeMetrics struct {
	CPUUsage      float64         `json:"cpu_usage"`
	MemoryUsage   MemoryUsageInfo `json:"memory_usage"`
	DiskUsage     []DiskUsageInfo `json:"disk_usage"`
	NetworkStatus []NetworkStatus `json:"network_status"`
}

type MemoryUsageInfo struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

type DiskUsageInfo struct {
	MountPoint  string  `json:"mount_point"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

type NetworkStatus struct {
	Interface   string `json:"interface"`
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
	ErrIn       uint64 `json:"err_in"`
	ErrOut      uint64 `json:"err_out"`
	DropIn      uint64 `json:"drop_in"`
	DropOut     uint64 `json:"drop_out"`
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

	if err := m.collectRealtimeMetrics(); err != nil {
		m.metrics = &RealtimeMetrics{}
	}

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

func (m *SystemModule) GetData() ([]map[string]interface{}, error) {
	info, err := m.GetInfo()
	if err != nil {
		return nil, err
	}
	result := []map[string]interface{}{
		{
			"hostname":     info.Hostname,
			"os_name":      info.OSName,
			"os_version":   info.OSVersion,
			"architecture": info.Architecture,
			"boot_time":    info.BootTime.Format(time.RFC3339),
			"current_user": info.CurrentUser,
			"cpu_count":    info.CPUCount,
			"memory_total": info.MemoryTotal,
			"disk_total":   info.DiskTotal,
			"is_domain":    info.IsDomain,
			"domain_name":  info.DomainName,
		},
	}

	if m.metrics != nil {
		result = append(result, map[string]interface{}{
			"type":           "realtime_metrics",
			"cpu_usage":      m.metrics.CPUUsage,
			"memory_usage":   m.metrics.MemoryUsage,
			"disk_usage":     m.metrics.DiskUsage,
			"network_status": m.metrics.NetworkStatus,
		})
	}

	return result, nil
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
		info.Uptime = hostInfo.Uptime
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
		for _, iface := range ifaces {
			for _, addr := range iface.Addrs {
				addrStr := addr.Addr
				if isPrivateIP(addrStr) {
					break
				}
			}
		}
	}

	m.checkDomainMembership(info)

	return info, nil
}

func (m *SystemModule) checkDomainMembership(info *model.SystemInfo) {
	cmd := exec.Command("powershell", "-Command",
		`(Get-CimInstance -ClassName Win32_ComputerSystem).PartOfDomain`)
	output, err := cmd.Output()
	if err == nil {
		result := strings.TrimSpace(string(output))
		info.IsDomain = strings.ToLower(result) == "true"
	}

	if info.IsDomain {
		cmd = exec.Command("powershell", "-Command",
			`(Get-CimInstance -ClassName Win32_ComputerSystem).Domain`)
		output, err := cmd.Output()
		if err == nil {
			info.DomainName = strings.TrimSpace(string(output))
		}
	}
}

func isPrivateIP(ip string) bool {
	return strings.HasPrefix(ip, "192.168.") ||
		strings.HasPrefix(ip, "10.") ||
		strings.HasPrefix(ip, "172.")
}

func (m *SystemModule) collectRealtimeMetrics() error {
	m.metrics = &RealtimeMetrics{}

	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		m.metrics.CPUUsage = cpuPercent[0]
	}

	if runtime.GOOS == "windows" {
		cmd := exec.Command("powershell", "-Command",
			`$cpu = Get-WmiObject Win32_Processor | Measure-Object -Property LoadPercentage -Average | Select-Object -ExpandProperty Average; if($cpu -eq $null) { $cpu = 0 }; Write-Output $cpu`)
		output, err := cmd.Output()
		if err == nil {
			var cpuVal float64
			if _, err := fmt.Sscanf(strings.TrimSpace(string(output)), "%f", &cpuVal); err == nil {
				m.metrics.CPUUsage = cpuVal
			}
		}
	}

	memInfo, err := mem.VirtualMemory()
	if err == nil {
		m.metrics.MemoryUsage = MemoryUsageInfo{
			Total:       memInfo.Total,
			Available:   memInfo.Available,
			Used:        memInfo.Used,
			UsedPercent: memInfo.UsedPercent,
		}
	}

	diskInfo, err := disk.Partitions(false)
	if err == nil {
		for _, d := range diskInfo {
			if d.Fstype == "" || d.Fstype == "devfs" {
				continue
			}
			usage, err := disk.Usage(d.Mountpoint)
			if err == nil {
				m.metrics.DiskUsage = append(m.metrics.DiskUsage, DiskUsageInfo{
					MountPoint:  d.Mountpoint,
					Total:       usage.Total,
					Used:        usage.Used,
					Free:        usage.Free,
					UsedPercent: usage.UsedPercent,
				})
			}
		}
	}

	ioCounters, err := net.IOCounters(true)
	if err == nil {
		for _, iface := range ioCounters {
			m.metrics.NetworkStatus = append(m.metrics.NetworkStatus, NetworkStatus{
				Interface:   iface.Name,
				BytesSent:   iface.BytesSent,
				BytesRecv:   iface.BytesRecv,
				PacketsSent: iface.PacketsSent,
				PacketsRecv: iface.PacketsRecv,
				ErrIn:       iface.Errin,
				ErrOut:      iface.Errout,
				DropIn:      iface.Dropin,
				DropOut:     iface.Dropout,
			})
		}
	}

	return nil
}

func (m *SystemModule) GetRealtimeMetrics() (*RealtimeMetrics, error) {
	if m.metrics == nil {
		return nil, fmt.Errorf("metrics not collected yet")
	}
	return m.metrics, nil
}

var ErrNotCollected = &CollectError{Message: "system info not collected yet"}

type CollectError struct {
	Message string
}

func (e *CollectError) Error() string {
	return e.Message
}
