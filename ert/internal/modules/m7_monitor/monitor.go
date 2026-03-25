//go:build windows

package m7_monitor

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
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
}

func New() *MonitorModule {
	return &MonitorModule{
		alerts: make([]model.AlertEvent, 0),
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

	if percent, err := cpu.Percent(time.Second, false); err == nil && len(percent) > 0 {
		m.cpu = percent[0]
		if m.cpu > 80 {
			m.alerts = append(m.alerts, model.AlertEvent{
				ID:        "cpu_alert_1",
				RuleID:    "cpu_threshold",
				RuleName:  "High CPU Usage",
				Severity:  model.RiskHigh,
				Message:   "CPU usage exceeded 80%",
				Value:     m.cpu,
				Threshold: 80,
				Timestamp: time.Now(),
				ModuleID:  7,
			})
		}
	}

	if memInfo, err := mem.VirtualMemory(); err == nil {
		m.mem = memInfo.Used
		m.memTotal = memInfo.Total
		if memInfo.UsedPercent > 85 {
			m.alerts = append(m.alerts, model.AlertEvent{
				ID:        "mem_alert_1",
				RuleID:    "mem_threshold",
				RuleName:  "High Memory Usage",
				Severity:  model.RiskHigh,
				Message:   "Memory usage exceeded 85%",
				Value:     memInfo.UsedPercent,
				Threshold: 85,
				Timestamp: time.Now(),
				ModuleID:  7,
			})
		}
	}

	if parts, err := disk.Partitions(false); err == nil {
		var totalUsed uint64
		for _, p := range parts {
			if usage, err := disk.Usage(p.Mountpoint); err == nil {
				totalUsed += usage.Used
			}
		}
		m.disk = totalUsed
	}

	ioCounters, err := net.IOCounters(true)
	if err == nil && len(ioCounters) > 0 {
		m.netIn = ioCounters[0].BytesRecv
		m.netOut = ioCounters[0].BytesSent
	}

	host.Info()

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
