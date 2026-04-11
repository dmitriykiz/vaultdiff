package vault

import (
	"errors"
	"fmt"
	"sort"
	"sync/atomic"
	"testing"
)

// mockReader is a SecretReader whose behaviour is driven by a map.
type mockReader struct {
	data    map[string]map[string]string
	callCount int32
}

func (m *mockReader) ReadSecret(path string) (map[string]string, error) {
	atomic.AddInt32(&m.callCount, 1)
	if d, ok := m.data[path]; ok {
		return d, nil
	}
	return nil, errors.New("not found: " + path)
}

func TestWorkerPool_FetchAll_ReturnsAllResults(t *testing.T) {
	reader := &mockReader{
		data: map[string]map[string]string{
			"secret/a": {"key": "alpha"},
			"secret/b": {"key": "beta"},
			"secret/c": {"key": "gamma"},
		},
	}
	paths := []string{"secret/a", "secret/b", "secret/c"}
	pool := NewWorkerPool(2)
	results := pool.FetchAll(reader, paths)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for %s: %v", r.Path, r.Err)
		}
	}
}

func TestWorkerPool_FetchAll_PropagatesErrors(t *testing.T) {
	reader := &mockReader{data: map[string]map[string]string{}}
	paths := []string{"secret/missing"}
	pool := NewWorkerPool(1)
	results := pool.FetchAll(reader, paths)

	if results[0].Err == nil {
		t.Error("expected error for missing path, got nil")
	}
}

func TestWorkerPool_FetchAll_OrderPreserved(t *testing.T) {
	data := map[string]map[string]string{}
	paths := make([]string, 10)
	for i := 0; i < 10; i++ {
		p := fmt.Sprintf("secret/%d", i)
		paths[i] = p
		data[p] = map[string]string{"index": fmt.Sprintf("%d", i)}
	}
	reader := &mockReader{data: data}
	pool := NewWorkerPool(4)
	results := pool.FetchAll(reader, paths)

	for i, r := range results {
		want := fmt.Sprintf("secret/%d", i)
		if r.Path != want {
			t.Errorf("position %d: got path %q, want %q", i, r.Path, want)
		}
	}
}

func TestWorkerPool_DefaultConcurrency(t *testing.T) {
	pool := NewWorkerPool(0)
	if pool.concurrency != 1 {
		t.Errorf("expected concurrency 1, got %d", pool.concurrency)
	}
}

func TestWorkerPool_FetchAll_Empty(t *testing.T) {
	reader := &mockReader{data: map[string]map[string]string{}}
	pool := NewWorkerPool(3)
	results := pool.FetchAll(reader, []string{})
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
	_ = sort.Search // suppress unused import warning
}
