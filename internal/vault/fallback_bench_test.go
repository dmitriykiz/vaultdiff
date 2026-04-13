package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newFallbackBenchServer(status int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func BenchmarkFallbackClient_PrimaryHit(b *testing.B) {
	srv := newFallbackBenchServer(200, `{"data":{"key":"value"}}`)
	defer srv.Close()
	client, _ := NewClient(srv.URL, "bench-token")
	c := NewFallbackClient(client, client, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.ReadSecret(ctx, "secret/bench")
	}
}

func BenchmarkFallbackClient_PrimaryMiss_FallbackHit(b *testing.B) {
	primSrv := newFallbackBenchServer(500, `{}`)
	defer primSrv.Close()
	secSrv := newFallbackBenchServer(200, `{"data":{"key":"value"}}`)
	defer secSrv.Close()

	prim, _ := NewClient(primSrv.URL, "bench-token")
	sec, _ := NewClient(secSrv.URL, "bench-token")
	c := NewFallbackClient(prim, sec, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.ReadSecret(ctx, "secret/bench")
	}
}
