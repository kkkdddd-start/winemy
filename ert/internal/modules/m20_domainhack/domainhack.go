//go:build windows

package m20_domainhack

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type DomainHackModule struct {
	ctx           context.Context
	storage       registry.Storage
	kerberoasting []map[string]interface{}
	asRepRoasting []map[string]interface{}
	goldenTicket  []map[string]interface{}
	silverTicket  []map[string]interface{}
}

func New() *DomainHackModule {
	return &DomainHackModule{
		kerberoasting: []map[string]interface{}{},
		asRepRoasting: []map[string]interface{}{},
		goldenTicket:  []map[string]interface{}{},
		silverTicket:  []map[string]interface{}{},
	}
}

func (m *DomainHackModule) ID() int       { return 20 }
func (m *DomainHackModule) Name() string  { return "domainhack" }
func (m *DomainHackModule) Priority() int { return 1 }

func (m *DomainHackModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	return nil
}

func (m *DomainHackModule) Collect(ctx context.Context) error {
	if err := m.detectKerberoasting(); err != nil {
		return err
	}
	if err := m.detectASREPRoasting(); err != nil {
		return err
	}
	m.detectGoldenTicket()
	m.detectSilverTicket()
	return nil
}

func (m *DomainHackModule) detectKerberoasting() error {
	m.kerberoasting = []map[string]interface{}{}

	cmd := exec.Command("powershell", "-Command", "Get-ADUser", "-Filter", "ServicePrincipalName -ne '$null'", "-Properties", "ServicePrincipalName,SamAccountName")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "Get-ADUser") {
			continue
		}
		if strings.Contains(line, "SPN:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				spn := strings.TrimSpace(parts[1])
				user := extractUserFromSPN(spn)
				m.kerberoasting = append(m.kerberoasting, map[string]interface{}{
					"type":        "kerberoasting",
					"account":     user,
					"spn":         spn,
					"risk_level":  model.RiskHigh,
					"description": "Account with SPN set - vulnerable to Kerberoasting",
					"detected_at": time.Now().Format(time.RFC3339),
				})
			}
		}
	}

	cmd2 := exec.Command("powershell", "-Command", "setspn", "-T", "domain", "-Q", "*/*")
	output2, err := cmd2.Output()
	if err == nil {
		lines2 := strings.Split(string(output2), "\n")
		for _, line := range lines2 {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "CN=") {
				exists := false
				for _, k := range m.kerberoasting {
					if k["account"] == extractCN(line) {
						exists = true
						break
					}
				}
				if !exists {
					m.kerberoasting = append(m.kerberoasting, map[string]interface{}{
						"type":        "kerberoasting",
						"account":     extractCN(line),
						"spn":         line,
						"risk_level":  model.RiskHigh,
						"description": "Account with SPN discovered via setspn",
						"detected_at": time.Now().Format(time.RFC3339),
					})
				}
			}
		}
	}

	return nil
}

func (m *DomainHackModule) detectASREPRoasting() error {
	m.asRepRoasting = []map[string]interface{}{}

	cmd := exec.Command("powershell", "-Command", "Get-ADUser", "-Filter", "DoesNotRequirePreAuth -eq $true", "-Properties", "DoesNotRequirePreAuth,SamAccountName")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "Get-ADUser") {
			continue
		}
		if strings.Contains(line, "True") || strings.Contains(line, "SamAccountName") {
			user := extractUserFromLine(line)
			if user != "" {
				m.asRepRoasting = append(m.asRepRoasting, map[string]interface{}{
					"type":        "as_rep_roasting",
					"account":     user,
					"risk_level":  model.RiskCritical,
					"description": "Pre-authentication not required - vulnerable to AS-REP Roasting",
					"detected_at": time.Now().Format(time.RFC3339),
				})
			}
		}
	}

	return nil
}

