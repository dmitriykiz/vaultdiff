package vault_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourusername/vaultdiff/internal/vault"
)

func TestCheckpointStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	store := vault.NewCheckpointStore(dir)

	cp := vault.Checkpoint{
		Path:    "secret/myapp",
		TakenAt: time.Now().Truncate(time.Second),
		Secrets: map[string]interface{}{
			"secret/myapp/db": "postgres://localhost/app",
		},
	}

	if err := store.Save("snap1", cp); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := store.Load("snap1")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Path != cp.Path {
		t.Errorf("path: got %q, want %q", got.Path, cp.Path)
	}
	if len(got.Secrets) != 1 {
		t.Errorf("secrets len: got %d, want 1", len(got.Secrets))
	}
}

func TestCheckpointStore_Load_NotFound(t *testing.T) {
	store := vault.NewCheckpointStore(t.TempDir())
	_, err := store.Load("missing")
	if err == nil {
		t.Fatal("expected error for missing checkpoint, got nil")
	}
}

func TestCheckpointStore_Delete(t *testing.T) {
	dir := t.TempDir()
	store := vault.NewCheckpointStore(dir)

	cp := vault.Checkpoint{Path: "secret/x", TakenAt: time.Now(), Secrets: map[string]interface{}{}}
	if err := store.Save("to-delete", cp); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := store.Delete("to-delete"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	path := filepath.Join(dir, "to-delete.json")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected file to be deleted, stat err: %v", err)
	}
}

func TestCheckpointStore_Save_CreatesDir(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "nested", "checkpoints")
	store := vault.NewCheckpointStore(dir)

	cp := vault.Checkpoint{Path: "secret/y", TakenAt: time.Now(), Secrets: map[string]interface{}{}}
	if err := store.Save("init", cp); err != nil {
		t.Fatalf("Save with nested dir: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("expected dir to be created: %v", err)
	}
}
