package vault

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newBatchTestServer(t *testing.T, secrets map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/v1/")
		val, ok := secrets[key]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, `{"data":{"value":%q}}`, val)
	}))
}

func TestBatchClient_ReadAll_ReturnsAllResults(t *testing.T) {
	srv := newBatchTestServer(t, map[string]string{
		"secret/a": "alpha",
		"secret/b": "beta",
		"secret/c": "gamma",
	})
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	batch := NewBatchClient(client, 3)

	results := batch.ReadAll(context.Background(), []string{"secret/a", "secret/b", "secret/c"})

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Error != nil {
			t.Errorf("unexpected error for %s: %v", r.Path, r.Error)
		}
	}
}

func TestBatchClient_ReadAll_OrderPreserved(t *testing.T) {
	srv := newBatchTestServer(t, map[string]string{
		"secret/x": "x-val",
		"secret/y": "y-val",
	})
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	batch := NewBatchClient(client, 2)

	paths := []string{"secret/x", "secret/y"}
	results := batch.ReadAll(context.Background(), paths)

	for i, r := range results {
		if r.Path != paths[i] {
			t.Errorf("index %d: expected path %s, got %s", i, paths[i], r.Path)
		}
	}
}

func TestBatchClient_ReadAll_PropagatesErrors(t *testing.T) {
	srv := newBatchTestServer(t, map[string]string{
		"secret/ok": "fine",
	})
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	batch := NewBatchClient(client, 2)

	results := batch.ReadAll(context.Background(), []string{"secret/ok", "secret/missing"})

	errs := Errors(results)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if errs[0].Path != "secret/missing" {
		t.Errorf("expected error on secret/missing, got %s", errs[0].Path)
	}
}

func TestBatchClient_DefaultConcurrency(t *testing.T) {
	srv := newBatchTestServer(t, map[string]string{})
	defer srv.Close()

	client := newTestClient(t, srv.URL)
	batch := NewBatchClient(client, 0)

	if batch.concurrency != 5 {
		t.Errorf("expected default concurrency 5, got %d", batch.concurrency)
	}
}

func TestJoinErrors_FormatsMessages(t *testing.T) {
	results := []BatchResult{
		{Path: "a", Error: errors.New("boom")},
		{Path: "b", Data: map[string]any{"k": "v"}},
	}
	out := JoinErrors(results)
	if !strings.Contains(out, "a: boom") {
		t.Errorf("expected error message in output, got: %s", out)
	}
	if strings.Contains(out, "b:") {
		t.Errorf("non-error path should not appear in JoinErrors output")
	}
}
