//go:build windows

package timeline

import (
	"sort"
	"time"

	"github.com/yourname/ert/internal/model"
)

type TimelineEngine struct {
	events []model.TimelineEvent
}

func New() *TimelineEngine {
	return &TimelineEngine{
		events: make([]model.TimelineEvent, 0),
	}
}

func (t *TimelineEngine) AddEvent(event model.TimelineEvent) {
	t.events = append(t.events, event)
}

func (t *TimelineEngine) Build() []model.TimelineEvent {
	sort.Slice(t.events, func(i, j int) bool {
		return t.events[i].Timestamp.Before(t.events[j].Timestamp)
	})
	return t.events
}

func (t *TimelineEngine) GetByTimeRange(start, end time.Time) []model.TimelineEvent {
	events := make([]model.TimelineEvent, 0)
	for _, e := range t.events {
		if e.Timestamp.After(start) && e.Timestamp.Before(end) {
			events = append(events, e)
		}
	}
	return events
}

func (t *TimelineEngine) GetBySeverity(level model.RiskLevel) []model.TimelineEvent {
	events := make([]model.TimelineEvent, 0)
	for _, e := range t.events {
		if e.Severity == level {
			events = append(events, e)
		}
	}
	return events
}

func (t *TimelineEngine) GetByModule(moduleID int) []model.TimelineEvent {
	events := make([]model.TimelineEvent, 0)
	for _, e := range t.events {
		if e.ModuleID == moduleID {
			events = append(events, e)
		}
	}
	return events
}

func (t *TimelineEngine) Clear() {
	t.events = make([]model.TimelineEvent, 0)
}
