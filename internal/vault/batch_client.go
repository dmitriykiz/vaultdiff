package vault

import (
	"context"
	"fmt"
	"sync"
)

// BatchResult holds the result of a single secret read in a batch operation.
type BatchResult struct {
	Path  string
	Data  map[string]any
	Error error
}

// BatchClient reads multiple secrets concurrently and returns all results.
type BatchClient struct {
	inner       SecretReader
	concurrency int
}

// NewBatchClient creates a BatchClient wrapping inner with the given concurrency.
// If concurrency is <= 0 it defaults to 5.
func NewBatchClient(inner SecretReader, concurrency int) *BatchClient {
	if concurrency <= 0 {
		concurrency = 5
	}
	return &BatchClient{inner: inner, concurrency: concurrency}
}

// ReadSecret delegates to the inner client for single-path reads.
func (b *BatchClient) ReadSecret(ctx context.Context, path string) (map[string]any, error) {
	return b.inner.ReadSecret(ctx, path)
}

// ReadAll reads all provided paths concurrently and returns a slice of BatchResult
// in the same order as the input paths.
func (b *BatchClient) ReadAll(ctx context.Context, paths []string) []BatchResult {
	results := make([]BatchResult, len(paths))

	type job struct {
		index int
		path  string
	}

	jobs := make(chan job, len(paths))
	for i, p := range paths {
		jobs <- job{index: i, path: p}
	}
	close(jobs)

	var wg sync.WaitGroup
	for w := 0; w < b.concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				data, err := b.inner.ReadSecret(ctx, j.path)
				results[j.index] = BatchResult{Path: j.path, Data: data, Error: err}
			}
		}()
	}
	wg.Wait()
	return results
}

// Errors returns only the failed results from a BatchResult slice.
func Errors(results []BatchResult) []BatchResult {
	var errs []BatchResult
	for _, r := range results {
		if r.Error != nil {
			errs = append(errs, r)
		}
	}
	return errs
}

// JoinErrors formats all errors from a BatchResult slice into a single string.
func JoinErrors(results []BatchResult) string {
	var msg string
	for _, r := range Errors(results) {
		msg += fmt.Sprintf("%s: %v\n", r.Path, r.Error)
	}
	return msg
}
