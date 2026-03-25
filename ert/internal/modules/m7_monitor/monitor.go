//go:build windows

package m7_monitor

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type MonitorModule struct {
	ctx      context.Context
	storage  registry.Storage
	mu       sync.RWMutex
	running  bool
	cpu      float64
	mem      uint64
	memTotal uint64
	disk     uint64
	netIn    uint64
	netOut   uint64
	alerts   []model.AlertEvent

	cpuHistory         []float64
	memHistory         []float64
	diskHistory        []float64
	netInHistory       []uint64
	netOutHistory      []uint64
	partitionStats     []map[string]interface{}
	diskIOStats        []map[string]interface{}
	lastCollection     time.Time
	collectionInterval time.Duration
	maxHistoryPoints   int
}

func New() *MonitorModule {
	return &MonitorModule{
		alerts:             make([]model.AlertEvent, 0),
		cpuHistory:         make([]float64, 0),
		memHistory:         make([]float64, 0),
		diskHistory:        make([]float64, 0),
		netInHistory:       make([]uint64, 0),
		netOutHistory:      make([]uint64, 0),
		collectionInterval: 2 * time.Second,
		maxHistoryPoints:   60,
	}
}

func (m *MonitorModule) ID() int       { return 7 }
func (m *MonitorModule) Name() string  { return "monitor" }
func (m *MonitorModule) Priority() int { return 0 }

func (m *MonitorModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	m.running = false
	return nil
}

func (m *MonitorModule) Collect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.running = true

	var cpuPercent float64
	if percent, err := cpu.Percent(time.Second, false); err == nil && len(percent) > 0 {
		cpuPercent = percent[0]
		m.cpu = cpuPercent
		m.cpuHistory = append(m.cpuHistory, cpuPercent)
		if len(m.cpuHistory) > m.maxHistoryPoints {
			m.cpuHistory = m.cpuHistory[1:]
		}
		if cpuPercent > 80 {
			m.alerts = append(m.alerts, model.AlertEvent{
				ID:        fmt.Sprintf("cpu_alert_%d", len(m.alerts)+1),
				RuleID:    "cpu_threshold",
				RuleName:  "High CPU Usage",
				Severity:  model.RiskHigh,
				Message:   "CPU usage exceeded 80%",
				Value:     cpuPercent,
				Threshold: 80,
				Timestamp: time.Now(),
				ModuleID:  7,
			})
		}
	}

	if memInfo, err := mem.VirtualMemory(); err == nil {
		m.mem = memInfo.Used
		m.memTotal = memInfo.Total
		memPercent := memInfo.UsedPercent
		m.memHistory = append(m.memHistory, memPercent)
		if len(m.memHistory) > m.maxHistoryPoints {
			m.memHistory = m.memHistory[1:]
		}
		if memPercent > 85 {
			m.alerts = append(m.alerts, model.AlertEvent{
				ID:        fmt.Sprintf("mem_alert_%d", len(m.alerts)+1),
				RuleID:    "mem_threshold",
				RuleName:  "High Memory Usage",
				Severity:  model.RiskHigh,
				Message:   "Memory usage exceeded 85%",
				Value:     memPercent,
				Threshold: 85,
				Timestamp: time.Now(),
				ModuleID:  7,
			})
		}
	}

	var totalUsed uint64
	if parts, err := disk.Partitions(false); err == nil {
		for _, p := range parts {
			if usage, err := disk.Usage(p.Mountpoint); err == nil {
				totalUsed += usage.Used
			}
		}
		m.disk = totalUsed
		diskPercent := float64(totalUsed) / float64(m.memTotal) * 100
		m.diskHistory = append(m.diskHistory, diskPercent)
		if len(m.diskHistory) > m.maxHistoryPoints {
			m.diskHistory = m.diskHistory[1:]
		}
	}

	ioCounters, err := net.IOCounters(true)
	if err == nil && len(ioCounters) > 0 {
		m.netIn = ioCounters[0].BytesRecv
		m.netOut = ioCounters[0].BytesSent
		m.netInHistory = append(m.netInHistory, ioCounters[0].BytesRecv)
		m.netOutHistory = append(m.netOutHistory, ioCounters[0].BytesSent)
		if len(m.netInHistory) > m.maxHistoryPoints {
			m.netInHistory = m.netInHistory[1:]
		}
		if len(m.netOutHistory) > m.maxHistoryPoints {
			m.netOutHistory = m.netOutHistory[1:]
		}
	}

	m.lastCollection = time.Now()
	m.running = false
	return nil
}

