package language

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// ============================================================
// Errors
//
// Sentinel errors, custom error types, wrapping, chains, errors.Is, and errors.As.
// ============================================================

// ErrorCode is the machine-readable half of an error.
//
// The split matters: the message is for a human reading a log, the code is for
// a program deciding what to do. The HTTP layer in L7 maps these codes onto
// status codes, which is only possible because the code survives wrapping while
// the message gets prefixes bolted onto it.
type ErrorCode string

const (
	CodeNotFound     ErrorCode = "NOT_FOUND"
	CodeConflict     ErrorCode = "CONFLICT"
	CodeValidation   ErrorCode = "VALIDATION"
	CodeUnauthorized ErrorCode = "UNAUTHORIZED"
	CodeInternal     ErrorCode = "INTERNAL"
	CodeTimeout      ErrorCode = "TIMEOUT"
	CodePrecondition ErrorCode = "PRECONDITION_FAILED"
)

// ---- DomainError ----

type DomainError struct {
	Code    ErrorCode
	Message string
	Field   string // See the corresponding tests for the intended behavior.
	Cause   error  // See the corresponding tests for the intended behavior.
}

// Error renders one of exactly four forms, depending on which optional fields
// are set:
//
//	[CODE] message
//	[CODE] field: message
//	[CODE] message: cause
//	[CODE] field: message: cause
//
// Every form starts with the code, and the pieces accumulate — no branch
// discards what an earlier branch produced.
//
// Worked examples, one per form:
//
//	{Code: NOT_FOUND, Message: "order not found"}
//	  → "[NOT_FOUND] order not found"
//
//	{Code: VALIDATION, Field: "email", Message: "invalid format"}
//	  → "[VALIDATION] email: invalid format"
//
//	{Code: CONFLICT, Message: "cannot update", Cause: errors.New("version changed")}
//	  → "[CONFLICT] cannot update: version changed"
//
//	{Code: VALIDATION, Field: "qty", Message: "too large", Cause: errBounds}
//	  → "[VALIDATION] qty: too large: bounds exceeded"
//
// Build it up: start with the code, append the field and a colon if there is
// one, append the message, then append the cause after a colon if there is one.
// Every branch *adds*; none replaces what came before.
func (e *DomainError) Error() string {
	panic("TODO")
}

// Unwrap exposes the cause so errors.Is and errors.As can walk the chain.
//
// Implementing this one method is what turns a wrapper into a link rather than
// a wall. Without it, every error underneath becomes invisible.
//
//	{Cause: errDB}.Unwrap() → errDB
//	{Cause: nil}.Unwrap()   → nil
func (e *DomainError) Unwrap() error {
	panic("TODO")
}

// Is makes two DomainErrors match when their codes match, so a caller can ask
// "is this a not-found?" without caring which entity was missing.
//
// This is the customisation hook errors.Is looks for. Note what it means: a
// bare &DomainError{Code: CodeNotFound} works as a *pattern* rather than as a
// real error value. Convenient, and worth being aware of — equality here is no
// longer reflexive with the message.
//
//	errors.Is(&DomainError{Code: NOT_FOUND, Message: "order 7"},
//	          &DomainError{Code: NOT_FOUND})                    → true
//	errors.Is(thatSameError, &DomainError{Code: VALIDATION})    → false
//	errors.Is(thatSameError, errors.New("order 7"))             → false
//
// The target may be any error, so type-assert with the comma-ok form: a
// non-DomainError target must return false, not panic.
func (e *DomainError) Is(target error) bool {
	panic("TODO")
}

// NewNotFound reports a missing entity.
//
//	NewNotFound("User", "123").Error()
//	  → "[NOT_FOUND] User with id '123' not found"
//
// Note the single quotes around the id. They are not decoration: without them
// an id of "" or " " produces a message nobody can interpret.
func NewNotFound(entity, id string) *DomainError {
	panic("TODO")
}

// NewValidation reports a bad field.
//
//	NewValidation("email", "invalid format").Error()
//	  → "[VALIDATION] email: invalid format"
//
// This is the only constructor that sets Field, which is what makes the second
// Error() form reachable.
func NewValidation(field, reason string) *DomainError {
	panic("TODO")
}

// NewConflict reports a state clash — a duplicate id, a stale version.
//
//	NewConflict("order already confirmed").Error()
//	  → "[CONFLICT] order already confirmed"
//
// This is the code the HTTP layer maps to 409 in L7.
func NewConflict(msg string) *DomainError {
	panic("TODO")
}

