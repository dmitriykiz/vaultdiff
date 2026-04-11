package vault

import (
	"context"
)

// RateLimitedClient wraps a SecretReader and enforces a per-second request
// limit using a RateLimiter. It satisfies the SecretReader interface so it
// can be composed transparently with other wrappers such as CachedClient.
type RateLimitedClient struct {
	inner   SecretReader
	limiter *RateLimiter
}

// NewRateLimitedClient returns a RateLimitedClient that delegates to inner
// after acquiring a rate-limit token. A rps value of zero disables limiting.
func NewRateLimitedClient(inner SecretReader, rps int) *RateLimitedClient {
	return &RateLimitedClient{
		inner:   inner,
		limiter: NewRateLimiter(rps),
	}
}

// ReadSecret blocks until a rate-limit token is available, then delegates to
// the underlying SecretReader.
func (r *RateLimitedClient) ReadSecret(ctx context.Context, path string) (map[string]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	r.limiter.Wait()
	return r.inner.ReadSecret(ctx, path)
}

// Stop releases resources held by the internal RateLimiter. It should be
// called when the client is no longer needed.
func (r *RateLimitedClient) Stop() {
	r.limiter.Stop()
}
