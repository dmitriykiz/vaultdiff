package vault

import (
	"context"
	"fmt"
	"strings"
)

// ScopedClient wraps a SecretReader and restricts all reads to a given path prefix.
// Any path that does not begin with the configured prefix will return an error.
type ScopedClient struct {
	inner  SecretReader
	prefix string
}

// NewScopedClient creates a ScopedClient that enforces the given path prefix.
// The prefix is normalised: leading/trailing slashes are trimmed.
func NewScopedClient(inner SecretReader, prefix string) *ScopedClient {
	return &ScopedClient{
		inner:  inner,
		prefix: strings.Trim(prefix, "/"),
	}
}

// ReadSecret delegates to the inner client only when path is within the scope.
// If the path is outside the prefix, a descriptive error is returned immediately
// without making any network call.
func (s *ScopedClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	normPath := strings.Trim(path, "/")

	if s.prefix != "" && !strings.HasPrefix(normPath, s.prefix+"/") && normPath != s.prefix {
		return nil, fmt.Errorf("scoped client: path %q is outside allowed prefix %q", path, s.prefix)
	}

	return s.inner.ReadSecret(ctx, path)
}

// Prefix returns the configured path prefix for inspection / logging.
func (s *ScopedClient) Prefix() string {
	return s.prefix
}
