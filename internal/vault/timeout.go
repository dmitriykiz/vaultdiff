package vault

import (
	"context"
	"fmt"
	"time"
)

// TimeoutClient wraps a SecretReader and enforces a per-request deadline.
type TimeoutClient struct {
	inner   SecretReader
	timeout time.Duration
}

// NewTimeoutClient returns a TimeoutClient that cancels any ReadSecret call
// that exceeds the given timeout. A zero or negative timeout disables the
// per-request deadline and delegates directly to inner.
func NewTimeoutClient(inner SecretReader, timeout time.Duration) *TimeoutClient {
	return &TimeoutClient{inner: inner, timeout: timeout}
}

// ReadSecret reads the secret at path, cancelling the request if the
// configured timeout elapses before the underlying client responds.
func (t *TimeoutClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	if t.timeout <= 0 {
		return t.inner.ReadSecret(ctx, path)
	}

	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	type result struct {
		data map[string]interface{}
		err  error
	}

	ch := make(chan result, 1)
	go func() {
		data, err := t.inner.ReadSecret(ctx, path)
		ch <- result{data, err}
	}()

	select {
	case res := <-ch:
		return res.data, res.err
	case <-ctx.Done():
		return nil, fmt.Errorf("vault: read %q exceeded timeout of %s: %w", path, t.timeout, ctx.Err())
	}
}
