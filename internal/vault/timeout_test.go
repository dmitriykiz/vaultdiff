package vault

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// slowServer returns a test server that sleeps for delay before responding.
func slowServer(t *testing.T, delay time.Duration) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"key":"value"}}`))
	}))
}

func TestTimeoutClient_PassesThroughOnSuccess(t *testing.T) {
	srv := newTestClient(t, map[string]interface{}{"username": "alice"})
	_ = srv

	inner := newTestClient(t, map[string]interface{}{"username": "alice"})
	client := NewTimeoutClient(inner, 2*time.Second)

	data, err := client.ReadSecret(context.Background(), "secret/data/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["username"] != "alice" {
		t.Errorf("expected username=alice, got %v", data["username"])
	}
}

func TestTimeoutClient_ZeroTimeoutDisablesDeadline(t *testing.T) {
	inner := newTestClient(t, map[string]interface{}{"token": "abc"})
	client := NewTimeoutClient(inner, 0)

	data, err := client.ReadSecret(context.Background(), "secret/data/myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["token"] != "abc" {
		t.Errorf("expected token=abc, got %v", data["token"])
	}
}

func TestTimeoutClient_ExceedsTimeout(t *testing.T) {
	// Use a fake inner reader that blocks until its context is cancelled.
	blocking := &blockingReader{block: make(chan struct{})}
	client := NewTimeoutClient(blocking, 50*time.Millisecond)

	_, err := client.ReadSecret(context.Background(), "secret/slow")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded in error chain, got: %v", err)
	}
}

func TestTimeoutClient_ParentContextCancelled(t *testing.T) {
	blocking := &blockingReader{block: make(chan struct{})}
	client := NewTimeoutClient(blocking, 5*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	_, err := client.ReadSecret(ctx, "secret/slow")
	if err == nil {
		t.Fatal("expected error due to cancelled context, got nil")
	}
}

// blockingReader is a SecretReader that blocks until its context is done.
type blockingReader struct {
	block chan struct{}
}

func (b *blockingReader) ReadSecret(ctx context.Context, _ string) (map[string]interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-b.block:
		return map[string]interface{}{}, nil
	}
}
