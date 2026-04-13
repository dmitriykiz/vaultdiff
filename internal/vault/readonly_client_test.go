package vault

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newReadOnlyTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"key":"value"}}`))
	}))
}

func TestReadOnlyClient_AllowsUnprotectedPath(t *testing.T) {
	srv := newReadOnlyTestServer(t)
	defer srv.Close()

	inner := newTestClient(t, srv.URL)
	client := NewReadOnlyClient(inner, []string{"secret/admin"})

	_, err := client.ReadSecret(context.Background(), "secret/app/config")
	if err != nil {
		t.Fatalf("expected no error for allowed path, got: %v", err)
	}
}

func TestReadOnlyClient_BlocksProtectedPath(t *testing.T) {
	srv := newReadOnlyTestServer(t)
	defer srv.Close()

	inner := newTestClient(t, srv.URL)
	client := NewReadOnlyClient(inner, []string{"secret/admin"})

	_, err := client.ReadSecret(context.Background(), "secret/admin/token")
	if err == nil {
		t.Fatal("expected error for protected path, got nil")
	}
}

func TestReadOnlyClient_NoProtectedPaths_AllowsAll(t *testing.T) {
	srv := newReadOnlyTestServer(t)
	defer srv.Close()

	inner := newTestClient(t, srv.URL)
	client := NewReadOnlyClient(inner, nil)

	_, err := client.ReadSecret(context.Background(), "secret/anything")
	if err != nil {
		t.Fatalf("expected no error with no protected paths, got: %v", err)
	}
}

func TestReadOnlyClient_ExactPrefixMatch(t *testing.T) {
	srv := newReadOnlyTestServer(t)
	defer srv.Close()

	inner := newTestClient(t, srv.URL)
	client := NewReadOnlyClient(inner, []string{"secret/admin"})

	// exact match on prefix boundary
	_, err := client.ReadSecret(context.Background(), "secret/admin")
	if err == nil {
		t.Fatal("expected error for exact protected prefix, got nil")
	}
}

func TestReadOnlyClient_Protected_ReturnsCopy(t *testing.T) {
	protected := []string{"secret/admin", "auth/"}
	client := NewReadOnlyClient(nil, protected)

	got := client.Protected()
	if len(got) != len(protected) {
		t.Fatalf("expected %d protected paths, got %d", len(protected), len(got))
	}
	// mutating returned slice must not affect client
	got[0] = "mutated"
	if client.Protected()[0] != protected[0] {
		t.Fatal("Protected() returned a reference to internal slice")
	}
}

func TestReadOnlyClient_InnerError_Propagated(t *testing.T) {
	inner := &errorReader{err: errors.New("vault unavailable")}
	client := NewReadOnlyClient(inner, nil)

	_, err := client.ReadSecret(context.Background(), "secret/app")
	if err == nil || err.Error() != "vault unavailable" {
		t.Fatalf("expected inner error to propagate, got: %v", err)
	}
}

// errorReader is a minimal SecretReader that always returns an error.
type errorReader struct{ err error }

func (e *errorReader) ReadSecret(_ context.Context, _ string) (map[string]interface{}, error) {
	return nil, e.err
}
