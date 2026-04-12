package vault_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/your-org/vaultdiff/internal/vault"
)

// staticReader is a minimal SecretReader that returns a fixed map.
type staticReader struct {
	data map[string]interface{}
	err  error
}

func (s *staticReader) ReadSecret(_ context.Context, _ string) (map[string]interface{}, error) {
	return s.data, s.err
}

func TestTransformedClient_NoRules_PassesThrough(t *testing.T) {
	inner := &staticReader{data: map[string]interface{}{"key": "value"}}
	client := vault.NewTransformedClient(inner, vault.NewTransformer(nil))

	got, err := client.ReadSecret(context.Background(), "secret/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["key"] != "value" {
		t.Errorf("expected 'value', got %v", got["key"])
	}
}

func TestTransformedClient_AppliesTransform(t *testing.T) {
	inner := &staticReader{
		data: map[string]interface{}{
			"db_password": "supersecret",
			"db_host":     "localhost",
		},
	}
	rules := []vault.TransformRule{vault.TruncateTransform("password", 5)}
	client := vault.NewTransformedClient(inner, vault.NewTransformer(rules))

	got, err := client.ReadSecret(context.Background(), "secret/db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["db_password"] != "super..." {
		t.Errorf("expected 'super...', got %v", got["db_password"])
	}
	if got["db_host"] != "localhost" {
		t.Errorf("host should be unchanged, got %v", got["db_host"])
	}
}

func TestTransformedClient_PropagatesError(t *testing.T) {
	expected := errors.New("vault unavailable")
	inner := &staticReader{err: expected}
	client := vault.NewTransformedClient(inner, vault.NewTransformer(nil))

	_, err := client.ReadSecret(context.Background(), "secret/missing")
	if !errors.Is(err, expected) {
		t.Errorf("expected propagated error, got %v", err)
	}
}

func TestTransformedClient_NilTransformer_PassesThrough(t *testing.T) {
	inner := &staticReader{data: map[string]interface{}{"api_key": "abc123"}}
	client := vault.NewTransformedClient(inner, nil)

	got, err := client.ReadSecret(context.Background(), "secret/api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["api_key"] != "abc123" {
		t.Errorf("expected 'abc123', got %v", got["api_key"])
	}
}

func TestTransformedClient_UpperCaseTransform(t *testing.T) {
	inner := &staticReader{data: map[string]interface{}{"env_region": "us-east-1"}}
	rules := []vault.TransformRule{vault.UpperCaseTransform("region")}
	client := vault.NewTransformedClient(inner, vault.NewTransformer(rules))

	got, err := client.ReadSecret(context.Background(), "secret/env")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["env_region"] != strings.ToUpper("us-east-1") {
		t.Errorf("expected upper-cased region, got %v", got["env_region"])
	}
}
