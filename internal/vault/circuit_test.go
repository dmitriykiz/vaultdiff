package vault

import (
	"testing"
	"time"
)

func TestCircuitBreaker_InitialStateClosed(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Second)
	if got := cb.State(); got != "closed" {
		t.Fatalf("expected closed, got %s", got)
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Second)
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	if got := cb.State(); got != "open" {
		t.Fatalf("expected open after threshold, got %s", got)
	}
	if err := cb.Allow(); err != ErrCircuitOpen {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_SuccessResetsClosed(t *testing.T) {
	cb := NewCircuitBreaker(2, time.Second)
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess()
	if got := cb.State(); got != "closed" {
		t.Fatalf("expected closed after success, got %s", got)
	}
	if err := cb.Allow(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCircuitBreaker_TransitionsToHalfOpen(t *testing.T) {
	now := time.Now()
	cb := NewCircuitBreaker(1, 10*time.Millisecond)
	cb.nowFn = func() time.Time { return now }
	cb.RecordFailure()
	if cb.State() != "open" {
		t.Fatal("expected open")
	}
	// Advance time past the reset timeout.
	cb.nowFn = func() time.Time { return now.Add(20 * time.Millisecond) }
	if err := cb.Allow(); err != nil {
		t.Fatalf("expected half-open allow, got %v", err)
	}
	if got := cb.State(); got != "half-open" {
		t.Fatalf("expected half-open, got %s", got)
	}
}

func TestCircuitBreaker_HalfOpenFailureReopens(t *testing.T) {
	now := time.Now()
	cb := NewCircuitBreaker(1, 10*time.Millisecond)
	cb.nowFn = func() time.Time { return now }
	cb.RecordFailure()
	cb.nowFn = func() time.Time { return now.Add(20 * time.Millisecond) }
	_ = cb.Allow() // transitions to half-open
	cb.RecordFailure()
	if got := cb.State(); got != "open" {
		t.Fatalf("expected open after half-open failure, got %s", got)
	}
}

func TestCircuitBreaker_DefaultThresholdAndTimeout(t *testing.T) {
	cb := NewCircuitBreaker(0, 0)
	if cb.threshold != 5 {
		t.Errorf("expected default threshold 5, got %d", cb.threshold)
	}
	if cb.resetTimeout != 30*time.Second {
		t.Errorf("expected default reset 30s, got %v", cb.resetTimeout)
	}
}
