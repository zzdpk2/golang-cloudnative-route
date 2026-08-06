package concurrency

import (
	"context"
	"sync"
	"time"
)

// ============================================================
// Channel Patterns
//
// Worker pools, pub/sub, and stream adapters — plus the
// ownership rules that stop them leaking goroutines.
//
// Three rules govern almost everything in this file:
//
//  1. Whoever sends on a channel closes it. Never the receiver, never a
//     bystander. Closing a channel someone is still sending on panics.
//  2. Every goroutine you start must have a guaranteed path to returning.
//     A goroutine blocked forever on a send nobody will receive is a leak, and
//     it holds everything it captured alive with it.
//  3. Anything that can block must also select on ctx.Done, or cancellation
//     cannot reach it.
//
// Run this package with -race, and think about goroutine leaks separately —
// the race detector does not report them.
// ============================================================

// ---- Worker Pool ----

type Job[T any, R any] struct {
	ID    int
	Input T
}

type JobResult[R any] struct {
	ID     int
	Output R
	Err    error
}

// WorkerPool is the long-lived, streaming counterpart to
// concurrency.BoundedWorkerPool.
//
// The difference is worth noticing: that one takes a slice, processes it, and
// returns. This one starts, accepts work over time, and has to be shut down.
// Lifecycle is the whole difficulty here — a pool that runs to completion never
// has to answer "what about work still in the queue?".
type WorkerPool[T any, R any] struct {
	workers   int
	processor func(context.Context, T) (R, error)
	jobs      chan Job[T, R]
	results   chan JobResult[R]
	wg        sync.WaitGroup
}

func NewWorkerPool[T any, R any](workers int, bufferSize int, fn func(context.Context, T) (R, error)) *WorkerPool[T, R] {
	panic("TODO")
}

// Start launches the workers.
//
// Each worker loops: take a job, process it, send the result. It must stop both
// when the jobs channel closes *and* when ctx is cancelled — the two-value
// receive tells you which, and a closed channel is always ready, so forgetting
// the ok check spins the CPU at 100%.
//
// The subtle one: sending the result can block if nobody is reading. That send
// needs to respect cancellation too, or cancelling the context leaves workers
// stuck on it forever.
func (p *WorkerPool[T, R]) Start(ctx context.Context) {
	panic("TODO")
}

func (p *WorkerPool[T, R]) Submit(job Job[T, R]) { panic("TODO") }

func (p *WorkerPool[T, R]) Results() <-chan JobResult[R] { panic("TODO") }

// Close shuts the pool down cleanly: no more jobs accepted, in-flight work
// finished, results channel closed once nothing more can arrive.
//
// The order is the exercise. Closing results before the workers have stopped
// panics on their next send. Closing jobs is what tells them to finish. And
// waiting has to happen somewhere between the two — work out whether Close can
// block, or whether the wait belongs in a goroutine.
//
// Note also that Submit sends on p.jobs with no protection: calling Submit
// after Close panics. Decide whether that is acceptable (it is how the standard
// library treats a closed channel) or whether the type should defend itself.
func (p *WorkerPool[T, R]) Close() {
	panic("TODO")
}

// ---- Pub/Sub ----

type Subscriber[T any] struct {
	ch     chan T
	id     int
	closed bool
}

// PubSub broadcasts each message to every subscriber.
//
// This is the mechanism behind the event bus in L8. The question that decides
// its whole design: what happens when one subscriber is slow?
//
// Blocking means one slow consumer stalls every other. Dropping means messages
// are silently lost. Buffering delays the choice without removing it. There is
// no fourth option, and every real message system is a position on that
// trade-off. Pick one, and be able to say which failure you chose.
type PubSub[T any] struct {
	mu          sync.RWMutex
	subscribers map[int]*Subscriber[T]
	nextID      int
	closed      bool
}

func NewPubSub[T any]() *PubSub[T] { panic("TODO") }

// Subscribe returns a receive channel and the function that cancels the
// subscription.
//
// Returning the unsubscribe function rather than exposing an Unsubscribe(id)
// method means the caller cannot lose the id or pass the wrong one. Same idea
// as context.WithCancel.
//
// Make the cancel function safe to call twice — callers will defer it *and*
// call it explicitly, and double-closing a channel panics.
func (ps *PubSub[T]) Subscribe(bufferSize int) (<-chan T, func()) {
	panic("TODO")
}

// Publish delivers a message to every current subscriber.
//
// Watch the locking. Holding the read lock while sending means a slow
// subscriber blocks anyone trying to subscribe or unsubscribe — and if that
// slow subscriber's own goroutine is what would call unsubscribe, you have a
// deadlock rather than a delay.
func (ps *PubSub[T]) Publish(msg T) {
	panic("TODO")
}

// Close shuts down every subscription.
//
// After this, Publish must not panic and must not send. The closed flag exists
// for that, and every method that touches it needs to check under the lock.
func (ps *PubSub[T]) Close() {
	panic("TODO")
}

// ForwardUntilDone forwards values until the input closes or ctx is cancelled, then
// closes its output.
//
// The most useful pattern in the file, and the one to internalise. Without it,
// `for v := range someChannel` cannot be cancelled: range has no case for
// ctx.Done, so a producer that never closes leaves the consumer stuck forever.
// ForwardUntilDone is the adapter that makes an uncancellable range cancellable.
//
// Note the nested select. Receiving from in has to respect cancellation, and so
// does sending downstream — miss the second and a cancelled pipeline still
// hangs on a full output channel.
func ForwardUntilDone[T any](ctx context.Context, in <-chan T) <-chan T {
	panic("TODO")
}

// Throttle forwards values no faster than one per interval.
//
// time.NewTicker must be stopped or it leaks its underlying timer; defer the
// Stop next to the creation.
//
// Two design questions the tests may not force but production will. Does a
// value that arrives during the wait get delayed or dropped? And does the first
// value wait for the first tick, or go straight through? Rate limiters are
// usually expected to let the first request pass immediately, which a naive
// ticker loop does not do.
func Throttle[T any](ctx context.Context, in <-chan T, interval time.Duration) <-chan T {
	panic("TODO")
}
