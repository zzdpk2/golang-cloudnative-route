// Package retry is L8's answer to "the dependency was briefly unavailable".
//
// Two mechanisms that look similar and solve opposite problems:
//
//	Retry           the failure is probably transient — try again
//	CircuitBreaker  the failure is clearly not transient — stop trying
//
// Using only the first turns a brief outage into a self-inflicted denial of
// service: every caller retries, the failing service is pinned under load, and
// it cannot recover. That interaction is why the two belong in one package.
package resilience

import (
	"context"
	"fmt"
	"sync"

	"time"
)

// ---- Retry ----

// RetryConfig controls how Do behaves.
type RetryConfig struct {
	// MaxAttempts is the total number of calls, not the number of *re*tries.
	// 3 means one attempt plus two retries.
	MaxAttempts int
	// InitDelay is how long to wait after the first failure.
	InitDelay time.Duration
	// MaxDelay caps the backoff, so exponential growth cannot run away.
	MaxDelay time.Duration
	// Multiplier is applied to the delay after each attempt: 2.0 doubles it.
	Multiplier float64
	// Jitter randomises the delay. See the note on Do for why this matters far
	// more than it looks.
	Jitter bool
	// RetryIf decides whether a given error is worth retrying at all.
	RetryIf func(error) bool
}

type RetryOption func(*RetryConfig)

func WithMaxAttempts(n int) RetryOption { panic("TODO") }

func WithInitDelay(d time.Duration) RetryOption { panic("TODO") }

func WithMaxDelay(d time.Duration) RetryOption { panic("TODO") }

func WithMultiplier(m float64) RetryOption { panic("TODO") }

func WithJitter(j bool) RetryOption { panic("TODO") }

func WithRetryIf(fn func(error) bool) RetryOption { panic("TODO") }

// Do calls fn until it succeeds, gives up, or the context ends.
//
// Behaviour, with MaxAttempts=3, InitDelay=100ms, Multiplier=2:
//
//	fn succeeds first time        → 1 call,  nil
//	fn fails then succeeds        → 2 calls, nil, having waited ~100ms
//	fn always fails               → 3 calls, an error wrapping the last one
//	RetryIf says no               → 1 call,  that error returned unwrapped
//	ctx cancelled while waiting   → returns ctx.Err() promptly
//
// The delays are 100ms then 200ms — growth happens *between* attempts, and
// there is no wait after the final one. Sleeping after the last failure is the
// most common bug here: it costs a full delay and buys nothing.
//
// # Four things to get right
//
// **Wait on the context, not on the clock.** `time.Sleep` cannot be cancelled,
// so a request that has already been abandoned still holds a goroutine for the
// full backoff. Select on a timer and ctx.Done together.
//
// **Honour RetryIf.** A 400 Bad Request will be a 400 next time too; retrying
// it wastes the caller's deadline. Return that error as-is, without the
// "after N attempts" wrapper, because it was not attempted N times.
//
// **Wrap the final error with %w** so the caller can still errors.Is it. The
// test checks that the original error is reachable through the wrapper.
//
// **Jitter is not decoration.** Without it, a thousand clients that failed
// together retry together — the thundering herd — and the recovering service is
// knocked over by the synchronised second wave. Spreading the delay randomly is
// what breaks the synchronisation, and it is the single most valuable line in
// this function.
func Do(ctx context.Context, fn func(ctx context.Context) error, opts ...RetryOption) error {
	panic("TODO")
}

// DoWithResult is Do for a function that returns a value.
//
//	DoWithResult(ctx, fetchOrder, WithMaxAttempts(3)) → (*Order, error)
//
// Build it on Do rather than duplicating the loop — capture the result in a
// closure. If that feels like a trick, note that it is the same trick
// Result.Catch and errors.ErrorGroup use, and that Go's lack of generic methods
// is why this has to be a free function at all.
func DoWithResult[T any](ctx context.Context, fn func(ctx context.Context) (T, error), opts ...RetryOption) (T, error) {
	panic("TODO")
}

// ---- Circuit Breaker ----

type CircuitState int32

const (
	// StateClosed is normal operation: calls go through, failures are counted.
	StateClosed CircuitState = iota
	// StateOpen fails fast without calling the dependency at all.
	StateOpen
	// StateHalfOpen lets a limited number of probe calls through to find out
	// whether the dependency has recovered.
	StateHalfOpen
)

func (s CircuitState) String() string { panic("TODO") }

var ErrCircuitOpen = fmt.Errorf("circuit breaker is open")

// CircuitBreaker stops calling a dependency that is clearly broken.
//
// The state machine:
//
//	closed    ──failureThreshold consecutive failures──→  open
//	open      ──openTimeout elapses──────────────────────→  half-open
//	half-open ──successThreshold consecutive successes──→  closed
//	half-open ──any failure──────────────────────────────→  open
//
// **Half-open is the state people leave out**, and without it recovery means
// sending full production traffic at a service that has just come back — which
// knocks it straight over again. Half-open sends one or two probes instead.
//
// Note "consecutive". A breaker that counts total failures eventually opens on
// any long-running healthy service.
//
// You built a version of this in L6's concurrency package. This one adds the
// success threshold and the half-open probe; the two should agree on everything
// they share.
type CircuitBreaker struct {
	failureThreshold int           // consecutive failures that open the circuit
	successThreshold int           // consecutive half-open successes that close it
	openTimeout      time.Duration // how long to stay open before probing

	state            int32 // CircuitState
	consecutiveFails int32
	consecutiveSucc  int32
	lastFailTime     int64 // unix nano

	mu sync.Mutex
}

func NewCircuitBreaker(failureThreshold, successThreshold int, openTimeout time.Duration) *CircuitBreaker {
	panic("TODO")
}

func (cb *CircuitBreaker) State() CircuitState { panic("TODO") }

// Execute runs fn through the breaker.
//
//	closed, fn succeeds        → nil
//	closed, fn fails           → that error; after `failureThreshold` in a row,
//	                             the circuit opens
//	open, before the timeout   → ErrCircuitOpen, and **fn is not called**
//	open, after the timeout    → the circuit becomes half-open and fn is called
//	half-open, fn succeeds     → nil; after `successThreshold` in a row, closed
//	half-open, fn fails        → that error, and straight back to open
//
// The test asserts `errors.Is(err, ErrCircuitOpen)`, so wrap rather than
// replace if you add context.
//
// # The locking question
//
// Holding cb.mu across the call to fn serialises every request through the
// breaker, which destroys the throughput you were protecting. So the lock has
// to be released around the call and retaken to record the outcome — and in
// that gap the state can change underneath you.
//
// Work out which decisions are still valid after the gap. "I checked the state
// and it was closed" is not the same as "the state is closed now", and the
// difference is what makes this harder than it looks.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	panic("TODO")
}

// recordSuccess counts a success and closes the circuit once there have been
// enough consecutive ones in half-open.
//
// Reset the failure counter — "consecutive" cuts both ways.
func (cb *CircuitBreaker) recordSuccess() {
	panic("TODO")
}

// recordFailure counts a failure, stamps the time, and opens the circuit when
// the threshold is reached.
//
// A failure in half-open goes back to open immediately, no matter what the
// count says: the probe was the test, and it failed.
func (cb *CircuitBreaker) recordFailure() {
	panic("TODO")
}

// transitionTo moves to a new state and resets whatever counters no longer
// apply.
//
// Centralising the transition is what stops the counters drifting out of step
// with the state — the bug that makes a breaker flap between open and closed
// under steady load.
func (cb *CircuitBreaker) transitionTo(state CircuitState) {
	panic("TODO")
}
