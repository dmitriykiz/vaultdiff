package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newPaginateBenchServer(b *testing.B, n int) *httptest.Server {
	b.Helper()
	keys := make([]string, n)
	for i := range keys {
		keys[i] = fmt.Sprintf("secret%d", i)
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "LIST" || r.URL.Query().Get("list") == "true" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": keys},
			})
			return
		}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		key := parts[len(parts)-1]
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"value": key},
		})
	}))
}

func benchmarkPaginate(b *testing.B, numSecrets, pageSize int) {
	b.Helper()
	srv := newPaginateBenchServer(b, numSecrets)
	defer srv.Close()
	client := NewPaginatedClient(newTestClient(b, srv), pageSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.EachPage(context.Background(), "secret", "", func(_ map[string]map[string]interface{}) error {
			return nil
		})
	}
}

func BenchmarkPaginatedClient_20Secrets_Page5(b *testing.B)  { benchmarkPaginate(b, 20, 5) }
func BenchmarkPaginatedClient_20Secrets_Page20(b *testing.B) { benchmarkPaginate(b, 20, 20) }
func BenchmarkPaginatedClient_50Secrets_Page10(b *testing.B) { benchmarkPaginate(b, 50, 10) }
