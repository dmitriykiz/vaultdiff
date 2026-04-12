package vault

import (
	"errors"
	"testing"
	"time"
)

func TestInMemoryMetrics_InitialState(t *testing.T) {
	m := &InMemoryMetrics{}
	snap := m.Snapshot()
	if snap.TotalRequests != 0 || snap.SuccessCount != 0 || snap.ErrorCount != 0 {
		t.Errorf("expected all zero, got %+v", snap)
	}
}

func TestInMemoryMetrics_RecordSuccess(t *testing.T) {
	m := &InMemoryMetrics{}
	m.Record("secret/a", 5*time.Millisecond, nil)

	snap := m.Snapshot()
	if snap.TotalRequests != 1 {
		t.Errorf("expected TotalRequests=1, got %d", snap.TotalRequests)
	}
	if snap.SuccessCount != 1 {
		t.Errorf("expected SuccessCount=1, got %d", snap.SuccessCount)
	}
	if snap.ErrorCount != 0 {
		t.Errorf("expected ErrorCount=0, got %d", snap.ErrorCount)
	}
	if snap.TotalLatencyMs != 5 {
		t.Errorf("expected TotalLatencyMs=5, got %d", snap.TotalLatencyMs)
	}
}

func TestInMemoryMetrics_RecordError(t *testing.T) {
	m := &InMemoryMetrics{}
	m.Record("secret/b", 2*time.Millisecond, errors.New("fail"))

	snap := m.Snapshot()
	if snap.ErrorCount != 1 {
		t.Errorf("expected ErrorCount=1, got %d", snap.ErrorCount)
	}
	if snap.SuccessCount != 0 {
		t.Errorf("expected SuccessCount=0, got %d", snap.SuccessCount)
	}
}

func TestInMemoryMetrics_Reset(t *testing.T) {
	m := &InMemoryMetrics{}
	m.Record("secret/a", 10*time.Millisecond, nil)
	m.Record("secret/b", 5*time.Millisecond, errors.New("err"))
	m.Reset()

	snap := m.Snapshot()
	if snap.TotalRequests != 0 || snap.SuccessCount != 0 || snap.ErrorCount != 0 || snap.TotalLatencyMs != 0 {
		t.Errorf("expected all zero after Reset, got %+v", snap)
	}
}

func TestInMemoryMetrics_AccumulatesMultiple(t *testing.T) {
	m := &InMemoryMetrics{}
	for i := 0; i < 5; i++ {
		m.Record("secret/x", time.Millisecond, nil)
	}
	m.Record("secret/y", time.Millisecond, errors.New("oops"))

	snap := m.Snapshot()
	if snap.TotalRequests != 6 {
		t.Errorf("expected 6 total, got %d", snap.TotalRequests)
	}
	if snap.SuccessCount != 5 {
		t.Errorf("expected 5 successes, got %d", snap.SuccessCount)
	}
	if snap.ErrorCount != 1 {
		t.Errorf("expected 1 error, got %d", snap.ErrorCount)
	}
}
