package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func newThrottleBenchServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"key":"value"}}`))
	}))
}

func benchmarkThrottle(b *testing.B, cap int) {
	b.Helper()
	srv := newThrottleBenchServer()
	defer srv.Close()

	inner := newTestClient(b, srv.URL)
	client := NewThrottledClient(inner, cap, 5*time.Second)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = client.ReadSecret(context.Background(), "secret/bench")
		}
	})
}

func BenchmarkThrottledClient_Cap1(b *testing.B)  { benchmarkThrottle(b, 1) }
func BenchmarkThrottledClient_Cap4(b *testing.B)  { benchmarkThrottle(b, 4) }
func BenchmarkThrottledClient_Cap16(b *testing.B) { benchmarkThrottle(b, 16) }

func BenchmarkThrottledClient_InFlight(b *testing.B) {
	srv := newThrottleBenchServer()
	defer srv.Close()

	inner := newTestClient(b, srv.URL)
	client := NewThrottledClient(inner, 8, 5*time.Second)

	var wg sync.WaitGroup
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = client.ReadSecret(context.Background(), "secret/bench")
		}()
	}
	wg.Wait()
}