func (m *DomainHackModule) detectGoldenTicket() error {
	m.goldenTicket = []map[string]interface{}{}

	m.goldenTicket = append(m.goldenTicket, map[string]interface{}{
		"type":           "golden_ticket",
		"risk_level":     model.RiskCritical,
		"description":    "Golden Ticket detection requires memory forensics or log analysis",
		"recommendation": "Check for abnormal TGT lifetimes, krbtgt password changes",
		"detected_at":    time.Now().Format(time.RFC3339),
	})

	return nil
}

func (m *DomainHackModule) detectSilverTicket() error {
	m.silverTicket = []map[string]interface{}{}

	m.silverTicket = append(m.silverTicket, map[string]interface{}{
		"type":           "silver_ticket",
		"risk_level":     model.RiskHigh,
		"description":    "Silver Ticket detection requires service log analysis",
		"recommendation": "Check for abnormal service ticket usage patterns",
		"detected_at":    time.Now().Format(time.RFC3339),
	})

	return nil
}

func extractUserFromSPN(spn string) string {
	parts := strings.Split(spn, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return spn
}

func extractCN(dn string) string {
	if strings.HasPrefix(dn, "CN=") {
		parts := strings.Split(dn, ",")
		if len(parts) > 0 {
			return strings.TrimPrefix(parts[0], "CN=")
		}
	}
	return dn
}

func extractUserFromLine(line string) string {
	if strings.Contains(line, "SamAccountName") {
		parts := strings.Split(line, ":")
		if len(parts) >= 2 {
			return strings.TrimSpace(parts[1])
		}
	}
	for _, word := range strings.Fields(line) {
		if !strings.Contains(word, "=") && word != "True" && word != "False" && word != "SamAccountName" {
			return word
		}
	}
	return ""
}

func (m *DomainHackModule) Stop() error {
	return nil
}

func (m *DomainHackModule) GetData() ([]map[string]interface{}, error) {
	result := []map[string]interface{}{}

	for _, k := range m.kerberoasting {
		result = append(result, map[string]interface{}{
			"category":    "kerberoasting",
			"account":     k["account"],
			"spn":         k["spn"],
			"risk_level":  k["risk_level"],
			"description": k["description"],
			"detected_at": k["detected_at"],
		})
	}

	for _, a := range m.asRepRoasting {
		result = append(result, map[string]interface{}{
			"category":    "as_rep_roasting",
			"account":     a["account"],
			"risk_level":  a["risk_level"],
			"description": a["description"],
			"detected_at": a["detected_at"],
		})
	}

	for _, g := range m.goldenTicket {
		result = append(result, map[string]interface{}{
			"category":       "golden_ticket",
			"risk_level":     g["risk_level"],
			"description":    g["description"],
			"recommendation": g["recommendation"],
			"detected_at":    g["detected_at"],
		})
	}

	for _, s := range m.silverTicket {
		result = append(result, map[string]interface{}{
			"category":       "silver_ticket",
			"risk_level":     s["risk_level"],
			"description":    s["description"],
			"recommendation": s["recommendation"],
			"detected_at":    s["detected_at"],
		})
	}

	return result, nil
}

func (m *DomainHackModule) GetKerberoasting() []map[string]interface{} {
	return m.kerberoasting
}

func (m *DomainHackModule) GetASREPRoasting() []map[string]interface{} {
	return m.asRepRoasting
}

func (m *DomainHackModule) DetectPasswordSpray() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	cmd := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
Get-WinEvent -FilterHashtable @{LogName='Security';ID=4625} -MaxEvents 100 -ErrorAction SilentlyContinue | ForEach-Object {
    $xml = [xml]$_.ToXml()
    $eventData = $xml.Event.EventData.Data
    $targetUser = ($eventData | Where-Object { $_.Name -eq 'TargetUserName' }).'#text'
    $targetDomain = ($eventData | Where-Object { $_.Name -eq 'TargetDomainName' }).'#text'
    $ipAddress = ($eventData | Where-Object { $_.Name -eq 'IpAddress' }).'#text'
    $time = $_.TimeCreated.ToString('yyyy-MM-dd HH:mm:ss')
    Write-Output ("$targetDomain\\$targetUser|$ipAddress|$time")
}`)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to detect password spray: %w", err)
	}

	attempts := make(map[string]int)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			key := parts[0] + "_" + parts[1]
			attempts[key]++
		}
	}

	for key, count := range attempts {
		if count >= 5 {
			parts := strings.Split(key, "_")
			if len(parts) >= 2 {
				results = append(results, map[string]interface{}{
					"account":     parts[0],
					"ip_address":  parts[1],
					"attempts":    count,
					"type":        "password_spray",
					"risk_level":  model.RiskCritical,
					"description": fmt.Sprintf("Potential password spray attack: %d failed login attempts", count),
				})
			}
		}
	}

	return results, nil
}

func (m *DomainHackModule) DetectAccountLockout() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	cmd := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
Get-WinEvent -FilterHashtable @{LogName='Security';ID=4740} -MaxEvents 50 -ErrorAction SilentlyContinue | ForEach-Object {
    $xml = [xml]$_.ToXml()
    $eventData = $xml.Event.EventData.Data
    $targetUser = ($eventData | Where-Object { $_.Name -eq 'TargetUserName' }).'#text'
    $targetDomain = ($eventData | Where-Object { $_.Name -eq 'TargetDomainName' }).'#text'
    $callerMachine = ($eventData | Where-Object { $_.Name -eq 'CallerMachineName' }).'#text'
    $callerUser = ($eventData | Where-Object { $_.Name -eq 'CallerUserName' }).'#text'
    $time = $_.TimeCreated.ToString('yyyy-MM-dd HH:mm:ss')
    Write-Output ("$targetDomain\\$targetUser|$callerMachine\\$callerUser|$time")
}`)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to detect account lockout: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			results = append(results, map[string]interface{}{
				"locked_account": parts[0],
				"lockout_source": parts[1],
				"timestamp":      parts[2],
				"type":           "account_lockout",
				"risk_level":     model.RiskHigh,
			})
		}
	}

	return results, nil
}

