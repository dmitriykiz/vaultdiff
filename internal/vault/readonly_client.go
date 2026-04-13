package vault

import (
	"context"
	"fmt"
)

// ReadOnlyClient wraps a SecretReader and enforces read-only access by
// rejecting any path that matches a protected prefix. It is intended to
// prevent accidental writes or access to sensitive mount roots during diff
// operations.
type ReadOnlyClient struct {
	inner      SecretReader
	protected  []string
}

// NewReadOnlyClient returns a ReadOnlyClient that delegates reads to inner
// but blocks access to any path that starts with a protected prefix.
// Passing a nil or empty protected list allows all paths.
func NewReadOnlyClient(inner SecretReader, protected []string) *ReadOnlyClient {
	guarded := make([]string, len(protected))
	copy(guarded, protected)
	return &ReadOnlyClient{inner: inner, protected: guarded}
}

// ReadSecret returns an error if path matches a protected prefix; otherwise
// it delegates to the underlying SecretReader.
func (c *ReadOnlyClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	if err := c.checkPath(path); err != nil {
		return nil, err
	}
	return c.inner.ReadSecret(ctx, path)
}

// ListSecrets returns an error if path matches a protected prefix; otherwise
// it delegates to the underlying SecretReader.
func (c *ReadOnlyClient) ListSecrets(ctx context.Context, path string) ([]string, error) {
	if err := c.checkPath(path); err != nil {
		return nil, err
	}
	type lister interface {
		ListSecrets(context.Context, string) ([]string, error)
	}
	if l, ok := c.inner.(lister); ok {
		return l.ListSecrets(ctx, path)
	}
	return nil, fmt.Errorf("readonly_client: inner client does not support ListSecrets")
}

// Protected returns a copy of the protected prefix list.
func (c *ReadOnlyClient) Protected() []string {
	out := make([]string, len(c.protected))
	copy(out, c.protected)
	return out
}

func (c *ReadOnlyClient) checkPath(path string) error {
	for _, p := range c.protected {
		if len(path) >= len(p) && path[:len(p)] == p {
			return fmt.Errorf("readonly_client: access to protected path %q denied", path)
		}
	}
	return nil
}
