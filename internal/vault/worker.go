package vault

import "sync"

// WorkerPool manages concurrent secret fetching with a bounded goroutine pool.
type WorkerPool struct {
	concurrency int
}

// NewWorkerPool creates a WorkerPool with the given concurrency limit.
// If concurrency is less than 1, it defaults to 1.
func NewWorkerPool(concurrency int) *WorkerPool {
	if concurrency < 1 {
		concurrency = 1
	}
	return &WorkerPool{concurrency: concurrency}
}

// SecretResult holds the outcome of a single secret fetch.
type SecretResult struct {
	Path   string
	Data   map[string]string
	Err    error
}

// FetchAll fetches secrets for all given paths concurrently using the pool.
// Results are returned in an unspecified order.
func (wp *WorkerPool) FetchAll(client SecretReader, paths []string) []SecretResult {
	results := make([]SecretResult, len(paths))
	sem := make(chan struct{}, wp.concurrency)
	var wg sync.WaitGroup

	for i, path := range paths {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, p string) {
			defer wg.Done()
			defer func() { <-sem }()
			data, err := client.ReadSecret(p)
			results[idx] = SecretResult{Path: p, Data: data, Err: err}
		}(i, path)
	}

	wg.Wait()
	return results
}
