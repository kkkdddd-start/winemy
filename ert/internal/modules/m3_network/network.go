//go:build windows

package m3_network

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	gopsutil_net "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type NetworkModule struct {
	connections  []model.NetworkConnDTO
	listening    []model.NetworkConnDTO
	stats        NetworkStats
	processNames map[uint32]string
}

type NetworkStats struct {
	TCPConnections   int            `json:"tcp_connections"`
	UDPConnections   int            `json:"udp_connections"`
	EstablishedCount int            `json:"established_count"`
	ListeningCount   int            `json:"listening_count"`
	StateCounts      map[string]int `json:"state_counts"`
}

func New() *NetworkModule {
	return &NetworkModule{
		processNames: make(map[uint32]string),
	}
}

func (m *NetworkModule) ID() int       { return 3 }
func (m *NetworkModule) Name() string  { return "network" }
func (m *NetworkModule) Priority() int { return 0 }

func (m *NetworkModule) Init(ctx context.Context, s registry.Storage) error {
	return nil
}

func (m *NetworkModule) Collect(ctx context.Context) error {
	netCons, err := gopsutil_net.Connections("")
	if err != nil {
		return err
	}

	m.buildProcessNameMap()

	m.connections = make([]model.NetworkConnDTO, 0, len(netCons))
	m.listening = make([]model.NetworkConnDTO, 0)
	m.stats = NetworkStats{
		StateCounts:      make(map[string]int),
		TCPConnections:   0,
		UDPConnections:   0,
		EstablishedCount: 0,
		ListeningCount:   0,
	}

	for _, c := range netCons {
		laddr := ""
		lport := uint16(0)
		if c.Laddr.IP != "" {
			laddr = c.Laddr.IP
			lport = uint16(c.Laddr.Port)
		}

		raddr := ""
		rport := uint16(0)
		if c.Raddr.IP != "" {
			raddr = c.Raddr.IP
			rport = uint16(c.Raddr.Port)
		}

		proto := "TCP"
		if c.Family == syscall.AF_INET6 {
			proto = "TCP6"
		}

		procName := m.processNames[uint32(c.Pid)]

		dto := model.NetworkConnDTO{
			PID:         uint32(c.Pid),
			ProcessName: procName,
			Protocol:    proto,
			LocalAddr:   laddr,
			LocalPort:   lport,
			RemoteAddr:  raddr,
			RemotePort:  rport,
			State:       parseState(c.Status),
			RiskLevel:   m.assessRiskLevel(raddr, rport),
		}
		m.connections = append(m.connections, dto)

		if c.Status == "LISTEN" {
			m.listening = append(m.listening, dto)
			m.stats.ListeningCount++
		}

		if strings.HasPrefix(proto, "TCP") {
			m.stats.TCPConnections++
		} else {
			m.stats.UDPConnections++
		}

		if c.Status == "ESTABLISHED" {
			m.stats.EstablishedCount++
		}

		m.stats.StateCounts[c.Status]++
	}
	return nil
}

func (m *NetworkModule) buildProcessNameMap() {
	procs, err := process.Processes()
	if err != nil {
		return
	}
	for _, p := range procs {
		name, _ := p.Name()
		m.processNames[uint32(p.Pid)] = name
	}
}

func (m *NetworkModule) GetListeningPorts() []model.NetworkConnDTO {
	return m.listening
}

func (m *NetworkModule) GetStats() NetworkStats {
	return m.stats
}

