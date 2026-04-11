package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func newRateLimitTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"key":"value"}}`))
	}))
}

func TestRateLimitedClient_PassesThroughResult(t *testing.T) {
	srv := newRateLimitTestServer(t)
	defer srv.Close()

	base, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	client := NewRateLimitedClient(base, 0) // unlimited
	defer client.Stop()

	data, err := client.ReadSecret(context.Background(), "secret/foo")
	if err != nil {
		t.Fatalf("ReadSecret: %v", err)
	}
	if data["key"] != "value" {
		t.Fatalf("expected key=value, got %v", data)
	}
}

func TestRateLimitedClient_RespectsContextCancellation(t *testing.T) {
	srv := newRateLimitTestServer(t)
	defer srv.Close()

	base, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	client := NewRateLimitedClient(base, 0)
	defer client.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err = client.ReadSecret(ctx, "secret/foo")
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

func TestRateLimitedClient_ThrottlesBurstRequests(t *testing.T) {
	var calls int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"x":"1"}}`))
	}))
	defer srv.Close()

	base, err := NewClient(srv.URL, "tok")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	const rps = 5
	client := NewRateLimitedClient(base, rps)
	defer client.Stop()

	start := time.Now()
	for i := 0; i < rps; i++ {
		_, _ = client.ReadSecret(context.Background(), "secret/x")
	}
	elapsed := time.Since(start)

	if atomic.LoadInt64(&calls) != rps {
		t.Fatalf("expected %d calls, got %d", rps, calls)
	}
	if elapsed > 600*time.Millisecond {
		t.Fatalf("burst took too long: %v", elapsed)
	}
}
