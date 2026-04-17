package vault

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newValidateTestServer(data map[string]interface{}) *httptest.Server {
	return newSnapshotTestServer(data) // reuse helper that returns KV v2 JSON
}

func TestValidatedClient_PassesThroughWhenRulesPass(t *testing.T) {
	srv := newMaskTestServer(map[string]interface{}{"user": "alice", "token": "abc"})
	defer srv.Close()
	c, _ := NewClient(srv.URL, "test-token")
	vc := NewValidatedClient(c, RequireKeys("user", "token"))
	data, err := vc.ReadSecret(context.Background(), "secret/data/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["user"] != "alice" {
		t.Errorf("expected user=alice, got %v", data["user"])
	}
}

func TestValidatedClient_ErrorsOnMissingRequiredKey(t *testing.T) {
	srv := newMaskTestServer(map[string]interface{}{"user": "alice"})
	defer srv.Close()
	c, _ := NewClient(srv.URL, "test-token")
	vc := NewValidatedClient(c, RequireKeys("user", "password"))
	_, err := vc.ReadSecret(context.Background(), "secret/data/test")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if !contains(err.Error(), "password") {
		t.Errorf("expected error to mention missing key, got: %v", err)
	}
}

func TestValidatedClient_ErrorsOnEmptyValue(t *testing.T) {
	srv := newMaskTestServer(map[string]interface{}{"user": "alice", "token": ""})
	defer srv.Close()
	c, _ := NewClient(srv.URL, "test-token")
	vc := NewValidatedClient(c, DenyEmptyValues())
	_, err := vc.ReadSecret(context.Background(), "secret/data/test")
	if err == nil {
		t.Fatal("expected error for empty value")
	}
	if !contains(err.Error(), "token") {
		t.Errorf("expected error to mention key with empty value, got: %v", err)
	}
}

func TestValidatedClient_PropagatesInnerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()
	c, _ := NewClient(srv.URL, "test-token")
	vc := NewValidatedClient(c, RequireKeys("x"))
	_, err := vc.ReadSecret(context.Background(), "secret/data/test")
	if err == nil {
		t.Fatal("expected error from inner client")
	}
}

func TestValidatedClient_NoRules_PassesThrough(t *testing.T) {
	srv := newMaskTestServer(map[string]interface{}{"k": "v"})
	defer srv.Close()
	c, _ := NewClient(srv.URL, "test-token")
	vc := NewValidatedClient(c)
	data, err := vc.ReadSecret(context.Background(), "secret/data/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["k"] != "v" {
		t.Errorf("expected k=v")
	}
}

func TestRequireKeys_MultipleErrors(t *testing.T) {
	rule := RequireKeys("a", "b", "c")
	err := rule("path", map[string]interface{}{"a": 1})
	if err == nil {
		t.Fatal("expected error")
	}
	if !contains(err.Error(), "b") || !contains(err.Error(), "c") {
		t.Errorf("error should mention both missing keys: %v", err)
	}
}

func contains(s, sub string) bool {
	return errors.New(s).Error() != "" && len(s) >= len(sub) &&
		(s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
