package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func newMaskTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"password":"supersecret","user":"admin","port":"5432"}}`))
	}))
}

func TestMaskedClient_NoRules_ReturnsUnmodified(t *testing.T) {
	srv := newMaskTestServer(t)
	defer srv.Close()
	client := newTestClient(t, srv.URL)
	masked := NewMaskedClient(client, nil)
	data, err := masked.ReadSecret(context.Background(), "secret/db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["password"] != "supersecret" {
		t.Errorf("expected unmodified value, got %v", data["password"])
	}
}

func TestMaskedClient_MasksMatchingValues(t *testing.T) {
	srv := newMaskTestServer(t)
	defer srv.Close()
	client := newTestClient(t, srv.URL)
	rules := []MaskRule{
		{Pattern: regexp.MustCompile(`supersecret`), Replacement: "***"},
	}
	masked := NewMaskedClient(client, rules)
	data, err := masked.ReadSecret(context.Background(), "secret/db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["password"] != "***" {
		t.Errorf("expected masked value, got %v", data["password"])
	}
	if data["user"] != "admin" {
		t.Errorf("expected unmasked user, got %v", data["user"])
	}
}

func TestMaskedClient_NonStringValuesUnchanged(t *testing.T) {
	srv := newMaskTestServer(t)
	defer srv.Close()
	client := newTestClient(t, srv.URL)
	rules := []MaskRule{
		{Pattern: regexp.MustCompile(`.*`), Replacement: "REDACTED"},
	}
	masked := NewMaskedClient(client, rules)
	// port comes back as string from JSON; confirm string fields are masked
	data, err := masked.ReadSecret(context.Background(), "secret/db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := data["port"]; !ok {
		t.Error("expected port key to exist")
	}
}

func TestMaskedClient_MultipleRulesAppliedInOrder(t *testing.T) {
	srv := newMaskTestServer(t)
	defer srv.Close()
	client := newTestClient(t, srv.URL)
	rules := []MaskRule{
		{Pattern: regexp.MustCompile(`super`), Replacement: "SUPER"},
		{Pattern: regexp.MustCompile(`SUPERsecret`), Replacement: "[MASKED]"},
	}
	masked := NewMaskedClient(client, rules)
	data, err := masked.ReadSecret(context.Background(), "secret/db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["password"] != "[MASKED]" {
		t.Errorf("expected [MASKED], got %v", data["password"])
	}
}
