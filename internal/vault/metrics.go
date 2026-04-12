package vault

import (
	"sync"
	"sync/atomic"
	"time"
)

// InMemoryMetrics is a thread-safe MetricsCollector that accumulates counters.
type InMemoryMetrics struct {
	totalRequests  atomic.Int64
	successCount   atomic.Int64
	errorCount     atomic.Int64
	latencyMu      sync.Mutex
	totalLatencyMs int64
}

// Record satisfies MetricsCollector.
func (m *InMemoryMetrics) Record(_ string, latency time.Duration, err error) {
	m.totalRequests.Add(1)
	if err != nil {
		m.errorCount.Add(1)
	} else {
		m.successCount.Add(1)
	}
	m.latencyMu.Lock()
	m.totalLatencyMs += latency.Milliseconds()
	m.latencyMu.Unlock()
}

// Snapshot returns a point-in-time copy of the collected metrics.
func (m *InMemoryMetrics) Snapshot() Metrics {
	m.latencyMu.Lock()
	latency := m.totalLatencyMs
	m.latencyMu.Unlock()
	return Metrics{
		TotalRequests:  m.totalRequests.Load(),
		SuccessCount:   m.successCount.Load(),
		ErrorCount:     m.errorCount.Load(),
		TotalLatencyMs: latency,
	}
}

// Reset clears all counters.
func (m *InMemoryMetrics) Reset() {
	m.totalRequests.Store(0)
	m.successCount.Store(0)
	m.errorCount.Store(0)
	m.latencyMu.Lock()
	m.totalLatencyMs = 0
	m.latencyMu.Unlock()
}
