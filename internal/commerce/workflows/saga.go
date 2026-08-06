package workflows

import (
	"context"

	"sync"
)

// ============================================================
// Saga
//
// Multi-step workflows that compensate completed steps in reverse order on failure.
// ============================================================

// SagaStep is one unit of forward work plus the compensation that undoes it.
//
// A saga exists because a database transaction cannot span services. You cannot
// BEGIN, charge a card at Stripe, ship via a carrier API, and ROLLBACK. So
// instead of preventing partial states, you accept them and define how to walk
// back out — trading isolation for the ability to cross a service boundary at
// all.
//
// The consequence is uncomfortable and worth sitting with: between step two
// succeeding and its compensation running, the system is in a state that is
// visible and wrong. A user refreshing the page may see it.
type SagaStep struct {
	Name       string
	Execute    func(ctx context.Context) error
	Compensate func(ctx context.Context) error
}

// SagaResult is a full account of what happened. When a distributed workflow
// goes wrong at 3am, "it failed" is not an answer.
//
// CompensateErrs is the field that matters most: a non-empty value means the
// rollback itself failed, so the system is now inconsistent and no amount of
// retrying will fix it. That is a page-a-human condition, not a log line.
type SagaResult struct {
	CompletedSteps []string
	FailedStep     string
	FailError      error
	CompensateErrs []error
}

func (r SagaResult) IsSuccess() bool { panic("TODO") }

func (r SagaResult) Error() string { panic("TODO") }

// Saga is a sequence of steps executed in order.
type Saga struct {
	mu    sync.Mutex
	steps []SagaStep
}

func NewSaga() *Saga { panic("TODO") }

// AddStep appends a step and returns the saga so calls can chain.
func (s *Saga) AddStep(step SagaStep) *Saga {
	panic("TODO")
}

// Execute runs the steps in order, and on the first failure compensates the
// steps that already completed, in reverse.
//
// The rules the tests pin down:
//   - the failing step is not compensated; its work never completed
//   - compensation continues even if one compensation itself fails
//   - a cancelled context stops forward progress but not the rollback
//
// You solved exactly this in L9's txdsl. Before writing anything, go and diff
// the two: this version has no shared state between steps, so it is strictly
// simpler. Ask what that buys and what it costs — a saga whose steps cannot
// pass data is a saga where step three cannot refund the payment step one made.
//
// The context question is the same one txdsl posed, and it has the same answer:
// handing an already-cancelled context to the compensations makes them all fail
// instantly, turning a recoverable failure into an unrecoverable one.
func (s *Saga) Execute(ctx context.Context) SagaResult {
	panic("TODO")
}

// StatefulSaga is the version whose steps can pass values along — a payment id
// created in step one that step two's compensation needs in order to refund it.
//
// The map is shared mutable state touched by every step, which is why there is
// a mutex. Note that the mutex protects the map but not the *logic*: two steps
// reading and writing the same key still race at the business level even when
// the map itself is safe. That distinction — memory safety versus logical
// consistency — is one Go beginners routinely conflate.
type StatefulSaga struct {
	steps []statefulStep
	state map[string]any
	mu    sync.Mutex
}

type statefulStep struct {
	Name       string
	Execute    func(ctx context.Context, state map[string]any) error
	Compensate func(ctx context.Context, state map[string]any) error
}

func NewStatefulSaga() *StatefulSaga { panic("TODO") }

// AddStep appends a step and returns the saga so calls can chain.
func (s *StatefulSaga) AddStep(name string,
	exec func(ctx context.Context, state map[string]any) error,
	comp func(ctx context.Context, state map[string]any) error,
) *StatefulSaga {
	panic("TODO")
}

// SetState seeds a value before execution starts.
func (s *StatefulSaga) SetState(key string, value any) *StatefulSaga {
	panic("TODO")
}

// GetState reads a value, reporting whether it was present.
//
// Compare this untyped map[string]any with txdsl.GetAs, which you gave a type
// parameter. Both work; one of them turns a typo into a compile error and the
// other into a nil at midnight. Having built both, decide which you would put
// in production.
func (s *StatefulSaga) GetState(key string) (any, bool) {
	panic("TODO")
}

// Execute is Saga.Execute with the state map threaded through every step and
// every compensation.
//
// Watch the locking. Steps receive the map directly and may hold it for as long
// as they run — so holding the saga's own mutex across a step's execution would
// serialise nothing useful and deadlock the moment a step calls back into
// SetState. Work out where the lock actually needs to be held.
func (s *StatefulSaga) Execute(ctx context.Context) SagaResult {
	panic("TODO")
}
