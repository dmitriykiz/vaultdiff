package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newPrefixTestServer(t *testing.T, expectedPath string, response string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/"+expectedPath {
			t.Errorf("unexpected path: got %q, want %q", r.URL.Path, "/v1/"+expectedPath)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
}

func TestPrefixedClient_PrependsPrefixToPath(t *testing.T) {
	srv := newPrefixTestServer(t, "team-a/secret/db", `{"data":{"user":"admin"}}`)
	defer srv.Close()

	inner := newTestClient(t, srv.URL)
	client := NewPrefixedClient(inner, "team-a")

	data, err := client.ReadSecret(context.Background(), "secret/db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["user"] != "admin" {
		t.Errorf("expected user=admin, got %v", data["user"])
	}
}

func TestPrefixedClient_EmptyPrefix_PassesThrough(t *testing.T) {
	srv := newPrefixTestServer(t, "secret/db", `{"data":{"key":"value"}}`)
	defer srv.Close()

	inner := newTestClient(t, srv.URL)
	client := NewPrefixedClient(inner, "")

	data, err := client.ReadSecret(context.Background(), "secret/db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["key"] != "value" {
		t.Errorf("expected key=value, got %v", data["key"])
	}
}

func TestPrefixedClient_TrimsSlashes(t *testing.T) {
	srv := newPrefixTestServer(t, "ns/secret/token", `{"data":{"val":"x"}}`)
	defer srv.Close()

	inner := newTestClient(t, srv.URL)
	// prefix and path both carry extra slashes — should be normalised
	client := NewPrefixedClient(inner, "/ns/")

	_, err := client.ReadSecret(context.Background(), "/secret/token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPrefixedClient_Prefix_ReturnsNormalisedPrefix(t *testing.T) {
	inner := newTestClient(t, "http://localhost")
	client := NewPrefixedClient(inner, "/my-team/")
	if client.Prefix() != "my-team" {
		t.Errorf("expected normalised prefix 'my-team', got %q", client.Prefix())
	}
}

func TestPrefixedClient_EmptyPath_ReturnsPrefix(t *testing.T) {
	srv := newPrefixTestServer(t, "root", `{"data":{"k":"v"}}`)
	defer srv.Close()

	inner := newTestClient(t, srv.URL)
	client := NewPrefixedClient(inner, "root")

	_, err := client.ReadSecret(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
