package vault

import (
	"errors"
	"net/http"
	"time"
)

// RetryConfig holds configuration for retry behaviour on transient errors.
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

// DefaultRetryConfig returns a RetryConfig with sensible defaults.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   200 * time.Millisecond,
		MaxDelay:    2 * time.Second,
	}
}

// isRetryable reports whether the given HTTP status code warrants a retry.
func isRetryable(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	}
	return false
}

// ErrMaxRetriesExceeded is returned when all retry attempts are exhausted.
var ErrMaxRetriesExceeded = errors.New("max retries exceeded")

// WithRetry executes fn up to cfg.MaxAttempts times, backing off between
// attempts. fn should return (statusCode, error); a zero statusCode signals
// that no HTTP response was received.
func WithRetry(cfg RetryConfig, sleep func(time.Duration), fn func() (int, error)) error {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 1
	}

	delay := cfg.BaseDelay

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		statusCode, err := fn()
		if err == nil {
			return nil
		}

		// Non-retryable HTTP error or final attempt — give up.
		if statusCode != 0 && !isRetryable(statusCode) {
			return err
		}
		if attempt == cfg.MaxAttempts {
			return ErrMaxRetriesExceeded
		}

		sleep(delay)
		delay *= 2
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}
	}

	return ErrMaxRetriesExceeded
}
