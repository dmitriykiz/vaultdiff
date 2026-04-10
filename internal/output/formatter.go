package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/user/vaultdiff/internal/diff"
)

// Format represents the output format type.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Formatter writes diff results to an output writer.
type Formatter struct {
	Writer io.Writer
	Format Format
}

// NewFormatter creates a new Formatter with the given writer and format.
func NewFormatter(w io.Writer, format Format) *Formatter {
	return &Formatter{Writer: w, Format: format}
}

// Write outputs the diff entries using the configured format.
func (f *Formatter) Write(entries []diff.Entry) error {
	switch f.Format {
	case FormatJSON:
		return f.writeJSON(entries)
	default:
		return f.writeText(entries)
	}
}

func (f *Formatter) writeText(entries []diff.Entry) error {
	for _, e := range entries {
		var line string
		switch e.Type {
		case diff.Added:
			line = fmt.Sprintf("+ %s = %s", e.Key, e.NewValue)
		case diff.Removed:
			line = fmt.Sprintf("- %s = %s", e.Key, e.OldValue)
		case diff.Modified:
			line = fmt.Sprintf("~ %s: %s -> %s", e.Key, e.OldValue, e.NewValue)
		case diff.Unchanged:
			line = fmt.Sprintf("  %s = %s", e.Key, e.OldValue)
		}
		if _, err := fmt.Fprintln(f.Writer, line); err != nil {
			return err
		}
	}
	return nil
}

func (f *Formatter) writeJSON(entries []diff.Entry) error {
	enc := json.NewEncoder(f.Writer)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}

// HasChanges returns true if any entry represents a change.
func HasChanges(entries []diff.Entry) bool {
	for _, e := range entries {
		if e.Type != diff.Unchanged {
			return true
		}
	}
	return false
}

// ParseFormat converts a string to a Format, returning an error if unknown.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "text", "":
		return FormatText, nil
	case "json":
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("unknown format %q: must be \"text\" or \"json\"", s)
	}
}
