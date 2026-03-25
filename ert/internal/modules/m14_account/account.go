//go:build windows

package m14_account

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type AccountModule struct {
	ctx      context.Context
	storage  registry.Storage
	accounts []model.AccountDTO
}

func New() *AccountModule {
	return &AccountModule{}
}

func (m *AccountModule) ID() int       { return 14 }
func (m *AccountModule) Name() string  { return "account" }
func (m *AccountModule) Priority() int { return 1 }

func (m *AccountModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	return nil
}

func (m *AccountModule) Collect(ctx context.Context) error {
	m.accounts = []model.AccountDTO{}

	output, err := exec.Command("net", "user").Output()
	if err != nil {
		m.accounts = append(m.accounts, model.AccountDTO{
			Name:      "Error",
			FullName:  fmt.Sprintf("Failed to enumerate accounts: %v", err),
			SID:       "",
			Domain:    "",
			Status:    "Unknown",
			LastLogon: time.Time{},
			RiskLevel: model.RiskLow,
		})
		return nil
	}

	lines := strings.Split(string(output), "\n")
	inUserSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "---") {
			inUserSection = true
			continue
		}

		if inUserSection && line != "" && !strings.HasPrefix(line, "The") && !strings.HasPrefix(line, "Command") {
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				username := parts[0]

				accountInfo := m.getAccountDetails(username)
				m.accounts = append(m.accounts, accountInfo)
			}
		}
	}

	m.collectDomainAccounts()

	return nil
}

func (m *AccountModule) getAccountDetails(username string) model.AccountDTO {
	output, err := exec.Command("net", "user", username).Output()
	if err != nil {
		return model.AccountDTO{
			Name:      username,
			FullName:  "",
			SID:       "",
			Domain:    "",
			Status:    "Unknown",
			LastLogon: time.Time{},
			RiskLevel: model.RiskLow,
		}
	}

	info := parseNetUserOutput(string(output), username)

	riskLevel := model.RiskLow
	isGuest := strings.ToLower(username) == "guest"
	isDisabled := strings.Contains(strings.ToLower(info.Status), "disabled")
	hasPassword := m.checkAccountHasPassword(username)

	if isGuest && !isDisabled {
		riskLevel = model.RiskHigh
	}

	if isSuspiciousAccount(username, info.FullName) {
		riskLevel = model.RiskHigh
	}

	return model.AccountDTO{
		Name:                 info.Name,
		FullName:             info.FullName,
		SID:                  info.SID,
		Domain:               info.Domain,
		Status:               info.Status,
		LastLogon:            info.LastLogon,
		PasswordExpired:      info.PasswordExpired,
		PasswordNeverExpires: info.PasswordNeverExpires,
		HasPassword:          hasPassword,
		IsDisabled:           isDisabled,
		IsGuest:              isGuest,
		RiskLevel:            riskLevel,
	}
}

func (m *AccountModule) checkAccountHasPassword(username string) bool {
	cmd := exec.Command("net", "user", username)
	output, err := cmd.Output()
	if err != nil {
		return true
	}
	return !strings.Contains(string(output), "Password not required")
}

type accountInfo struct {
	Name                 string
	FullName             string
	SID                  string
	Domain               string
	Status               string
	LastLogon            time.Time
	PasswordExpired      bool
	PasswordNeverExpires bool
}

func parseNetUserOutput(output, defaultName string) accountInfo {
	info := accountInfo{
		Name:   defaultName,
		Status: "Unknown",
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Full Name") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				info.FullName = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Account active") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				status := strings.TrimSpace(parts[1])
				info.Status = status
			}
		} else if strings.HasPrefix(line, "Last logon") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				lastLogonStr := strings.TrimSpace(parts[1])
				if lastLogonStr != "Never" {
					info.LastLogon = parseLogonTime(lastLogonStr)
				}
			}
		} else if strings.HasPrefix(line, "SID") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				info.SID = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Password expires") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				expiresStr := strings.TrimSpace(parts[1])
				info.PasswordNeverExpires = expiresStr == "Never"
			}
		} else if strings.HasPrefix(line, "Password last set") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				lastSetStr := strings.TrimSpace(parts[1])
				info.PasswordExpired = lastSetStr == "Never"
			}
		}
	}

	return info
}

func parseLogonTime(timeStr string) time.Time {
	timeStr = strings.TrimSpace(timeStr)

	formats := []string{
		"1/2/2006 3:04:05 PM",
		"1/2/2006 15:04:05",
		"01/02/2006 15:04:05",
		"01/02/2006 3:04:05 PM",
		"1/2/2006",
		"01/02/2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t
		}
	}
	return time.Time{}
}

func (m *AccountModule) collectDomainAccounts() {
	output, err := exec.Command("net", "user", "/domain").Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	inUserSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "---") {
			inUserSection = true
			continue
		}

		if inUserSection && line != "" && !strings.HasPrefix(line, "The") && !strings.HasPrefix(line, "Command") {
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				username := parts[0]
				if username != "" {
					m.accounts = append(m.accounts, model.AccountDTO{
						Name:      username,
						FullName:  "",
						SID:       "",
						Domain:    "DOMAIN",
						Status:    "Unknown",
						LastLogon: time.Time{},
						RiskLevel: model.RiskLow,
					})
				}
			}
		}
	}
}

func isSuspiciousAccount(name, fullName string) bool {
	nameLower := strings.ToLower(name)
	fullNameLower := strings.ToLower(fullName)

	suspiciousPatterns := []string{
		"hidden",
		"temp",
		"secret",
		"backdoor",
		"malware",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(nameLower, pattern) || strings.Contains(fullNameLower, pattern) {
			return true
		}
	}

	return false
}

