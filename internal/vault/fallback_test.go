package vault

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newFallbackTestServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func TestFallbackClient_PrimarySucceeds_NoFallback(t *testing.T) {
	primary := newTestClient(t, newFallbackTestServer(t, 200,
		`{"data":{"key":"primary"}}`).URL)
	secondary := newTestClient(t, newFallbackTestServer(t, 200,
		`{"data":{"key":"secondary"}}`).URL)

	c := NewFallbackClient(primary, secondary, nil)
	data, err := c.ReadSecret(context.Background(), "secret/foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["key"] != "primary" {
		t.Errorf("expected primary data, got %v", data)
	}
}

func TestFallbackClient_PrimaryFails_UsesSecondary(t *testing.T) {
	primary := newTestClient(t, newFallbackTestServer(t, 500, `{}`).URL)
	secondary := newTestClient(t, newFallbackTestServer(t, 200,
		`{"data":{"key":"fallback"}}`).URL)

	c := NewFallbackClient(primary, secondary, nil)
	data, err := c.ReadSecret(context.Background(), "secret/foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["key"] != "fallback" {
		t.Errorf("expected fallback data, got %v", data)
	}
}

func TestFallbackClient_BothFail_ReturnsJoinedError(t *testing.T) {
	primary := newTestClient(t, newFallbackTestServer(t, 500, `{}`).URL)
	secondary := newTestClient(t, newFallbackTestServer(t, 404, `{}`).URL)

	c := NewFallbackClient(primary, secondary, nil)
	_, err := c.ReadSecret(context.Background(), "secret/foo")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFallbackClient_ShouldFallbackFalse_ReturnsImmediately(t *testing.T) {
	primary := newTestClient(t, newFallbackTestServer(t, 500, `{}`).URL)
	secondary := newTestClient(t, newFallbackTestServer(t, 200,
		`{"data":{"key":"secondary"}}`).URL)

	neverFallback := func(error) bool { return false }
	c := NewFallbackClient(primary, secondary, neverFallback)
	_, err := c.ReadSecret(context.Background(), "secret/foo")
	if err == nil {
		t.Fatal("expected error when fallback disabled")
	}
}

func TestFallbackClient_ShouldFallback_MatchesSpecificError(t *testing.T) {
	sentinel := errors.New("not found")
	var called bool

	primary := &staticErrorReader{err: sentinel}
	secondary := &staticDataReader{data: map[string]interface{}{"k": "v"}}

	onlyNotFound := func(err error) bool { return errors.Is(err, sentinel) }
	c := NewFallbackClient(primary, secondary, onlyNotFound)
	data, err := c.ReadSecret(context.Background(), "secret/bar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["k"] != "v" {
		t.Errorf("expected secondary data")
	}
	_ = called
}

// helpers

type staticErrorReader struct{ err error }

func (s *staticErrorReader) ReadSecret(_ context.Context, _ string) (map[string]interface{}, error) {
	return nil, s.err
}

type staticDataReader struct{ data map[string]interface{} }

func (s *staticDataReader) ReadSecret(_ context.Context, _ string) (map[string]interface{}, error) {
	return s.data, nil
}
