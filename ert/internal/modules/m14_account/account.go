//go:build windows

package m14_account

import (
	"context"
	"fmt"
	"os/exec"
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
	if isSuspiciousAccount(username, info.FullName) {
		riskLevel = model.RiskHigh
	}

	return model.AccountDTO{
		Name:      info.Name,
		FullName:  info.FullName,
		SID:       info.SID,
		Domain:    info.Domain,
		Status:    info.Status,
		LastLogon: info.LastLogon,
		RiskLevel: riskLevel,
	}
}

type accountInfo struct {
	Name      string
	FullName  string
	SID       string
	Domain    string
	Status    string
	LastLogon time.Time
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
				info.Status = strings.TrimSpace(parts[1])
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
