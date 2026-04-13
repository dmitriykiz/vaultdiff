package vault

import (
	"context"
	"errors"
	"fmt"
)

// ChainClient tries a sequence of SecretReaders in order, returning the first
// successful result. If all readers fail, a combined error is returned.
type ChainClient struct {
	readers []SecretReader
}

// NewChainClient returns a ChainClient that consults each reader in order.
// At least one reader must be provided.
func NewChainClient(readers ...SecretReader) (*ChainClient, error) {
	if len(readers) == 0 {
		return nil, fmt.Errorf("vault: NewChainClient requires at least one reader")
	}
	return &ChainClient{readers: readers}, nil
}

// ReadSecret walks the chain of readers. The first reader to return a nil
// error wins. If every reader fails, all errors are joined and returned.
func (c *ChainClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	errs := make([]error, 0, len(c.readers))
	for _, r := range c.readers {
		data, err := r.ReadSecret(ctx, path)
		if err == nil {
			return data, nil
		}
		errs = append(errs, err)
	}
	return nil, fmt.Errorf("vault: all readers failed: %w", errors.Join(errs...))
}

// Len returns the number of readers in the chain.
func (c *ChainClient) Len() int { return len(c.readers) }
