package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// BenchmarkRecurseSecrets_Flat benchmarks a flat listing of 50 leaf secrets.
func BenchmarkRecurseSecrets_Flat(b *testing.B) {
	keys := make([]string, 50)
	for i := range keys {
		keys[i] = "secret-key"
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"keys": keys},
		})
	}))
	defer server.Close()

	client := newTestClient(b, server.URL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.RecurseSecrets("secret", "", KVv1)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// newTestClientB is a helper that accepts testing.TB.
func newTestClient(tb testing.TB, addr string) *Client {
	tb.Helper()
	c, err := NewClient(addr, "test-token")
	if err != nil {
		tb.Fatalf("NewClient: %v", err)
	}
	return c
}
