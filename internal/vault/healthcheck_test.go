package vault

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newHealthTestServer(statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if statusCode != http.StatusOK {
			http.Error(w, "vault error", statusCode)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"status":"ok"}}`))
	}))
}

func TestHealthChecker_DefaultProbePath(t *testing.T) {
	client := &fakeReader{}
	hc := NewHealthChecker(client, "")
	if hc.probePath != "sys/health" {
		t.Errorf("expected default probe path 'sys/health', got %q", hc.probePath)
	}
}

func TestHealthChecker_CustomProbePath(t *testing.T) {
	client := &fakeReader{}
	hc := NewHealthChecker(client, "secret/ping")
	if hc.probePath != "secret/ping" {
		t.Errorf("expected probe path 'secret/ping', got %q", hc.probePath)
	}
}

func TestHealthChecker_Check_Healthy(t *testing.T) {
	client := &fakeReader{data: map[string]interface{}{"status": "ok"}}
	hc := NewHealthChecker(client, "sys/health")

	status := hc.Check(context.Background())

	if !status.Healthy {
		t.Errorf("expected healthy, got unhealthy: %v", status.Error)
	}
	if status.Error != nil {
		t.Errorf("expected no error, got %v", status.Error)
	}
	if status.Latency < 0 {
		t.Errorf("expected non-negative latency")
	}
}

func TestHealthChecker_Check_Unhealthy(t *testing.T) {
	client := &fakeReader{err: errors.New("connection refused")}
	hc := NewHealthChecker(client, "sys/health")

	status := hc.Check(context.Background())

	if status.Healthy {
		t.Error("expected unhealthy status")
	}
	if status.Error == nil {
		t.Error("expected non-nil error")
	}
	if !strings.Contains(status.Error.Error(), "health probe failed") {
		t.Errorf("unexpected error message: %v", status.Error)
	}
}

func TestHealthStatus_String_Healthy(t *testing.T) {
	server := newHealthTestServer(http.StatusOK)
	defer server.Close()

	client := &fakeReader{data: map[string]interface{}{"ok": true}}
	hc := NewHealthChecker(client, "sys/health")
	status := hc.Check(context.Background())

	s := status.String()
	if !strings.Contains(s, "healthy") {
		t.Errorf("expected 'healthy' in string output, got: %s", s)
	}
}

func TestHealthStatus_String_Unhealthy(t *testing.T) {
	client := &fakeReader{err: errors.New("timeout")}
	hc := NewHealthChecker(client, "sys/health")
	status := hc.Check(context.Background())

	s := status.String()
	if !strings.Contains(s, "unhealthy") {
		t.Errorf("expected 'unhealthy' in string output, got: %s", s)
	}
}

// fakeReader is a minimal SecretReader used in health check tests.
type fakeReader struct {
	data map[string]interface{}
	err  error
}

func (f *fakeReader) ReadSecret(_ context.Context, _ string) (map[string]interface{}, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.data, nil
}
