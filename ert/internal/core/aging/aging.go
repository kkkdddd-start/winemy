package aging

import (
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

type Task struct {
	ID        string
	Priority  int
	WaitCount int
	CreatedAt time.Time
	Handler   func() error
}

type AgingController struct {
	maxWaitTime    time.Duration
	priorityBoost  int
	checkInterval  time.Duration
	taskWaitCounts map[string]int
	mu             sync.RWMutex
	stopCh         chan struct{}
}

func NewAgingController(maxWaitTime time.Duration, boostThreshold, priorityBoost int, checkInterval time.Duration) *AgingController {
	return &AgingController{
		maxWaitTime:    maxWaitTime,
		priorityBoost:  priorityBoost,
		checkInterval:  checkInterval,
		taskWaitCounts: make(map[string]int),
		stopCh:         make(chan struct{}),
	}
}

func (ac *AgingController) Start() {
	go ac.run()
}

func (ac *AgingController) Stop() {
	close(ac.stopCh)
}

func (ac *AgingController) RecordWait(taskID string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.taskWaitCounts[taskID]++
}

func (ac *AgingController) GetPriority(taskID string, currentPriority int) int {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	waitCount := ac.taskWaitCounts[taskID]
	if waitCount > 10 {
		newPriority := currentPriority - ac.priorityBoost
		if newPriority < 1 {
			newPriority = 1
		}
		return newPriority
	}

	return currentPriority
}

func (ac *AgingController) CheckMaxWait(task *Task) bool {
	if time.Since(task.CreatedAt) > ac.maxWaitTime {
		return true
	}
	return false
}

func (ac *AgingController) run() {
	ticker := time.NewTicker(ac.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ac.stopCh:
			return
		case <-ticker.C:
			ac.cleanup()
		}
	}
}

func (ac *AgingController) cleanup() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	for id, count := range ac.taskWaitCounts {
		if count > 100 {
			delete(ac.taskWaitCounts, id)
		}
	}
}

type AgingScheduler struct {
	controller *AgingController
	eg         *errgroup.Group
}

func NewAgingScheduler(controller *AgingController) *AgingScheduler {
	return &AgingScheduler{
		controller: controller,
		eg:         &errgroup.Group{},
	}
}

func (s *AgingScheduler) Schedule(tasks []*Task) error {
	for _, task := range tasks {
		s.controller.RecordWait(task.ID)
		task.Priority = s.controller.GetPriority(task.ID, task.Priority)
	}

	return nil
}
