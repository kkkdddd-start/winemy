package m23_baseline

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type BaselineModule struct {
	ctx             context.Context
	storage         registry.Storage
	passwordPolicy  map[string]interface{}
	accountPolicy   map[string]interface{}
	auditPolicy     []map[string]interface{}
	networkSecurity []map[string]interface{}
	serviceConfig   []map[string]interface{}
}

func New() *BaselineModule {
	return &BaselineModule{
		passwordPolicy:  map[string]interface{}{},
		accountPolicy:   map[string]interface{}{},
		auditPolicy:     []map[string]interface{}{},
		networkSecurity: []map[string]interface{}{},
		serviceConfig:   []map[string]interface{}{},
	}
}

func (m *BaselineModule) ID() int       { return 23 }
func (m *BaselineModule) Name() string  { return "baseline" }
func (m *BaselineModule) Priority() int { return 1 }

func (m *BaselineModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	return nil
}

func (m *BaselineModule) Collect(ctx context.Context) error {
	if err := m.collectPasswordPolicy(); err != nil {
		return err
	}
	if err := m.collectAccountPolicy(); err != nil {
		return err
	}
	if err := m.collectAuditPolicy(); err != nil {
		return err
	}
	if err := m.collectNetworkSecurity(); err != nil {
		return err
	}
	if err := m.collectServiceConfig(); err != nil {
		return err
	}
	return nil
}

func (m *BaselineModule) collectPasswordPolicy() error {
	m.passwordPolicy = map[string]interface{}{}

	cmd := exec.Command("net", "accounts")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Minimum password length:") {
			value := strings.TrimPrefix(line, "Minimum password length:")
			value = strings.TrimSpace(value)
			m.passwordPolicy["min_length"] = value
		} else if strings.Contains(line, "Maximum password age (days):") {
			value := strings.TrimPrefix(line, "Maximum password age (days):")
			value = strings.TrimSpace(value)
			m.passwordPolicy["max_age"] = value
		} else if strings.Contains(line, "Minimum password age (days):") {
			value := strings.TrimPrefix(line, "Minimum password age (days):")
			value = strings.TrimSpace(value)
			m.passwordPolicy["min_age"] = value
		} else if strings.Contains(line, "Lockout threshold:") {
			value := strings.TrimPrefix(line, "Lockout threshold:")
			value = strings.TrimSpace(value)
			m.passwordPolicy["lockout_threshold"] = value
			if value == "Never" {
				m.passwordPolicy["lockout_risk"] = model.RiskHigh
			} else {
				m.passwordPolicy["lockout_risk"] = model.RiskLow
			}
		} else if strings.Contains(line, "Lockout duration (minutes):") {
			value := strings.TrimPrefix(line, "Lockout duration (minutes):")
			value = strings.TrimSpace(value)
			m.passwordPolicy["lockout_duration"] = value
		}
	}

	if m.passwordPolicy["min_length"] == nil || m.passwordPolicy["min_length"] == "0" {
		m.passwordPolicy["min_length_risk"] = model.RiskHigh
	} else {
		minLen := 0
		fmt.Sscanf(m.passwordPolicy["min_length"].(string), "%d", &minLen)
		if minLen < 8 {
			m.passwordPolicy["min_length_risk"] = model.RiskMedium
		} else {
			m.passwordPolicy["min_length_risk"] = model.RiskLow
		}
	}

	return nil
}

func (m *BaselineModule) collectAccountPolicy() error {
	m.accountPolicy = map[string]interface{}{}

	cmd := exec.Command("net", "user")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(output), "\n")
	userCount := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "Users for") && !strings.HasPrefix(line, "-------") && !strings.HasPrefix(line, "The command") {
			userCount++
		}
	}
	m.accountPolicy["total_users"] = userCount - 2

	cmd2 := exec.Command("powershell", "-Command", "Get-LocalUser | Where-Object {$_.Enabled -eq $true} | Measure-Object | Select-Object -ExpandProperty Count")
	output2, err := cmd2.Output()
	if err == nil {
		enabledUsers := strings.TrimSpace(string(output2))
		m.accountPolicy["enabled_users"] = enabledUsers
	}

	m.accountPolicy["risk_level"] = model.RiskLow

	return nil
}

func (m *BaselineModule) collectAuditPolicy() error {
	m.auditPolicy = []map[string]interface{}{}

	cmd := exec.Command("auditpol", "/get", "/category:*")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "Audit Pol") || strings.Contains(line, "System Log") {
			continue
		}

		if strings.Contains(line, "No Auditing") {
			m.auditPolicy = append(m.auditPolicy, map[string]interface{}{
				"category":   "Unknown",
				"setting":    "No Auditing",
				"risk_level": model.RiskHigh,
			})
		} else if strings.Contains(line, "Success") || strings.Contains(line, "Failure") {
			parts := strings.Split(line, " ")
			category := ""
			setting := ""
			for i, p := range parts {
				if p == "Success" || p == "Failure" {
					category = strings.Join(parts[:i], " ")
					setting = strings.Join(parts[i:], " ")
					break
				}
			}
			if category != "" {
				risk := model.RiskLow
				if strings.Contains(category, "Logon") || strings.Contains(category, "Account") {
					risk = model.RiskMedium
				}
				m.auditPolicy = append(m.auditPolicy, map[string]interface{}{
					"category":   category,
					"setting":    setting,
					"risk_level": risk,
				})
			}
		}
	}

	return nil
}

