package cmd

import (
	"bytes"
	"testing"
)

func TestExecute_MissingArgs(t *testing.T) {
	// Capture cobra error for missing arguments
	rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for missing arguments, got nil")
	}
}

func TestExecute_TooManyArgs(t *testing.T) {
	rootCmd.SetArgs([]string{"a", "b", "c"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for too many arguments, got nil")
	}
}

func TestInit_DefaultFlagValues(t *testing.T) {
	// Reset flags to defaults
	vaultAddr = ""
	vaultToken = ""
	format = "text"
	mount = "secret"

	if format != "text" {
		t.Errorf("expected default format 'text', got %q", format)
	}
	if mount != "secret" {
		t.Errorf("expected default mount 'secret', got %q", mount)
	}
}

func TestRootCmd_HelpOutput(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"--help"})
	// Help should not return an error
	_ = rootCmd.Execute()
	if buf.Len() == 0 {
		t.Error("expected help output, got empty buffer")
	}
}

func TestExecute_InvalidFormat(t *testing.T) {
	// Providing an unsupported format value should return an error
	rootCmd.SetArgs([]string{"--format", "xml", "path/a", "path/b"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for invalid format 'xml', got nil")
	}
}
