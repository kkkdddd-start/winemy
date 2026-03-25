package concurrency

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type Priority int

const (
	PriorityHigh   Priority = 0
	PriorityMedium Priority = 1
	PriorityLow    Priority = 2
)

type Task struct {
	ID        string
	Priority  Priority
	Handler   func(ctx context.Context) error
	CreatedAt time.Time
	Result    error
}

type SemaphorePool struct {
	high   *semaphore.Weighted
	medium *semaphore.Weighted
	low    *semaphore.Weighted
}

func NewSemaphorePool(highWorkers, mediumWorkers, lowWorkers int64) *SemaphorePool {
	return &SemaphorePool{
		high:   semaphore.NewWeighted(highWorkers),
		medium: semaphore.NewWeighted(mediumWorkers),
		low:    semaphore.NewWeighted(lowWorkers),
	}
}

func (p *SemaphorePool) Execute(ctx context.Context, task *Task) error {
	var s *semaphore.Weighted
	switch task.Priority {
	case PriorityHigh:
		s = p.high
	case PriorityMedium:
		s = p.medium
	default:
		s = p.low
	}

	if err := s.Acquire(ctx, 1); err != nil {
		return err
	}
	defer s.Release(1)

	return task.Handler(ctx)
}

func (p *SemaphorePool) ExecuteAndWait(ctx context.Context, tasks []*Task) error {
	eg, ctx := errgroup.WithContext(ctx)

	for _, task := range tasks {
		t := task
		eg.Go(func() error {
			return p.Execute(ctx, t)
		})
	}

	return eg.Wait()
}

type TaskQueue struct {
	mu      sync.Mutex
	tasks   map[string]*Task
	pending []*Task
}

func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		tasks:   make(map[string]*Task),
		pending: make([]*Task, 0),
	}
}

func (q *TaskQueue) Add(task *Task) {
	q.mu.Lock()
	defer q.mu.Unlock()

	task.CreatedAt = time.Now()
	q.tasks[task.ID] = task
	q.pending = append(q.pending, task)
}

func (q *TaskQueue) Get() *Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.pending) == 0 {
		return nil
	}

	task := q.pending[0]
	q.pending = q.pending[1:]
	return task
}

func (q *TaskQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.pending)
}
