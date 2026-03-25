//go:build windows

package m19_domain

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type DomainModule struct {
	ctx            context.Context
	storage        registry.Storage
	domainInfo     map[string]interface{}
	domainUsers    []model.AccountDTO
	domainGroups   []map[string]interface{}
	ouStruct       []map[string]interface{}
	gpoList        []map[string]interface{}
	trustRelations []map[string]interface{}
}

func New() *DomainModule {
	return &DomainModule{
		domainUsers:    []model.AccountDTO{},
		domainGroups:   []map[string]interface{}{},
		ouStruct:       []map[string]interface{}{},
		gpoList:        []map[string]interface{}{},
		trustRelations: []map[string]interface{}{},
	}
}

func (m *DomainModule) ID() int       { return 19 }
func (m *DomainModule) Name() string  { return "domain" }
func (m *DomainModule) Priority() int { return 1 }

func (m *DomainModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	return nil
}

func (m *DomainModule) Collect(ctx context.Context) error {
	if err := m.collectDomainInfo(); err != nil {
		return err
	}
	if err := m.collectDomainUsers(); err != nil {
		return err
	}
	if err := m.collectDomainGroups(); err != nil {
		return err
	}
	if err := m.collectOUStructure(); err != nil {
		return err
	}
	if err := m.collectGPOList(); err != nil {
		return err
	}
	if err := m.collectTrustRelations(); err != nil {
		return err
	}
	return nil
}

func (m *DomainModule) collectDomainInfo() error {
	m.domainInfo = make(map[string]interface{})

	cmd := exec.Command("systeminfo")
	output, err := cmd.Output()
	if err != nil {
		m.domainInfo["domain"] = "WORKGROUP"
		m.domainInfo["domain_role"] = "Standalone"
		return nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Domain:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				m.domainInfo["domain"] = strings.TrimSpace(parts[1])
			}
		}
		if strings.Contains(line, "Domain Role:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				m.domainInfo["domain_role"] = strings.TrimSpace(parts[1])
			}
		}
	}

	if m.domainInfo["domain"] == "WORKGROUP" {
		m.domainInfo["is_domain_joined"] = false
	} else {
		m.domainInfo["is_domain_joined"] = true
	}

	return nil
}

func (m *DomainModule) collectDomainUsers() error {
	m.domainUsers = []model.AccountDTO{}

	cmd := exec.Command("net", "user", "/domain")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(output), "\n")
	inUserList := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "-----------") || strings.Contains(line, "The command") {
			inUserList = true
			continue
		}
		if inUserList && line != "" && !strings.HasPrefix(line, "The") {
			user := model.AccountDTO{
				Name:      line,
				Domain:    m.domainInfo["domain"].(string),
				RiskLevel: model.RiskLow,
			}
			m.domainUsers = append(m.domainUsers, user)
		}
	}

	return nil
}

func (m *DomainModule) collectDomainGroups() error {
	m.domainGroups = []map[string]interface{}{}

	cmd := exec.Command("net", "group", "/domain")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(output), "\n")
	inGroupList := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "-----------") || strings.Contains(line, "Group name") {
			inGroupList = true
			continue
		}
		if inGroupList && line != "" && !strings.HasPrefix(line, "The") && !strings.Contains(line, "command completed") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				group := map[string]interface{}{
					"name":       parts[0],
					"comment":    "",
					"members":    []string{},
					"risk_level": model.RiskLow,
				}
				if len(parts) > 1 {
					group["comment"] = strings.Join(parts[1:], " ")
				}
				m.domainGroups = append(m.domainGroups, group)
			}
		}
	}

	return nil
}

func (m *DomainModule) collectOUStructure() error {
	m.ouStruct = []map[string]interface{}{}

	cmd := exec.Command("dsquery", "ou")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "dsquery") {
			continue
		}
		ou := map[string]interface{}{
			"dn":         line,
			"name":       extractCN(line),
			"risk_level": model.RiskLow,
		}
		m.ouStruct = append(m.ouStruct, ou)
	}

	return nil
}

func (m *DomainModule) collectGPOList() error {
	m.gpoList = []map[string]interface{}{}

	cmd := exec.Command("gpresult", "/R")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Group Policy Objects") {
			continue
		}
		if strings.Contains(line, "{") && strings.Contains(line, "}") {
			gpo := map[string]interface{}{
				"guid":       extractGUID(line),
				"name":       strings.TrimSpace(line),
				"risk_level": model.RiskLow,
			}
			m.gpoList = append(m.gpoList, gpo)
		}
	}

	return nil
}

func (m *DomainModule) collectTrustRelations() error {
	m.trustRelations = []map[string]interface{}{}

	cmd := exec.Command("nltest", "/domain_trusts", "/all_trusts")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "List of") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) >= 2 {
			trust := map[string]interface{}{
				"domain":     strings.TrimSpace(parts[0]),
				"trust_dir":  "",
				"trust_type": "",
				"risk_level": model.RiskMedium,
			}
			if len(parts) >= 3 {
				trust["trust_dir"] = strings.TrimSpace(parts[1])
				trust["trust_type"] = strings.TrimSpace(parts[2])
			}
			m.trustRelations = append(m.trustRelations, trust)
		}
	}

	return nil
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

func extractGUID(s string) string {
	start := strings.Index(s, "{")
	end := strings.Index(s, "}")
	if start != -1 && end != -1 && end > start {
		return s[start : end+1]
	}
	return ""
}

