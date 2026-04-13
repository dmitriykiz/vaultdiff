package vault

import (
	"errors"
	"sync"
	"time"
)

// ErrCircuitOpen is returned when the circuit breaker is in the open state.
var ErrCircuitOpen = errors.New("circuit breaker is open")

type circuitState int

const (
	stateClosed circuitState = iota
	stateOpen
	stateHalfOpen
)

// CircuitBreaker tracks failures and opens the circuit after a threshold.
type CircuitBreaker struct {
	mu           sync.Mutex
	state        circuitState
	failures     int
	threshold    int
	resetTimeout time.Duration
	openedAt     time.Time
	nowFn        func() time.Time
}

// NewCircuitBreaker creates a CircuitBreaker with the given failure threshold
// and reset timeout. After threshold consecutive failures the circuit opens;
// after resetTimeout it moves to half-open and allows one probe request.
func NewCircuitBreaker(threshold int, resetTimeout time.Duration) *CircuitBreaker {
	if threshold <= 0 {
		threshold = 5
	}
	if resetTimeout <= 0 {
		resetTimeout = 30 * time.Second
	}
	return &CircuitBreaker{
		threshold:    threshold,
		resetTimeout: resetTimeout,
		nowFn:        time.Now,
	}
}

// Allow reports whether a request should be allowed through.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case stateClosed:
		return nil
	case stateHalfOpen:
		return nil
	case stateOpen:
		if cb.nowFn().Sub(cb.openedAt) >= cb.resetTimeout {
			cb.state = stateHalfOpen
			return nil
		}
		return ErrCircuitOpen
	}
	return nil
}

// RecordSuccess resets the failure counter and closes the circuit.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = stateClosed
}

// RecordFailure increments the failure counter and may open the circuit.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	if cb.state == stateHalfOpen || cb.failures >= cb.threshold {
		cb.state = stateOpen
		cb.openedAt = cb.nowFn()
	}
}

// State returns the current circuit state as a string (for diagnostics).
func (cb *CircuitBreaker) State() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case stateClosed:
		return "closed"
	case stateOpen:
		return "open"
	case stateHalfOpen:
		return "half-open"
	}
	return "unknown"
}
