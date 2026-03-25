//go:build windows

package m13_logging

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/yourname/ert/internal/model"
	"github.com/yourname/ert/internal/registry"
)

type LoggingModule struct {
	ctx     context.Context
	storage registry.Storage
	logs    []EventLogEntry
}

type EventLogEntry struct {
	EventID     int
	EventType   string
	Level       string
	Source      string
	Channel     string
	Computer    string
	TimeCreated time.Time
	RawXML      string
	Message     string
	RiskLevel   model.RiskLevel
}

func New() *LoggingModule {
	return &LoggingModule{}
}

func (m *LoggingModule) ID() int       { return 13 }
func (m *LoggingModule) Name() string  { return "logging" }
func (m *LoggingModule) Priority() int { return 1 }

func (m *LoggingModule) Init(ctx context.Context, s registry.Storage) error {
	m.ctx = ctx
	m.storage = s
	return nil
}

func (m *LoggingModule) Collect(ctx context.Context) error {
	m.logs = []EventLogEntry{}

	logChannels := []string{"Security", "System", "Application"}

	for _, channel := range logChannels {
		m.queryEventLog(channel, 20)
	}

	return nil
}

func (m *LoggingModule) queryEventLog(channel string, maxEvents int) {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`Get-WinEvent -LogName '%s' -MaxEvents %d 2>$null | ForEach-Object { [xml]$_.ToXml() | Select-Object -ExpandProperty Event | ConvertTo-Json -Compress }`, channel, maxEvents))

	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "null" {
			continue
		}

		entry := parseEventJSON(line, channel)
		if entry != nil {
			m.logs = append(m.logs, *entry)
		}
	}
}

type winEvent struct {
	System struct {
		EventID  int `json:"EventID"`
		Provider struct {
			Name string `json:"Name"`
		} `json:"Provider"`
		Computer    string `json:"Computer"`
		Level       string `json:"Level"`
		TimeCreated struct {
			Attributes struct {
				SystemTime string `json:"SystemTime"`
			} `json:"@xsi.type"`
		} `json:"TimeCreated"`
	} `json:"System"`
	EventData struct {
		Data string `json:"Data"`
	} `json:"EventData"`
}

func parseEventJSON(jsonData, channel string) *EventLogEntry {
	if !strings.HasPrefix(jsonData, "{") && !strings.HasPrefix(jsonData, "[") {
		return nil
	}

	event := &winEvent{}
	if err := json.Unmarshal([]byte(jsonData), event); err != nil {
		return nil
	}

	level := "Unknown"
	switch event.System.Level {
	case "0":
		level = "Information"
	case "1":
		level = "Critical"
	case "2":
		level = "Error"
	case "3":
		level = "Warning"
	case "4":
		level = "Information"
	case "5":
		level = "Verbose"
	}

	timeCreated := time.Now()
	if event.System.TimeCreated.Attributes.SystemTime != "" {
		if parsed, err := time.Parse(time.RFC3339, event.System.TimeCreated.Attributes.SystemTime); err == nil {
			timeCreated = parsed
		}
	}

	eventType := channel
	riskLevel := model.RiskLow
	if isHighRiskEvent(event.System.EventID, channel) {
		riskLevel = model.RiskHigh
	} else if isMediumRiskEvent(event.System.EventID, channel) {
		riskLevel = model.RiskMedium
	}

	return &EventLogEntry{
		EventID:     event.System.EventID,
		EventType:   eventType,
		Level:       level,
		Source:      event.System.Provider.Name,
		Channel:     channel,
		Computer:    event.System.Computer,
		TimeCreated: timeCreated,
		RawXML:      jsonData,
		Message:     extractMessage(event.EventData.Data),
		RiskLevel:   riskLevel,
	}
}

func extractMessage(data string) string {
	if data == "" {
		return ""
	}
	if len(data) > 500 {
		return data[:500] + "..."
	}
	return data
}

func isHighRiskEvent(eventID int, channel string) bool {
	highRiskIDs := map[string][]int{
		"Security":    {4624, 4625, 4634, 4648, 4672, 4719, 4720, 4722, 4726, 4732, 4756, 4757},
		"System":      {7045, 7034, 7036},
		"Application": {1000, 1001, 1002},
	}

	ids, ok := highRiskIDs[channel]
	if !ok {
		return false
	}

	for _, id := range ids {
		if id == eventID {
			return true
		}
	}
	return false
}

func isMediumRiskEvent(eventID int, channel string) bool {
	mediumRiskIDs := map[string][]int{
		"Security":    {4728, 4729, 4730, 4733},
		"System":      {6000, 6001, 6002},
		"Application": {1003, 1004},
	}

	ids, ok := mediumRiskIDs[channel]
	if !ok {
		return false
	}

	for _, id := range ids {
		if id == eventID {
			return true
		}
	}
	return false
}

func (m *LoggingModule) Stop() error {
	return nil
}

func (m *LoggingModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.logs))
	for _, l := range m.logs {
		timeCreatedStr := ""
		if !l.TimeCreated.IsZero() {
			timeCreatedStr = l.TimeCreated.Format(time.RFC3339)
		}
		result = append(result, map[string]interface{}{
			"event_id":     l.EventID,
			"event_type":   l.EventType,
			"level":        l.Level,
			"source":       l.Source,
			"channel":      l.Channel,
			"computer":     l.Computer,
			"time_created": timeCreatedStr,
			"raw_xml":      l.RawXML,
			"message":      l.Message,
			"risk_level":   l.RiskLevel,
		})
	}
	return result, nil
}
