package vault

import (
	"context"
	"fmt"
	"strings"
)

// VersionedSecret wraps a secret map with its version metadata.
type VersionedSecret struct {
	Data    map[string]interface{}
	Version int
	Path    string
}

// VersionClient reads a specific version of a KVv2 secret.
type VersionClient struct {
	inner  SecretReader
	version int
}

// NewVersionClient returns a VersionClient that appends a version query to KVv2 paths.
// Pass version=0 to always fetch the latest.
func NewVersionClient(inner SecretReader, version int) *VersionClient {
	return &VersionClient{inner: inner, version: version}
}

// ReadSecret delegates to the inner client, rewriting the path to include the
// version parameter for KVv2 mounts (paths containing "/data/").
func (c *VersionClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	target := path
	if c.version > 0 && strings.Contains(path, "/data/") {
		target = fmt.Sprintf("%s?version=%d", strings.TrimRight(path, "/"), c.version)
	}
	return c.inner.ReadSecret(ctx, target)
}

// Version returns the pinned version number (0 means latest).
func (c *VersionClient) Version() int {
	return c.version
}

// WithVersion returns a new VersionClient with the given version pinned.
func (c *VersionClient) WithVersion(v int) *VersionClient {
	return NewVersionClient(c.inner, v)
}
