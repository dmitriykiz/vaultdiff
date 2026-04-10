package vault

import (
	"sync"
	"time"
)

// CacheEntry holds a cached secret value with an expiry timestamp.
type CacheEntry struct {
	Data      map[string]interface{}
	FetchedAt time.Time
	TTL       time.Duration
}

// IsExpired reports whether the cache entry has passed its TTL.
func (e *CacheEntry) IsExpired() bool {
	if e.TTL <= 0 {
		return false
	}
	return time.Since(e.FetchedAt) > e.TTL
}

// SecretCache is a thread-safe in-memory cache for Vault secret data.
type SecretCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	ttl     time.Duration
}

// NewSecretCache creates a new SecretCache with the given TTL.
// A TTL of zero disables expiry.
func NewSecretCache(ttl time.Duration) *SecretCache {
	return &SecretCache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
	}
}

// Get retrieves a cached entry by path. Returns nil if absent or expired.
func (c *SecretCache) Get(path string) map[string]interface{} {
	c.mu.RLock()
	entry, ok := c.entries[path]
	c.mu.RUnlock()
	if !ok || entry.IsExpired() {
		return nil
	}
	return entry.Data
}

// Set stores secret data under the given path.
func (c *SecretCache) Set(path string, data map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[path] = &CacheEntry{
		Data:      data,
		FetchedAt: time.Now(),
		TTL:       c.ttl,
	}
}

// Invalidate removes a single path from the cache.
func (c *SecretCache) Invalidate(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, path)
}

// Flush clears all entries from the cache.
func (c *SecretCache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*CacheEntry)
}

// Len returns the number of entries currently held in the cache.
func (c *SecretCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}
