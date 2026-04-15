package vault

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newBatchBenchServer(b *testing.B, n int) (*httptest.Server, []string) {
	b.Helper()
	secrets := make(map[string]string, n)
	paths := make([]string, n)
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("secret/bench-%d", i)
		secrets[key] = fmt.Sprintf("value-%d", i)
		paths[i] = key
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Path[len("/v1/"):]
		if val, ok := secrets[key]; ok {
			fmt.Fprintf(w, `{"data":{"value":%q}}`, val)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	return srv, paths
}

func benchmarkBatch(b *testing.B, n, concurrency int) {
	b.Helper()
	srv, paths := newBatchBenchServer(b, n)
	defer srv.Close()

	client := newTestClient(b, srv.URL)
	batch := NewBatchClient(client, concurrency)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batch.ReadAll(ctx, paths)
	}
}

func BenchmarkBatchClient_10Secrets_Workers1(b *testing.B)  { benchmarkBatch(b, 10, 1) }
func BenchmarkBatchClient_10Secrets_Workers4(b *testing.B)  { benchmarkBatch(b, 10, 4) }
func BenchmarkBatchClient_50Secrets_Workers4(b *testing.B)  { benchmarkBatch(b, 50, 4) }
func BenchmarkBatchClient_50Secrets_Workers16(b *testing.B) { benchmarkBatch(b, 50, 16) }
