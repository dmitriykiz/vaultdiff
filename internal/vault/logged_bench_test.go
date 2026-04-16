package vault

import (
	"context"
	"io"
	"testing"
)

func newLoggedBenchServer(b *testing.B) SecretReader {
	b.Helper()
	return newTestClient(b, map[string]any{
		"username": "admin",
		"password": "s3cr3t",
	})
}

func BenchmarkLoggedClient_ReadSecret_Discard(b *testing.B) {
	inner := newLoggedBenchServer(b)
	logger := NewStdLogger(io.Discard)
	c := NewLoggedClient(inner, logger)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.ReadSecret(ctx, "secret/bench")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLoggedClient_ReadSecret_CaptureLogger(b *testing.B) {
	inner := newLoggedBenchServer(b)
	log := &captureLogger{}
	c := NewLoggedClient(inner, log)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.entries = log.entries[:0]
		_, err := c.ReadSecret(ctx, "secret/bench")
		if err != nil {
			b.Fatal(err)
		}
	}
}
