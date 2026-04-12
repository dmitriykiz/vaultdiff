package vault

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// Snapshot holds a point-in-time capture of all secrets under a Vault path.
type Snapshot struct {
	Path      string
	CapturedAt time.Time
	Secrets   map[string]map[string]interface{}
}

// SnapshotTaker captures a full recursive snapshot of secrets from a Vault mount.
type SnapshotTaker struct {
	client    SecretReader
	workers   int
}

// NewSnapshotTaker creates a SnapshotTaker using the given client and worker concurrency.
func NewSnapshotTaker(client SecretReader, workers int) *SnapshotTaker {
	if workers <= 0 {
		workers = DefaultConcurrency
	}
	return &SnapshotTaker{client: client, workers: workers}
}

// Take recursively reads all secrets under path and returns a Snapshot.
func (s *SnapshotTaker) Take(ctx context.Context, mount, path string, engine EngineType) (*Snapshot, error) {
	keys, err := RecurseSecrets(ctx, s.client, mount, path, engine)
	if err != nil {
		return nil, fmt.Errorf("snapshot recurse %s: %w", path, err)
	}

	sort.Strings(keys)

	pool := NewWorkerPool(s.client, s.workers)
	results, err := pool.FetchAll(ctx, keys)
	if err != nil {
		return nil, fmt.Errorf("snapshot fetch %s: %w", path, err)
	}

	secrets := make(map[string]map[string]interface{}, len(results))
	for k, v := range results {
		secrets[k] = v
	}

	return &Snapshot{
		Path:       path,
		CapturedAt: time.Now().UTC(),
		Secrets:    secrets,
	}, nil
}
