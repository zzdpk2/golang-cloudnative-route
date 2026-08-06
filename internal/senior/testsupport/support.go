// Package testsupport contains deterministic helpers for senior-track tests.
//
// These helpers coordinate tests; they do not belong on the production path.
// Production code should accept a narrow injected callback and use Noop when
// no test hook is needed.
package testsupport

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Hook func(context.Context) error

func Noop(context.Context) error {
	return nil
}

// Barrier pauses an injected Hook at a deterministic point. A test can wait
// until the point is reached, perform a concurrent action, then release it.
//
// A Barrier is single-use. Create a new one for each fault occurrence so a test
// cannot accidentally depend on invocation order.
type Barrier struct {
	reached     chan struct{}
	release     chan struct{}
	reachedOnce sync.Once
	releaseOnce sync.Once
}

func NewBarrier() *Barrier {
	return &Barrier{
		reached: make(chan struct{}),
		release: make(chan struct{}),
	}
}

func (b *Barrier) Reach(ctx context.Context) error {
	if b == nil {
		return errors.New("nil barrier")
	}
	b.reachedOnce.Do(func() { close(b.reached) })
	select {
	case <-b.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (b *Barrier) WaitUntilReached(ctx context.Context) error {
	if b == nil {
		return errors.New("nil barrier")
	}
	select {
	case <-b.reached:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (b *Barrier) Release() {
	if b == nil {
		return
	}
	b.releaseOnce.Do(func() { close(b.release) })
}

type CallResult[T any] struct {
	Worker int
	Value  T
	Err    error
}

// RunConcurrent starts every worker behind the same gate, then releases all of
// them together. Results retain worker order even when completion order differs.
func RunConcurrent[T any](
	ctx context.Context,
	workers int,
	fn func(context.Context, int) (T, error),
) []CallResult[T] {
	if workers <= 0 {
		return nil
	}

	start := make(chan struct{})
	results := make([]CallResult[T], workers)
	var ready sync.WaitGroup
	var done sync.WaitGroup
	ready.Add(workers)
	done.Add(workers)

	for worker := 0; worker < workers; worker++ {
		go func(worker int) {
			defer done.Done()
			ready.Done()
			select {
			case <-start:
			case <-ctx.Done():
				results[worker] = CallResult[T]{Worker: worker, Err: ctx.Err()}
				return
			}
			value, err := fn(ctx, worker)
			results[worker] = CallResult[T]{
				Worker: worker,
				Value:  value,
				Err:    err,
			}
		}(worker)
	}

	ready.Wait()
	close(start)
	done.Wait()
	return results
}

// Eventually retries check until it succeeds or the context ends. It uses a
// timer rather than time.Sleep so cancellation is prompt.
func Eventually(
	ctx context.Context,
	interval time.Duration,
	check func() error,
) error {
	if interval <= 0 {
		return errors.New("interval must be positive")
	}
	if check == nil {
		return errors.New("check is required")
	}

	var last error
	for {
		if err := check(); err == nil {
			return nil
		} else {
			last = err
		}

		timer := time.NewTimer(interval)
		select {
		case <-timer.C:
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			if last == nil {
				return ctx.Err()
			}
			return fmt.Errorf("%w; last check: %v", ctx.Err(), last)
		}
	}
}
