package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newScopeTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"key":"value"}}`))
	}))
}

func TestScopedClient_AllowsMatchingPath(t *testing.T) {
	srv := newScopeTestServer(t)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	scoped := NewScopedClient(client, "secret/myapp")

	data, err := scoped.ReadSecret(context.Background(), "secret/myapp/db")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if data["key"] != "value" {
		t.Errorf("expected key=value, got %v", data)
	}
}

func TestScopedClient_AllowsExactPrefixPath(t *testing.T) {
	srv := newScopeTestServer(t)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	scoped := NewScopedClient(client, "secret/myapp")

	_, err := scoped.ReadSecret(context.Background(), "secret/myapp")
	if err != nil {
		t.Fatalf("expected exact prefix path to be allowed, got %v", err)
	}
}

func TestScopedClient_BlocksOutOfScopePath(t *testing.T) {
	srv := newScopeTestServer(t)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	scoped := NewScopedClient(client, "secret/myapp")

	_, err := scoped.ReadSecret(context.Background(), "secret/otherapp/creds")
	if err == nil {
		t.Fatal("expected error for out-of-scope path, got nil")
	}
}

func TestScopedClient_EmptyPrefix_AllowsAll(t *testing.T) {
	srv := newScopeTestServer(t)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	scoped := NewScopedClient(client, "")

	_, err := scoped.ReadSecret(context.Background(), "any/path/at/all")
	if err != nil {
		t.Fatalf("empty prefix should allow all paths, got %v", err)
	}
}

func TestScopedClient_Prefix_ReturnsConfiguredValue(t *testing.T) {
	client := newTestClient(t, "http://127.0.0.1")
	scoped := NewScopedClient(client, "/secret/myapp/")

	if scoped.Prefix() != "secret/myapp" {
		t.Errorf("expected trimmed prefix 'secret/myapp', got %q", scoped.Prefix())
	}
}
