package vault

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newVersionTestServer(t *testing.T, capturedPath *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*capturedPath = r.URL.RequestURI()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"value":"secret"}}`))
	}))
}

func TestVersionClient_LatestVersion_NoRewrite(t *testing.T) {
	var captured string
	srv := newVersionTestServer(t, &captured)
	defer srv.Close()

	inner, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatal(err)
	}
	vc := NewVersionClient(inner, 0)
	_, _ = vc.ReadSecret(context.Background(), "secret/data/myapp")

	if strings.Contains(captured, "version=") {
		t.Errorf("expected no version param, got %s", captured)
	}
}

func TestVersionClient_PinnedVersion_AppendsParam(t *testing.T) {
	var captured string
	srv := newVersionTestServer(t, &captured)
	defer srv.Close()

	inner, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatal(err)
	}
	vc := NewVersionClient(inner, 3)
	_, _ = vc.ReadSecret(context.Background(), "secret/data/myapp")

	if !strings.Contains(captured, "version=3") {
		t.Errorf("expected version=3 in path, got %s", captured)
	}
}

func TestVersionClient_NonDataPath_NoRewrite(t *testing.T) {
	var captured string
	srv := newVersionTestServer(t, &captured)
	defer srv.Close()

	inner, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatal(err)
	}
	vc := NewVersionClient(inner, 2)
	_, _ = vc.ReadSecret(context.Background(), "secret/myapp") // KVv1 style

	if strings.Contains(captured, "version=") {
		t.Errorf("expected no version param for non-data path, got %s", captured)
	}
}

func TestVersionClient_Version_Accessor(t *testing.T) {
	inner, _ := NewClient("http://127.0.0.1:8200", "tok")
	vc := NewVersionClient(inner, 5)
	if vc.Version() != 5 {
		t.Errorf("expected version 5, got %d", vc.Version())
	}
}

func TestVersionClient_WithVersion_ReturnsNew(t *testing.T) {
	inner, _ := NewClient("http://127.0.0.1:8200", "tok")
	vc := NewVersionClient(inner, 1)
	nv := vc.WithVersion(9)
	if nv.Version() != 9 {
		t.Errorf("expected version 9, got %d", nv.Version())
	}
	if vc.Version() != 1 {
		t.Error("original version client should be unchanged")
	}
}