func (m *BaselineModule) collectNetworkSecurity() error {
	m.networkSecurity = []map[string]interface{}{}

	securitySettings := []struct {
		name        string
		checkCmd    string
		expected    string
		riskIfNot   model.RiskLevel
		description string
	}{
		{
			"Lanman Authentication Level",
			`reg query "HKLM\SYSTEM\CurrentControlSet\Control\Lsa" /vLmCompatibilityLevel`,
			"3",
			model.RiskHigh,
			"Network security: LAN Manager authentication level",
		},
		{
			"LDAP Client Signing",
			`reg query "HKLM\SYSTEM\CurrentControlSet\Services\LDAP" /v LDAPClientIntegrity`,
			"1",
			model.RiskMedium,
			"LDAP client signing requirements",
		},
		{
			"SMBv1 Enabled",
			`reg query "HKLM\SYSTEM\CurrentControlSet\Services\LanmanServer\Parameters" /v SMB1`,
			"0",
			model.RiskCritical,
			"SMB version 1 should be disabled",
		},
	}

	for _, setting := range securitySettings {
		cmd := exec.Command("cmd", "/c", setting.checkCmd)
		output, err := cmd.Output()
		if err != nil {
			m.networkSecurity = append(m.networkSecurity, map[string]interface{}{
				"name":        setting.name,
				"status":      "Not Configured",
				"risk_level":  setting.riskIfNot,
				"description": setting.description,
			})
			continue
		}

		outputStr := strings.TrimSpace(string(output))
		isCompliant := strings.Contains(outputStr, setting.expected)

		riskLevel := model.RiskLow
		if !isCompliant {
			riskLevel = setting.riskIfNot
		}

		m.networkSecurity = append(m.networkSecurity, map[string]interface{}{
			"name":        setting.name,
			"status":      isCompliant,
			"risk_level":  riskLevel,
			"description": setting.description,
		})
	}

	return nil
}

func (m *BaselineModule) collectServiceConfig() error {
	m.serviceConfig = []map[string]interface{}{}

	dangerousServices := []struct {
		name        string
		description string
		riskLevel   model.RiskLevel
	}{
		{"TelnetService", "Telnet service - unencrypted communication", model.RiskHigh},
		{"FTPService", "FTP service - unencrypted communication", model.RiskHigh},
		{"SNMPService", "SNMP service - unencrypted communication", model.RiskMedium},
		{"RemoteRegistry", "Remote Registry - allows remote access to registry", model.RiskHigh},
		{"WinRM", "Windows Remote Management - potential remote access", model.RiskMedium},
		{"RDS", "Remote Desktop Services - potential remote access", model.RiskMedium},
	}

	cmd := exec.Command("sc", "query", "state=", "all")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(output), "\n")
	var currentService string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SERVICE_NAME:") {
			currentService = strings.TrimPrefix(line, "SERVICE_NAME:")
			currentService = strings.TrimSpace(currentService)
		} else if strings.HasPrefix(line, "STATE") && currentService != "" {
			stateStr := strings.TrimSpace(strings.Split(line, ":")[1])

			for _, dangerous := range dangerousServices {
				if strings.ToLower(currentService) == strings.ToLower(dangerous.name) {
					m.serviceConfig = append(m.serviceConfig, map[string]interface{}{
						"name":        currentService,
						"state":       stateStr,
						"description": dangerous.description,
						"risk_level":  dangerous.riskLevel,
					})
					break
				}
			}
			currentService = ""
		}
	}

	return nil
}

func (m *BaselineModule) Stop() error {
	return nil
}

func (m *BaselineModule) GetData() ([]map[string]interface{}, error) {
	result := []map[string]interface{}{}

	result = append(result, map[string]interface{}{
		"type":      "password_policy",
		"data":      m.passwordPolicy,
		"timestamp": time.Now().Format(time.RFC3339),
	})

	result = append(result, map[string]interface{}{
		"type":      "account_policy",
		"data":      m.accountPolicy,
		"timestamp": time.Now().Format(time.RFC3339),
	})

	for _, a := range m.auditPolicy {
		result = append(result, map[string]interface{}{
			"type":       "audit_policy",
			"category":   a["category"],
			"setting":    a["setting"],
			"risk_level": a["risk_level"],
		})
	}

	for _, n := range m.networkSecurity {
		result = append(result, map[string]interface{}{
			"type":        "network_security",
			"name":        n["name"],
			"status":      n["status"],
			"risk_level":  n["risk_level"],
			"description": n["description"],
		})
	}

	for _, s := range m.serviceConfig {
		result = append(result, map[string]interface{}{
			"type":        "service_config",
			"name":        s["name"],
			"state":       s["state"],
			"description": s["description"],
			"risk_level":  s["risk_level"],
		})
	}

	return result, nil
}

func (m *BaselineModule) GetPasswordPolicy() map[string]interface{} {
	return m.passwordPolicy
}

func (m *BaselineModule) GetAuditPolicy() []map[string]interface{} {
	return m.auditPolicy
}

func (m *BaselineModule) GetNetworkSecurity() []map[string]interface{} {
	return m.networkSecurity
}

func (m *BaselineModule) GetServiceConfig() []map[string]interface{} {
	return m.serviceConfig
}
