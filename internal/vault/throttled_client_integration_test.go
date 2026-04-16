package vault_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/your-org/vaultdiff/internal/vault"
)

// TestThrottledClient_Integration_WithObservedClient verifies that
// ThrottledClient composes correctly with ObservedClient.
func TestThrottledClient_Integration_WithObservedClient(t *testing.T) {
	var hits int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&hits, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"env":"prod"}}`))
	}))
	defer srv.Close()

	cfg := vault.DefaultRetryConfig()
	cfg.MaxAttempts = 1

	base, err := vault.NewClient(srv.URL, "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	metrics := vault.NewInMemoryMetrics()
	observed := vault.NewObservedClient(base, metrics)
	throttled := vault.NewThrottledClient(observed, 4, 3*time.Second)

	const requests = 20
	var wg sync.WaitGroup
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = throttled.ReadSecret(context.Background(), "secret/data/env")
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt64(&hits); got != requests {
		t.Errorf("expected %d server hits, got %d", requests, got)
	}
	if s := metrics.Successes(); s != requests {
		t.Errorf("expected %d successes recorded, got %d", requests, s)
	}
}
