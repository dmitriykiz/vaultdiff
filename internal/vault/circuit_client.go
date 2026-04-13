package vault

import (
	"context"
	"errors"
)

// CircuitBreakerClient wraps a SecretReader and short-circuits requests when
// the underlying service appears to be unavailable.
type CircuitBreakerClient struct {
	inner   SecretReader
	breaker *CircuitBreaker
}

// NewCircuitBreakerClient returns a SecretReader that applies the supplied
// CircuitBreaker to every read. If the breaker is nil a default one is used.
func NewCircuitBreakerClient(inner SecretReader, cb *CircuitBreaker) *CircuitBreakerClient {
	if cb == nil {
		cb = NewCircuitBreaker(5, 0)
	}
	return &CircuitBreakerClient{inner: inner, breaker: cb}
}

// ReadSecret checks the circuit state before delegating to the inner client.
// A successful read resets the failure counter; any error increments it.
func (c *CircuitBreakerClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	if err := c.breaker.Allow(); err != nil {
		return nil, err
	}

	data, err := c.inner.ReadSecret(ctx, path)
	if err != nil {
		// Do not penalise the circuit for context cancellation.
		if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			c.breaker.RecordFailure()
		}
		return nil, err
	}

	c.breaker.RecordSuccess()
	return data, nil
}

// Breaker exposes the underlying CircuitBreaker for inspection.
func (c *CircuitBreakerClient) Breaker() *CircuitBreaker {
	return c.breaker
}
