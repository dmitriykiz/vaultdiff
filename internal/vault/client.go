package vault

import (
	"fmt"
	"os"

	vaultapi "github.com/hashicorp/vault/api"
)

// Client wraps the Vault API client with helper methods for secret retrieval.
type Client struct {
	vc *vaultapi.Client
}

// NewClient creates a new Vault client using environment variables or provided address.
// It respects VAULT_ADDR and VAULT_TOKEN environment variables.
func NewClient(address string) (*Client, error) {
	cfg := vaultapi.DefaultConfig()

	if address != "" {
		cfg.Address = address
	} else if addr := os.Getenv("VAULT_ADDR"); addr != "" {
		cfg.Address = addr
	}

	if err := cfg.ReadEnvironment(); err != nil {
		return nil, fmt.Errorf("reading vault environment: %w", err)
	}

	vc, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault client: %w", err)
	}

	if token := os.Getenv("VAULT_TOKEN"); token != "" {
		vc.SetToken(token)
	}

	return &Client{vc: vc}, nil
}

// ReadSecret reads a KV secret at the given path and returns its data as a
// map of string keys to string values. Supports both KV v1 and v2 paths.
func (c *Client) ReadSecret(path string) (map[string]string, error) {
	secret, err := c.vc.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("reading secret at %q: %w", path, err)
	}
	if secret == nil {
		return nil, fmt.Errorf("no secret found at path %q", path)
	}

	// KV v2 wraps data under a "data" key.
	raw := secret.Data
	if nested, ok := secret.Data["data"]; ok {
		if nestedMap, ok := nested.(map[string]interface{}); ok {
			raw = nestedMap
		}
	}

	result := make(map[string]string, len(raw))
	for k, v := range raw {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result, nil
}
