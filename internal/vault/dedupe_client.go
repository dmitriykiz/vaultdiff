package vault

import (
	"context"
	"sync"
)

// inflightCall represents a single in-flight or completed ReadSecret call.
type inflightCall struct {
	wg  sync.WaitGroup
	val map[string]interface{}
	err error
}

// DedupeClient wraps a SecretReader and collapses concurrent identical
// ReadSecret calls for the same path into a single upstream request.
// This is useful when multiple goroutines race to read the same secret.
type DedupeClient struct {
	inner  SecretReader
	mu     sync.Mutex
	calls  map[string]*inflightCall
}

// NewDedupeClient returns a DedupeClient backed by inner.
func NewDedupeClient(inner SecretReader) *DedupeClient {
	return &DedupeClient{
		inner: inner,
		calls: make(map[string]*inflightCall),
	}
}

// ReadSecret issues at most one upstream read per path for concurrent callers.
// All callers sharing the same in-flight request receive the same result.
func (d *DedupeClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	d.mu.Lock()
	if call, ok := d.calls[path]; ok {
		d.mu.Unlock()
		call.wg.Wait()
		return call.val, call.err
	}

	call := &inflightCall{}
	call.wg.Add(1)
	d.calls[path] = call
	d.mu.Unlock()

	call.val, call.err = d.inner.ReadSecret(ctx, path)
	call.wg.Done()

	d.mu.Lock()
	delete(d.calls, path)
	d.mu.Unlock()

	return call.val, call.err
}