func (m *DomainModule) Stop() error {
	return nil
}

func (m *DomainModule) GetData() ([]map[string]interface{}, error) {
	result := []map[string]interface{}{}

	result = append(result, map[string]interface{}{
		"type":   "domain_info",
		"domain": m.domainInfo["domain"],
		"role":   m.domainInfo["domain_role"],
		"joined": m.domainInfo["is_domain_joined"],
	})

	for _, u := range m.domainUsers {
		result = append(result, map[string]interface{}{
			"type":       "domain_user",
			"name":       u.Name,
			"full_name":  u.FullName,
			"domain":     u.Domain,
			"sid":        u.SID,
			"status":     u.Status,
			"last_logon": u.LastLogon.Format(time.RFC3339),
			"risk_level": u.RiskLevel,
		})
	}

	for _, g := range m.domainGroups {
		result = append(result, map[string]interface{}{
			"type":       "domain_group",
			"name":       g["name"],
			"comment":    g["comment"],
			"members":    g["members"],
			"risk_level": g["risk_level"],
		})
	}

	for _, ou := range m.ouStruct {
		result = append(result, map[string]interface{}{
			"type":       "organizational_unit",
			"dn":         ou["dn"],
			"name":       ou["name"],
			"risk_level": ou["risk_level"],
		})
	}

	for _, gpo := range m.gpoList {
		result = append(result, map[string]interface{}{
			"type":       "gpo",
			"guid":       gpo["guid"],
			"name":       gpo["name"],
			"risk_level": gpo["risk_level"],
		})
	}

	for _, t := range m.trustRelations {
		result = append(result, map[string]interface{}{
			"type":       "trust_relation",
			"domain":     t["domain"],
			"trust_dir":  t["trust_dir"],
			"trust_type": t["trust_type"],
			"risk_level": t["risk_level"],
		})
	}

	return result, nil
}

func (m *DomainModule) GetDomainInfo() map[string]interface{} {
	return m.domainInfo
}

func (m *DomainModule) GetDomainUsers() []model.AccountDTO {
	return m.domainUsers
}

func (m *DomainModule) CollectGroupMembers() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	for _, g := range m.domainGroups {
		groupName, ok := g["name"].(string)
		if !ok {
			continue
		}

		cmd := exec.Command("net", "group", groupName, "/domain")
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		lines := strings.Split(string(output), "\n")
		inMemberSection := false

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "---") {
				inMemberSection = true
				continue
			}
			if inMemberSection && line != "" && !strings.HasPrefix(line, "The") && !strings.HasPrefix(line, "Group") && !strings.HasPrefix(line, "Comment") && !strings.HasPrefix(line, "Members") && !strings.HasPrefix(line, "command completed") {
				results = append(results, map[string]interface{}{
					"group_name": groupName,
					"member":     line,
					"type":       "group_member",
				})
			}
		}
	}

	return results, nil
}

func (m *DomainModule) GetGroupDescription(groupName string) (string, error) {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$ErrorActionPreference='SilentlyContinue'
$group = Get-ADGroup -Identity '%s' -Properties Description -ErrorAction SilentlyContinue
if($group) { Write-Output $group.Description } else { Write-Output '' }`, groupName))

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get group description: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func (m *DomainModule) DetectOfflineMode() (map[string]interface{}, error) {
	result := map[string]interface{}{
		"is_offline":      false,
		"domain_joined":   m.domainInfo["is_domain_joined"],
		"domain_name":     m.domainInfo["domain"],
		"check_timestamp": time.Now().Format(time.RFC3339),
	}

	cmd := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
$dc = Get-ADDomainController -Discover -DomainName (Get-ADDomain).Name -ErrorAction SilentlyContinue
if($dc) {
    $status = Test-Connection -ComputerName $dc.HostName -Count 1 -Quiet -ErrorAction SilentlyContinue
    if($status) { Write-Output 'Online' } else { Write-Output 'Offline' }
} else {
    Write-Output 'NoDC'
}`)

	output, err := cmd.Output()
	if err == nil {
		status := strings.TrimSpace(string(output))
		result["is_offline"] = status == "Offline" || status == "NoDC"
		result["dc_status"] = status
	}

	cmd2 := exec.Command("nltest", "/dsgetdc:.")
	output2, err := cmd2.Output()
	if err == nil {
		result["dc_info"] = strings.TrimSpace(string(output2))
	}

	return result, nil
}

func (m *DomainModule) Search(keyword string) ([]map[string]interface{}, error) {
	results := []map[string]interface{}{}
	keywordLower := strings.ToLower(keyword)

	for _, u := range m.domainUsers {
		if strings.Contains(strings.ToLower(u.Name), keywordLower) ||
			strings.Contains(strings.ToLower(u.FullName), keywordLower) {
			results = append(results, map[string]interface{}{
				"type": "domain_user",
				"name": u.Name,
			})
		}
	}

	for _, g := range m.domainGroups {
		if name, ok := g["name"].(string); ok {
			if strings.Contains(strings.ToLower(name), keywordLower) {
				results = append(results, map[string]interface{}{
					"type": "domain_group",
					"name": name,
				})
			}
		}
	}

	for _, ou := range m.ouStruct {
		if dn, ok := ou["dn"].(string); ok {
			if strings.Contains(strings.ToLower(dn), keywordLower) {
				results = append(results, map[string]interface{}{
					"type": "organizational_unit",
					"dn":   dn,
				})
			}
		}
	}

	return results, nil
}
