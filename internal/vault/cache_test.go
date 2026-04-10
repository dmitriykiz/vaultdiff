package vault

import (
	"testing"
	"time"
)

func TestSecretCache_SetAndGet(t *testing.T) {
	c := NewSecretCache(0)
	data := map[string]interface{}{"key": "value"}
	c.Set("secret/foo", data)

	got := c.Get("secret/foo")
	if got == nil {
		t.Fatal("expected cached data, got nil")
	}
	if got["key"] != "value" {
		t.Errorf("expected 'value', got %v", got["key"])
	}
}

func TestSecretCache_MissReturnsNil(t *testing.T) {
	c := NewSecretCache(0)
	if got := c.Get("secret/missing"); got != nil {
		t.Errorf("expected nil for missing key, got %v", got)
	}
}

func TestSecretCache_Expiry(t *testing.T) {
	c := NewSecretCache(10 * time.Millisecond)
	c.Set("secret/ttl", map[string]interface{}{"x": 1})

	time.Sleep(20 * time.Millisecond)

	if got := c.Get("secret/ttl"); got != nil {
		t.Error("expected expired entry to return nil")
	}
}

func TestSecretCache_NoExpiryWhenTTLZero(t *testing.T) {
	c := NewSecretCache(0)
	c.Set("secret/persistent", map[string]interface{}{"a": "b"})

	time.Sleep(5 * time.Millisecond)

	if got := c.Get("secret/persistent"); got == nil {
		t.Error("expected entry to persist with zero TTL")
	}
}

func TestSecretCache_Invalidate(t *testing.T) {
	c := NewSecretCache(0)
	c.Set("secret/a", map[string]interface{}{"k": "v"})
	c.Invalidate("secret/a")

	if got := c.Get("secret/a"); got != nil {
		t.Error("expected nil after invalidation")
	}
}

func TestSecretCache_Flush(t *testing.T) {
	c := NewSecretCache(0)
	c.Set("secret/a", map[string]interface{}{})
	c.Set("secret/b", map[string]interface{}{})
	c.Flush()

	if c.Len() != 0 {
		t.Errorf("expected 0 entries after flush, got %d", c.Len())
	}
}

func TestSecretCache_Len(t *testing.T) {
	c := NewSecretCache(0)
	if c.Len() != 0 {
		t.Errorf("expected 0, got %d", c.Len())
	}
	c.Set("a", map[string]interface{}{})
	c.Set("b", map[string]interface{}{})
	if c.Len() != 2 {
		t.Errorf("expected 2, got %d", c.Len())
	}
}

func TestCacheEntry_IsExpired(t *testing.T) {
	entry := &CacheEntry{
		FetchedAt: time.Now().Add(-1 * time.Second),
		TTL:       500 * time.Millisecond,
	}
	if !entry.IsExpired() {
		t.Error("expected entry to be expired")
	}
}

func TestCacheEntry_NotExpiredWithZeroTTL(t *testing.T) {
	entry := &CacheEntry{
		FetchedAt: time.Now().Add(-10 * time.Second),
		TTL:       0,
	}
	if entry.IsExpired() {
		t.Error("expected entry with zero TTL to never expire")
	}
}
