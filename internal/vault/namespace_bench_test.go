package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newNamespaceBenchServer(b *testing.B) *httptest.Server {
	b.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"key": "value"},
		})
	}))
}

func BenchmarkNamespacedClient_ReadSecret(b *testing.B) {
	srv := newNamespaceBenchServer(b)
	defer srv.Close()

	inner := newTestClient(b, srv.URL)
	client := NewNamespacedClient(inner, "benchns")
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.ReadSecret(ctx, "secret/bench")
	}
}

func BenchmarkNamespacedClient_QualifiedPath(b *testing.B) {
	inner := newTestClient(b, "http://127.0.0.1")
	client := NewNamespacedClient(inner, "org/team")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.qualifiedPath("secret/myapp/config")
	}
}
