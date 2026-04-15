package vault

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Checkpoint represents a saved snapshot of secrets at a point in time.
type Checkpoint struct {
	Path      string                 `json:"path"`
	TakenAt   time.Time              `json:"taken_at"`
	Secrets   map[string]interface{} `json:"secrets"`
}

// CheckpointStore persists and loads Checkpoint values to/from disk.
type CheckpointStore struct {
	dir string
}

// NewCheckpointStore returns a CheckpointStore rooted at dir.
func NewCheckpointStore(dir string) *CheckpointStore {
	return &CheckpointStore{dir: dir}
}

// Save writes a checkpoint to <dir>/<name>.json.
func (s *CheckpointStore) Save(name string, cp Checkpoint) error {
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return fmt.Errorf("checkpoint: mkdir %s: %w", s.dir, err)
	}
	path := s.filePath(name)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("checkpoint: create %s: %w", path, err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cp); err != nil {
		return fmt.Errorf("checkpoint: encode: %w", err)
	}
	return nil
}

// Load reads a checkpoint from <dir>/<name>.json.
func (s *CheckpointStore) Load(name string) (Checkpoint, error) {
	path := s.filePath(name)
	f, err := os.Open(path)
	if err != nil {
		return Checkpoint{}, fmt.Errorf("checkpoint: open %s: %w", path, err)
	}
	defer f.Close()
	var cp Checkpoint
	if err := json.NewDecoder(f).Decode(&cp); err != nil {
		return Checkpoint{}, fmt.Errorf("checkpoint: decode: %w", err)
	}
	return cp, nil
}

// Delete removes the checkpoint file for name.
func (s *CheckpointStore) Delete(name string) error {
	return os.Remove(s.filePath(name))
}

func (s *CheckpointStore) filePath(name string) string {
	return fmt.Sprintf("%s/%s.json", s.dir, name)
}