func (m *MonitorModule) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = false
	return nil
}

func (m *MonitorModule) GetData() ([]map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uptime := time.Now().Unix()

	result := []map[string]interface{}{
		{
			"metric": "cpu",
			"value":  m.cpu,
			"unit":   "percent",
			"status": getStatus(m.cpu, 80),
		},
		{
			"metric": "memory",
			"value":  m.mem,
			"unit":   "bytes",
			"status": getStatus(float64(m.mem), float64(m.memTotal)*0.85),
		},
		{
			"metric": "disk",
			"value":  m.disk,
			"unit":   "bytes",
			"status": "normal",
		},
		{
			"metric": "network_in",
			"value":  m.netIn,
			"unit":   "bytes",
			"status": "normal",
		},
		{
			"metric": "network_out",
			"value":  m.netOut,
			"unit":   "bytes",
			"status": "normal",
		},
		{
			"metric": "uptime",
			"value":  uptime,
			"unit":   "seconds",
			"status": "normal",
		},
	}

	for _, alert := range m.alerts {
		result = append(result, map[string]interface{}{
			"metric":    "alert",
			"rule_id":   alert.RuleID,
			"rule_name": alert.RuleName,
			"severity":  alert.Severity,
			"message":   alert.Message,
			"value":     alert.Value,
			"timestamp": alert.Timestamp.Format(time.RFC3339),
		})
	}

	return result, nil
}

func getStatus(value float64, threshold float64) string {
	if value > threshold {
		return "warning"
	}
	return "normal"
}

func (m *MonitorModule) ExportJSON(filePath string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"metrics":   m.getMetricsData(),
		"alerts":    m.alerts,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(filePath, jsonData, 0644)
}

func (m *MonitorModule) ExportCSV(filePath string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Metric", "Value", "Unit", "Status", "Timestamp"})

	for _, metric := range m.getMetricsData() {
		writer.Write([]string{
			metric["metric"].(string),
			fmt.Sprintf("%v", metric["value"]),
			metric["unit"].(string),
			metric["status"].(string),
			time.Now().Format(time.RFC3339),
		})
	}

	for _, alert := range m.alerts {
		writer.Write([]string{
			"alert",
			fmt.Sprintf("%v", alert.Value),
			"",
			fmt.Sprintf("%v", alert.Severity),
			alert.Timestamp.Format(time.RFC3339),
		})
	}

	return nil
}

func (m *MonitorModule) getMetricsData() []map[string]interface{} {
	uptime := time.Now().Unix()

	return []map[string]interface{}{
		{
			"metric": "cpu",
			"value":  m.cpu,
			"unit":   "percent",
			"status": getStatus(m.cpu, 80),
		},
		{
			"metric": "memory",
			"value":  m.mem,
			"unit":   "bytes",
			"status": getStatus(float64(m.mem), float64(m.memTotal)*0.85),
		},
		{
			"metric": "disk",
			"value":  m.disk,
			"unit":   "bytes",
			"status": "normal",
		},
		{
			"metric": "network_in",
			"value":  m.netIn,
			"unit":   "bytes",
			"status": "normal",
		},
		{
			"metric": "network_out",
			"value":  m.netOut,
			"unit":   "bytes",
			"status": "normal",
		},
		{
			"metric": "uptime",
			"value":  uptime,
			"unit":   "seconds",
			"status": "normal",
		},
	}
}

