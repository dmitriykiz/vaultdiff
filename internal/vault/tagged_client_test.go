package vault

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTagTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"username":"admin","password":"s3cr3t"}}`))
	}))
}

func TestTaggedClient_NoTags_ReturnsUnmodified(t *testing.T) {
	srv := newTagTestServer(t)
	defer srv.Close()

	inner := newTestClient(t, srv.URL)
	client := NewTaggedClient(inner, nil)

	data, err := client.ReadSecret(context.Background(), "secret/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := data["_tags"]; ok {
		t.Error("expected no _tags key when no tags provided")
	}
	if data["username"] != "admin" {
		t.Errorf("expected username=admin, got %v", data["username"])
	}
}

func TestTaggedClient_InjectsTagKey(t *testing.T) {
	srv := newTagTestServer(t)
	defer srv.Close()

	inner := newTestClient(t, srv.URL)
	client := NewTaggedClient(inner, map[string]string{"env": "prod", "team": "platform"})

	data, err := client.ReadSecret(context.Background(), "secret/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tagVal, ok := data["_tags"]
	if !ok {
		t.Fatal("expected _tags key in result")
	}
	tagStr, ok := tagVal.(string)
	if !ok {
		t.Fatalf("expected _tags to be string, got %T", tagVal)
	}
	if !strings.Contains(tagStr, "env=prod") {
		t.Errorf("expected tag string to contain env=prod, got %q", tagStr)
	}
	if !strings.Contains(tagStr, "team=platform") {
		t.Errorf("expected tag string to contain team=platform, got %q", tagStr)
	}
}

func TestTaggedClient_DoesNotMutateInnerResult(t *testing.T) {
	original := map[string]interface{}{"key": "value"}
	stub := &stubReader{data: original}
	client := NewTaggedClient(stub, map[string]string{"x": "1"})

	_, err := client.ReadSecret(context.Background(), "any/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := original["_tags"]; ok {
		t.Error("inner map was mutated by TaggedClient")
	}
}

func TestTaggedClient_PropagatesError(t *testing.T) {
	stub := &stubReader{err: errors.New("vault unavailable")}
	client := NewTaggedClient(stub, map[string]string{"env": "staging"})

	_, err := client.ReadSecret(context.Background(), "secret/broken")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "vault unavailable") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestTaggedClient_Tags_ReturnsCopy(t *testing.T) {
	stub := &stubReader{data: map[string]interface{}{}}
	client := NewTaggedClient(stub, map[string]string{"region": "us-east-1"})

	tags := client.Tags()
	tags["injected"] = "yes"

	origTags := client.Tags()
	if _, ok := origTags["injected"]; ok {
		t.Error("Tags() should return a copy, not a reference")
	}
}

// stubReader is a minimal SecretReader for unit tests.
type stubReader struct {
	data map[string]interface{}
	err  error
}

func (s *stubReader) ReadSecret(_ context.Context, _ string) (map[string]interface{}, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.data, nil
}
