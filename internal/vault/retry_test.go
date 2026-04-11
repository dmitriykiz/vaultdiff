package vault

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func noSleep(_ time.Duration) {}

func TestWithRetry_SuccessOnFirstAttempt(t *testing.T) {
	calls := 0
	err := WithRetry(DefaultRetryConfig(), noSleep, func() (int, error) {
		calls++
		return http.StatusOK, nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestWithRetry_SuccessAfterTransientFailure(t *testing.T) {
	calls := 0
	err := WithRetry(DefaultRetryConfig(), noSleep, func() (int, error) {
		calls++
		if calls < 3 {
			return http.StatusServiceUnavailable, errors.New("unavailable")
		}
		return http.StatusOK, nil
	})
	if err != nil {
		t.Fatalf("expected nil error after retry, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestWithRetry_NonRetryableErrorStopsImmediately(t *testing.T) {
	calls := 0
	sentinel := errors.New("forbidden")
	err := WithRetry(DefaultRetryConfig(), noSleep, func() (int, error) {
		calls++
		return http.StatusForbidden, sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call for non-retryable error, got %d", calls)
	}
}

func TestWithRetry_ExhaustsAttemptsReturnsMaxRetriesExceeded(t *testing.T) {
	cfg := RetryConfig{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}
	calls := 0
	err := WithRetry(cfg, noSleep, func() (int, error) {
		calls++
		return http.StatusTooManyRequests, errors.New("rate limited")
	})
	if !errors.Is(err, ErrMaxRetriesExceeded) {
		t.Fatalf("expected ErrMaxRetriesExceeded, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestWithRetry_BackoffDelaysAreApplied(t *testing.T) {
	cfg := RetryConfig{MaxAttempts: 3, BaseDelay: 100 * time.Millisecond, MaxDelay: time.Second}
	var delays []time.Duration
	sleepFn := func(d time.Duration) { delays = append(delays, d) }

	_ = WithRetry(cfg, sleepFn, func() (int, error) {
		return http.StatusServiceUnavailable, errors.New("err")
	})

	// 3 attempts → 2 sleeps (after attempt 1 and 2)
	if len(delays) != 2 {
		t.Fatalf("expected 2 sleep calls, got %d", len(delays))
	}
	if delays[1] != 200*time.Millisecond {
		t.Fatalf("expected second delay 200ms (doubled), got %v", delays[1])
	}
}

func TestIsRetryable(t *testing.T) {
	retryable := []int{429, 500, 502, 503, 504}
	for _, code := range retryable {
		if !isRetryable(code) {
			t.Errorf("expected %d to be retryable", code)
		}
	}
	non := []int{200, 400, 403, 404}
	for _, code := range non {
		if isRetryable(code) {
			t.Errorf("expected %d to NOT be retryable", code)
		}
	}
}
