package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockVaultServer(t *testing.T, path string, payload map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"data": payload}); err != nil {
			t.Fatalf("encoding mock response: %v", err)
		}
	}))
}

func TestNewClient_DefaultAddress(t *testing.T) {
	t.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")
	t.Setenv("VAULT_TOKEN", "test-token")

	client, err := NewClient("")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_ExplicitAddress(t *testing.T) {
	t.Setenv("VAULT_TOKEN", "test-token")

	client, err := NewClient("http://localhost:9200")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestReadSecret_KVv1(t *testing.T) {
	server := mockVaultServer(t, "/v1/secret/myapp", map[string]interface{}{
		"username": "admin",
		"password": "s3cr3t",
	})
	defer server.Close()

	t.Setenv("VAULT_TOKEN", "test-token")
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}

	secrets, err := client.ReadSecret("secret/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secrets["username"] != "admin" {
		t.Errorf("expected username=admin, got %q", secrets["username"])
	}
	if secrets["password"] != "s3cr3t" {
		t.Errorf("expected password=s3cr3t, got %q", secrets["password"])
	}
}

func TestReadSecret_KVv2(t *testing.T) {
	server := mockVaultServer(t, "/v1/secret/data/myapp", map[string]interface{}{
		"data": map[string]interface{}{
			"api_key": "abc123",
		},
	})
	defer server.Close()

	t.Setenv("VAULT_TOKEN", "test-token")
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("creating client: %v", err)
	}

	secrets, err := client.ReadSecret("secret/data/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secrets["api_key"] != "abc123" {
		t.Errorf("expected api_key=abc123, got %q", secrets["api_key"])
	}
}
