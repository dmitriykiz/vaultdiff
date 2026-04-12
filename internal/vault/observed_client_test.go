package vault

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newObserveTestServer(t *testing.T, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if status != http.StatusOK {
			http.Error(w, "error", status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"key":"val"}}`))
	}))
}

func TestObservedClient_RecordsSuccess(t *testing.T) {
	srv := newObserveTestServer(t, http.StatusOK)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	m := &InMemoryMetrics{}
	obs := NewObservedClient(client, m)

	_, err := obs.ReadSecret(context.Background(), "secret/foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	snap := m.Snapshot()
	if snap.TotalRequests != 1 {
		t.Errorf("expected 1 request, got %d", snap.TotalRequests)
	}
	if snap.SuccessCount != 1 {
		t.Errorf("expected 1 success, got %d", snap.SuccessCount)
	}
	if snap.ErrorCount != 0 {
		t.Errorf("expected 0 errors, got %d", snap.ErrorCount)
	}
}

func TestObservedClient_RecordsError(t *testing.T) {
	srv := newObserveTestServer(t, http.StatusForbidden)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	m := &InMemoryMetrics{}
	obs := NewObservedClient(client, m)

	_, err := obs.ReadSecret(context.Background(), "secret/forbidden")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	snap := m.Snapshot()
	if snap.ErrorCount != 1 {
		t.Errorf("expected 1 error, got %d", snap.ErrorCount)
	}
	if snap.SuccessCount != 0 {
		t.Errorf("expected 0 successes, got %d", snap.SuccessCount)
	}
}

func TestObservedClient_NilCollector_DoesNotPanic(t *testing.T) {
	srv := newObserveTestServer(t, http.StatusOK)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	obs := NewObservedClient(client, nil)

	_, err := obs.ReadSecret(context.Background(), "secret/foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestObservedClient_RecordsLatency(t *testing.T) {
	srv := newObserveTestServer(t, http.StatusOK)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	m := &InMemoryMetrics{}
	obs := NewObservedClient(client, m)

	for i := 0; i < 3; i++ {
		_, _ = obs.ReadSecret(context.Background(), "secret/foo")
	}

	snap := m.Snapshot()
	if snap.TotalRequests != 3 {
		t.Errorf("expected 3 requests, got %d", snap.TotalRequests)
	}
}

func TestObservedClient_PropagatesError(t *testing.T) {
	expected := errors.New("inner error")
	inner := &stubReader{err: expected}
	m := &InMemoryMetrics{}
	obs := NewObservedClient(inner, m)

	_, err := obs.ReadSecret(context.Background(), "any/path")
	if !errors.Is(err, expected) {
		t.Errorf("expected propagated error, got %v", err)
	}
}

type stubReader struct{ err error }

func (s *stubReader) ReadSecret(_ context.Context, _ string) (map[string]interface{}, error) {
	return nil, s.err
}
