package vault

import (
	"context"
	"fmt"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// SecretReader defines the interface for reading secrets from Vault.
type SecretReader interface {
	ReadSecret(ctx context.Context, path string) (map[string]string, error)
}

// Client wraps the Vault API client.
type Client struct {
	logical *vaultapi.Logical
}

// ReadSecret reads a secret at the given path and returns its key-value data.
// It supports both KV v1 and KV v2 secret engines.
func (c *Client) ReadSecret(ctx context.Context, path string) (map[string]string, error) {
	secret, err := c.logical.ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("reading secret at %q: %w", path, err)
	}
	if secret == nil {
		return nil, fmt.Errorf("secret not found at path %q", path)
	}

	data, err := extractData(path, secret)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(data))
	for k, v := range data {
		str, ok := v.(string)
		if !ok {
			str = fmt.Sprintf("%v", v)
		}
		result[k] = str
	}
	return result, nil
}

// extractData handles both KV v1 (data at top-level) and KV v2 (data nested under "data").
func extractData(path string, secret *vaultapi.Secret) (map[string]interface{}, error) {
	if secret.Data == nil {
		return nil, fmt.Errorf("secret at %q has no data", path)
	}

	// KV v2 paths typically contain "/data/" segment; check for nested data map.
	if strings.Contains(path, "/data/") {
		nested, ok := secret.Data["data"]
		if ok && nested != nil {
			if nestedMap, ok := nested.(map[string]interface{}); ok {
				return nestedMap, nil
			}
		}
	}

	return secret.Data, nil
}
