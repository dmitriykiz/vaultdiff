package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newObserveBenchServer(b *testing.B) *httptest.Server {
	b.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"key":"value"}}`))
	}))
}

func BenchmarkObservedClient_ReadSecret(b *testing.B) {
	srv := newObserveBenchServer(b)
	defer srv.Close()

	client := &Client{address: srv.URL, httpClient: &http.Client{Timeout: 5 * time.Second}}
	m := &InMemoryMetrics{}
	obs := NewObservedClient(client, m)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = obs.ReadSecret(ctx, "secret/bench")
	}
}

func BenchmarkInMemoryMetrics_Record(b *testing.B) {
	m := &InMemoryMetrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Record("secret/bench", time.Millisecond, nil)
	}
}

func BenchmarkInMemoryMetrics_RecordConcurrent(b *testing.B) {
	m := &InMemoryMetrics{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Record("secret/concurrent", time.Millisecond, nil)
		}
	})
}
