package vault

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

type captureLogger struct {
	entries []map[string]any
}

func (l *captureLogger) Log(level, msg string, fields map[string]any) {
	e := map[string]any{"level": level, "msg": msg}
	for k, v := range fields {
		e[k] = v
	}
	l.entries = append(l.entries, e)
}

func TestLoggedClient_LogsSuccessfulRead(t *testing.T) {
	srv := newTestClient(t, map[string]any{"foo": "bar"})
	log := &captureLogger{}
	c := NewLoggedClient(srv, log)

	data, err := c.ReadSecret(context.Background(), "secret/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["foo"] != "bar" {
		t.Errorf("expected foo=bar, got %v", data["foo"])
	}
	if len(log.entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(log.entries))
	}
	if log.entries[0]["level"] != "info" {
		t.Errorf("expected level=info")
	}
}

func TestLoggedClient_LogsErrorRead(t *testing.T) {
	failing := &errorReader{err: errors.New("vault down")}
	log := &captureLogger{}
	c := NewLoggedClient(failing, log)

	_, err := c.ReadSecret(context.Background(), "secret/missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if len(log.entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(log.entries))
	}
	if log.entries[0]["level"] != "error" {
		t.Errorf("expected level=error")
	}
	if _, ok := log.entries[0]["error"]; !ok {
		t.Errorf("expected error field in log entry")
	}
}

func TestLoggedClient_NilLogger_UsesDiscard(t *testing.T) {
	srv := newTestClient(t, map[string]any{"k": "v"})
	c := NewLoggedClient(srv, nil)
	_, err := c.ReadSecret(context.Background(), "secret/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStdLogger_WritesToWriter(t *testing.T) {
	var buf bytes.Buffer
	l := NewStdLogger(&buf)
	l.Log("info", "hello", map[string]any{"path": "secret/x"})
	out := buf.String()
	if !strings.Contains(out, "level=info") {
		t.Errorf("expected level=info in output: %s", out)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("expected msg in output: %s", out)
	}
}

type errorReader struct{ err error }

func (e *errorReader) ReadSecret(_ context.Context, _ string) (map[string]any, error) {
	return nil, e.err
}
