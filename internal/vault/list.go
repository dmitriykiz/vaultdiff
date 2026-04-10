package vault

import (
	"context"
	"fmt"
	"strings"
)

// ListSecrets returns the keys available under the given path for a KV v1 or
// KV v2 mount. For KV v2 the metadata sub-path is used automatically.
func (c *Client) ListSecrets(ctx context.Context, secretPath string, engine EngineType) ([]string, error) {
	listPath := buildListPath(secretPath, engine)

	secret, err := c.logical.ListWithContext(ctx, listPath)
	if err != nil {
		return nil, fmt.Errorf("vault: list %q: %w", listPath, err)
	}
	if secret == nil || secret.Data == nil {
		return []string{}, nil
	}

	raw, ok := secret.Data["keys"]
	if !ok {
		return []string{}, nil
	}

	iface, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("vault: unexpected keys type for path %q", listPath)
	}

	keys := make([]string, 0, len(iface))
	for _, v := range iface {
		s, ok := v.(string)
		if !ok {
			continue
		}
		keys = append(keys, s)
	}
	return keys, nil
}

// buildListPath returns the correct Vault API path for a LIST operation
// depending on the engine version.
func buildListPath(secretPath string, engine EngineType) string {
	secretPath = strings.TrimPrefix(secretPath, "/")
	if engine != EngineKVv2 {
		return secretPath
	}
	// KV v2: insert "metadata" after the mount segment.
	parts := strings.SplitN(secretPath, "/", 2)
	if len(parts) == 1 {
		return fmt.Sprintf("%s/metadata", parts[0])
	}
	return fmt.Sprintf("%s/metadata/%s", parts[0], parts[1])
}
