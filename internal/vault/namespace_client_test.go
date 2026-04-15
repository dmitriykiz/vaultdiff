package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newNamespaceTestServer(t *testing.T, wantPath string, data map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/"+wantPath {
			http.Error(w, "unexpected path: "+r.URL.Path, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
	}))
}

func TestNamespacedClient_PrependsSingleNamespace(t *testing.T) {
	srv := newNamespaceTestServer(t, "acme/secret/db", map[string]interface{}{"user": "admin"})
	defer srv.Close()

	inner := newTestClient(t, srv.URL)
	client := NewNamespacedClient(inner, "acme")

	got, err := client.ReadSecret(context.Background(), "secret/db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["user"] != "admin" {
		t.Errorf("expected user=admin, got %v", got["user"])
	}
}

func TestNamespacedClient_TrimsSlashesFromNamespace(t *testing.T) {
	srv := newNamespaceTestServer(t, "corp/secret/key", map[string]interface{}{"val": "x"})
	defer srv.Close()

	inner := newTestClient(t, srv.URL)
	client := NewNamespacedClient(inner, "/corp/")

	if client.Namespace() != "corp" {
		t.Errorf("expected namespace=corp, got %q", client.Namespace())
	}

	_, err := client.ReadSecret(context.Background(), "secret/key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNamespacedClient_EmptyNamespace_PassesThrough(t *testing.T) {
	srv := newNamespaceTestServer(t, "secret/plain", map[string]interface{}{"k": "v"})
	defer srv.Close()

	inner := newTestClient(t, srv.URL)
	client := NewNamespacedClient(inner, "")

	_, err := client.ReadSecret(context.Background(), "secret/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNamespacedClient_EmptyPath_ReturnsError(t *testing.T) {
	inner := newTestClient(t, "http://127.0.0.1")
	client := NewNamespacedClient(inner, "ns")

	_, err := client.ReadSecret(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty path, got nil")
	}
}

func TestNamespacedClient_NestedNamespace(t *testing.T) {
	srv := newNamespaceTestServer(t, "org/team/secret/cfg", map[string]interface{}{"env": "prod"})
	defer srv.Close()

	inner := newTestClient(t, srv.URL)
	client := NewNamespacedClient(inner, "org/team")

	got, err := client.ReadSecret(context.Background(), "secret/cfg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["env"] != "prod" {
		t.Errorf("expected env=prod, got %v", got["env"])
	}
}
