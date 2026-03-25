package circuit

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type State int32

const (
	StateClosed   State = 0
	StateOpen     State = 1
	StateHalfOpen State = 2
)

type CircuitBreaker struct {
	name             string
	failureThreshold int
	resetTimeout     time.Duration
	state            atomic.Int32
	failures         atomic.Int32
	lastFailure      int64
	lastFailureMu    sync.Mutex
}

func New(name string, failureThreshold int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
	}
}

func (cb *CircuitBreaker) State() State {
	return State(cb.state.Load())
}

func (cb *CircuitBreaker) onFailure() {
	cb.failures.Add(1)

	cb.lastFailureMu.Lock()
	cb.lastFailure = time.Now().UnixNano()
	cb.lastFailureMu.Unlock()

	if cb.failures.Load() >= int32(cb.failureThreshold) {
		cb.state.Store(int32(StateOpen))
	}
}

func (cb *CircuitBreaker) onSuccess() {
	cb.failures.Store(0)
	cb.state.Store(int32(StateClosed))
}

func (cb *CircuitBreaker) Execute(ctx context.Context, op func() error) error {
	state := cb.State()

	switch state {
	case StateOpen:
		cb.lastFailureMu.Lock()
		lastFailure := cb.lastFailure
		cb.lastFailureMu.Unlock()

		elapsed := time.Since(time.Unix(0, lastFailure))

		if elapsed < cb.resetTimeout {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return ErrCircuitOpen
			}
		}

		cb.state.CompareAndSwap(int32(StateOpen), int32(StateHalfOpen))

	case StateHalfOpen:
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	err := op()
	if err != nil {
		cb.onFailure()
		return err
	}

	if cb.State() == StateHalfOpen {
		cb.onSuccess()
	}

	return nil
}

var ErrCircuitOpen = &CircuitError{Name: "circuit open"}

type CircuitError struct {
	Name string
}

func (e *CircuitError) Error() string {
	return e.Name + ": circuit is open"
}
