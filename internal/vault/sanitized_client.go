package vault

import "context"

// SanitizedClient wraps a SecretReader and applies sanitization rules
// to all returned secret data before it reaches the caller.
type SanitizedClient struct {
	inner     SecretReader
	sanitizer *Sanitizer
}

// NewSanitizedClient creates a SanitizedClient wrapping inner with the
// provided Sanitizer. If sanitizer is nil, data is passed through unchanged.
func NewSanitizedClient(inner SecretReader, sanitizer *Sanitizer) *SanitizedClient {
	return &SanitizedClient{inner: inner, sanitizer: sanitizer}
}

// ReadSecret reads from the underlying client and sanitizes the result.
func (c *SanitizedClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	data, err := c.inner.ReadSecret(ctx, path)
	if err != nil {
		return nil, err
	}
	if c.sanitizer == nil {
		return data, nil
	}
	return c.sanitizer.Apply(data), nil
}
