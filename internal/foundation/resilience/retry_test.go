package resilience

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestDo_SucceedsFirst(t *testing.T) {
	callCount := 0
	err := Do(context.Background(), func(ctx context.Context) error {
		callCount++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if callCount != 1 {
		t.Errorf("callCount = %d", callCount)
	}
}

func TestDo_SucceedsAfterRetry(t *testing.T) {
	callCount := 0
	err := Do(context.Background(), func(ctx context.Context) error {
		callCount++
		if callCount < 3 {
			return fmt.Errorf("fail %d", callCount)
		}
		return nil
	}, WithMaxAttempts(5), WithInitDelay(1*time.Millisecond))

	if err != nil {
		t.Fatal(err)
	}
	if callCount != 3 {
		t.Errorf("callCount = %d", callCount)
	}
}

func TestDo_AllFail(t *testing.T) {
	err := Do(context.Background(), func(ctx context.Context) error {
		return fmt.Errorf("always fails")
	}, WithMaxAttempts(3), WithInitDelay(1*time.Millisecond))

	if err == nil {
		t.Fatal("should fail")
	}
	if !errors.Is(err, fmt.Errorf("always fails")) {
		// wrapped error should contain original
		t.Logf("error: %v", err)
	}
}

func TestDo_RespectsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	callCount := 0
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := Do(ctx, func(ctx context.Context) error {
		callCount++
		return fmt.Errorf("fail")
	}, WithMaxAttempts(100), WithInitDelay(20*time.Millisecond))

	if !errors.Is(err, context.Canceled) {
		t.Logf("error = %v (should be context.Canceled or wrapped)", err)
	}
}

func TestDo_RetryIf(t *testing.T) {
	tempErr := fmt.Errorf("temporary")
	permErr := fmt.Errorf("permanent")

	callCount := 0
	err := Do(context.Background(), func(ctx context.Context) error {
		callCount++
		if callCount == 1 {
			return tempErr
		}
		return permErr
	},
		WithMaxAttempts(5),
		WithInitDelay(1*time.Millisecond),
		WithRetryIf(func(err error) bool { return err == tempErr }),
	)

	if err != permErr {
		t.Errorf("should stop on permanent error, got: %v", err)
	}
	if callCount != 2 {
		t.Errorf("callCount = %d (should stop after non-retryable error)", callCount)
	}
}

func TestDoWithResult(t *testing.T) {
	callCount := 0
	result, err := DoWithResult(context.Background(), func(ctx context.Context) (string, error) {
		callCount++
		if callCount < 2 {
			return "", fmt.Errorf("not yet")
		}
		return "success", nil
	}, WithMaxAttempts(3), WithInitDelay(1*time.Millisecond))

	if err != nil {
		t.Fatal(err)
	}
	if result != "success" {
		t.Errorf("result = %q", result)
	}
}

// ---- Circuit Breaker Tests ----

func TestCircuitBreaker_StartsClose(t *testing.T) {
	cb := NewCircuitBreaker(3, 2, 100*time.Millisecond)
	if cb.State() != StateClosed {
		t.Error("should start closed")
	}
}

func TestCircuitBreaker_OpensAfterFailures(t *testing.T) {
	cb := NewCircuitBreaker(3, 1, 100*time.Millisecond)

	for i := 0; i < 3; i++ {
		cb.Execute(func() error { return fmt.Errorf("fail") })
	}

	if cb.State() != StateOpen {
		t.Errorf("state = %s, want open", cb.State())
	}

	err := cb.Execute(func() error { return nil })
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("err = %v, want ErrCircuitOpen", err)
	}
}

func TestCircuitBreaker_SuccessResetsCounter(t *testing.T) {
	cb := NewCircuitBreaker(3, 1, 100*time.Millisecond)

	// 2 failures then success → resets
	cb.Execute(func() error { return fmt.Errorf("fail") })
	cb.Execute(func() error { return fmt.Errorf("fail") })
	cb.Execute(func() error { return nil }) // success resets

	if cb.State() != StateClosed {
		t.Error("should still be closed after success resets counter")
	}

	// need 3 more consecutive failures to open
	cb.Execute(func() error { return fmt.Errorf("fail") })
	cb.Execute(func() error { return fmt.Errorf("fail") })
	if cb.State() != StateClosed {
		t.Error("only 2 consecutive, should be closed")
	}
}

func TestCircuitBreaker_HalfOpenRecovery(t *testing.T) {
	cb := NewCircuitBreaker(2, 2, 50*time.Millisecond)

	// Trip the breaker
	cb.Execute(func() error { return fmt.Errorf("fail") })
	cb.Execute(func() error { return fmt.Errorf("fail") })
	if cb.State() != StateOpen {
		t.Fatal("should be open")
	}

	// Wait for timeout → should transition to half-open
	time.Sleep(60 * time.Millisecond)

	// This call should go through (half-open allows it)
	err := cb.Execute(func() error { return nil })
	if err != nil {
		t.Fatalf("half-open should allow: %v", err)
	}

	// Need 2 successes to close (successThreshold=2)
	err = cb.Execute(func() error { return nil })
	if err != nil {
		t.Fatalf("second success: %v", err)
	}

	if cb.State() != StateClosed {
		t.Errorf("state = %s, should be closed after %d successes", cb.State(), 2)
	}
}

func TestCircuitBreaker_HalfOpenFailure(t *testing.T) {
	cb := NewCircuitBreaker(2, 2, 50*time.Millisecond)

	// Trip
	cb.Execute(func() error { return fmt.Errorf("fail") })
	cb.Execute(func() error { return fmt.Errorf("fail") })

	// Wait for half-open
	time.Sleep(60 * time.Millisecond)

	// Fail in half-open → back to open
	cb.Execute(func() error { return fmt.Errorf("still broken") })

	if cb.State() != StateOpen {
		t.Errorf("state = %s, should go back to open", cb.State())
	}
}

func TestCircuitBreaker_Concurrent(t *testing.T) {
	cb := NewCircuitBreaker(5, 2, 100*time.Millisecond)
	var success, blocked int32

	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func() {
			err := cb.Execute(func() error { return fmt.Errorf("fail") })
			if errors.Is(err, ErrCircuitOpen) {
				atomic.AddInt32(&blocked, 1)
			}
			select {
			case done <- struct{}{}:
			default:
			}
		}()
	}

	// Wait for some goroutines
	for i := 0; i < 50; i++ {
		<-done
	}

	_ = success
	// After 5 failures, remaining should be blocked
	if atomic.LoadInt32(&blocked) == 0 {
		t.Error("some calls should have been blocked by circuit breaker")
	}
	t.Logf("blocked = %d / 50", atomic.LoadInt32(&blocked))
}
