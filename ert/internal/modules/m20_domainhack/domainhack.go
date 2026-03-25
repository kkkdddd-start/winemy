//go:build windows

package m20_domainhack

import (
	"context"
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
