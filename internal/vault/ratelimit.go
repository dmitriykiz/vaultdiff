package vault

import (
	"sync"
	"time"
)

// RateLimiter enforces a maximum number of requests per second using a
// token-bucket style approach backed by a ticker.
type RateLimiter struct {
	mu      sync.Mutex
	tokens  int
	max     int
	ticker  *time.Ticker
	done    chan struct{}
}

// NewRateLimiter creates a RateLimiter that allows up to rps requests per
// second. A rps value of zero disables rate limiting.
func NewRateLimiter(rps int) *RateLimiter {
	if rps <= 0 {
		return &RateLimiter{}
	}
	rl := &RateLimiter{
		tokens: rps,
		max:    rps,
		done:   make(chan struct{}),
	}
	rl.ticker = time.NewTicker(time.Second)
	go rl.refill()
	return rl
}

// refill restores the token bucket to its maximum capacity every second.
func (rl *RateLimiter) refill() {
	for {
		select {
		case <-rl.ticker.C:
			rl.mu.Lock()
			rl.tokens = rl.max
			rl.mu.Unlock()
		case <-rl.done:
			return
		}
	}
}

// Wait blocks until a token is available. If rate limiting is disabled it
// returns immediately.
func (rl *RateLimiter) Wait() {
	if rl.max == 0 {
		return
	}
	for {
		rl.mu.Lock()
		if rl.tokens > 0 {
			rl.tokens--
			rl.mu.Unlock()
			return
		}
		rl.mu.Unlock()
		time.Sleep(5 * time.Millisecond)
	}
}

// Stop releases resources held by the RateLimiter.
func (rl *RateLimiter) Stop() {
	if rl.ticker != nil {
		rl.ticker.Stop()
		close(rl.done)
	}
}
