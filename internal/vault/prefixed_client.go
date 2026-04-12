package vault

import (
	"context"
	"fmt"
	"strings"
)

// PrefixedClient wraps a SecretReader and automatically prepends a namespace
// prefix to every path before delegating to the inner client. This is useful
// when all operations should be scoped under a common mount or team prefix
// without callers needing to repeat it.
type PrefixedClient struct {
	inner  SecretReader
	prefix string
}

// NewPrefixedClient returns a PrefixedClient that prepends prefix to every
// path passed to ReadSecret. The prefix is normalised: leading/trailing
// slashes are trimmed so that the final path is always "prefix/path".
func NewPrefixedClient(inner SecretReader, prefix string) *PrefixedClient {
	return &PrefixedClient{
		inner:  inner,
		prefix: strings.Trim(prefix, "/"),
	}
}

// ReadSecret prepends the configured prefix to path and delegates to the
// inner SecretReader. If the prefix is empty the path is passed through
// unchanged.
func (p *PrefixedClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	effective := p.effectivePath(path)
	return p.inner.ReadSecret(ctx, effective)
}

// Prefix returns the normalised prefix string used by this client.
func (p *PrefixedClient) Prefix() string {
	return p.prefix
}

// effectivePath combines the prefix and the caller-supplied path.
func (p *PrefixedClient) effectivePath(path string) string {
	path = strings.Trim(path, "/")
	if p.prefix == "" {
		return path
	}
	if path == "" {
		return p.prefix
	}
	return fmt.Sprintf("%s/%s", p.prefix, path)
}
