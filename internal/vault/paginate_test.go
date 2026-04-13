package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// pagedTestServer returns a mock Vault server that handles LIST and GET for
// a flat set of secrets under secret/data/myapp.
func pagedTestServer(t *testing.T, keys []string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "LIST" || r.URL.Query().Get("list") == "true" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"keys": keys},
			})
			return
		}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		key := parts[len(parts)-1]
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"value": key + "_val"},
		})
	}))
}

func TestNewPaginatedClient_DefaultPageSize(t *testing.T) {
	c := NewPaginatedClient(newTestClient(t, httptest.NewServer(http.NewServeMux())), 0)
	if c.pageSize != 20 {
		t.Fatalf("expected default page size 20, got %d", c.pageSize)
	}
}

func TestPaginatedClient_EachPage_AllKeys(t *testing.T) {
	keys := []string{"a", "b", "c", "d", "e"}
	srv := pagedTestServer(t, keys)
	defer srv.Close()

	inner := newTestClient(t, srv)
	client := NewPaginatedClient(inner, 2)

	var collected []string
	err := client.EachPage(context.Background(), "secret", "", func(page map[string]map[string]interface{}) error {
		for k := range page {
			collected = append(collected, k)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(collected) != len(keys) {
		t.Errorf("expected %d keys, got %d", len(keys), len(collected))
	}
}

func TestPaginatedClient_EachPage_StopsOnCallbackError(t *testing.T) {
	keys := []string{"x", "y", "z"}
	srv := pagedTestServer(t, keys)
	defer srv.Close()

	client := NewPaginatedClient(newTestClient(t, srv), 1)

	pages := 0
	err := client.EachPage(context.Background(), "secret", "", func(_ map[string]map[string]interface{}) error {
		pages++
		return fmt.Errorf("stop")
	})
	if err == nil {
		t.Fatal("expected error from callback, got nil")
	}
	if pages != 1 {
		t.Errorf("expected 1 page processed, got %d", pages)
	}
}

func TestPaginatedClient_EachPage_EmptyKeys(t *testing.T) {
	srv := pagedTestServer(t, []string{})
	defer srv.Close()

	client := NewPaginatedClient(newTestClient(t, srv), 5)

	called := false
	err := client.EachPage(context.Background(), "secret", "", func(_ map[string]map[string]interface{}) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("callback should not be called for empty key list")
	}
}

// TestPaginatedClient_EachPage_ExactPageBoundary verifies that pagination
// works correctly when the total number of keys is an exact multiple of
// the page size (no partial last page).
func TestPaginatedClient_EachPage_ExactPageBoundary(t *testing.T) {
	keys := []string{"k1", "k2", "k3", "k4"}
	srv := pagedTestServer(t, keys)
	defer srv.Close()

	client := NewPaginatedClient(newTestClient(t, srv), 2)

	var collected []string
	err := client.EachPage(context.Background(), "secret", "", func(page map[string]map[string]interface{}) error {
		for k := range page {
			collected = append(collected, k)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(collected) != len(keys) {
		t.Errorf("expected %d keys, got %d", len(keys), len(collected))
	}
}
