package vault

import (
	"context"
	"fmt"
	"time"
)

// HealthStatus represents the result of a Vault health probe.
type HealthStatus struct {
	Healthy     bool
	Latency     time.Duration
	Error       error
	CheckedAt   time.Time
}

// HealthChecker probes a SecretReader to verify connectivity.
type HealthChecker struct {
	client  SecretReader
	probePath string
}

// NewHealthChecker creates a HealthChecker that probes the given path.
// If probePath is empty it defaults to "sys/health".
func NewHealthChecker(client SecretReader, probePath string) *HealthChecker {
	if probePath == "" {
		probePath = "sys/health"
	}
	return &HealthChecker{client: client, probePath: probePath}
}

// Check performs a single health probe and returns a HealthStatus.
func (h *HealthChecker) Check(ctx context.Context) HealthStatus {
	start := time.Now()
	status := HealthStatus{CheckedAt: start}

	_, err := h.client.ReadSecret(ctx, h.probePath)
	status.Latency = time.Since(start)

	if err != nil {
		status.Healthy = false
		status.Error = fmt.Errorf("health probe failed for %q: %w", h.probePath, err)
		return status
	}

	status.Healthy = true
	return status
}

// String returns a human-readable summary of the health status.
func (s HealthStatus) String() string {
	if s.Healthy {
		return fmt.Sprintf("healthy (latency=%s, checked_at=%s)", s.Latency.Round(time.Millisecond), s.CheckedAt.Format(time.RFC3339))
	}
	return fmt.Sprintf("unhealthy (latency=%s, error=%v, checked_at=%s)", s.Latency.Round(time.Millisecond), s.Error, s.CheckedAt.Format(time.RFC3339))
}
