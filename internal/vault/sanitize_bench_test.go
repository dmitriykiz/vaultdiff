package vault

import (
	"context"
	"regexp"
	"testing"
)

func BenchmarkSanitizer_Apply_SmallMap(b *testing.B) {
	s := NewSanitizer(DefaultSanitizeRules())
	data := map[string]interface{}{
		"password": "hunter2",
		"username": "admin",
		"host":     "localhost",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Apply(data)
	}
}

func BenchmarkSanitizer_Apply_LargeMap(b *testing.B) {
	s := NewSanitizer([]SanitizeRule{
		{KeyPattern: regexp.MustCompile(`(password|token|secret)`), Replacement: "[sanitized]"},
	})
	data := make(map[string]interface{}, 50)
	for i := 0; i < 50; i++ {
		data["key"] = "value"
	}
	data["password"] = "secret"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Apply(data)
	}
}

func BenchmarkSanitizedClient_ReadSecret(b *testing.B) {
	stub := &stubSanitizeReader{
		data: map[string]interface{}{
			"password": "secret",
			"api_key":  "tok_abc",
			"host":     "localhost",
		},
	}
	client := NewSanitizedClient(stub, NewSanitizer(DefaultSanitizeRules()))
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.ReadSecret(ctx, "secret/bench")
	}
}
