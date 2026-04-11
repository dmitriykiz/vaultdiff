package vault

import (
	"context"
	"testing"
)

func BenchmarkAuditedClient_ReadSecret(b *testing.B) {
	stub := &stubReader{data: map[string]string{"k": "v"}}
	log := NewAuditLog(nil)
	client := NewAuditedClient(stub, log)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//nolint:errcheck
		client.ReadSecret(ctx, "secret/bench")
	}
}

func BenchmarkAuditLog_Record(b *testing.B) {
	log := NewAuditLog(nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Record("read", "secret/bench", nil)
	}
}

func BenchmarkAuditLog_RecordConcurrent(b *testing.B) {
	log := NewAuditLog(nil)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			log.Record("read", "secret/concurrent", nil)
		}
	})
}
