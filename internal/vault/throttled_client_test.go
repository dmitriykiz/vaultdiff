package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func newThrottleTestServer(t *testing.T, delay time.Duration) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"key":"val"}}`))
	}))
}

func TestThrottledClient_PassesThroughResult(t *testing.T) {
	srv := newThrottleTestServer(t, 0)
	defer srv.Close()

	inner := newTestClient(t, srv.URL)
	client := NewThrottledClient(inner, 5, 0)

	data, err := client.ReadSecret(context.Background(), "secret/foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["key"] != "val" {
		t.Errorf("expected val, got %v", data["key"])
	}
}

func TestThrottledClient_LimitsInFlight(t *testing.T) {
	const maxInFlight = 3
	var peak int64
	var mu sync.Mutex
	var current int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := atomic.AddInt64(&current, 1)
		mu.Lock()
		if v > atomic.LoadInt64(&peak) {
			atomic.StoreInt64(&peak, v)
		}
		mu.Unlock()
		time.Sleep(20 * time.Millisecond)
		atomic.AddInt64(&current, -1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"k":"v"}}`))
	}))
	defer srv.Close()

	inner := newTestClient(t, srv.URL)
	client := NewThrottledClient(inner, maxInFlight, 2*time.Second)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = client.ReadSecret(context.Background(), "secret/x")
		}()
	}
	wg.Wait()

	if p := atomic.LoadInt64(&peak); p > int64(maxInFlight) {
		t.Errorf("peak in-flight %d exceeded cap %d", p, maxInFlight)
	}
}

func TestThrottledClient_TimeoutOnFullSlots(t *testing.T) {
	srv := newThrottleTestServer(t, 200*time.Millisecond)
	defer srv.Close()

	inner := newTestClient(t, srv.URL)
	client := NewThrottledClient(inner, 1, 10*time.Millisecond)

	// Fill the single slot.
	go func() { _, _ = client.ReadSecret(context.Background(), "secret/slow") }()
	time.Sleep(5 * time.Millisecond)

	_, err := client.ReadSecret(context.Background(), "secret/blocked")
	if err == nil {
		t.Fatal("expected throttle timeout error, got nil")
	}
}

func TestThrottledClient_DefaultCap(t *testing.T) {
	inner := newTestClient(t, "http://localhost")
	client := NewThrottledClient(inner, 0, 0)
	if client.Cap() != 10 {
		t.Errorf("expected default cap 10, got %d", client.Cap())
	}
}
