//go:build windows

package m23_baseline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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

func (m *BaselineModule) CheckFirewallStatus() (map[string]interface{}, error) {
	result := map[string]interface{}{
		"domain_profile":  "Unknown",
		"private_profile": "Unknown",
		"public_profile":  "Unknown",
		"check_timestamp": time.Now().Format(time.RFC3339),
	}

	cmd := exec.Command("netsh", "advfirewall", "show", "allprofiles", "state")
	output, err := cmd.Output()
	if err != nil {
		return result, fmt.Errorf("failed to check firewall status: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	currentProfile := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Domain Profile") {
			currentProfile = "domain_profile"
		} else if strings.Contains(line, "Private Profile") {
			currentProfile = "private_profile"
		} else if strings.Contains(line, "Public Profile") {
			currentProfile = "public_profile"
		} else if strings.Contains(line, "State") && currentProfile != "" {
			if strings.Contains(line, "ON") {
				result[currentProfile] = "Enabled"
			} else if strings.Contains(line, "OFF") {
				result[currentProfile] = "Disabled"
				result["risk_level"] = model.RiskHigh
			}
		}
	}

	if result["risk_level"] == nil {
		result["risk_level"] = model.RiskLow
	}

	return result, nil
}

func (m *BaselineModule) CheckUACSettings() (map[string]interface{}, error) {
	result := map[string]interface{}{
		"uac_enabled":     false,
		"consent_admin":   false,
		"consent_ui":      false,
		"enable_lua":      false,
		"check_timestamp": time.Now().Format(time.RFC3339),
	}

	cmd := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
$uac = Get-ItemProperty -Path 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System'
Write-Output ("EnableLUA:" + $uac.EnableLUA)
Write-Output ("ConsentAdmin:" + $uac.ConsentAdminBehaviorLast)
Write-Output ("ConsentUI:" + $uac.EnableUI)
Write-Output ("FilterAdmin:" + $uac.FilterAdministratorToken)`)

	output, err := cmd.Output()
	if err != nil {
		return result, fmt.Errorf("failed to check UAC settings: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "EnableLUA:") {
			val := strings.TrimPrefix(line, "EnableLUA:")
			result["uac_enabled"] = strings.TrimSpace(val) == "1"
		} else if strings.HasPrefix(line, "ConsentAdmin:") {
			val := strings.TrimPrefix(line, "ConsentAdmin:")
			result["consent_admin"] = strings.TrimSpace(val) == "2"
		} else if strings.HasPrefix(line, "ConsentUI:") {
			val := strings.TrimPrefix(line, "ConsentUI:")
			result["consent_ui"] = strings.TrimSpace(val) == "1"
		} else if strings.HasPrefix(line, "FilterAdmin:") {
			val := strings.TrimPrefix(line, "FilterAdmin:")
			result["enable_lua"] = strings.TrimSpace(val) == "1"
		}
	}

	if !result["uac_enabled"].(bool) || !result["enable_lua"].(bool) {
		result["risk_level"] = model.RiskCritical
	} else if !result["consent_admin"].(bool) {
		result["risk_level"] = model.RiskMedium
	} else {
		result["risk_level"] = model.RiskLow
	}

	return result, nil
}

func (m *BaselineModule) DetectLMHashStorage() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	cmd := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
$lmHash = Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Control\Lsa' -Name 'NoLMHash' -ErrorAction SilentlyContinue
if($lmHash) {
    if($lmHash.NoLMHash -eq 0) {
        Write-Output 'LMHash:Enabled'
    } else {
        Write-Output 'LMHash:Disabled'
    }
} else {
    Write-Output 'LMHash:Unknown'
}`)

	output, err := cmd.Output()
	if err == nil {
		result := strings.TrimSpace(string(output))
		if result == "LMHash:Enabled" {
			results = append(results, map[string]interface{}{
				"setting":        "LM Hash Storage",
				"status":         "Enabled - LM hashes stored in SAM",
				"risk_level":     model.RiskHigh,
				"recommendation": "Disable LM hash storage via NoLMHash registry key",
			})
		} else if result == "LMHash:Disabled" {
			results = append(results, map[string]interface{}{
				"setting":    "LM Hash Storage",
				"status":     "Disabled",
				"risk_level": model.RiskLow,
			})
		}
	}

	cmd2 := exec.Command("net", "accounts")
	output2, err := cmd2.Output()
	if err == nil {
		lines := strings.Split(string(output2), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Lannman Authentication Level") {
				results = append(results, map[string]interface{}{
					"setting":    "LM Authentication",
					"status":     strings.TrimSpace(strings.Split(line, ":")[1]),
					"risk_level": model.RiskMedium,
				})
			}
		}
	}

	return results, nil
}

func (m *BaselineModule) ExportReport(filePath string) error {
	data := map[string]interface{}{
		"timestamp":        time.Now().Format(time.RFC3339),
		"password_policy":  m.passwordPolicy,
		"account_policy":   m.accountPolicy,
		"audit_policy":     m.auditPolicy,
		"network_security": m.networkSecurity,
		"service_config":   m.serviceConfig,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal baseline data: %w", err)
	}

	return os.WriteFile(filePath, jsonData, 0644)
}
