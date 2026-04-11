package vault

import (
	"fmt"
	"testing"
)

func BenchmarkWorkerPool_FetchAll_Concurrency1(b *testing.B) {
	benchmarkWorkerPool(b, 1, 100)
}

func BenchmarkWorkerPool_FetchAll_Concurrency4(b *testing.B) {
	benchmarkWorkerPool(b, 4, 100)
}

func BenchmarkWorkerPool_FetchAll_Concurrency16(b *testing.B) {
	benchmarkWorkerPool(b, 16, 100)
}

func benchmarkWorkerPool(b *testing.B, concurrency, numPaths int) {
	b.Helper()
	data := make(map[string]map[string]string, numPaths)
	paths := make([]string, numPaths)
	for i := 0; i < numPaths; i++ {
		p := fmt.Sprintf("secret/bench/%d", i)
		paths[i] = p
		data[p] = map[string]string{"value": fmt.Sprintf("v%d", i)}
	}
	reader := &mockReader{data: data}
	pool := NewWorkerPool(concurrency)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pool.FetchAll(reader, paths)
	}
}
