package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newSnapshotBenchServer(n int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// list endpoint
		if r.URL.Query().Get("list") == "true" || r.Method == "LIST" {
			keys := make([]string, n)
			for i := 0; i < n; i++ {
				keys[i] = fmt.Sprintf("secret-%d", i)
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": keys},
			})
			return
		}
		// read endpoint
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"data": map[string]interface{}{"value": "bench"},
			},
		})
	}))
}

func benchmarkSnapshot(b *testing.B, secrets, workers int) {
	b.Helper()
	srv := newSnapshotBenchServer(secrets)
	defer srv.Close()

	client := newTestClient(b, srv.URL)
	taker := NewSnapshotTaker(client, workers)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = taker.Take(ctx, "secret", "bench", EngineKVv2)
	}
}

func BenchmarkSnapshotTaker_10Secrets_Workers1(b *testing.B)  { benchmarkSnapshot(b, 10, 1) }
func BenchmarkSnapshotTaker_10Secrets_Workers4(b *testing.B)  { benchmarkSnapshot(b, 10, 4) }
func BenchmarkSnapshotTaker_50Secrets_Workers4(b *testing.B)  { benchmarkSnapshot(b, 50, 4) }
func BenchmarkSnapshotTaker_50Secrets_Workers16(b *testing.B) { benchmarkSnapshot(b, 50, 16) }
