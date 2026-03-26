//go:build windows

package m13_logging

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
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

func (m *LoggingModule) CollectDNSLogs() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	cmd := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
Get-WinEvent -LogName 'Microsoft-Windows-DNS Client Events/Operational' -MaxEvents 100 -ErrorAction SilentlyContinue | ForEach-Object {
    Write-Output ($_.TimeCreated.ToString('yyyy-MM-dd HH:mm:ss') + "|" + $_.Id.ToString() + "|" + $_.LevelDisplayName + "|" + $_.Message.Substring(0, [Math]::Min(200, $_.Message.Length)))
}`)

	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.Split(line, "|")
			if len(parts) >= 4 {
				results = append(results, map[string]interface{}{
					"timestamp": parts[0],
					"event_id":  parts[1],
					"level":     parts[2],
					"message":   parts[3],
					"type":      "dns_event",
				})
			}
		}
	}

	cmd2 := exec.Command("powershell", "-Command",
		`$ErrorActionPreference='SilentlyContinue'
Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Services\EventLog\DNS Client Events' -ErrorAction SilentlyContinue | Select-Object -ExpandProperty File`)

	output2, err := cmd2.Output()
	if err == nil {
		dnsLogPath := strings.TrimSpace(string(output2))
		if dnsLogPath != "" {
			cmd3 := exec.Command("cmd", "/c", fmt.Sprintf("type \"%s\" 2>nul | findstr /i \"query request\" | findstr /v \"Cache\"", dnsLogPath))
			output3, _ := cmd3.Output()
			lines := strings.Split(string(output3), "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					results = append(results, map[string]interface{}{
						"message": strings.TrimSpace(line),
						"type":    "dns_query",
					})
				}
			}
		}
	}

	return results, nil
}

func (m *LoggingModule) FilterByLevel(level string) []EventLogEntry {
	var results []EventLogEntry
	for _, l := range m.logs {
		if strings.ToLower(l.Level) == strings.ToLower(level) {
			results = append(results, l)
		}
	}
	return results
}

func (m *LoggingModule) FilterBySource(source string) []EventLogEntry {
	var results []EventLogEntry
	sourceLower := strings.ToLower(source)
	for _, l := range m.logs {
		if strings.Contains(strings.ToLower(l.Source), sourceLower) {
			results = append(results, l)
		}
	}
	return results
}

func (m *LoggingModule) FilterByEventID(eventID int) []EventLogEntry {
	var results []EventLogEntry
	for _, l := range m.logs {
		if l.EventID == eventID {
			results = append(results, l)
		}
	}
	return results
}

func (m *LoggingModule) Search(keyword string) []EventLogEntry {
	var results []EventLogEntry
	keywordLower := strings.ToLower(keyword)
	for _, l := range m.logs {
		if strings.Contains(strings.ToLower(l.Message), keywordLower) ||
			strings.Contains(strings.ToLower(l.Source), keywordLower) ||
			strings.Contains(strconv.Itoa(l.EventID), keywordLower) {
			results = append(results, l)
		}
	}
	return results
}

func (m *LoggingModule) ParseEVTX(filePath string) ([]EventLogEntry, error) {
	var results []EventLogEntry

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`$ErrorActionPreference='SilentlyContinue'
$events = Get-WinEvent -Path '%s' -ErrorAction SilentlyContinue
foreach($e in $events) {
    [xml]$xml = $e.ToXml()
    $eventData = $xml.Event.EventData.Data
    $dataStr = ''
    foreach($d in $eventData) {
        $dataStr += $d.'#text' + ' '
    }
    Write-Output ($e.TimeCreated.ToString('yyyy-MM-ddTHH:mm:ssZ') + "|" + $e.Id.ToString() + "|" + $e.LevelDisplayName + "|" + $e.ProviderName + "|" + $e.LogName + "|" + $e.MachineName + "|" + $dataStr.Substring(0, [Math]::Min(500, $dataStr.Length)))
}`, filePath))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to parse EVTX file: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 7)
		if len(parts) >= 7 {
			t, _ := time.Parse("2006-01-02T15:04:05Z", parts[0])
			eventID, _ := strconv.Atoi(parts[1])
			results = append(results, EventLogEntry{
				EventID:     eventID,
				EventType:   parts[4],
				Level:       parts[2],
				Source:      parts[3],
				Channel:     parts[4],
				Computer:    parts[5],
				TimeCreated: t,
				Message:     parts[6],
				RawXML:      line,
			})
		}
	}

	return results, nil
}

func (m *LoggingModule) ExportHTML(filePath string) error {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>ERT Event Log Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #333; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #4CAF50; color: white; }
        .risk-high { color: red; font-weight: bold; }
        .risk-medium { color: orange; }
        .risk-low { color: green; }
        .filter-bar { margin: 20px 0; }
    </style>
</head>
<body>
    <h1>ERT Event Log Report</h1>
    <p>Generated: ` + time.Now().Format(time.RFC3339) + `</p>
    <p>Total Events: ` + fmt.Sprintf("%d", len(m.logs)) + `</p>
    <table>
        <tr>
            <th>Timestamp</th>
            <th>Event ID</th>
            <th>Level</th>
            <th>Source</th>
            <th>Channel</th>
            <th>Message</th>
            <th>Risk</th>
        </tr>`

	for _, l := range m.logs {
		riskClass := "risk-low"
		if l.RiskLevel == model.RiskHigh || l.RiskLevel == model.RiskCritical {
			riskClass = "risk-high"
		} else if l.RiskLevel == model.RiskMedium {
			riskClass = "risk-medium"
		}

		message := l.Message
		if len(message) > 100 {
			message = message[:100] + "..."
		}

		html += fmt.Sprintf(`        <tr>
            <td>%s</td>
            <td>%d</td>
            <td>%s</td>
            <td>%s</td>
            <td>%s</td>
            <td>%s</td>
            <td class="%s">%d</td>
        </tr>`, l.TimeCreated.Format(time.RFC3339), l.EventID, l.Level, l.Source, l.Channel, message, riskClass, l.RiskLevel)
	}

	html += `    </table>
</body>
</html>`

	return os.WriteFile(filePath, []byte(html), 0644)
}

func (m *LoggingModule) ExportJSON(filePath string) error {
	data := map[string]interface{}{
		"timestamp":    time.Now().Format(time.RFC3339),
		"total_events": len(m.logs),
		"logs":         m.logs,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(filePath, jsonData, 0644)
}
