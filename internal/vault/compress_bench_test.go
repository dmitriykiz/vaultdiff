package vault

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"testing"
)

func newCompressBenchData(t *testing.B, n int) map[string]any {
	t.Helper()
	data := make(map[string]any, n)
	for i := 0; i < n; i++ {
		var buf bytes.Buffer
		w := gzip.NewWriter(&buf)
		_, _ = fmt.Fprintf(w, "value-%d", i)
		_ = w.Close()
		data[fmt.Sprintf("key-%d", i)] = buf.String()
	}
	return data
}

func BenchmarkCompressedClient_ReadSecret_10Keys(b *testing.B) {
	data := newCompressBenchData(b, 10)
	inner := &staticReader{data: data}
	c := NewCompressedClient(inner, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.ReadSecret(context.Background(), "secret/bench")
	}
}

func BenchmarkCompressedClient_ReadSecret_100Keys(b *testing.B) {
	data := newCompressBenchData(b, 100)
	inner := &staticReader{data: data}
	c := NewCompressedClient(inner, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.ReadSecret(context.Background(), "secret/bench")
	}
}

func BenchmarkCompressedClient_Disabled_100Keys(b *testing.B) {
	data := newCompressBenchData(b, 100)
	inner := &staticReader{data: data}
	c := NewCompressedClient(inner, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.ReadSecret(context.Background(), "secret/bench")
	}
}
