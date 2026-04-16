package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newVersionBenchServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"key":"value"}}`))
	}))
}

func BenchmarkVersionClient_ReadSecret_Latest(b *testing.B) {
	srv := newVersionBenchServer()
	defer srv.Close()
	inner, _ := NewClient(srv.URL, "bench-token")
	vc := NewVersionClient(inner, 0)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = vc.ReadSecret(ctx, "secret/data/bench")
	}
}

func BenchmarkVersionClient_ReadSecret_Pinned(b *testing.B) {
	srv := newVersionBenchServer()
	defer srv.Close()
	inner, _ := NewClient(srv.URL, "bench-token")
	vc := NewVersionClient(inner, 4)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = vc.ReadSecret(ctx, "secret/data/bench")
	}
}
