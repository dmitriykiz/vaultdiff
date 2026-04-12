package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newSnapshotTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/secret/metadata/myapp":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"keys": []string{"db", "api"},
				},
			})
		case "/v1/secret/data/myapp/db":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"host": "localhost"},
				},
			})
		case "/v1/secret/data/myapp/api":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"data": map[string]interface{}{"key": "abc123"},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestSnapshotTaker_Take_ReturnsAllSecrets(t *testing.T) {
	srv := newSnapshotTestServer(t)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	taker := NewSnapshotTaker(client, 2)

	snap, err := taker.Take(context.Background(), "secret", "myapp", EngineKVv2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if snap == nil {
		t.Fatal("expected snapshot, got nil")
	}
	if len(snap.Secrets) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(snap.Secrets))
	}
	if snap.Path != "myapp" {
		t.Errorf("expected path myapp, got %s", snap.Path)
	}
	if snap.CapturedAt.IsZero() {
		t.Error("expected CapturedAt to be set")
	}
}

func TestSnapshotTaker_DefaultConcurrency(t *testing.T) {
	srv := newSnapshotTestServer(t)
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	taker := NewSnapshotTaker(client, 0)

	if taker.workers != DefaultConcurrency {
		t.Errorf("expected default concurrency %d, got %d", DefaultConcurrency, taker.workers)
	}
}

func TestSnapshotTaker_Take_EmptyMount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"keys": []string{}},
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	taker := NewSnapshotTaker(client, 1)

	snap, err := taker.Take(context.Background(), "secret", "", EngineKVv2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snap.Secrets) != 0 {
		t.Errorf("expected empty snapshot, got %d entries", len(snap.Secrets))
	}
}
