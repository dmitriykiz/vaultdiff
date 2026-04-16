package vault

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTraceTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"key":"value"}}`))
	}))
}

func TestTracedClient_RecordsSuccessfulRead(t *testing.T) {
	srv := newTraceTestServer(t)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	tracer := NewTracer(nil)
	traced := NewTracedClient(client, tracer)

	_, err := traced.ReadSecret(context.Background(), "secret/foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := tracer.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Path != "secret/foo" {
		t.Errorf("expected path secret/foo, got %s", events[0].Path)
	}
	if events[0].Err != nil {
		t.Errorf("expected nil error, got %v", events[0].Err)
	}
	if events[0].Duration <= 0 {
		t.Error("expected positive duration")
	}
}

func TestTracedClient_RecordsErrorRead(t *testing.T) {
	errReader := &stubErrorReader{err: errors.New("vault down")}
	tracer := NewTracer(nil)
	traced := NewTracedClient(errReader, tracer)

	_, err := traced.ReadSecret(context.Background(), "secret/bar")
	if err == nil {
		t.Fatal("expected error")
	}

	events := tracer.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Err == nil {
		t.Error("expected non-nil error in event")
	}
}

func TestTracedClient_NilTracer_UsesNoop(t *testing.T) {
	srv := newTraceTestServer(t)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	traced := NewTracedClient(client, nil)

	_, err := traced.ReadSecret(context.Background(), "secret/baz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(traced.Tracer().Events()) != 1 {
		t.Error("expected event even with nil tracer arg")
	}
}

func TestTracedClient_EventsReturnsCopy(t *testing.T) {
	tracer := NewTracer(nil)
	events := tracer.Events()
	events = append(events, TraceEvent{Path: "injected"})
	if len(tracer.Events()) != 0 {
		t.Error("Events() should return a copy, not the internal slice")
	}
}

// stubErrorReader is a minimal SecretReader that always returns an error.
type stubErrorReader struct{ err error }

func (s *stubErrorReader) ReadSecret(_ context.Context, _ string) (map[string]interface{}, error) {
	return nil, s.err
}
