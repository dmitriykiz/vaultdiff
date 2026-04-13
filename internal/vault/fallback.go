package vault

import (
	"context"
	"errors"
)

// FallbackClient tries a primary SecretReader and, on error, delegates to a
// secondary SecretReader. Only errors that satisfy isFallbackable are forwarded
// to the secondary; all other errors are returned immediately.
type FallbackClient struct {
	primary   SecretReader
	secondary SecretReader
	should    func(error) bool
}

// NewFallbackClient returns a FallbackClient that consults secondary whenever
// the primary returns an error for which shouldFallback returns true.
// If shouldFallback is nil, every error triggers the fallback.
func NewFallbackClient(primary, secondary SecretReader, shouldFallback func(error) bool) *FallbackClient {
	if shouldFallback == nil {
		shouldFallback = func(error) bool { return true }
	}
	return &FallbackClient{
		primary:   primary,
		secondary: secondary,
		should:    shouldFallback,
	}
}

// ReadSecret attempts to read from the primary client. If the primary returns
// an error and shouldFallback(err) is true, the secondary client is tried.
// The secondary error, if any, is wrapped with context about both failures.
func (f *FallbackClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	data, err := f.primary.ReadSecret(ctx, path)
	if err == nil {
		return data, nil
	}
	if !f.should(err) {
		return nil, err
	}
	secData, secErr := f.secondary.ReadSecret(ctx, path)
	if secErr != nil {
		return nil, errors.Join(err, secErr)
	}
	return secData, nil
}
