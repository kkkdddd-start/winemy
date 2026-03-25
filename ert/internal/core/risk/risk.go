//go:build windows

package risk

import (
	"regexp"
	"strings"

	"github.com/yourname/ert/internal/model"
)

type RiskEngine struct {
	ruleGroups map[string]*RuleGroup
}

type RuleGroup struct {
	Name    string
	Rules   []*RiskRule
	Enabled bool
}

type RiskRule struct {
	ID        string
	Name      string
	Type      string
	Pattern   string
	Score     int
	Regex     *regexp.Regexp
	RiskLevel model.RiskLevel
}

func NewRiskEngine() *RiskEngine {
	return &RiskEngine{
		ruleGroups: make(map[string]*RuleGroup),
	}
}

func (re *RiskEngine) RegisterGroup(group *RuleGroup) {
	if group.Enabled {
		for _, rule := range group.Rules {
			if rule.Pattern != "" {
				rule.Regex = regexp.MustCompile(rule.Pattern)
			}
		}
	}
	re.ruleGroups[group.Name] = group
}

func (re *RiskEngine) EvaluateProcess(proc *model.ProcessDTO) model.RiskLevel {
	score := 0

	for _, group := range re.ruleGroups {
		if !group.Enabled {
			continue
		}

		for _, rule := range group.Rules {
			switch rule.Type {
			case "path":
				if re.matchPattern(proc.Path, rule.Regex) {
					score += rule.Score
				}
			case "name":
				if re.matchPattern(proc.Name, rule.Regex) {
					score += rule.Score
				}
			case "commandline":
				if re.matchPattern(proc.CommandLine, rule.Regex) {
					score += rule.Score
				}
			case "user":
				if re.matchPattern(proc.User, rule.Regex) {
					score += rule.Score
				}
			}
		}
	}

	return re.scoreToLevel(score)
}

func (re *RiskEngine) EvaluateNetwork(conn *model.NetworkConnDTO) model.RiskLevel {
	score := 0

	for _, group := range re.ruleGroups {
		if !group.Enabled {
			continue
		}

		for _, rule := range group.Rules {
			switch rule.Type {
			case "remote_addr":
				if re.matchPattern(conn.RemoteAddr, rule.Regex) {
					score += rule.Score
				}
			case "remote_port":
				if re.matchPort(conn.RemotePort, rule.Regex) {
					score += rule.Score
				}
			}
		}
	}

	return re.scoreToLevel(score)
}

func (re *RiskEngine) EvaluateRegistry(key *model.RegistryKeyDTO) model.RiskLevel {
	score := 0

	for _, group := range re.ruleGroups {
		if !group.Enabled {
			continue
		}

		for _, rule := range group.Rules {
			switch rule.Type {
			case "path":
				if re.matchPattern(key.Path, rule.Regex) {
					score += rule.Score
				}
			case "value":
				if re.matchPattern(key.Value, rule.Regex) {
					score += rule.Score
				}
			}
		}
	}

	return re.scoreToLevel(score)
}

func (re *RiskEngine) scoreToLevel(score int) model.RiskLevel {
	switch {
	case score >= 70:
		return model.RiskCritical
	case score >= 40:
		return model.RiskHigh
	case score >= 20:
		return model.RiskMedium
	default:
		return model.RiskLow
	}
}

func (re *RiskEngine) matchPattern(s string, regex *regexp.Regexp) bool {
	if regex == nil || s == "" {
		return false
	}
	return regex.MatchString(s)
}

func (re *RiskEngine) matchPort(port uint16, regex *regexp.Regexp) bool {
	if regex == nil {
		return false
	}
	return regex.MatchString(strings.TrimSpace(strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ToLower(regex.String()),
			" ", ""),
		"(", ""),
	))
}

func DefaultRiskRules() []*RuleGroup {
	return []*RuleGroup{
		{
			Name:    "suspicious_process",
			Enabled: true,
			Rules: []*RiskRule{
				{Type: "name", Pattern: `(?i)(mimikatz|pwdump|procdump|lsass)`, Score: 50, RiskLevel: model.RiskCritical},
				{Type: "name", Pattern: `(?i)(cmd|powershell).*\.tmp`, Score: 30, RiskLevel: model.RiskHigh},
				{Type: "path", Pattern: `(?i)(temp|appdata).*`, Score: 20, RiskLevel: model.RiskMedium},
				{Type: "commandline", Pattern: `(?i)(-enc|-encodedcommand)`, Score: 40, RiskLevel: model.RiskHigh},
			},
		},
		{
			Name:    "suspicious_network",
			Enabled: true,
			Rules: []*RiskRule{
				{Type: "remote_port", Pattern: `(4444|5555|6666|7777|8888|9999)`, Score: 50, RiskLevel: model.RiskCritical},
				{Type: "remote_addr", Pattern: `(?i)(tunnel|tor|onion)`, Score: 40, RiskLevel: model.RiskHigh},
			},
		},
		{
			Name:    "persistence",
			Enabled: true,
			Rules: []*RiskRule{
				{Type: "path", Pattern: `(?i)(run|runonce|autorun)`, Score: 30, RiskLevel: model.RiskHigh},
				{Type: "path", Pattern: `(?i)(schtasks|wmi)`, Score: 30, RiskLevel: model.RiskHigh},
			},
		},
	}
}
