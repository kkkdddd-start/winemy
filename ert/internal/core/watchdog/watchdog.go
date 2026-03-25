package watchdog

import (
	"context"
	"sync"
	"time"
)

type Watchdog struct {
	timeout   time.Duration
	onTimeout func(taskID string)
	tasks     map[string]*watchTask
	mu        sync.RWMutex
	stopCh    chan struct{}
}

type watchTask struct {
	ID        string
	StartTime time.Time
	Timeout   time.Duration
	TaskCtx   context.Context
	Cancel    context.CancelFunc
}

func NewWatchdog(timeout time.Duration, onTimeout func(string)) *Watchdog {
	return &Watchdog{
		timeout:   timeout,
		onTimeout: onTimeout,
		tasks:     make(map[string]*watchTask),
		stopCh:    make(chan struct{}),
	}
}

func (w *Watchdog) Start() {
	go w.run()
}

func (w *Watchdog) Stop() {
	close(w.stopCh)
}

func (w *Watchdog) Watch(taskID string, ctx context.Context) (context.Context, context.CancelFunc) {
	w.mu.Lock()
	defer w.mu.Unlock()

	taskCtx, cancel := context.WithTimeout(ctx, w.timeout)
	w.tasks[taskID] = &watchTask{
		ID:        taskID,
		StartTime: time.Now(),
		Timeout:   w.timeout,
		TaskCtx:   taskCtx,
		Cancel:    cancel,
	}

	return taskCtx, func() {
		w.mu.Lock()
		delete(w.tasks, taskID)
		w.mu.Unlock()
		cancel()
	}
}

func (w *Watchdog) run() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.checkTimeouts()
		}
	}
}

func (w *Watchdog) checkTimeouts() {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	for id, task := range w.tasks {
		if now.Sub(task.StartTime) > task.Timeout {
			if w.onTimeout != nil {
				w.onTimeout(id)
			}
		}
	}
}
