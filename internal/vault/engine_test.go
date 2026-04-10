package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMountFromPath(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"secret/my/app", "secret"},
		{"/secret/my/app", "secret"},
		{"kv", "kv"},
		{"", ""},
	}
	for _, tc := range cases {
		got := MountFromPath(tc.input)
		if got != tc.want {
			t.Errorf("MountFromPath(%q) = %q; want %q", tc.input, got, tc.want)
		}
	}
}

func TestDetectEngineType_KVv2(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/secret/metadata" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c, err := NewClient(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	engine, err := DetectEngineType(context.Background(), c, "secret")
	if err != nil {
		t.Fatalf("DetectEngineType: %v", err)
	}
	if engine != EngineKVv2 {
		t.Errorf("expected EngineKVv2, got %v", engine)
	}
}

func TestDetectEngineType_KVv1(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/kv" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"foo":"bar"}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c, err := NewClient(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	engine, err := DetectEngineType(context.Background(), c, "kv")
	if err != nil {
		t.Fatalf("DetectEngineType: %v", err)
	}
	if engine != EngineKVv1 {
		t.Errorf("expected EngineKVv1, got %v", engine)
	}
}

func TestDetectEngineType_Unknown(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c, err := NewClient(ts.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = DetectEngineType(context.Background(), c, "nonexistent")
	if err == nil {
		t.Error("expected error for unknown engine, got nil")
	}
}
