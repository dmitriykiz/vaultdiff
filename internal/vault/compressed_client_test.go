package vault

import (
	"bytes"
	"compress/gzip"
	"context"
	"testing"
)

func gzipString(t *testing.T, s string) string {
	t.Helper()
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, _ = w.Write([]byte(s))
	_ = w.Close()
	return buf.String()
}

type staticReader struct{ data map[string]any }

func (s *staticReader) ReadSecret(_ context.Context, _ string) (map[string]any, error) {
	return s.data, nil
}

func TestCompressedClient_Disabled_PassesThrough(t *testing.T) {
	inner := &staticReader{data: map[string]any{"key": gzipString(t, "hello")}}
	c := NewCompressedClient(inner, false)
	got, err := c.ReadSecret(context.Background(), "secret/foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["key"] != inner.data["key"] {
		t.Errorf("expected raw value, got %v", got["key"])
	}
}

func TestCompressedClient_DecompressesGzippedString(t *testing.T) {
	compressed := gzipString(t, "hello world")
	inner := &staticReader{data: map[string]any{"msg": compressed}}
	c := NewCompressedClient(inner, true)
	got, err := c.ReadSecret(context.Background(), "secret/foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["msg"] != "hello world" {
		t.Errorf("expected 'hello world', got %v", got["msg"])
	}
}

func TestCompressedClient_NonCompressedStringUnchanged(t *testing.T) {
	inner := &staticReader{data: map[string]any{"plain": "not compressed"}}
	c := NewCompressedClient(inner, true)
	got, err := c.ReadSecret(context.Background(), "secret/foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["plain"] != "not compressed" {
		t.Errorf("expected original value, got %v", got["plain"])
	}
}

func TestCompressedClient_NonStringValuesUnchanged(t *testing.T) {
	inner := &staticReader{data: map[string]any{"count": 42}}
	c := NewCompressedClient(inner, true)
	got, err := c.ReadSecret(context.Background(), "secret/foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["count"] != 42 {
		t.Errorf("expected 42, got %v", got["count"])
	}
}

func TestCompressedClient_DecompressesJSON(t *testing.T) {
	compressed := gzipString(t, `{"nested":true}`)
	inner := &staticReader{data: map[string]any{"obj": compressed}}
	c := NewCompressedClient(inner, true)
	got, err := c.ReadSecret(context.Background(), "secret/foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := got["obj"].(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", got["obj"])
	}
	if m["nested"] != true {
		t.Errorf("expected true, got %v", m["nested"])
	}
}
