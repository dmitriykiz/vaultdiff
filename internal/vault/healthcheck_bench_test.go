package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newHealthBenchServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"status":"ok"}}`))
	}))
}

func BenchmarkHealthChecker_Check_Healthy(b *testing.B) {
	client := &fakeReader{data: map[string]interface{}{"status": "ok"}}
	hc := NewHealthChecker(client, "sys/health")
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hc.Check(ctx)
	}
}

func BenchmarkHealthChecker_Check_Unhealthy(b *testing.B) {
	client := &fakeReader{err: errUnhealthy}
	hc := NewHealthChecker(client, "sys/health")
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hc.Check(ctx)
	}
}

func BenchmarkHealthStatus_String(b *testing.B) {
	client := &fakeReader{data: map[string]interface{}{"ok": true}}
	hc := NewHealthChecker(client, "sys/health")
	status := hc.Check(context.Background())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = status.String()
	}
}

var errUnhealthy = &unhealthyError{}

type unhealthyError struct{}

func (e *unhealthyError) Error() string { return "vault unreachable" }
