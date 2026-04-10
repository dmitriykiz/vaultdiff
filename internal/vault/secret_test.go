package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = server.URL
	api, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("creating vault client: %v", err)
	}
	api.SetToken("test-token")
	return &Client{logical: api.Logical()}
}

func TestReadSecret_KVv1_ReturnsData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := map[string]interface{}{
			"data": map[string]interface{}{"foo": "bar", "baz": "qux"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	got, err := client.ReadSecret(context.Background(), "secret/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["foo"] != "bar" {
		t.Errorf("expected foo=bar, got %q", got["foo"])
	}
	if got["baz"] != "qux" {
		t.Errorf("expected baz=qux, got %q", got["baz"])
	}
}

func TestReadSecret_KVv2_UnwrapsNestedData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"data": map[string]interface{}{"username": "admin", "password": "secret"},
				"metadata": map[string]interface{}{"version": 3},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	client := newTestClient(t, server)
	got, err := client.ReadSecret(context.Background(), "secret/data/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["username"] != "admin" {
		t.Errorf("expected username=admin, got %q", got["username"])
	}
	if got["password"] != "secret" {
		t.Errorf("expected password=secret, got %q", got["password"])
	}
}

func TestReadSecret_NotFound_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errors":[]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server)
	_, err := client.ReadSecret(context.Background(), "secret/missing")
	if err == nil {
		t.Fatal("expected error for missing secret, got nil")
	}
}
