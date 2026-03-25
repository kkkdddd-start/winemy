package m16_threat

import (
	"context"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type ThreatModule struct {
	ctx           context.Context
	storage       registry.Storage
	malProcesses  []model.ProcessDTO
	suspiciousNet []model.NetworkConnDTO
	sensitiveActs []map[string]interface{}
	threatRules   []model.AlertRule
}

func New() *ThreatModule {
	return &ThreatModule{
		malProcesses:  []model.ProcessDTO{},
		suspiciousNet: []model.NetworkConnDTO{},
		sensitiveActs: []map[string]interface{}{},
		threatRules:   []model.AlertRule{},
	}
}

func (m *ThreatModule) ID() int       { return 16 }
func (m *ThreatModule) Name() string  { return "threat" }
func (m *ThreatModule) Priority() int { return 0 }

func (m *ThreatModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	m.loadThreatRules()
	return nil
}

func (m *ThreatModule) loadThreatRules() {
	m.threatRules = []model.AlertRule{
		{ID: "rule1", Name: "Suspicious Process", ModuleID: 16, Severity: model.RiskHigh, Enabled: true},
		{ID: "rule2", Name: "Known Malware Hash", ModuleID: 16, Severity: model.RiskCritical, Enabled: true},
		{ID: "rule3", Name: "Suspicious Network", ModuleID: 16, Severity: model.RiskMedium, Enabled: true},
		{ID: "rule4", Name: "Sensitive Behavior", ModuleID: 16, Severity: model.RiskMedium, Enabled: true},
	}
}

func (m *ThreatModule) Collect(ctx context.Context) error {
	if err := m.detectMaliciousProcesses(); err != nil {
		return err
	}
	if err := m.detectSuspiciousNetwork(); err != nil {
		return err
	}
	m.detectSensitiveBehaviors()
	return nil
}

func (m *ThreatModule) detectMaliciousProcesses() error {
	procs, err := process.Processes()
	if err != nil {
		return err
	}

	m.malProcesses = []model.ProcessDTO{}
	suspiciousNames := []string{
		"mimikatz", "pwdump", "procdump", "lsass", "cmd",
		"powershell", "nc", "netcat", "psexec", "wce",
		"gsecdump", "fgdump", "hashdump", "metasploit",
	}

	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}

		nameLower := strings.ToLower(name)
		for _, sus := range suspiciousNames {
			if strings.Contains(nameLower, sus) {
				cpu, _ := p.CPUPercent()
				mem, _ := p.MemoryInfo()
				createTime, _ := p.CreateTime()
				ppid, _ := p.Ppid()

				dto := model.ProcessDTO{
					PID:         uint32(p.Pid),
					Name:        name,
					Path:        "",
					CommandLine: "",
					User:        "",
					CPU:         cpu,
					Memory:      mem.RSS,
					StartTime:   time.Unix(int64(createTime)/1000, 0),
					RiskLevel:   model.RiskHigh,
				}
				_ = ppid
				m.malProcesses = append(m.malProcesses, dto)
				break
			}
		}
	}
	return nil
}

func (m *ThreatModule) detectSuspiciousNetwork() error {
	conns, err := net.Connections("tcp")
	if err != nil {
		return err
	}

	m.suspiciousNet = []model.NetworkConnDTO{}
	suspiciousPorts := map[uint32]bool{
		4444:  true,
		5555:  true,
		6666:  true,
		7777:  true,
		31337: true,
	}

	for _, c := range conns {
		if suspiciousPorts[c.Raddr.Port] {
			proto := "tcp"
			if c.Type == 2 {
				proto = "udp"
			}
			dto := model.NetworkConnDTO{
				PID:        uint32(c.Pid),
				Protocol:   proto,
				LocalAddr:  c.Laddr.IP,
				LocalPort:  uint16(c.Laddr.Port),
				RemoteAddr: c.Raddr.IP,
				RemotePort: uint16(c.Raddr.Port),
				State:      c.Status,
				RiskLevel:  model.RiskHigh,
			}
			m.suspiciousNet = append(m.suspiciousNet, dto)
		}

		if c.Status == "listening" && c.Raddr.Port == 0 {
			continue
		}

		remoteStr := c.Raddr.IP
		if remoteStr == "127.0.0.1" || remoteStr == "0.0.0.0" {
			continue
		}

		suspiciousIPs := []string{"192.0.2.", "198.51.100.", "203.0.113."}
		for _, sip := range suspiciousIPs {
			if strings.HasPrefix(remoteStr, sip) {
				proto := "tcp"
				if c.Type == 2 {
					proto = "udp"
				}
				dto := model.NetworkConnDTO{
					PID:        uint32(c.Pid),
					Protocol:   proto,
					LocalAddr:  c.Laddr.IP,
					LocalPort:  uint16(c.Laddr.Port),
					RemoteAddr: c.Raddr.IP,
					RemotePort: uint16(c.Raddr.Port),
					State:      c.Status,
					RiskLevel:  model.RiskMedium,
				}
				m.suspiciousNet = append(m.suspiciousNet, dto)
				break
			}
		}
	}
	return nil
}

