package vault

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// ThrottledClient wraps a SecretReader and enforces a maximum number of
// in-flight requests at any given time.
type ThrottledClient struct {
	inner    SecretReader
	sem      chan struct{}
	timeout  time.Duration
	inflight int64
}

// NewThrottledClient creates a ThrottledClient that allows at most maxInFlight
// concurrent reads. If timeout > 0 the client will wait at most that long to
// acquire a slot before returning an error.
func NewThrottledClient(inner SecretReader, maxInFlight int, timeout time.Duration) *ThrottledClient {
	if maxInFlight <= 0 {
		maxInFlight = 10
	}
	return &ThrottledClient{
		inner:   inner,
		sem:     make(chan struct{}, maxInFlight),
		timeout: timeout,
	}
}

// ReadSecret acquires a concurrency slot, delegates to the inner client, then
// releases the slot.
func (t *ThrottledClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	waitCtx := ctx
	var cancel context.CancelFunc
	if t.timeout > 0 {
		waitCtx, cancel = context.WithTimeout(ctx, t.timeout)
		defer cancel()
	}

	select {
	case t.sem <- struct{}{}:
		// slot acquired
	case <-waitCtx.Done():
		return nil, fmt.Errorf("throttled: could not acquire slot: %w", waitCtx.Err())
	}

	atomic.AddInt64(&t.inflight, 1)
	defer func() {
		<-t.sem
		atomic.AddInt64(&t.inflight, -1)
	}()

	return t.inner.ReadSecret(ctx, path)
}

// InFlight returns the current number of in-flight requests.
func (t *ThrottledClient) InFlight() int {
	return int(atomic.LoadInt64(&t.inflight))
}

// Cap returns the maximum number of concurrent requests allowed.
func (t *ThrottledClient) Cap() int {
	return cap(t.sem)
}
