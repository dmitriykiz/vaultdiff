package vault

import (
	"context"
	"errors"
	"testing"
)

// stubReader is a minimal SecretReader for testing.
type stubReader struct {
	data map[string]string
	err  error
}

func (s *stubReader) ReadSecret(_ context.Context, _ string) (map[string]string, error) {
	return s.data, s.err
}

func TestAuditedClient_RecordsSuccessfulRead(t *testing.T) {
	stub := &stubReader{data: map[string]string{"key": "val"}}
	log := NewAuditLog(nil)
	client := NewAuditedClient(stub, log)

	data, err := client.ReadSecret(context.Background(), "secret/foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["key"] != "val" {
		t.Errorf("unexpected data: %v", data)
	}
	if log.Len() != 1 {
		t.Fatalf("expected 1 audit event, got %d", log.Len())
	}
	events := log.Events()
	if events[0].Path != "secret/foo" {
		t.Errorf("expected path secret/foo, got %s", events[0].Path)
	}
	if events[0].Err != nil {
		t.Errorf("expected nil error in event")
	}
}

func TestAuditedClient_RecordsErrorRead(t *testing.T) {
	expectedErr := errors.New("not found")
	stub := &stubReader{err: expectedErr}
	log := NewAuditLog(nil)
	client := NewAuditedClient(stub, log)

	_, err := client.ReadSecret(context.Background(), "secret/missing")
	if err == nil {
		t.Fatal("expected error")
	}
	events := log.Events()
	if events[0].Err == nil {
		t.Error("expected error to be recorded in audit log")
	}
}

func TestAuditedClient_NilLogDefaultsToDiscard(t *testing.T) {
	stub := &stubReader{data: map[string]string{}}
	client := NewAuditedClient(stub, nil)
	_, err := client.ReadSecret(context.Background(), "secret/x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.Log().Len() != 1 {
		t.Error("expected event to be recorded even with nil log")
	}
}

func TestAuditedClient_LogAccessor(t *testing.T) {
	log := NewAuditLog(nil)
	client := NewAuditedClient(&stubReader{}, log)
	if client.Log() != log {
		t.Error("Log() should return the same AuditLog instance")
	}
}
