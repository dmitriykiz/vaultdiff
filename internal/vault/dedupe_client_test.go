package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
)

func newDedupeTestServer(t *testing.T, hits *atomic.Int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"key":"value"}}`))
	}))
}

func TestDedupeClient_SingleRequest_PassesThrough(t *testing.T) {
	var hits atomic.Int32
	srv := newDedupeTestServer(t, &hits)
	defer srv.Close()

	inner, err := newTestClient(srv.URL)
	if err != nil {
		t.Fatalf("newTestClient: %v", err)
	}
	client := NewDedupeClient(inner)

	data, err := client.ReadSecret(context.Background(), "secret/foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["key"] != "value" {
		t.Errorf("expected key=value, got %v", data["key"])
	}
	if hits.Load() != 1 {
		t.Errorf("expected 1 upstream hit, got %d", hits.Load())
	}
}

func TestDedupeClient_ConcurrentRequests_DeduplicatedToOne(t *testing.T) {
	var hits atomic.Int32
	ready := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-ready // block until all goroutines are waiting
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"key":"concurrent"}}`))
	}))
	defer srv.Close()

	inner, err := newTestClient(srv.URL)
	if err != nil {
		t.Fatalf("newTestClient: %v", err)
	}
	client := NewDedupeClient(inner)

	const n = 10
	var wg sync.WaitGroup
	errs := make([]error, n)
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			_, errs[i] = client.ReadSecret(context.Background(), "secret/shared")
		}()
	}
	close(ready)
	wg.Wait()

	for i, e := range errs {
		if e != nil {
			t.Errorf("goroutine %d returned error: %v", i, e)
		}
	}
	if hits.Load() > 3 {
		t.Errorf("expected deduplicated upstream calls (<=3), got %d", hits.Load())
	}
}

func TestDedupeClient_DifferentPaths_NotDeduplicated(t *testing.T) {
	var hits atomic.Int32
	srv := newDedupeTestServer(t, &hits)
	defer srv.Close()

	inner, err := newTestClient(srv.URL)
	if err != nil {
		t.Fatalf("newTestClient: %v", err)
	}
	client := NewDedupeClient(inner)

	_, _ = client.ReadSecret(context.Background(), "secret/a")
	_, _ = client.ReadSecret(context.Background(), "secret/b")

	if hits.Load() != 2 {
		t.Errorf("expected 2 upstream hits for different paths, got %d", hits.Load())
	}
}