func (m *DomainHackModule) DetectSensitiveGroupChange() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	sensitiveGroups := []string{
		"Domain Admins",
		"Enterprise Admins",
		"Schema Admins",
		"Administrators",
		"Account Operators",
		"Backup Operators",
		"Server Operators",
	}

	for _, groupName := range sensitiveGroups {
		cmd := exec.Command("powershell", "-Command",
			fmt.Sprintf(`$ErrorActionPreference='SilentlyContinue'
Get-WinEvent -FilterHashtable @{LogName='Security';ID=4728 -or ID=4729 -or ID=4732 -or ID=4733} -MaxEvents 50 -ErrorAction SilentlyContinue | ForEach-Object {{
    $xml = [xml]$_.ToXml()
    $eventData = $xml.Event.EventData.Data
    $member = ($eventData | Where-Object {{ $_.Name -eq 'Member' }}).'#text'
    $targetUser = ($eventData | Where-Object {{ $_.Name -eq 'TargetUserName' }}).'#text'
    $domain = ($eventData | Where-Object {{ $_.Name -eq 'TargetDomainName' }}).'#text'
    if($targetUser -eq '%s') {{
        Write-Output ("$domain\\$targetUser|$member|{0}" -f $_.TimeCreated.ToString('yyyy-MM-dd HH:mm:ss'))
    }}
}}`, groupName))

		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				parts := strings.Split(line, "|")
				if len(parts) >= 3 {
					results = append(results, map[string]interface{}{
						"group_name":   groupName,
						"member_added": parts[1],
						"timestamp":    parts[2],
						"type":         "sensitive_group_change",
						"risk_level":   model.RiskCritical,
					})
				}
			}
		}
	}

	return results, nil
}

