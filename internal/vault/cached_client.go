package vault

import (
	"context"
	"fmt"
	"time"
)

// CachedClient wraps a Vault client with an in-memory secret cache.
type CachedClient struct {
	client *Client
	cache  *SecretCache
}

// NewCachedClient creates a CachedClient using the provided Client and TTL.
func NewCachedClient(client *Client, ttl time.Duration) *CachedClient {
	return &CachedClient{
		client: client,
		cache:  NewSecretCache(ttl),
	}
}

// ReadSecret returns secret data for the given path, using the cache when
// available. On a cache miss the underlying client is called and the result
// is stored before returning.
func (cc *CachedClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	if cached := cc.cache.Get(path); cached != nil {
		return cached, nil
	}

	engineType, err := DetectEngineType(cc.client, MountFromPath(path))
	if err != nil {
		return nil, fmt.Errorf("detect engine for %q: %w", path, err)
	}

	data, err := cc.client.ReadSecret(ctx, path, engineType)
	if err != nil {
		return nil, err
	}

	cc.cache.Set(path, data)
	return data, nil
}

// InvalidatePath removes a single path from the underlying cache.
func (cc *CachedClient) InvalidatePath(path string) {
	cc.cache.Invalidate(path)
}

// FlushCache clears all cached entries.
func (cc *CachedClient) FlushCache() {
	cc.cache.Flush()
}

// CacheLen returns the number of entries currently in the cache.
func (cc *CachedClient) CacheLen() int {
	return cc.cache.Len()
}
