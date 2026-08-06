package testsupport

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestBarrierPausesAtInjectedPoint(t *testing.T) {
	barrier := NewBarrier()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	completed := make(chan error, 1)
	go func() {
		completed <- barrier.Reach(ctx)
	}()

	if err := barrier.WaitUntilReached(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case <-completed:
		t.Fatal("hook completed before release")
	default:
	}

	barrier.Release()
	if err := <-completed; err != nil {
		t.Fatal(err)
	}
	barrier.Release()
}

func TestBarrierHonorsCancellation(t *testing.T) {
	barrier := NewBarrier()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := barrier.Reach(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

func TestRunConcurrentReleasesWorkersTogetherAndRetainsOrder(t *testing.T) {
	var active int32
	var maximum int32

	results := RunConcurrent(context.Background(), 12,
		func(_ context.Context, worker int) (int, error) {
			now := atomic.AddInt32(&active, 1)
			for {
				seen := atomic.LoadInt32(&maximum)
				if now <= seen || atomic.CompareAndSwapInt32(&maximum, seen, now) {
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
			atomic.AddInt32(&active, -1)
			return worker * 2, nil
		})

	if maximum < 2 {
		t.Fatalf("workers did not overlap: maximum active = %d", maximum)
	}
	for worker, result := range results {
		if result.Worker != worker || result.Value != worker*2 || result.Err != nil {
			t.Errorf("result[%d] = %+v", worker, result)
		}
	}
}

func TestEventuallyRetriesAndReportsLastFailure(t *testing.T) {
	var attempts int
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := Eventually(ctx, time.Millisecond, func() error {
		attempts++
		if attempts < 3 {
			return errors.New("not yet")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d", attempts)
	}

	expired, cancelExpired := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancelExpired()
	err = Eventually(expired, time.Millisecond, func() error {
		return errors.New("still unavailable")
	})
	if err == nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v", err)
	}
	if !errors.Is(errors.Unwrap(err), context.DeadlineExceeded) {
		t.Fatalf("deadline was not wrapped: %v", err)
	}
}
