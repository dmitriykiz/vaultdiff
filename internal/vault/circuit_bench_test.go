package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newCircuitBenchServer(b *testing.B) (*httptest.Server, SecretReader) {
	b.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"k":"v"}}`))
	}))
	b.Cleanup(srv.Close)
	client, err := NewClient(srv.URL, "bench-token")
	if err != nil {
		b.Fatalf("NewClient: %v", err)
	}
	return srv, client
}

func BenchmarkCircuitBreakerClient_ReadSecret(b *testing.B) {
	_, inner := newCircuitBenchServer(b)
	cb := NewCircuitBreaker(100, time.Minute)
	client := NewCircuitBreakerClient(inner, cb)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.ReadSecret(ctx, "secret/bench")
	}
}

func BenchmarkCircuitBreaker_AllowRecord(b *testing.B) {
	cb := NewCircuitBreaker(1000, time.Minute)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.Allow()
		cb.RecordSuccess()
	}
}