func (m *MonitorModule) GetCPUPerCore() ([]map[string]interface{}, error) {
	perCPU, err := cpu.Percent(time.Second, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU per core: %w", err)
	}

	var results []map[string]interface{}
	for i, percent := range perCPU {
		results = append(results, map[string]interface{}{
			"core":      i,
			"usage":     percent,
			"unit":      "percent",
			"status":    getStatus(percent, 80),
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
	return results, nil
}

func (m *MonitorModule) GetMemoryPageRate() (map[string]interface{}, error) {
	cmd := exec.Command("powershell", "-Command",
		`$perfCounters = Get-Counter '\Memory\Pages Input/sec','\Memory\Pages Output/sec' -ErrorAction SilentlyContinue
if($perfCounters) {
    $input = $perfCounters.CounterSamples | Where-Object { $_.Path -like '*Pages Input*' } | Select-Object -ExpandProperty CookedValue
    $output = $perfCounters.CounterSamples | Where-Object { $_.Path -like '*Pages Output*' } | Select-Object -ExpandProperty CookedValue
    Write-Output "$input|$output"
} else {
    Write-Output "0|0"
}`)

	output, err := cmd.Output()
	if err != nil {
		return map[string]interface{}{
			"pages_input_per_sec":  0,
			"pages_output_per_sec": 0,
			"timestamp":            time.Now().Format(time.RFC3339),
		}, nil
	}

	parts := strings.Split(strings.TrimSpace(string(output)), "|")
	pageIn := uint64(0)
	pageOut := uint64(0)
	if len(parts) >= 2 {
		fmt.Sscanf(parts[0], "%d", &pageIn)
		fmt.Sscanf(parts[1], "%d", &pageOut)
	}

	return map[string]interface{}{
		"pages_input_per_sec":  pageIn,
		"pages_output_per_sec": pageOut,
		"timestamp":            time.Now().Format(time.RFC3339),
	}, nil
}

func (m *MonitorModule) GetDiskIOStats() (map[string]interface{}, error) {
	ioCounters, err := disk.IOCounters()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk IO stats: %w", err)
	}

	var results []map[string]interface{}
	for name, io := range ioCounters {
		results = append(results, map[string]interface{}{
			"disk_name":     name,
			"read_count":    io.ReadCount,
			"write_count":   io.WriteCount,
			"read_bytes":    io.ReadBytes,
			"write_bytes":   io.WriteBytes,
			"read_time_ms":  io.ReadTime,
			"write_time_ms": io.WriteTime,
			"timestamp":     time.Now().Format(time.RFC3339),
		})
	}

	stats := map[string]interface{}{
		"disks":      results,
		"disk_count": len(results),
	}
	return stats, nil
}

func (m *MonitorModule) GetPartitionStats() ([]map[string]interface{}, error) {
	parts, err := disk.Partitions(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get partitions: %w", err)
	}

	var results []map[string]interface{}
	for _, p := range parts {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}

		results = append(results, map[string]interface{}{
			"device":       p.Device,
			"mountpoint":   p.Mountpoint,
			"fstype":       p.Fstype,
			"total":        usage.Total,
			"used":         usage.Used,
			"free":         usage.Free,
			"used_percent": usage.UsedPercent,
			"timestamp":    time.Now().Format(time.RFC3339),
		})
	}
	return results, nil
}

func (m *MonitorModule) GetRealtimeHistory() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"cpu_history":     m.cpuHistory,
		"mem_history":     m.memHistory,
		"disk_history":    m.diskHistory,
		"net_in_history":  m.netInHistory,
		"net_out_history": m.netOutHistory,
		"cpu_current":     m.cpu,
		"mem_current":     float64(m.mem),
		"mem_total":       float64(m.memTotal),
		"mem_percent":     len(m.memHistory) > 0,
		"disk_current":    float64(m.disk),
		"net_in_current":  m.netIn,
		"net_out_current": m.netOut,
		"timestamp":       time.Now().Format(time.RFC3339),
	}
}

func (m *MonitorModule) GetNetworkConnectionCount() (map[string]interface{}, error) {
	netIO, err := net.IOCounters(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get network IO counters: %w", err)
	}

	var totalConn uint64
	for _, n := range netIO {
		totalConn += n.Dropin + n.Dropout
	}

	connections, err := net.Connections("tcp")
	tcpCount := len(connections)
	if err != nil {
		tcpCount = 0
	}

	udpConnections, err := net.Connections("udp")
	udpCount := len(udpConnections)
	if err != nil {
		udpCount = 0
	}

	return map[string]interface{}{
		"total_connections": totalConn,
		"tcp_count":         tcpCount,
		"udp_count":         udpCount,
		"timestamp":         time.Now().Format(time.RFC3339),
	}, nil
}

func (m *MonitorModule) SetAlertThreshold(metric string, threshold float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.alerts {
		if m.alerts[i].RuleID == metric+"_threshold" {
			m.alerts[i].Threshold = threshold
			return
		}
	}

	m.alerts = append(m.alerts, model.AlertEvent{
		RuleID:    metric + "_threshold",
		RuleName:  "Custom " + metric + " Threshold",
		Severity:  model.RiskMedium,
		Threshold: threshold,
		Timestamp: time.Now(),
		ModuleID:  7,
	})
}

func (m *MonitorModule) GetAlerts() []model.AlertEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.alerts
}

func (m *MonitorModule) ClearAlerts() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alerts = make([]model.AlertEvent, 0)
}
