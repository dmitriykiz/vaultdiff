package vault

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newCircuitTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, SecretReader) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return srv, client
}

func TestCircuitBreakerClient_PassesThroughOnSuccess(t *testing.T) {
	_, inner := newCircuitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"key":"val"}}`))
	})
	cb := NewCircuitBreaker(3, time.Second)
	client := NewCircuitBreakerClient(inner, cb)

	data, err := client.ReadSecret(context.Background(), "secret/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["key"] != "val" {
		t.Errorf("expected val, got %v", data["key"])
	}
	if cb.State() != "closed" {
		t.Errorf("expected closed, got %s", cb.State())
	}
}

func TestCircuitBreakerClient_OpensAfterFailures(t *testing.T) {
	_, inner := newCircuitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	cb := NewCircuitBreaker(2, time.Second)
	client := NewCircuitBreakerClient(inner, cb)

	for i := 0; i < 2; i++ {
		_, _ = client.ReadSecret(context.Background(), "secret/test")
	}
	if cb.State() != "open" {
		t.Fatalf("expected open, got %s", cb.State())
	}
	_, err := client.ReadSecret(context.Background(), "secret/test")
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreakerClient_ContextCancelDoesNotTrip(t *testing.T) {
	_, inner := newCircuitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	cb := NewCircuitBreaker(1, time.Second)
	client := NewCircuitBreakerClient(inner, cb)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	_, _ = client.ReadSecret(ctx, "secret/test")
	if cb.State() != "closed" {
		t.Errorf("circuit should remain closed on ctx cancel, got %s", cb.State())
	}
}

func TestCircuitBreakerClient_NilBreakerUsesDefault(t *testing.T) {
	_, inner := newCircuitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{}}`))
	})
	client := NewCircuitBreakerClient(inner, nil)
	if client.Breaker() == nil {
		t.Fatal("expected non-nil breaker")
	}
}
