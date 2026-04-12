package vault

import (
	"context"
)

// pagedVaultClient adapts the existing vault Client so it satisfies PagedLister,
// bridging ListSecrets (from list.go) and ReadSecret (from secret.go) into a
// single interface consumed by PaginatedClient.
type pagedVaultClient struct {
	client *Client
	engine EngineType
}

// NewPagedVaultClient wraps a *Client and detected engine type to produce a
// PagedLister ready for use with PaginatedClient.
func NewPagedVaultClient(c *Client, engine EngineType) PagedLister {
	return &pagedVaultClient{client: c, engine: engine}
}

func (p *pagedVaultClient) ReadSecret(ctx context.Context, mount, path string) (map[string]interface{}, error) {
	return p.client.ReadSecret(ctx, mount, path, p.engine)
}

func (p *pagedVaultClient) ListSecrets(ctx context.Context, mount, path string) ([]string, error) {
	return p.client.ListSecrets(ctx, mount, path, p.engine)
}
