package vault

import (
	"context"
	"fmt"
	"time"
)

// CheckpointClient wraps a SecretReader and can persist a named snapshot of
// all secrets under a mount path to a CheckpointStore.
type CheckpointClient struct {
	inner SecretReader
	store *CheckpointStore
}

// NewCheckpointClient returns a CheckpointClient backed by inner and store.
func NewCheckpointClient(inner SecretReader, store *CheckpointStore) *CheckpointClient {
	return &CheckpointClient{inner: inner, store: store}
}

// ReadSecret delegates to the inner client.
func (c *CheckpointClient) ReadSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	return c.inner.ReadSecret(ctx, path)
}

// TakeCheckpoint snapshots all secrets returned by taker under mountPath and
// persists them under name in the store.
func (c *CheckpointClient) TakeCheckpoint(ctx context.Context, name, mountPath string, taker *SnapshotTaker) error {
	secrets, err := taker.Take(ctx, mountPath)
	if err != nil {
		return fmt.Errorf("checkpoint %q: take snapshot: %w", name, err)
	}
	cp := Checkpoint{
		Path:    mountPath,
		TakenAt: time.Now().UTC(),
		Secrets: make(map[string]interface{}, len(secrets)),
	}
	for k, v := range secrets {
		cp.Secrets[k] = v
	}
	if err := c.store.Save(name, cp); err != nil {
		return fmt.Errorf("checkpoint %q: save: %w", name, err)
	}
	return nil
}

// LoadCheckpoint retrieves a previously saved checkpoint by name.
func (c *CheckpointClient) LoadCheckpoint(name string) (Checkpoint, error) {
	cp, err := c.store.Load(name)
	if err != nil {
		return Checkpoint{}, fmt.Errorf("checkpoint %q: load: %w", name, err)
	}
	return cp, nil
}