func (m *NetworkModule) GetIPLocation(ip string) (country, city string, err error) {
	if ip == "" || net.ParseIP(ip) == nil {
		return "", "", nil
	}

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$result = Invoke-RestMethod -Uri "http://ip-api.com/json/%s" -TimeoutSec 5 -ErrorAction SilentlyContinue; if($result) { Write-Output "$($result.country)|$($result.city)" } else { Write-Output "Unknown|Unknown" }`, ip))
	output, err := cmd.Output()
	if err != nil {
		return "Unknown", "Unknown", nil
	}

	parts := strings.Split(strings.TrimSpace(string(output)), "|")
	if len(parts) >= 2 {
		return parts[0], parts[1], nil
	}
	return "Unknown", "Unknown", nil
}

func (m *NetworkModule) parseState(status string) string {
	switch status {
	case "ESTABLISHED":
		return "ESTABLISHED"
	case "LISTEN":
		return "LISTENING"
	case "TIME_WAIT":
		return "TIME_WAIT"
	case "CLOSE_WAIT":
		return "CLOSE_WAIT"
	case "SYN_SENT":
		return "SYN_SENT"
	case "SYN_RECV":
		return "SYN_RECV"
	case "FIN_WAIT1":
		return "FIN_WAIT1"
	case "FIN_WAIT2":
		return "FIN_WAIT2"
	case "CLOSING":
		return "CLOSING"
	case "LAST_ACK":
		return "LAST_ACK"
	case "DELETE_TCB":
		return "DELETE_TCB"
	default:
		return status
	}
}

func (m *NetworkModule) assessRiskLevel(remoteAddr string, remotePort uint16) model.RiskLevel {
	suspiciousPorts := map[uint16]bool{
		4444: true, 5555: true, 6666: true, 6667: true,
		31337: true, 12345: true, 54321: true,
		11211: true, 27017: true,
	}

	if suspiciousPorts[remotePort] {
		return model.RiskHigh
	}

	if remoteAddr == "" {
		return model.RiskLow
	}

	ip := net.ParseIP(remoteAddr)
	if ip == nil {
		return model.RiskMedium
	}

	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() {
		return model.RiskLow
	}

	if m.isPrivateIP(remoteAddr) {
		return model.RiskLow
	}

	if m.isForeignIP(remoteAddr) {
		return model.RiskHigh
	}

	return model.RiskMedium
}

func (m *NetworkModule) isPrivateIP(ipStr string) bool {
	parts := strings.Split(ipStr, ".")
	if len(parts) != 4 {
		return false
	}
	first, _ := strconv.Atoi(parts[0])
	second, _ := strconv.Atoi(parts[1])

	if first == 10 {
		return true
	}
	if first == 172 && second >= 16 && second <= 31 {
		return true
	}
	if first == 192 && second == 168 {
		return true
	}
	return false
}

func (m *NetworkModule) isForeignIP(ipStr string) bool {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$result = Invoke-RestMethod -Uri "http://ip-api.com/json/%s" -TimeoutSec 3 -ErrorAction SilentlyContinue; if($result -and $result.countryCode -ne "CN") { Write-Output "true" } else { Write-Output "false" }`, ipStr))
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(output)), "true")
}

func (m *NetworkModule) Search(keyword string) []model.NetworkConnDTO {
	results := []model.NetworkConnDTO{}
	keywordLower := strings.ToLower(keyword)
	for _, c := range m.connections {
		if strings.Contains(strings.ToLower(c.ProcessName), keywordLower) ||
			strings.Contains(c.RemoteAddr, keyword) ||
			strings.Contains(strconv.FormatUint(uint64(c.RemotePort), 10), keyword) ||
			strings.Contains(strconv.FormatUint(uint64(c.LocalPort), 10), keyword) {
			results = append(results, c)
		}
	}
	return results
}

func parseState(status string) string {
	switch status {
	case "ESTABLISHED":
		return "ESTABLISHED"
	case "LISTEN":
		return "LISTENING"
	case "TIME_WAIT":
		return "TIME_WAIT"
	case "CLOSE_WAIT":
		return "CLOSE_WAIT"
	default:
		return status
	}
}

func assessRiskLevel(remoteAddr string) model.RiskLevel {
	if remoteAddr == "" {
		return model.RiskLow
	}
	ip := net.ParseIP(remoteAddr)
	if ip == nil {
		return model.RiskMedium
	}
	if ip.IsLoopback() || ip.IsPrivate() {
		return model.RiskLow
	}
	return model.RiskMedium
}

func (m *NetworkModule) Stop() error {
	return nil
}

func (m *NetworkModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.connections))
	for _, c := range m.connections {
		country, city, _ := m.GetIPLocation(c.RemoteAddr)
		result = append(result, map[string]interface{}{
			"pid":          c.PID,
			"process_name": c.ProcessName,
			"protocol":     c.Protocol,
			"local_addr":   c.LocalAddr,
			"local_port":   c.LocalPort,
			"remote_addr":  c.RemoteAddr,
			"remote_port":  c.RemotePort,
			"country":      country,
			"city":         city,
			"state":        c.State,
			"risk_level":   c.RiskLevel,
		})
	}
	return result, nil
}

func (m *NetworkModule) DetectPortForwarding() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	cmd := exec.Command("netsh", "interface", "portproxy", "show", "all")
	output, err := cmd.Output()
	if err != nil {
		cmd2 := exec.Command("powershell", "-Command",
			`$ErrorActionPreference='SilentlyContinue'
Get-NetPortMapping | ForEach-Object {
    Write-Output ($_.LocalAddress + '|' + $_.LocalPort + '|' + $_.RemoteAddress + '|' + $_.RemotePort)
}`)

		output2, err2 := cmd2.Output()
		if err2 != nil {
			return results, nil
		}

		lines := strings.Split(string(output2), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.Split(line, "|")
			if len(parts) >= 4 {
				results = append(results, map[string]interface{}{
					"type":           "port_forwarding",
					"local_address":  parts[0],
					"local_port":     parts[1],
					"remote_address": parts[2],
					"remote_port":    parts[3],
				})
			}
		}
		return results, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Listen") || strings.Contains(line, "Connect") {
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				results = append(results, map[string]interface{}{
					"type":            "port_forwarding",
					"listen_address":  fields[1],
					"listen_port":     fields[2],
					"connect_address": fields[4],
					"connect_port":    fields[5],
				})
			}
		}
	}

	return results, nil
}
