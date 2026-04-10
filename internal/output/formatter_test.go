package output_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/vaultdiff/internal/diff"
	"github.com/user/vaultdiff/internal/output"
)

func sampleEntries() []diff.Entry {
	return []diff.Entry{
		{Key: "db_host", Type: diff.Unchanged, OldValue: "localhost", NewValue: "localhost"},
		{Key: "db_pass", Type: diff.Modified, OldValue: "old_secret", NewValue: "new_secret"},
		{Key: "new_key", Type: diff.Added, OldValue: "", NewValue: "added_val"},
		{Key: "old_key", Type: diff.Removed, OldValue: "removed_val", NewValue: ""},
	}
}

func TestWriteText(t *testing.T) {
	var buf bytes.Buffer
	f := output.NewFormatter(&buf, output.FormatText)
	if err := f.Write(sampleEntries()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "+ new_key = added_val") {
		t.Errorf("expected added line, got:\n%s", out)
	}
	if !strings.Contains(out, "- old_key = removed_val") {
		t.Errorf("expected removed line, got:\n%s", out)
	}
	if !strings.Contains(out, "~ db_pass: old_secret -> new_secret") {
		t.Errorf("expected modified line, got:\n%s", out)
	}
	if !strings.Contains(out, "  db_host = localhost") {
		t.Errorf("expected unchanged line, got:\n%s", out)
	}
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	f := output.NewFormatter(&buf, output.FormatJSON)
	if err := f.Write(sampleEntries()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"Key"`) && !strings.Contains(out, `"key"`) {
		t.Errorf("expected JSON output with keys, got:\n%s", out)
	}
}

func TestHasChanges(t *testing.T) {
	if !output.HasChanges(sampleEntries()) {
		t.Error("expected HasChanges to return true")
	}
	unchanged := []diff.Entry{
		{Key: "k", Type: diff.Unchanged, OldValue: "v", NewValue: "v"},
	}
	if output.HasChanges(unchanged) {
		t.Error("expected HasChanges to return false for unchanged entries")
	}
}

func TestParseFormat(t *testing.T) {
	cases := []struct {
		input   string
		want    output.Format
		wantErr bool
	}{
		{"text", output.FormatText, false},
		{"", output.FormatText, false},
		{"json", output.FormatJSON, false},
		{"JSON", output.FormatJSON, false},
		{"xml", "", true},
	}
	for _, tc := range cases {
		got, err := output.ParseFormat(tc.input)
		if tc.wantErr && err == nil {
			t.Errorf("ParseFormat(%q): expected error", tc.input)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("ParseFormat(%q): unexpected error: %v", tc.input, err)
		}
		if !tc.wantErr && got != tc.want {
			t.Errorf("ParseFormat(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