func (m *DomainHackModule) DetectDCSync() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	cmd := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
Get-WinEvent -FilterHashtable @{LogName='Security';ID=4662} -MaxEvents 100 -ErrorAction SilentlyContinue | ForEach-Object {
    $xml = [xml]$_.ToXml()
    $eventData = $xml.Event.EventData.Data
    $objectType = ($eventData | Where-Object { $_.Name -eq 'ObjectType' }).'#text'
    $objectName = ($eventData | Where-Object { $_.Name -eq 'ObjectName' }).'#text'
    $accessMask = ($eventData | Where-Object { $_.Name -eq 'AccessMask' }).'#text'
    if($objectType -match 'ds-replication' -or $objectName -match 'DC=DomainDnsZones|DC=ForestDnsZones' -or $accessMask -match '0x100|0x10000') {
        $user = ($eventData | Where-Object { $_.Name -eq 'SubjectUserName' }).'#text'
        $domain = ($eventData | Where-Object { $_.Name -eq 'SubjectDomainName' }).'#text'
        Write-Output ("$domain\\$user|$objectType|$accessMask|" + $_.TimeCreated.ToString('yyyy-MM-dd HH:mm:ss'))
    }
}`)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to detect DCSync: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			results = append(results, map[string]interface{}{
				"account_used": parts[0],
				"object_type":  parts[1],
				"access_mask":  parts[2],
				"timestamp":    parts[3],
				"type":         "dcsync_attack",
				"risk_level":   model.RiskCritical,
			})
		}
	}

	return results, nil
}

func (m *DomainHackModule) DetectPtH() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	cmd := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
Get-WinEvent -FilterHashtable @{LogName='Security';ID=4624} -MaxEvents 200 -ErrorAction SilentlyContinue | Where-Object {
    $_.Message -match 'LogonType.*3' -and $_.Message -match 'NTLM'
} | ForEach-Object {
    $xml = [xml]$_.ToXml()
    $eventData = $xml.Event.EventData.Data
    $targetUser = ($eventData | Where-Object { $_.Name -eq 'TargetUserName' }).'#text'
    $ipAddress = ($eventData | Where-Object { $_.Name -eq 'IpAddress' }).'#text'
    $time = $_.TimeCreated.ToString('yyyy-MM-dd HH:mm:ss')
    if($ipAddress -and $ipAddress -ne '-') {
        Write-Output ("$targetUser|$ipAddress|$time")
    }
}`)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to detect PtH: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			results = append(results, map[string]interface{}{
				"account":    parts[0],
				"ip_address": parts[1],
				"timestamp":  parts[2],
				"type":       "pth_attack",
				"risk_level": model.RiskCritical,
			})
		}
	}

	return results, nil
}

