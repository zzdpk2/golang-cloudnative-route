// Package pipeline composes fallible processing steps.
//
// It is the fourth middleware system in this codebase, and putting all four
// side by side is most of what it teaches:
//
//	minigin           Context.Next, one type for handler and middleware
//	adapter/http      func(http.Handler) http.Handler, the stdlib convention
//	adapter/middleware generic over Req/Resp, transport-agnostic
//	pipeline          generic over In/Out, and the *types change per stage*
//
// That last difference is the interesting one. A Step turns an In into an Out,
// so chaining two steps means the first one's output type must be the second
// one's input type — and the compiler enforces it. An HTTP middleware chain
// cannot express that, because every stage has the same type.
//
// This is what the bulk-import path in L9 runs on: parse → validate → price →
// persist, each stage producing a different shape.
package resilience

import (
	"context"

	"time"
)

// Step turns an input into an output, or fails.
type Step[In any, Out any] func(ctx context.Context, input In) (Out, error)

// Then chains two steps, feeding the first's output into the second.
//
//	parse  := Step[string, Order]
//	price  := Step[Order, PricedOrder]
//	Then(parse, price) → Step[string, PricedOrder]
//
//	if parse fails → the error is returned and price is never called
//
// Note what the signature buys: `Then(price, parse)` will not compile. The type
// parameters make an out-of-order pipeline a build error rather than a runtime
// surprise.
//
// On the failure path you must return a zero Out. A generic function cannot
// return nil for an unconstrained type parameter, so `var zero C` is the idiom
// — you will write it in every function in this file.
func Then[A, B, C any](first Step[A, B], second Step[B, C]) Step[A, C] {
	panic("TODO")
}

// ---- Middleware ----

// Middleware wraps a Step in another Step with the same types.
type Middleware[In any, Out any] func(Step[In, Out]) Step[In, Out]

// WithLogging logs before and after the step.
//
//	on success → logFn("start: input=5"), then logFn("done: output=a:5")
//	on failure → logFn("start: input=5"), then logFn("error: boom")
//
// Check the test for the exact strings. Note that it logs the *input value*,
// which is fine for an int and a privacy incident for a customer record — the
// same warning as adapter/http's LoggingMiddleware.
func WithLogging[In any, Out any](logFn func(string)) Middleware[In, Out] {
	panic("TODO")
}

// WithTiming reports how long the step took.
//
//	recordFn is called exactly once per invocation, on success *and* on failure
//
// Recording only successes is the classic mistake: it makes a service that
// times out on every request look instantaneous, because the slow calls never
// get counted.
func WithTiming[In any, Out any](recordFn func(time.Duration)) Middleware[In, Out] {
	panic("TODO")
}

// WithTimeout gives the step a deadline.
//
//	step takes 50ms, timeout 10ms → the step's ctx is cancelled at 10ms
//
// It cancels the *context*; it cannot stop a step that ignores ctx. Same
// limitation as concurrency.WithTimeout in L6 — a timeout is a request, not a
// kill.
//
// Always defer the cancel, including on the success path, or the timer survives
// until it fires.
func WithTimeout[In any, Out any](timeout time.Duration) Middleware[In, Out] {
	panic("TODO")
}

// WithRetry retries a failing step.
//
//	maxRetries=2 → up to 3 calls in total
//	all fail     → an error wrapping the last one
//	ctx ends     → ctx.Err(), promptly
//
// Note that maxRetries counts *re*tries here, while retry.Do's MaxAttempts
// counts total calls. Two functions in the same codebase counting the same
// thing differently is a genuine API smell — decide which you would keep, and
// what it would cost to align them.
//
// Wait on the context, not on time.Sleep. And ask why this exists at all when
// platform/retry does the same job better: the answer is that this one composes
// into a Step chain, which retry.Do cannot. That is a real reason, and "we
// already have one" is a real counter-argument.
func WithRetry[In any, Out any](maxRetries int, delay time.Duration) Middleware[In, Out] {
	panic("TODO")
}

// WithRecover turns a panic in the step into an error.
//
//	step panics with an error  → that error, wrapped
//	step panics with a string  → an error carrying the string
//	step returns normally      → unchanged
//
// Three things to get right, and you have met all three before:
//   - recover only works in a deferred function called directly by the
//     panicking one
//   - the recovered value is an `any` and may be anything at all
//   - the result must go to the **named** return values, because the return
//     statement has already run — the mechanism from Customer.AuditLog
//
// This belongs at a boundary, not everywhere. A panic means an invariant broke,
// and converting it to an error can hide the bug rather than handle it.
func WithRecover[In any, Out any]() Middleware[In, Out] {
	panic("TODO")
}

// Apply wraps a step so the first middleware listed is the outermost.
//
//	Apply(step, WithLogging(log), WithTiming(rec))
//	  → logging sees the call first and the result last;
//	    timing measures only the step
//
// Fourth time you have written this ordering, after minigin, adapter/http, and
// adapter/middleware. The failure mode is always the same and always quiet: the
// chain runs backwards, nothing errors, and recovery ends up inside the thing
// it was supposed to protect.
func Apply[In any, Out any](step Step[In, Out], middlewares ...Middleware[In, Out]) Step[In, Out] {
	panic("TODO")
}

// ---- Parallel and Fallback ----

// Parallel runs every step on the same input and collects all the results.
//
//	Parallel(stepA, stepB)(ctx, 5) → ["a:5", "b:10"], nil
//
// **Results stay in step order**, not completion order — the test indexes them.
// Preallocating a slice and having each goroutine write to its own index needs
// no lock at all: distinct elements are distinct memory. Appending from several
// goroutines would race *and* scramble the order.
//
// Then the decision the signature leaves open: what happens when one step
// fails? Returning the first error discards the successful results; returning
// them all with a joined error keeps everything but forces the caller to
// inspect. Check the test, then decide what you would want for a real
// fan-out — the "charge card, reserve stock, send email" case makes the
// question concrete.
func Parallel[In any, Out any](steps ...Step[In, Out]) Step[In, []Out] {
	panic("TODO")
}

// Fallback tries each step in turn and returns the first success.
//
//	Fallback(fromCache, fromDB, fromUpstream)
//	  → cache hit          → the cached value, DB never touched
//	  → cache miss         → falls through to the DB
//	  → everything fails   → an error wrapping the last failure
//
// The opposite of Parallel: sequential, lazy, and it stops at the first thing
// that works. This is the read-through cache pattern, and also how a client
// falls back from a fast unreliable endpoint to a slow reliable one.
//
// Decide whether it should stop on a cancelled context rather than dutifully
// trying four more sources that will all fail the same way.
func Fallback[In any, Out any](steps ...Step[In, Out]) Step[In, Out] {
	panic("TODO")
}
