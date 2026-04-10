package diff_test

import (
	"testing"

	"github.com/yourusername/vaultdiff/internal/diff"
)

func TestCompare_Added(t *testing.T) {
	src := map[string]string{}
	tgt := map[string]string{"foo": "bar"}

	result := diff.Compare(src, tgt)
	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result.Entries))
	}
	if result.Entries[0].Change != diff.Added {
		t.Errorf("expected Added, got %s", result.Entries[0].Change)
	}
}

func TestCompare_Removed(t *testing.T) {
	src := map[string]string{"foo": "bar"}
	tgt := map[string]string{}

	result := diff.Compare(src, tgt)
	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result.Entries))
	}
	if result.Entries[0].Change != diff.Removed {
		t.Errorf("expected Removed, got %s", result.Entries[0].Change)
	}
}

func TestCompare_Modified(t *testing.T) {
	src := map[string]string{"key": "old"}
	tgt := map[string]string{"key": "new"}

	result := diff.Compare(src, tgt)
	if result.Entries[0].Change != diff.Modified {
		t.Errorf("expected Modified, got %s", result.Entries[0].Change)
	}
	if result.Entries[0].OldValue != "old" || result.Entries[0].NewValue != "new" {
		t.Errorf("unexpected values: old=%s new=%s", result.Entries[0].OldValue, result.Entries[0].NewValue)
	}
}

func TestCompare_Unchanged(t *testing.T) {
	src := map[string]string{"key": "val"}
	tgt := map[string]string{"key": "val"}

	result := diff.Compare(src, tgt)
	if result.Entries[0].Change != diff.Unchanged {
		t.Errorf("expected Unchanged, got %s", result.Entries[0].Change)
	}
	if result.HasChanges() {
		t.Error("expected HasChanges to return false")
	}
}

func TestCompare_SortedOutput(t *testing.T) {
	src := map[string]string{"z": "1", "a": "2", "m": "3"}
	tgt := map[string]string{"z": "1", "a": "2", "m": "3"}

	result := diff.Compare(src, tgt)
	keys := make([]string, len(result.Entries))
	for i, e := range result.Entries {
		keys[i] = e.Key
	}
	if keys[0] != "a" || keys[1] != "m" || keys[2] != "z" {
		t.Errorf("entries not sorted: %v", keys)
	}
}
