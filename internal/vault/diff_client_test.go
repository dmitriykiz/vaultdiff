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

func newDiffTestServer(mounts map[string][]string, secrets map[string]map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/")

		// sys/mounts — return kv-v1 for all mounts
		if r.URL.Path == "/v1/sys/mounts" {
			mountsResp := map[string]interface{}{}
			for mount := range mounts {
				mountsResp[mount+"/"] = map[string]interface{}{"type": "kv", "options": map[string]interface{}{"version": "1"}}
			}
			_ = json.NewEncoder(w).Encode(mountsResp)
			return
		}

		// LIST request
		if r.Method == "LIST" {
			for mount, keys := range mounts {
				if path == mount {
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"keys": keys}})
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// GET secret
		if data, ok := secrets[path]; ok {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestDiffClient_TakeSnapshots_ReturnsBothPaths(t *testing.T) {
	server := newDiffTestServer(
		map[string][]string{
			"secret/a": {"key1"},
			"secret/b": {"key2"},
		},
		map[string]map[string]interface{}{
			"secret/a/key1": {"val": "alpha"},
			"secret/b/key2": {"val": "beta"},
		},
	)
	defer server.Close()

	client := newTestClient(t, server.URL)
	dc := NewDiffClient(client, 2)

	snA, snB, err := dc.TakeSnapshots(context.Background(), "secret/a", "secret/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snA.Path != "secret/a" {
		t.Errorf("expected path secret/a, got %s", snA.Path)
	}
	if snB.Path != "secret/b" {
		t.Errorf("expected path secret/b, got %s", snB.Path)
	}
	if _, ok := snA.Secrets["secret/a/key1"]; !ok {
		t.Error("expected secret/a/key1 in snapshot A")
	}
	if _, ok := snB.Secrets["secret/b/key2"]; !ok {
		t.Error("expected secret/b/key2 in snapshot B")
	}
}

func TestDiffClient_TakeSnapshots_PropagatesError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	dc := NewDiffClient(client, 1)

	_, _, err := dc.TakeSnapshots(context.Background(), "secret/a", "secret/b")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(fmt.Sprintf("%v", err), "snapshot") {
		t.Errorf("expected error to mention snapshot, got: %v", err)
	}
}
