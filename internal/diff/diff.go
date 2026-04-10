package diff

import "sort"

// ChangeType represents the type of change between two secret values.
type ChangeType string

const (
	Added    ChangeType = "added"
	Removed  ChangeType = "removed"
	Modified ChangeType = "modified"
	Unchanged ChangeType = "unchanged"
)

// Entry represents a single key-level diff result.
type Entry struct {
	Key      string
	Change   ChangeType
	OldValue string
	NewValue string
}

// Result holds the full diff between two secret maps.
type Result struct {
	Entries []Entry
}

// HasChanges returns true if any entry is not unchanged.
func (r *Result) HasChanges() bool {
	for _, e := range r.Entries {
		if e.Change != Unchanged {
			return true
		}
	}
	return false
}

// Compare computes the diff between two maps of secret key/value pairs.
func Compare(source, target map[string]string) *Result {
	seen := make(map[string]bool)
	var entries []Entry

	for key, srcVal := range source {
		seen[key] = true
		tgtVal, exists := target[key]
		switch {
		case !exists:
			entries = append(entries, Entry{Key: key, Change: Removed, OldValue: srcVal})
		case srcVal != tgtVal:
			entries = append(entries, Entry{Key: key, Change: Modified, OldValue: srcVal, NewValue: tgtVal})
		default:
			entries = append(entries, Entry{Key: key, Change: Unchanged, OldValue: srcVal, NewValue: tgtVal})
		}
	}

	for key, tgtVal := range target {
		if !seen[key] {
			entries = append(entries, Entry{Key: key, Change: Added, NewValue: tgtVal})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Key < entries[j].Key
	})

	return &Result{Entries: entries}
}
