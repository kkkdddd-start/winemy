package m3_network

import (
	"context"
	"net"
	"syscall"

	gopsutil_net "github.com/shirou/gopsutil/v4/net"
	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type NetworkModule struct {
	connections []model.NetworkConnDTO
}

func New() *NetworkModule {
	return &NetworkModule{}
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

	m.connections = make([]model.NetworkConnDTO, 0, len(netCons))
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

		proto := "tcp"
		if c.Family == syscall.AF_INET6 {
			proto = "tcp6"
		}

		dto := model.NetworkConnDTO{
			PID:        uint32(c.Pid),
			Protocol:   proto,
			LocalAddr:  laddr,
			LocalPort:  lport,
			RemoteAddr: raddr,
			RemotePort: rport,
			State:      parseState(c.Status),
			RiskLevel:  assessRiskLevel(raddr),
		}
		m.connections = append(m.connections, dto)
	}
	return nil
}

func (m *NetworkModule) Stop() error {
	return nil
}

func (m *NetworkModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.connections))
	for _, c := range m.connections {
		result = append(result, map[string]interface{}{
			"pid":         c.PID,
			"protocol":    c.Protocol,
			"local_addr":  c.LocalAddr,
			"local_port":  c.LocalPort,
			"remote_addr": c.RemoteAddr,
			"remote_port": c.RemotePort,
			"state":       c.State,
			"risk_level":  c.RiskLevel,
		})
	}
	return result, nil
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
