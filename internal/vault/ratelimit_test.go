package vault

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRateLimiter_DisabledAllowsAll(t *testing.T) {
	rl := NewRateLimiter(0)
	defer rl.Stop()

	// Should never block when disabled.
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			rl.Wait()
		}
		close(done)
	}()

	select {
	case <-done:
		// ok
	case <-time.After(500 * time.Millisecond):
		t.Fatal("disabled rate limiter blocked unexpectedly")
	}
}

func TestRateLimiter_LimitsRequests(t *testing.T) {
	const rps = 10
	rl := NewRateLimiter(rps)
	defer rl.Stop()

	var count int64
	var wg sync.WaitGroup

	start := time.Now()
	for i := 0; i < rps; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.Wait()
			atomic.AddInt64(&count, 1)
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	if atomic.LoadInt64(&count) != rps {
		t.Fatalf("expected %d completions, got %d", rps, count)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("first burst took too long: %v", elapsed)
	}
}

func TestRateLimiter_RefillsTokens(t *testing.T) {
	rl := NewRateLimiter(5)
	defer rl.Stop()

	// Drain the initial tokens.
	for i := 0; i < 5; i++ {
		rl.Wait()
	}

	// Force a refill by waiting just over a second.
	time.Sleep(1100 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		rl.Wait()
		close(done)
	}()

	select {
	case <-done:
		// ok — tokens were refilled
	case <-time.After(200 * time.Millisecond):
		t.Fatal("token was not refilled after one second")
	}
}

func TestRateLimiter_StopIsIdempotent(t *testing.T) {
	rl := NewRateLimiter(0)
	// Calling Stop on a disabled limiter should not panic.
	rl.Stop()
	rl.Stop()
}
