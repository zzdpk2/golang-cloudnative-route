// Package txdsl is the L9 exercise: a declarative transaction DSL.
//
// The problem it exists to solve: placing an order must deduct stock, lock a
// coupon, and create a payment record. Any step can fail, and everything
// already done has to be undone. Written by hand that becomes a pyramid of
//
//	if err := deductStock(); err != nil { return err }
//	if err := lockCoupon(); err != nil { releaseStock(); return err }
//	if err := createPayment(); err != nil { unlockCoupon(); releaseStock(); return err }
//
// where the rollback logic is duplicated, ordered by hand, and wrong the moment
// somebody inserts a fourth step.
//
// You are building the thing that makes that read as a declaration instead:
//
//	New().
//	    Step("deduct stock", deduct, release).
//	    Step("lock coupon", lock, unlock).
//	    Step("create payment", create, void).
//	    Run(ctx)
//
// This is a design exercise. The signatures below are fixed because the tests
// call them, but every decision inside is yours, and the tests deliberately
// leave some of them open. Read DESIGN.md in this directory before you start.
package workflows

import (
	"context"
	"errors"
)

// ErrNilStep reports a step registered without an action.
var ErrNilStep = errors.New("txdsl: step has no action")

// State carries values between steps. Step 3 usually needs the payment id that
// step 2 produced, and a compensation almost always needs to know what its
// forward action actually did.
//
// It is shared across a single Run and touched by every step, so decide what
// happens if a compensation runs while the state is being read.
type State struct {
	// Whatever a key/value carrier needs.
}

// Set stores a value under key, replacing any previous value.
func (s *State) Set(key string, value any) {
	panic("TODO")
}

// Get returns the value stored under key, and whether it was present.
func (s *State) Get(key string) (any, bool) {
	panic("TODO")
}

// GetAs is the typed read: it succeeds only when the key exists *and* holds a T.
//
// Free functions carry the type parameter because Go methods cannot. Once you
// have written it, look at how much noise this saves at the call sites in the
// tests.
func GetAs[T any](s *State, key string) (T, bool) {
	panic("TODO")
}

// StepFunc is one unit of forward work, or the compensation that undoes it.
type StepFunc func(ctx context.Context, state *State) error

// Result describes what a Run did. It is deliberately verbose: when a
// distributed transaction goes wrong at 3am, "it failed" is not an answer.
type Result struct {
	// Completed lists the steps whose action returned nil, in execution order.
	Completed []string
	// FailedStep names the step that broke the chain, empty on success.
	FailedStep string
	// Err is the error that step returned, nil on success.
	Err error
	// Compensated lists the steps that were rolled back, in the order the
	// rollbacks ran.
	Compensated []string
	// CompensateErrs collects failures from the rollbacks themselves. A
	// non-empty value here means the system is now inconsistent and a human
	// has to look at it.
	CompensateErrs []error
}

// OK reports whether every step completed.
func (r Result) OK() bool {
	panic("TODO")
}

// Tx is the builder. Steps run in registration order; compensations run in
// reverse.
type Tx struct {
	// What the builder has to accumulate, plus somewhere to park a
	// registration error until Run is called — see the note on Step.
}

// New starts an empty transaction.
func New() *Tx {
	panic("TODO")
}

// Step registers a unit of work and the compensation that undoes it.
//
// A nil compensate is legal and means "this step needs no undo" — reading is
// the obvious case. A nil action is a programming error.
//
// This returns *Tx so calls can chain, which means it cannot return an error.
// Park the problem and surface it from Run as ErrNilStep. That deferred-error
// pattern is worth recognising: you have seen it in bufio.Writer and in the
// standard library's SQL builders.
func (t *Tx) Step(name string, action, compensate StepFunc) *Tx {
	panic("TODO")
}

// Run executes the steps in order, and on the first failure compensates the
// steps that already completed, in reverse.
//
// The rules the tests pin down:
//   - the failing step is never compensated; its action did not complete
//   - compensation continues through the remaining steps even if one
//     compensation itself fails, because stopping strands more state than
//     carrying on
//   - a cancelled context stops forward progress, and what has been done so far
//     is still compensated
//   - compensation is not abandoned just because ctx is already cancelled;
//     think about which context the rollbacks should actually run under
//
// The last one is the interesting decision. If you pass the cancelled ctx
// straight to the compensations, every one of them fails immediately and you
// have made the inconsistency worse. What would you pass instead, and what new
// risk does that introduce?
func (t *Tx) Run(ctx context.Context) Result {
	panic("TODO")
}
