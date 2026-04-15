package vault_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yourusername/vaultdiff/internal/vault"
)

func newCheckpointTestServer(secrets map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/sys/internal/ui/mounts/") {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"data":{"options":{"version":"1"}}}`)
			return
		}
		if r.URL.Query().Get("list") == "true" {
			keys := make([]string, 0, len(secrets))
			for k := range secrets {
				keys = append(keys, strings.TrimPrefix(k, "secret/"))
			}
			body := `{"data":{"keys":[`
			for i, k := range keys {
				if i > 0 {
					body += ","
				}
				body += `"` + k + `"`
			}
			body += `]}}`
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, body)
			return
		}
		for path, val := range secrets {
			if strings.HasSuffix(r.URL.Path, path) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprintf(w, `{"data":{"value":%q}}`, val)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestCheckpointClient_TakeAndLoad(t *testing.T) {
	srv := newCheckpointTestServer(map[string]string{
		"secret/app/token": "abc123",
	})
	defer srv.Close()

	client, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	store := vault.NewCheckpointStore(t.TempDir())
	cc := vault.NewCheckpointClient(client, store)
	taker := vault.NewSnapshotTaker(client, 2)

	if err := cc.TakeCheckpoint(context.Background(), "snap-a", "secret", taker); err != nil {
		t.Fatalf("TakeCheckpoint: %v", err)
	}
	cp, err := cc.LoadCheckpoint("snap-a")
	if err != nil {
		t.Fatalf("LoadCheckpoint: %v", err)
	}
	if cp.Path != "secret" {
		t.Errorf("path: got %q, want %q", cp.Path, "secret")
	}
	if cp.TakenAt.IsZero() {
		t.Error("TakenAt should not be zero")
	}
}

func TestCheckpointClient_LoadCheckpoint_Missing(t *testing.T) {
	srv := newCheckpointTestServer(nil)
	defer srv.Close()
	client, _ := vault.NewClient(srv.URL, "test-token")
	store := vault.NewCheckpointStore(t.TempDir())
	cc := vault.NewCheckpointClient(client, store)

	_, err := cc.LoadCheckpoint("nonexistent")
	if err == nil {
		t.Fatal("expected error loading missing checkpoint")
	}
}

func TestCheckpointClient_ReadSecret_Delegates(t *testing.T) {
	srv := newCheckpointTestServer(map[string]string{
		"secret/app/token": "xyz",
	})
	defer srv.Close()
	client, _ := vault.NewClient(srv.URL, "test-token")
	store := vault.NewCheckpointStore(t.TempDir())
	cc := vault.NewCheckpointClient(client, store)

	data, err := cc.ReadSecret(context.Background(), "secret/app/token")
	if err != nil {
		t.Fatalf("ReadSecret: %v", err)
	}
	if data == nil {
		t.Error("expected non-nil data")
	}
}
