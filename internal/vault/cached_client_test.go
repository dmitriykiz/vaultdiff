package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func newCacheTestServer(t *testing.T, calls *atomic.Int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/sys/internal/ui/mounts/secret/foo":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"options": map[string]interface{}{"version": "1"}},
			})
		case "/v1/secret/foo":
			calls.Add(1)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"password": "s3cr3t"},
			})
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestCachedClient_CacheHitAvoidsDuplicateRequest(t *testing.T) {
	var calls atomic.Int32
	srv := newCacheTestServer(t, &calls)
	defer srv.Close()

	client, err := newTestClient(srv.URL)
	if err != nil {
		t.Fatalf("newTestClient: %v", err)
	}
	cc := NewCachedClient(client, time.Minute)

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		data, err := cc.ReadSecret(ctx, "secret/foo")
		if err != nil {
			t.Fatalf("ReadSecret call %d: %v", i, err)
		}
		if data["password"] != "s3cr3t" {
			t.Errorf("unexpected data: %v", data)
		}
	}

	if calls.Load() != 1 {
		t.Errorf("expected 1 upstream call, got %d", calls.Load())
	}
}

func TestCachedClient_InvalidateForcesFetch(t *testing.T) {
	var calls atomic.Int32
	srv := newCacheTestServer(t, &calls)
	defer srv.Close()

	client, err := newTestClient(srv.URL)
	if err != nil {
		t.Fatalf("newTestClient: %v", err)
	}
	cc := NewCachedClient(client, time.Minute)
	ctx := context.Background()

	if _, err := cc.ReadSecret(ctx, "secret/foo"); err != nil {
		t.Fatalf("first read: %v", err)
	}
	cc.InvalidatePath("secret/foo")
	if _, err := cc.ReadSecret(ctx, "secret/foo"); err != nil {
		t.Fatalf("second read: %v", err)
	}

	if calls.Load() != 2 {
		t.Errorf("expected 2 upstream calls after invalidation, got %d", calls.Load())
	}
}

func TestCachedClient_FlushCache(t *testing.T) {
	var calls atomic.Int32
	srv := newCacheTestServer(t, &calls)
	defer srv.Close()

	client, err := newTestClient(srv.URL)
	if err != nil {
		t.Fatalf("newTestClient: %v", err)
	}
	cc := NewCachedClient(client, time.Minute)
	ctx := context.Background()

	cc.ReadSecret(ctx, "secret/foo") //nolint:errcheck
	if cc.CacheLen() != 1 {
		t.Errorf("expected 1 cached entry, got %d", cc.CacheLen())
	}
	cc.FlushCache()
	if cc.CacheLen() != 0 {
		t.Errorf("expected 0 after flush, got %d", cc.CacheLen())
	}
}
