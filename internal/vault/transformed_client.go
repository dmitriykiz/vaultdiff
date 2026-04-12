package vault

import "context"

// TransformedClient wraps a SecretReader and applies a Transformer to every
// secret it reads before returning it to the caller.
type TransformedClient struct {
	inner       SecretReader
	transformer *Transformer
}

// NewTransformedClient returns a TransformedClient wrapping inner.
func NewTransformedClient(inner SecretReader, transformer *Transformer) *TransformedClient {
	return &TransformedClient{
		inner:       inner,
		transformer: transformer,
	}
}

// ReadSecret reads from the underlying client and applies all configured
// transformation rules to the returned data map.
func (c *TransformedClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	data, err := c.inner.ReadSecret(ctx, path)
	if err != nil {
		return nil, err
	}
	if c.transformer == nil {
		return data, nil
	}
	return c.transformer.Apply(data), nil
}
