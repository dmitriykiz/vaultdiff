package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestBuildListPath_KVv1(t *testing.T) {
	got := buildListPath("secret/apps", EngineKVv1)
	if got != "secret/apps" {
		t.Errorf("got %q, want %q", got, "secret/apps")
	}
}

func TestBuildListPath_KVv2_WithSubpath(t *testing.T) {
	got := buildListPath("secret/apps", EngineKVv2)
	want := "secret/metadata/apps"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuildListPath_KVv2_MountOnly(t *testing.T) {
	got := buildListPath("secret", EngineKVv2)
	want := "secret/metadata"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestListSecrets_ReturnsKeys(t *testing.T) {
	expectedKeys := []string{"alpha", "beta", "gamma"}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "LIST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"keys": expectedKeys},
		})
	}))
	defer ts.Close()

	c, err := NewClient(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	keys, err := c.ListSecrets(context.Background(), "secret/apps", EngineKVv1)
	if err != nil {
		t.Fatalf("ListSecrets: %v", err)
	}
	if !reflect.DeepEqual(keys, expectedKeys) {
		t.Errorf("got %v, want %v", keys, expectedKeys)
	}
}

func TestListSecrets_EmptyMount(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{}})
	}))
	defer ts.Close()

	c, err := NewClient(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	keys, err := c.ListSecrets(context.Background(), "secret", EngineKVv1)
	if err != nil {
		t.Fatalf("ListSecrets: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected empty keys, got %v", keys)
	}
}
