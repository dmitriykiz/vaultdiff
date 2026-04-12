package vault

import (
	"context"
	"errors"
	"regexp"
	"testing"
)

type stubSanitizeReader struct {
	data map[string]interface{}
	err  error
}

func (s *stubSanitizeReader) ReadSecret(_ context.Context, _ string) (map[string]interface{}, error) {
	return s.data, s.err
}

func TestSanitizedClient_NoSanitizer_PassesThrough(t *testing.T) {
	stub := &stubSanitizeReader{data: map[string]interface{}{"password": "secret"}}
	client := NewSanitizedClient(stub, nil)
	result, err := client.ReadSecret(context.Background(), "secret/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["password"] != "secret" {
		t.Fatalf("expected 'secret', got %v", result["password"])
	}
}

func TestSanitizedClient_AppliesSanitizer(t *testing.T) {
	stub := &stubSanitizeReader{data: map[string]interface{}{"api_key": "abc123", "env": "prod"}}
	s := NewSanitizer([]SanitizeRule{
		{KeyPattern: regexp.MustCompile(`api_key`), Replacement: "[sanitized]"},
	})
	client := NewSanitizedClient(stub, s)
	result, err := client.ReadSecret(context.Background(), "secret/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["api_key"] != "[sanitized]" {
		t.Fatalf("expected '[sanitized]', got %v", result["api_key"])
	}
	if result["env"] != "prod" {
		t.Fatalf("expected 'prod', got %v", result["env"])
	}
}

func TestSanitizedClient_PropagatesError(t *testing.T) {
	expected := errors.New("vault unavailable")
	stub := &stubSanitizeReader{err: expected}
	client := NewSanitizedClient(stub, NewSanitizer(DefaultSanitizeRules()))
	_, err := client.ReadSecret(context.Background(), "secret/test")
	if !errors.Is(err, expected) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestSanitizedClient_DefaultRules_MasksPassword(t *testing.T) {
	stub := &stubSanitizeReader{data: map[string]interface{}{"password": "hunter2", "host": "localhost"}}
	client := NewSanitizedClient(stub, NewSanitizer(DefaultSanitizeRules()))
	result, err := client.ReadSecret(context.Background(), "secret/db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["password"] == "hunter2" {
		t.Fatal("password was not sanitized")
	}
	if result["host"] != "localhost" {
		t.Fatalf("expected 'localhost', got %v", result["host"])
	}
}
