package vault

import (
	"context"
	"fmt"
)

// PagedLister is a SecretReader that also supports listing keys at a path.
type PagedLister interface {
	SecretReader
	ListSecrets(ctx context.Context, mount, path string) ([]string, error)
}

// PaginatedClient wraps a PagedLister and fetches secrets in fixed-size pages,
// invoking a callback for each page of resolved secret maps.
type PaginatedClient struct {
	inner    PagedLister
	pageSize int
}

// NewPaginatedClient creates a PaginatedClient with the given page size.
// If pageSize <= 0 it defaults to 20.
func NewPaginatedClient(inner PagedLister, pageSize int) *PaginatedClient {
	if pageSize <= 0 {
		pageSize = 20
	}
	return &PaginatedClient{inner: inner, pageSize: pageSize}
}

// EachPage lists all keys under mount/path, splits them into pages of
// pageSize, reads each secret, and calls fn with the resulting map.
// Iteration stops early if fn returns a non-nil error.
func (p *PaginatedClient) EachPage(
	ctx context.Context,
	mount, path string,
	fn func(page map[string]map[string]interface{}) error,
) error {
	keys, err := p.inner.ListSecrets(ctx, mount, path)
	if err != nil {
		return fmt.Errorf("paginate list %s/%s: %w", mount, path, err)
	}

	for start := 0; start < len(keys); start += p.pageSize {
		end := start + p.pageSize
		if end > len(keys) {
			end = len(keys)
		}
		chunk := keys[start:end]

		page := make(map[string]map[string]interface{}, len(chunk))
		for _, key := range chunk {
			secretPath := key
			if path != "" {
				secretPath = path + "/" + key
			}
			data, err := p.inner.ReadSecret(ctx, mount, secretPath)
			if err != nil {
				return fmt.Errorf("paginate read %s/%s: %w", mount, secretPath, err)
			}
			page[key] = data
		}

		if err := fn(page); err != nil {
			return err
		}
	}
	return nil
}
