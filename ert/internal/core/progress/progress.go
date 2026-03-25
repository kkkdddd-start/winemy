package progress

import (
	"sync"
	"time"
)

type Progress struct {
	ModuleID   int       `json:"module_id"`
	Current    int       `json:"current"`
	Total      int       `json:"total"`
	Percentage float64   `json:"percentage"`
	Message    string    `json:"message"`
	ETA        time.Time `json:"eta"`
	StartedAt  time.Time `json:"started_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ProgressManager struct {
	subscribers map[int]chan *Progress
	mu          sync.RWMutex
}

func New() *ProgressManager {
	return &ProgressManager{
		subscribers: make(map[int]chan *Progress),
	}
}

func (pm *ProgressManager) Subscribe(moduleID int) chan *Progress {
	ch := make(chan *Progress, 100)
	pm.mu.Lock()
	pm.subscribers[moduleID] = ch
	pm.mu.Unlock()
	return ch
}

func (pm *ProgressManager) Unsubscribe(moduleID int) {
	pm.mu.Lock()
	if ch, ok := pm.subscribers[moduleID]; ok {
		close(ch)
		delete(pm.subscribers, moduleID)
	}
	pm.mu.Unlock()
}

func (pm *ProgressManager) Update(progress *Progress) {
	progress.UpdatedAt = time.Now()
	progress.Percentage = progress.Calculate()
	progress.ETA = progress.EstimateETA()

	pm.mu.RLock()
	if ch, ok := pm.subscribers[progress.ModuleID]; ok {
		select {
		case ch <- progress:
		default:
		}
	}
	pm.mu.RUnlock()
}

func (p *Progress) Calculate() float64 {
	if p.Total == 0 {
		return 0
	}
	return float64(p.Current) / float64(p.Total) * 100
}

func (p *Progress) EstimateETA() time.Time {
	if p.Current == 0 {
		return time.Time{}
	}

	elapsed := time.Since(p.StartedAt)
	rate := float64(p.Current) / elapsed.Seconds()
	remaining := float64(p.Total-p.Current) / rate

	return time.Now().Add(time.Duration(remaining) * time.Second)
}
