package vault

import (
	"context"
	"fmt"
	"strings"
)

// NamespacedClient wraps a SecretReader and injects a Vault namespace header
// into every request by prepending the namespace to the path and storing it
// for use by HTTP-level middleware. For path-based namespace routing the
// client prepends the namespace segment to every secret path.
type NamespacedClient struct {
	inner     SecretReader
	namespace string
}

// NewNamespacedClient returns a SecretReader that prepends namespace to every
// path passed to ReadSecret. namespace must be a non-empty string; leading and
// trailing slashes are trimmed automatically.
func NewNamespacedClient(inner SecretReader, namespace string) *NamespacedClient {
	return &NamespacedClient{
		inner:     inner,
		namespace: strings.Trim(namespace, "/"),
	}
}

// Namespace returns the configured namespace.
func (c *NamespacedClient) Namespace() string {
	return c.namespace
}

// qualifiedPath prepends the namespace to path, producing "namespace/path".
func (c *NamespacedClient) qualifiedPath(path string) (string, error) {
	if c.namespace == "" {
		return path, nil
	}
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return "", fmt.Errorf("namespacedclient: path must not be empty")
	}
	return c.namespace + "/" + path, nil
}

// ReadSecret implements SecretReader. It prepends the namespace to path before
// delegating to the inner client.
func (c *NamespacedClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	qualified, err := c.qualifiedPath(path)
	if err != nil {
		return nil, err
	}
	return c.inner.ReadSecret(ctx, qualified)
}
