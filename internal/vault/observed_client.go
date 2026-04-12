package vault

import (
	"context"
	"time"
)

// Metrics holds counters and timing data collected by ObservedClient.
type Metrics struct {
	TotalRequests  int64
	SuccessCount   int64
	ErrorCount     int64
	TotalLatencyMs int64
}

// MetricsCollector receives per-request observations.
type MetricsCollector interface {
	Record(path string, latency time.Duration, err error)
}

// ObservedClient wraps a SecretReader and records request metrics.
type ObservedClient struct {
	inner     SecretReader
	collector MetricsCollector
}

// NewObservedClient returns a SecretReader that records metrics via collector.
// If collector is nil, observations are silently discarded.
func NewObservedClient(inner SecretReader, collector MetricsCollector) *ObservedClient {
	return &ObservedClient{inner: inner, collector: collector}
}

// ReadSecret reads the secret at path and records latency and outcome.
func (o *ObservedClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	start := time.Now()
	data, err := o.inner.ReadSecret(ctx, path)
	latency := time.Since(start)
	if o.collector != nil {
		o.collector.Record(path, latency, err)
	}
	return data, err
}