// NewInternal wraps an unexpected failure.
//
//	NewInternal(errors.New("connection refused")).Error()
//	  → "[INTERNAL] connection refused"
//
// It sets Cause and no Message, so the underlying text is all the caller sees.
// Decide whether that is right: an INTERNAL error is the one thing that reaches
// an end user, and "connection refused" tells them nothing while telling an
// attacker something.
func NewInternal(cause error) *DomainError {
	panic("TODO")
}

// NewTimeout reports an operation that ran out of time.
//
//	NewTimeout("checkout", context.DeadlineExceeded).Error()
//	  → "[TIMEOUT] operation 'checkout' timed out: context deadline exceeded"
//
// Both a Message and a Cause, so this is the fourth Error() form minus the
// field — a useful one to check your formatting against.
func NewTimeout(operation string, cause error) *DomainError {
	panic("TODO")
}

// ---- MultiError ----

// MultiError collects several failures so they can be reported together.
//
// It is what policy.ValidateAll needed: a form that surfaces one error, then
// the next after you fix it, then a third, is a form people hate.
//
// The mutex is there because ErrorGroup below adds to it from several
// goroutines at once. Check that *every* method touching errs takes it —
// a read that skips the lock still races, and `go test -race` will say so.
type MultiError struct {
	mu   sync.Mutex
	errs []error
}

// NewMultiError returns an empty collector.
//
// A zero MultiError is already usable, so ask yourself what this constructor is
// actually for before writing it — sometimes the honest answer is "consistency
// with the rest of the package", and that is allowed.
func NewMultiError() *MultiError {
	panic("TODO")
}

// Add records a failure, ignoring nil.
//
//	Add(nil)      → nothing recorded
//	Add(errFoo)   → one error recorded
//
// Ignoring nil is what lets callers write `m.Add(doThing())` without a guard at
// every call site, which is most of this type's ergonomic value.
func (m *MultiError) Add(err error) {
	panic("TODO")
}

// HasErrors reports whether anything was collected.
//
// **Take the lock.** ErrorGroup calls Add from several goroutines, so an
// unsynchronised read here is a genuine data race even though it only reads a
// length. `go test -race` will catch it; reasoning that "reading an int is
// atomic anyway" will not save you, because the Go memory model does not
// promise that.
func (m *MultiError) HasErrors() bool {
	panic("TODO")
}

// ToError returns the collector as an error, or nil when nothing was collected.
//
//	nothing added   → nil
//	one error added → a non-nil error
//
// **This is the nil-interface trap, and it is the reason this method exists.**
// Returning `m` unconditionally gives the caller a non-nil error interface even
// when there are no errors, because the interface holds a type. Every
// `if err != nil` downstream then fires on success.
//
// You met this in entity.CheckProduct and again in lab/types. Third time: the
// fix is to return a literal nil, not a typed nil.
func (m *MultiError) ToError() error {
	panic("TODO")
}

// Error joins the collected messages into one string.
//
// Check the test for the separator and for what a single collected error should
// look like — "1 error occurred: ..." reads badly, and so does a bare list when
// there are twelve.
func (m *MultiError) Error() string {
	panic("TODO")
}

// Unwrap returns all the collected errors.
//
// The []error form, not the single-error one. Go 1.20 added support for an
// error that wraps *several* others, and errors.Is walks all the branches. That
// is what makes errors.Is(multi, ErrNotFound) work when only the third
// collected error was a not-found.
//
// Having both Unwrap() error and Unwrap() []error on one type is not allowed —
// look up what happens if you try, and why the language made that choice.
func (m *MultiError) Unwrap() []error {
	panic("TODO")
}

// Errors exposes the collected errors to a caller.
//
// Third time you have met this question: Order.Lines, Customer.Addresses, and
// now here. Same answer, and the mutex means there is a second reason for it
// beyond mutation — handing out the slice lets a caller read it while another
// goroutine appends.
func (m *MultiError) Errors() []error {
	panic("TODO")
}

// ErrorGroup runs functions concurrently and collects every failure.
//
// Compare it with golang.org/x/sync/errgroup, which keeps only the *first*
// error and cancels the rest. Both are right for different jobs: errgroup for
// "these all have to succeed, stop as soon as one does not", this for
// "run everything, tell me all the problems". Validating ten fields against ten
// services wants the second.
type ErrorGroup struct {
	wg     sync.WaitGroup
	errors *MultiError
}

