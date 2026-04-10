package diff

import "sort"

// ChangeType describes the kind of change for a secret key.
type ChangeType string

const (
	Added     ChangeType = "added"
	Removed   ChangeType = "removed"
	Modified  ChangeType = "modified"
	Unchanged ChangeType = "unchanged"
)

// Entry represents a single key-level diff result.
type Entry struct {
	Key      string     `json:"key"`
	Type     ChangeType `json:"type"`
	OldValue string     `json:"old_value,omitempty"`
	NewValue string     `json:"new_value,omitempty"`
}

// Compare computes the diff between two maps of secret key/value pairs.
// The returned slice is sorted alphabetically by key.
func Compare(src, dst map[string]string) []Entry {
	keySet := make(map[string]struct{})
	for k := range src {
		keySet[k] = struct{}{}
	}
	for k := range dst {
		keySet[k] = struct{}{}
	}

	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	entries := make([]Entry, 0, len(keys))
	for _, k := range keys {
		srcVal, inSrc := src[k]
		dstVal, inDst := dst[k]

		switch {
		case inSrc && !inDst:
			entries = append(entries, Entry{Key: k, Type: Removed, OldValue: srcVal})
		case !inSrc && inDst:
			entries = append(entries, Entry{Key: k, Type: Added, NewValue: dstVal})
		case srcVal != dstVal:
			entries = append(entries, Entry{Key: k, Type: Modified, OldValue: srcVal, NewValue: dstVal})
		default:
			entries = append(entries, Entry{Key: k, Type: Unchanged, OldValue: srcVal, NewValue: dstVal})
		}
	}
	return entries
}
