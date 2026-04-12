package vault

import (
	"context"
	"fmt"
)

// DiffClient compares secrets between two Vault paths using a SnapshotTaker.
type DiffClient struct {
	taker *SnapshotTaker
}

// NewDiffClient creates a DiffClient backed by the given SecretReader.
func NewDiffClient(reader SecretReader, workers int) *DiffClient {
	return &DiffClient{
		taker: NewSnapshotTaker(reader, workers),
	}
}

// PathSnapshot holds the resolved secrets for a single Vault path.
type PathSnapshot struct {
	Path    string
	Secrets map[string]map[string]interface{}
}

// TakeSnapshots concurrently fetches secrets from both paths and returns
// two PathSnapshot values ready for diffing.
func (d *DiffClient) TakeSnapshots(ctx context.Context, pathA, pathB string) (*PathSnapshot, *PathSnapshot, error) {
	type result struct {
		snap *PathSnapshot
		err  error
	}

	fetch := func(path string) result {
		secrets, err := d.taker.Take(ctx, path)
		if err != nil {
			return result{err: fmt.Errorf("snapshot %q: %w", path, err)}
		}
		return result{snap: &PathSnapshot{Path: path, Secrets: secrets}}
	}

	chA := make(chan result, 1)
	chB := make(chan result, 1)

	go func() { chA <- fetch(pathA) }()
	go func() { chB <- fetch(pathB) }()

	ra := <-chA
	if ra.err != nil {
		return nil, nil, ra.err
	}

	rb := <-chB
	if rb.err != nil {
		return nil, nil, rb.err
	}

	return ra.snap, rb.snap, nil
}