func NewErrorGroup() *ErrorGroup { panic("TODO") }

// Go runs fn in a new goroutine.
//
// The Add-before-launch ordering is not stylistic. Calling wg.Add inside the
// new goroutine is a real race: Wait can observe a zero counter and return
// before that goroutine has started. Write it the wrong way once and run it
// under -race — whether the detector catches it is itself instructive, because
// a race that only shows up under load is the worst kind.
func (g *ErrorGroup) Go(fn func() error) {
	panic("TODO")
}

// Wait blocks until every goroutine has finished, then returns the collected
// errors, or nil when they all succeeded.
func (g *ErrorGroup) Wait() error {
	panic("TODO")
}

// ---- Error Wrapping Helpers ----

// Wrap adds context while keeping the original error reachable:
//
//	Wrap(errors.New("connection refused"), "save order")
//	=> "save order: connection refused"
//
// A nil error must wrap to nil, or every caller needs a guard.
//
// The whole point is the difference between %v and %w in fmt.Errorf. Both
// produce the same string; only one keeps errors.Is working. Write it with %v
// first, watch the test fail, and note that the message looked perfectly fine.
// That is why this bug survives code review.
func Wrap(err error, msg string) error {
	panic("TODO")
}

// WrapIf wraps only when condition holds, and otherwise returns err untouched.
//
// Useful for "add this context only when we are the outermost layer". Also
// worth a moment's suspicion: a helper whose behaviour depends on a boolean
// argument at the call site can be harder to read than the if statement it
// replaced.
func WrapIf(err error, condition bool, msg string) error {
	panic("TODO")
}

// Recover searches an error chain for one of type T and returns it.
//
// It is errors.As with the awkwardness removed — no pre-declared variable, no
// pointer-to-pointer at the call site, and the type stated where you read it:
//
//	de, ok := Recover[*DomainError](err)
//
// The constraint is `T error`, so T is already an error type. Think about
// whether *DomainError or DomainError is the right thing to ask for, and what
// happens if a caller picks the wrong one.
func Recover[T error](err error) (T, bool) {
	panic("TODO")
}

// ---- Validation Builder ----

type ValidationBuilder struct {
	entity string
	errors *MultiError
}

func NewValidationBuilder(entity string) *ValidationBuilder { panic("TODO") }

// Check records a validation failure when condition is false.
//
// Note the polarity: the argument is what must be *true*, so the error is
// recorded when it is not. Getting this backwards inverts every rule in the
// codebase at once, so make sure the first test you run covers both directions.
//
// Returning *ValidationBuilder is what allows chaining:
//
//	NewValidationBuilder("Order").
//	    CheckNotEmpty(cmd.CustomerID, "customer_id").
//	    CheckRange(len(cmd.Lines), 1, 20, "lines").
//	    Build()
//
// Compare that with txdsl's chain from L9. Same shape, same deferred-error
// trick, different problem — worth noticing when a pattern is genuinely
// recurring rather than being copied.
func (vb *ValidationBuilder) Check(condition bool, field, reason string) *ValidationBuilder {
	panic("TODO")
}

// CheckNotEmpty records a failure when value is empty.
//
// Decide whether a string of nothing but spaces counts as empty here. Whichever
// you pick, pick the same thing vo.NewEmail did.
func (vb *ValidationBuilder) CheckNotEmpty(value, field string) *ValidationBuilder {
	panic("TODO")
}

// CheckRange records a failure when value falls outside [min, max].
//
// Inclusive or exclusive bounds is a decision, and an off-by-one here rejects a
// perfectly good order at exactly the boundary — the case least likely to be
// covered by a hand-written test and most likely to be hit in production.
func (vb *ValidationBuilder) CheckRange(value, min, max int, field string) *ValidationBuilder {
	panic("TODO")
}

// Build returns the accumulated failures, or nil when everything passed.
//
// Returning nil on success is the whole contract. A builder that hands back a
// non-nil empty MultiError makes every caller's `if err != nil` fire — the same
// nil-interface trap you met in entity.CheckProduct, arrived at from a
// different direction.
func (vb *ValidationBuilder) Build() error {
	panic("TODO")
}

var (
	_ = errors.As
	_ = fmt.Sprintf
	_ = strings.Join
)
