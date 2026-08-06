// Package functional covers closures, defer, panic, and recover — the four Go
// mechanisms most likely to produce code that looks obviously right and is not.
//
// Almost every exercise here has a wrong-looking answer that works and a
// right-looking answer that does not. Predict the output before running the
// test, every time. Being surprised is the point; guessing correctly without
// understanding is not.
package language

// MakeCounters_Wrong is named for what it *used* to be.
//
// Before Go 1.22 the loop variable was declared once for the whole loop, so
// every closure captured the same i and they all returned n. Since Go 1.22 each
// iteration gets a fresh i, and this now does what it looks like it does.
//
// So before drawing any conclusion, check the `go` directive in go.mod, then
// run the test and see which world you are in. Then find the version of this
// bug that still exists today: the change fixed `for i := range n`, but it did
// not fix a variable declared *outside* the loop and captured inside it.
//
// This is the same trap as Order.LineProcessors in L2.
func MakeCounters_Wrong(n int) []func() int { panic("TODO") }

// MakeCounters_Correct returns n closures where closure i returns i, and does
// so regardless of language version.
//
// Two classic techniques: shadow the variable inside the loop body, or pass it
// as an argument to an immediately-invoked function. Both work. Write whichever
// you find clearer, then write the other one too and decide which you would
// rather find in a code review.
func MakeCounters_Correct(n int) []func() int {
	panic("TODO")
}

// Accumulator returns a function that adds to a running total and returns it.
//
// The total survives between calls because the closure captured the variable
// itself, not a copy of its value. That is the whole idea of a closure, and
// it is also why two calls to Accumulator produce two independent totals.
func Accumulator(initial int) func(int) int {
	panic("TODO")
}

// DeferOrder demonstrates that defers run last-in-first-out.
//
// The expected result is:
//
//	main body, third defer, second defer, first defer
//
// LIFO is not arbitrary: defers usually release things acquired in order, and
// releasing in reverse is what makes nested acquisition safe.
//
// There is a catch in the signature. The return value is not named, so the
// value is copied out *before* the defers run. Getting the test to pass will
// force you to notice — and to fix it in a way that also explains
// DeferModifyReturn below.
func DeferOrder() []string {
	panic("TODO")
}

// DeferArgEval shows when a deferred call's *arguments* are evaluated.
//
// They are evaluated at the moment `defer` executes, not when the deferred
// function eventually runs. So a deferred call capturing a value as an argument
// freezes it there and then, even if the variable changes afterwards.
func DeferArgEval() (result string) {
	panic("TODO")
}

// DeferClosureVsArg puts the two side by side: one defer reads a variable
// through the closure, the other receives it as an argument. Change the
// variable in between, and they disagree.
//
// This single difference explains most defer bugs in the wild. Make sure you
// can state the rule in one sentence before moving on.
func DeferClosureVsArg() (closureResult, argResult string) {
	panic("TODO")
}

// DeferModifyReturn shows a deferred function changing the value the caller
// receives, after `return` has already run.
//
// Only possible because the return value is named. `return 1` with a named
// result assigns 1 to n, *then* runs the defers, *then* hands n back — so a
// defer that touches n changes what the caller sees.
//
// This is the mechanism behind Customer.AuditLog and fp.Catch. It is also how
// every recover-and-convert-to-error helper in Go works.
func DeferModifyReturn() (n int) {
	panic("TODO")
}

// ---- Panic / Recover ----

// SafeDiv divides, converting a division-by-zero panic into an error.
//
// recover only works inside a deferred function, and only one called directly
// by the panicking function. Calling recover in a helper that the defer calls
// returns nil and the panic keeps going — try it once so the rule sticks.
func SafeDiv(a, b int) (result int, err error) {
	panic("TODO")
}

// MustParse parses an integer and panics on failure.
//
// Third Must* in the codebase, after vo.MustNewMoney and fp.Result.Unwrap. Same
// rule each time: acceptable when a failure means the program is wrong,
// never for data that came from outside.
func MustParse(s string) int {
	panic("TODO")
}

// SafeCall runs fn and converts any panic into an error.
//
// The recovered value is an `any`. It is often an error, frequently a string,
// and can be any value at all — including nil from a deliberate `panic(nil)`.
// Handle the non-error case rather than asserting and hoping.
//
// This is the guard that belongs at a goroutine boundary and in an HTTP
// handler, because a panic in a goroutine takes down the whole process and no
// caller can catch it. It does not belong scattered through ordinary code,
// where it converts "an invariant broke" into "something returned an error".
func SafeCall(fn func()) (err error) {
	panic("TODO")
}

// Apply maps fn over items.
//
// Decide whether to build a new slice or transform in place, and note that the
// signature does not tell the caller which you chose. That ambiguity is a
// design flaw in the signature — think about what would fix it.
func Apply[T any](items []T, fn func(T) T) []T {
	panic("TODO")
}

type Pred[T any] func(T) bool

// Negate inverts a predicate.
func Negate[T any](p Pred[T]) Pred[T] {
	panic("TODO")
}

// ComposeAll chains any number of same-typed transformations, left to right.
//
// Variadic composition is possible here only because every function has the
// identical type. That is exactly the restriction fp.Pipe2 and Pipe3 could not
// escape — and seeing both makes the reason concrete rather than abstract.
func ComposeAll[T any](fns ...func(T) T) func(T) T {
	panic("TODO")
}

// SortBy returns items sorted by less.
//
// sort.Slice sorts in place, so returning a sorted slice means deciding whether
// the caller's slice may be reordered underneath them. Surprising a caller this
// way is a genuinely nasty bug — the same aliasing question as Order.Lines,
// arriving from yet another direction.
func SortBy[T any](items []T, less func(a, b T) bool) []T {
	panic("TODO")
}

type MethodCounter struct {
	n int
}

func (c *MethodCounter) Inc()       { panic("TODO") }
func (c *MethodCounter) Value() int { panic("TODO") }

// MethodValueDemo uses a *method value*: c.Inc with the receiver already bound,
// producing a func() you can call later.
//
// Increment twice through the method value and return the count, so the test's
// expected 2 comes out.
//
// The receiver is captured when the method value is created. Work out what that
// means for a *value* receiver versus a pointer receiver — one of them takes a
// copy, and the copy is where the surprise lives.
func MethodValueDemo() int {
	panic("TODO")
}

// MethodExprDemo uses a *method expression*: (*MethodCounter).Inc, an ordinary
// function that takes the receiver as its first argument.
//
// Same result as MethodValueDemo, reached the other way. The pair is worth
// holding in your head together — a method value is a method expression with
// the first argument already applied, which is exactly fp.Partial from the
// previous package, built into the language.
func MethodExprDemo() int {
	panic("TODO")
}