func (m *AccountModule) Stop() error {
	return nil
}

func (m *AccountModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.accounts))
	for _, a := range m.accounts {
		lastLogonStr := ""
		if !a.LastLogon.IsZero() {
			lastLogonStr = a.LastLogon.Format(time.RFC3339)
		}
		result = append(result, map[string]interface{}{
			"name":       a.Name,
			"full_name":  a.FullName,
			"sid":        a.SID,
			"domain":     a.Domain,
			"status":     a.Status,
			"last_logon": lastLogonStr,
			"risk_level": a.RiskLevel,
		})
	}
	return result, nil
}

func (m *AccountModule) CollectGroups() ([]map[string]interface{}, error) {
	var groups []map[string]interface{}

	cmd := exec.Command("net", "localgroup")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate groups: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "The") || strings.HasPrefix(line, "Command") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			groupName := parts[0]
			if groupName != "" {
				groups = append(groups, map[string]interface{}{
					"name":       groupName,
					"domain":     "LOCAL",
					"risk_level": model.RiskLow,
				})
			}
		}
	}

	return groups, nil
}

func (m *AccountModule) GetGroupMembers(groupName string) ([]string, error) {
	var members []string

	cmd := exec.Command("net", "localgroup", groupName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	inMemberSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "---") {
			inMemberSection = true
			continue
		}
		if inMemberSection && line != "" && !strings.HasPrefix(line, "The") && !strings.HasPrefix(line, "Command") {
			members = append(members, line)
		}
	}

	return members, nil
}

func (m *AccountModule) ParseAccountSID(sid string) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"sid":           sid,
		"account_type":  "Unknown",
		"domain":        "Unknown",
		"rid":           0,
		"is_well_known": false,
	}

	if sid == "" {
		return result, nil
	}

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$obj = New-Object System.Security.Principal.SecurityIdentifier('%s'); $acc = $obj.Translate([System.Security.Principal.NTAccount]); Write-Output $acc.Value`, sid))
	output, err := cmd.Output()
	if err == nil {
		accountName := strings.TrimSpace(string(output))
		if accountName != "" && !strings.Contains(accountName, "System.Security") {
			result["account_name"] = accountName
			if strings.Contains(accountName, "\\") {
				parts := strings.Split(accountName, "\\")
				result["domain"] = parts[0]
				result["account_name"] = parts[1]
			}
		}
	}

	parts := strings.Split(sid, "-")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		rid, err := strconv.ParseUint(lastPart, 10, 32)
		if err == nil {
			result["rid"] = rid

			switch {
			case rid == 500:
				result["account_type"] = "Administrator"
				result["is_well_known"] = true
			case rid == 501:
				result["account_type"] = "Guest"
				result["is_well_known"] = true
			case rid == 502:
				result["account_type"] = "Krbtgt"
				result["is_well_known"] = true
			case rid >= 512 && rid <= 519:
				result["account_type"] = "Well-Known Group"
				result["is_well_known"] = true
			case rid >= 1000 && rid <= 1100:
				result["account_type"] = "User Account"
				result["is_well_known"] = false
			default:
				result["account_type"] = "Other"
			}
		}
	}

	return result, nil
}

func (m *AccountModule) DetectEmptyPassword() ([]model.AccountDTO, error) {
	var results []model.AccountDTO

	cmd := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
Get-LocalUser | Where-Object { $_.PasswordRequired -eq $false } | ForEach-Object {
    Write-Output $_.Name
}`)

	output, err := cmd.Output()
	if err != nil {
		return results, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		username := strings.TrimSpace(line)
		if username != "" {
			for _, a := range m.accounts {
				if strings.ToLower(a.Name) == strings.ToLower(username) {
					a.RiskLevel = model.RiskCritical
					results = append(results, a)
					break
				}
			}
		}
	}

	return results, nil
}

func (m *AccountModule) DetectNeverExpirePassword() ([]model.AccountDTO, error) {
	var results []model.AccountDTO

	cmd := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
Get-LocalUser | Where-Object { $_.PasswordNeverExpires -eq $true } | ForEach-Object {
    Write-Output $_.Name
}`)

	output, err := cmd.Output()
	if err != nil {
		return results, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		username := strings.TrimSpace(line)
		if username != "" {
			for _, a := range m.accounts {
				if strings.ToLower(a.Name) == strings.ToLower(username) {
					a.RiskLevel = model.RiskHigh
					results = append(results, a)
					break
				}
			}
		}
	}

	return results, nil
}

func (m *AccountModule) DetectGuestEnabled() ([]model.AccountDTO, error) {
	var results []model.AccountDTO

	cmd := exec.Command("net", "user", "guest")
	output, err := cmd.Output()
	if err != nil {
		return results, nil
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "Account active") && !strings.Contains(outputStr, "Account active    No") {
		for _, a := range m.accounts {
			if strings.ToLower(a.Name) == "guest" {
				a.RiskLevel = model.RiskHigh
				results = append(results, a)
				break
			}
		}
	}

	return results, nil
}

func (m *AccountModule) Search(keyword string) ([]model.AccountDTO, error) {
	results := []model.AccountDTO{}
	keywordLower := strings.ToLower(keyword)

	for _, a := range m.accounts {
		if strings.Contains(strings.ToLower(a.Name), keywordLower) ||
			strings.Contains(strings.ToLower(a.FullName), keywordLower) ||
			strings.Contains(strings.ToLower(a.Domain), keywordLower) ||
			strings.Contains(strings.ToLower(a.SID), keywordLower) {
			results = append(results, a)
		}
	}

	return results, nil
}