func (m *DomainHackModule) DetectRDPHijack() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	cmd := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
Get-WinEvent -FilterHashtable @{LogName='Security';ID=4624} -MaxEvents 100 -ErrorAction SilentlyContinue | Where-Object {
    $_.Message -match 'LogonType.*10'
} | ForEach-Object {
    $xml = [xml]$_.ToXml()
    $eventData = $xml.Event.EventData.Data
    $targetUser = ($eventData | Where-Object { $_.Name -eq 'TargetUserName' }).'#text'
    $ipAddress = ($eventData | Where-Object { $_.Name -eq 'IpAddress' }).'#text'
    $time = $_.TimeCreated.ToString('yyyy-MM-dd HH:mm:ss')
    Write-Output ("$targetUser|$ipAddress|$time")
}`)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to detect RDP hijack: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			results = append(results, map[string]interface{}{
				"account":    parts[0],
				"ip_address": parts[1],
				"timestamp":  parts[2],
				"type":       "rdp_hijack",
				"risk_level": model.RiskHigh,
			})
		}
	}

	return results, nil
}

func (m *DomainHackModule) DetectWMIExec() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	cmd := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
Get-WinEvent -FilterHashtable @{LogName='Microsoft-Windows-Sysmon/Operational';ID=1} -MaxEvents 100 -ErrorAction SilentlyContinue | Where-Object {
    $_.Message -match 'powershell.*-c.*New-Object.*WMISystem|WMIC.*process.*call.*create'
} | ForEach-Object {
    $commandLine = $_.Message -match 'CommandLine.*powershell'
    $time = $_.TimeCreated.ToString('yyyy-MM-dd HH:mm:ss')
    Write-Output ("WMIExec|$commandLine|$time")
}`)

	output, err := cmd.Output()
	if err != nil {
		cmd2 := exec.Command("powershell", "-Command",
			`$ErrorActionPreference='SilentlyContinue'
Get-WinEvent -FilterHashtable @{LogName='Security';ID=4688} -MaxEvents 100 -ErrorAction SilentlyContinue | Where-Object {
    $_.Message -match 'wmic.*process.*call.*create'
} | ForEach-Object {
    $time = $_.TimeCreated.ToString('yyyy-MM-dd HH:mm:ss')
    Write-Output ("WMIExec|SuspiciousWMICall|$time")
}`)

		output2, err2 := cmd2.Output()
		if err2 == nil {
			lines2 := strings.Split(string(output2), "\n")
			for _, line := range lines2 {
				line = strings.TrimSpace(line)
				if line != "" {
					parts := strings.Split(line, "|")
					if len(parts) >= 2 {
						results = append(results, map[string]interface{}{
							"type":       parts[0],
							"indicator":  parts[1],
							"timestamp":  parts[2],
							"risk_level": model.RiskHigh,
						})
					}
				}
			}
		}
		return results, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			parts := strings.Split(line, "|")
			if len(parts) >= 3 {
				results = append(results, map[string]interface{}{
					"type":       parts[0],
					"indicator":  parts[1],
					"timestamp":  parts[2],
					"risk_level": model.RiskHigh,
				})
			}
		}
	}

	return results, nil
}

func (m *DomainHackModule) DetectPSRemoting() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	cmd := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
Get-WinEvent -FilterHashtable @{LogName='Security';ID=4624} -MaxEvents 100 -ErrorAction SilentlyContinue | Where-Object {
    $_.Message -match 'LogonType.*3' -and $_.Message -match 'LocalTokenActivation'
} | ForEach-Object {
    $xml = [xml]$_.ToXml()
    $eventData = $xml.Event.EventData.Data
    $targetUser = ($eventData | Where-Object { $_.Name -eq 'TargetUserName' }).'#text'
    $ipAddress = ($eventData | Where-Object { $_.Name -eq 'IpAddress' }).'#text'
    $time = $_.TimeCreated.ToString('yyyy-MM-dd HH:mm:ss')
    if($ipAddress -and $ipAddress -ne '-') {
        Write-Output ("$targetUser|$ipAddress|$time|PSRemoting")
    }
}`)

	output, err := cmd.Output()
	if err != nil {
		return results, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			results = append(results, map[string]interface{}{
				"account":    parts[0],
				"ip_address": parts[1],
				"timestamp":  parts[2],
				"type":       "psremoting",
				"risk_level": model.RiskMedium,
			})
		}
	}

	return results, nil
}

func (m *DomainHackModule) Search(keyword string) ([]map[string]interface{}, error) {
	results := []map[string]interface{}{}
	keywordLower := strings.ToLower(keyword)

	for _, k := range m.kerberoasting {
		if str, ok := k["account"].(string); ok {
			if strings.Contains(strings.ToLower(str), keywordLower) {
				results = append(results, map[string]interface{}{
					"category": "kerberoasting",
					"account":  k["account"],
					"spn":      k["spn"],
				})
			}
		}
	}

	for _, a := range m.asRepRoasting {
		if str, ok := a["account"].(string); ok {
			if strings.Contains(strings.ToLower(str), keywordLower) {
				results = append(results, map[string]interface{}{
					"category": "as_rep_roasting",
					"account":  a["account"],
				})
			}
		}
	}

	return results, nil
}
