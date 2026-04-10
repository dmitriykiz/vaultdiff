package vault

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkSecretCache_SetGet(b *testing.B) {
	c := NewSecretCache(time.Minute)
	data := map[string]interface{}{"password": "hunter2", "user": "admin"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("secret/bench/%d", i%100)
		c.Set(path, data)
		_ = c.Get(path)
	}
}

func BenchmarkSecretCache_ConcurrentGet(b *testing.B) {
	c := NewSecretCache(time.Minute)
	for i := 0; i < 100; i++ {
		c.Set(fmt.Sprintf("secret/path/%d", i), map[string]interface{}{"k": i})
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = c.Get(fmt.Sprintf("secret/path/%d", i%100))
			i++
		}
	})
}

func BenchmarkSecretCache_Flush(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		c := NewSecretCache(0)
		for j := 0; j < 500; j++ {
			c.Set(fmt.Sprintf("secret/%d", j), map[string]interface{}{})
		}
		b.StartTimer()
		c.Flush()
	}
}
