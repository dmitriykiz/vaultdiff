package vault

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestAuditLog_RecordOK(t *testing.T) {
	var buf bytes.Buffer
	log := NewAuditLog(&buf)
	log.Record("read", "secret/foo", nil)

	if log.Len() != 1 {
		t.Fatalf("expected 1 event, got %d", log.Len())
	}
	events := log.Events()
	if events[0].Operation != "read" {
		t.Errorf("unexpected operation: %s", events[0].Operation)
	}
	if events[0].Path != "secret/foo" {
		t.Errorf("unexpected path: %s", events[0].Path)
	}
	if events[0].Err != nil {
		t.Errorf("expected nil error, got %v", events[0].Err)
	}
	if !strings.Contains(buf.String(), "secret/foo") {
		t.Errorf("expected path in output, got: %s", buf.String())
	}
}

func TestAuditLog_RecordError(t *testing.T) {
	var buf bytes.Buffer
	log := NewAuditLog(&buf)
	err := errors.New("permission denied")
	log.Record("read", "secret/bar", err)

	events := log.Events()
	if events[0].Err == nil {
		t.Fatal("expected error to be recorded")
	}
	if !strings.Contains(buf.String(), "permission denied") {
		t.Errorf("expected error in output, got: %s", buf.String())
	}
}

func TestAuditLog_NilWriter(t *testing.T) {
	log := NewAuditLog(nil)
	log.Record("list", "secret/", nil)
	if log.Len() != 1 {
		t.Fatalf("expected 1 event, got %d", log.Len())
	}
}

func TestAuditLog_MultipleEvents(t *testing.T) {
	log := NewAuditLog(nil)
	paths := []string{"a", "b", "c"}
	for _, p := range paths {
		log.Record("read", p, nil)
	}
	if log.Len() != 3 {
		t.Fatalf("expected 3 events, got %d", log.Len())
	}
}

func TestAuditLog_EventsReturnsCopy(t *testing.T) {
	log := NewAuditLog(nil)
	log.Record("read", "x", nil)
	events := log.Events()
	events[0].Path = "mutated"
	original := log.Events()
	if original[0].Path == "mutated" {
		t.Error("Events() should return a copy, not a reference")
	}
}

func TestAuditEvent_String_OK(t *testing.T) {
	log := NewAuditLog(nil)
	log.Record("read", "secret/test", nil)
	events := log.Events()
	s := events[0].String()
	if !strings.Contains(s, "read") || !strings.Contains(s, "secret/test") || !strings.Contains(s, "ok") {
		t.Errorf("unexpected String() output: %s", s)
	}
}
