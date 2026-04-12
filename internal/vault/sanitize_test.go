package vault

import (
	"regexp"
	"testing"
)

func TestSanitizer_NilSanitizer_ReturnsOriginal(t *testing.T) {
	var s *Sanitizer
	data := map[string]interface{}{"key": "value"}
	result := s.Apply(data)
	if result["key"] != "value" {
		t.Fatalf("expected 'value', got %v", result["key"])
	}
}

func TestSanitizer_NoRules_ReturnsOriginal(t *testing.T) {
	s := NewSanitizer(nil)
	data := map[string]interface{}{"password": "secret123"}
	result := s.Apply(data)
	if result["password"] != "secret123" {
		t.Fatalf("expected 'secret123', got %v", result["password"])
	}
}

func TestSanitizer_ReplacesMatchingKey(t *testing.T) {
	s := NewSanitizer([]SanitizeRule{
		{KeyPattern: regexp.MustCompile(`password`), Replacement: "[redacted]"},
	})
	data := map[string]interface{}{"password": "hunter2", "username": "admin"}
	result := s.Apply(data)
	if result["password"] != "[redacted]" {
		t.Fatalf("expected '[redacted]', got %v", result["password"])
	}
	if result["username"] != "admin" {
		t.Fatalf("expected 'admin', got %v", result["username"])
	}
}

func TestSanitizer_CaseInsensitiveMatch(t *testing.T) {
	s := NewSanitizer([]SanitizeRule{
		{KeyPattern: regexp.MustCompile(`token`), Replacement: "[sanitized]"},
	})
	data := map[string]interface{}{"API_TOKEN": "tok_abc123"}
	result := s.Apply(data)
	if result["API_TOKEN"] != "[sanitized]" {
		t.Fatalf("expected '[sanitized]', got %v", result["API_TOKEN"])
	}
}

func TestSanitizer_DoesNotMutateOriginal(t *testing.T) {
	s := NewSanitizer(DefaultSanitizeRules())
	original := map[string]interface{}{"password": "original"}
	_ = s.Apply(original)
	if original["password"] != "original" {
		t.Fatal("original data was mutated")
	}
}

func TestDefaultSanitizeRules_MasksCommonKeys(t *testing.T) {
	s := NewSanitizer(DefaultSanitizeRules())
	cases := []string{"password", "api_key", "token", "secret", "private_key", "credential"}
	for _, key := range cases {
		data := map[string]interface{}{key: "sensitive"}
		result := s.Apply(data)
		if result[key] == "sensitive" {
			t.Errorf("key %q was not sanitized", key)
		}
	}
}
