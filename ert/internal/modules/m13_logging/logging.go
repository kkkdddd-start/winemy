package m13_logging

import (
	"context"
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
	m.logs = []EventLogEntry{
		{
			EventID:     4624,
			EventType:   "Security",
			Level:       "Information",
			Source:      "Microsoft-Windows-Security-Auditing",
			Channel:     "Security",
			Computer:    "DESKTOP-ABC123",
			TimeCreated: time.Now().Add(-10 * time.Minute),
			RawXML:      "<Event><EventID>4624</EventID></Event>",
			Message:     "An account was successfully logged on.",
			RiskLevel:   model.RiskLow,
		},
		{
			EventID:     4625,
			EventType:   "Security",
			Level:       "Warning",
			Source:      "Microsoft-Windows-Security-Auditing",
			Channel:     "Security",
			Computer:    "DESKTOP-ABC123",
			TimeCreated: time.Now().Add(-30 * time.Minute),
			RawXML:      "<Event><EventID>4625</EventID></Event>",
			Message:     "An account failed to log on.",
			RiskLevel:   model.RiskMedium,
		},
		{
			EventID:     1000,
			EventType:   "Application",
			Level:       "Error",
			Source:      "Application Error",
			Channel:     "Application",
			Computer:    "DESKTOP-ABC123",
			TimeCreated: time.Now().Add(-1 * time.Hour),
			RawXML:      "<Event><EventID>1000</EventID></Event>",
			Message:     "Application error: crash detected",
			RiskLevel:   model.RiskMedium,
		},
		{
			EventID:     7036,
			EventType:   "System",
			Level:       "Information",
			Source:      "Service Control Manager",
			Channel:     "System",
			Computer:    "DESKTOP-ABC123",
			TimeCreated: time.Now().Add(-2 * time.Hour),
			RawXML:      "<Event><EventID>7036</EventID></Event>",
			Message:     "The service entered the running state.",
			RiskLevel:   model.RiskLow,
		},
		{
			EventID:     1,
			EventType:   "Custom",
			Level:       "Critical",
			Source:      "SuspiciousActivity",
			Channel:     "Custom",
			Computer:    "DESKTOP-ABC123",
			TimeCreated: time.Now().Add(-5 * time.Minute),
			RawXML:      "<Event><EventID>1</EventID></Event>",
			Message:     "Suspicious process execution detected",
			RiskLevel:   model.RiskCritical,
		},
	}

	return nil
}

func (m *LoggingModule) Stop() error {
	return nil
}

func (m *LoggingModule) GetData() ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, len(m.logs))
	for _, l := range m.logs {
		result = append(result, map[string]interface{}{
			"event_id":     l.EventID,
			"event_type":   l.EventType,
			"level":        l.Level,
			"source":       l.Source,
			"channel":      l.Channel,
			"computer":     l.Computer,
			"time_created": l.TimeCreated.Format(time.RFC3339),
			"raw_xml":      l.RawXML,
			"message":      l.Message,
			"risk_level":   l.RiskLevel,
		})
	}
	return result, nil
}
