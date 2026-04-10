package vault

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
)

func TestRecurseSecrets_FlatMount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "LIST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		body := map[string]interface{}{
			"data": map[string]interface{}{
				"keys": []string{"alpha", "beta"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(body)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	paths, err := client.RecurseSecrets("secret", "", KVv1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"alpha", "beta"}
	if len(paths) != len(expected) {
		t.Fatalf("expected %d paths, got %d", len(expected), len(paths))
	}
	for i, p := range paths {
		if p != expected[i] {
			t.Errorf("path[%d]: expected %q, got %q", i, expected[i], p)
		}
	}
}

func TestRecurseSecrets_NestedDirectories(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "LIST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		var keys []string
		if calls == 0 {
			keys = []string{"top", "nested/"}
		} else {
			keys = []string{"child"}
		}
		calls++
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"keys": keys},
		})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	paths, err := client.RecurseSecrets("secret", "", KVv1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sort.Strings(paths)
	expected := []string{"nested/child", "top"}
	sort.Strings(expected)

	if len(paths) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, paths)
	}
	for i := range expected {
		if paths[i] != expected[i] {
			t.Errorf("index %d: expected %q, got %q", i, expected[i], paths[i])
		}
	}
}

func TestRecurseSecrets_EmptyMount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.RecurseSecrets("secret", "", KVv1)
	if err == nil {
		t.Fatal("expected error for empty/missing mount, got nil")
	}
}