func (m *ThreatModule) detectSensitiveBehaviors() {
	m.sensitiveActs = []map[string]interface{}{}

	sensitivePatterns := []struct {
		name        string
		pattern     string
		description string
		severity    model.RiskLevel
	}{
		{"PowerShell Encoded Command", "enc", "PowerShell encoded command detected", model.RiskHigh},
		{"Registry Modification", "reg add", "Registry modification detected", model.RiskMedium},
		{"Service Creation", "sc create", "Windows service creation detected", model.RiskMedium},
		{"WMI Execution", "wmic ", "WMI command execution detected", model.RiskMedium},
		{"Scheduled Task", "schtasks", "Scheduled task creation detected", model.RiskMedium},
	}

	procs, _ := process.Processes()
	for _, p := range procs {
		cmdline, err := p.CmdlineSlice()
		if err != nil || len(cmdline) == 0 {
			continue
		}
		fullCmd := strings.ToLower(strings.Join(cmdline, " "))

		for _, pattern := range sensitivePatterns {
			if strings.Contains(fullCmd, pattern.pattern) {
				name, _ := p.Name()
				m.sensitiveActs = append(m.sensitiveActs, map[string]interface{}{
					"pid":         p.Pid,
					"name":        name,
					"pattern":     pattern.name,
					"description": pattern.description,
					"command":     strings.Join(cmdline, " "),
					"risk_level":  pattern.severity,
					"timestamp":   time.Now().Format(time.RFC3339),
				})
			}
		}
	}
}

func (m *ThreatModule) Stop() error {
	return nil
}

func (m *ThreatModule) GetData() ([]map[string]interface{}, error) {
	result := []map[string]interface{}{}

	for _, p := range m.malProcesses {
		result = append(result, map[string]interface{}{
			"type":       "malicious_process",
			"pid":        p.PID,
			"name":       p.Name,
			"path":       p.Path,
			"cpu":        p.CPU,
			"memory":     p.Memory,
			"start_time": p.StartTime.Format(time.RFC3339),
			"risk_level": p.RiskLevel,
		})
	}

	for _, n := range m.suspiciousNet {
		result = append(result, map[string]interface{}{
			"type":        "suspicious_network",
			"pid":         n.PID,
			"protocol":    n.Protocol,
			"local_addr":  n.LocalAddr,
			"local_port":  n.LocalPort,
			"remote_addr": n.RemoteAddr,
			"remote_port": n.RemotePort,
			"state":       n.State,
			"risk_level":  n.RiskLevel,
		})
	}

	for _, a := range m.sensitiveActs {
		result = append(result, map[string]interface{}{
			"type":        "sensitive_behavior",
			"pid":         a["pid"],
			"name":        a["name"],
			"pattern":     a["pattern"],
			"description": a["description"],
			"command":     a["command"],
			"risk_level":  a["risk_level"],
			"timestamp":   a["timestamp"],
		})
	}

	return result, nil
}

func (m *ThreatModule) GetThreatRules() []model.AlertRule {
	return m.threatRules
}

func (m *ThreatModule) GetMaliciousProcesses() []model.ProcessDTO {
	return m.malProcesses
}

func (m *ThreatModule) GetSuspiciousNetwork() []model.NetworkConnDTO {
	return m.suspiciousNet
}
