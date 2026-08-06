// Package iter is a pull-based, lazy iterator built out of closures.
//
// Everything here rests on one idea: a sequence is not a data structure, it is
// a function you can call again. State lives in a closure, and nothing is
// computed until somebody asks — which is what makes an infinite sequence like
// Fibonacci possible at all.
//
// Two consequences to keep in mind throughout, because most bugs in this
// package come from forgetting them:
//
//   - a Seq is *stateful and single-use*. Consuming it advances it. Passing the
//     same Seq to two consumers does not give both the same elements; it splits
//     them. Compare that with a slice, which you can range over twice.
//
//   - laziness means side effects happen when the value is *pulled*, not when
//     the pipeline is built. Map(seq, fn) calls fn zero times until something
//     downstream asks for an element.
//
// Go 1.23 added iter.Seq and range-over-func to the standard library, solving
// the same problem with a push-based design instead. Once this package is
// green, go read that API and work out why the language team chose push. The
// comparison is more valuable than the exercise.
package iter

// Seq yields the next element and reports whether there was one.
//
// Once it has returned false, it must keep returning false. Consumers below
// depend on that; a Seq that "recovers" after ending breaks all of them.
type Seq[T any] func() (T, bool)

// FromSlice yields each element of items in order.
//
// This is the simplest closure-with-state in the package. Write it first: every
// other constructor is the same shape with different state.
func FromSlice[T any](items []T) Seq[T] {
	panic("TODO")
}

// Range yields the integers from start up to but not including end.
func Range(start, end int) Seq[int] {
	panic("TODO")
}

// RangeStep is Range with an arbitrary stride.
//
// Decide what a negative step means, and what a zero step means. A zero step
// with the obvious implementation produces an infinite sequence of the same
// number, which is a hang rather than an error — decide whether that is
// acceptable before you ship it.
func RangeStep(start, end, step int) Seq[int] {
	panic("TODO")
}

// Repeat yields value forever.
//
// The first genuinely infinite sequence. It only terminates because something
// downstream — Take, TakeWhile — stops pulling. Calling ToSlice on it hangs,
// and that is not a bug in Repeat.
func Repeat[T any](value T) Seq[T] {
	panic("TODO")
}

// RepeatN yields value exactly n times.
func RepeatN[T any](value T, n int) Seq[T] {
	panic("TODO")
}

// Generate yields seed, then fn(seed), then fn(fn(seed)), forever.
//
// Watch the first element: the seed itself is emitted before fn is ever
// applied. Off-by-one here is the most common mistake in the file, and the test
// checks the first value specifically.
func Generate[T any](seed T, fn func(T) T) Seq[T] {
	panic("TODO")
}

// Fibonacci yields 0, 1, 1, 2, 3, 5, ... forever.
//
// Go's multiple assignment evaluates the whole right-hand side before assigning,
// which makes advancing two variables at once a one-liner. Try it with two
// separate statements and see what breaks.
func Fibonacci() Seq[int] {
	panic("TODO")
}

// Map yields fn applied to each element.
//
// The returned Seq must not call fn until it is itself called. If you find
// yourself looping here, you have written an eager version — the whole point is
// that this returns immediately, having computed nothing.
func Map[T any, U any](seq Seq[T], fn func(T) U) Seq[U] {
	panic("TODO")
}

// Filter yields only the elements satisfying pred.
//
// The one combinator that must loop internally: a single pull from the result
// may need several pulls from the source before it finds a match, or may
// exhaust it entirely. Everything else here pulls at most once per call.
func Filter[T any](seq Seq[T], pred func(T) bool) Seq[T] {
	panic("TODO")
}

// Take yields at most the first n elements.
//
// This is what makes infinite sequences usable. Note that it must stop pulling
// from the source once its own count is reached — pulling and discarding would
// still run the source's side effects.
func Take[T any](seq Seq[T], n int) Seq[T] {
	panic("TODO")
}

// TakeWhile yields elements until pred first fails, then stops.
//
// It must *stay* stopped. The element that failed the predicate has already
// been pulled out of the source and cannot be put back — so once you have seen
// it, the sequence is over even if later elements would pass. That "no
// pushback" limitation is inherent to pull-based iterators, and it is worth
// noticing that a slice-based implementation would not have it.
func TakeWhile[T any](seq Seq[T], pred func(T) bool) Seq[T] {
	panic("TODO")
}

// Skip discards the first n elements and yields the rest.
//
// When should the skipping happen — while building, or on the first pull? Only
// one of those answers is lazy, and the difference is observable when the
// source has side effects.
func Skip[T any](seq Seq[T], n int) Seq[T] {
	panic("TODO")
}

// FlatMap maps each element to a sequence and concatenates the results.
//
// The hard one. You are holding a source sequence and a current inner sequence,
// and a single pull may have to exhaust several inner sequences before finding
// an element — or discover the source is finished. Empty inner sequences are
// the case that breaks naive implementations; make sure the test covers one.
func FlatMap[T any, U any](seq Seq[T], fn func(T) Seq[U]) Seq[U] {
	panic("TODO")
}

// Zip pairs elements from two sequences, ending when either ends.
//
// Careful about what happens at the end: pulling from both before checking
// either means you consume an element from the longer sequence and throw it
// away. Usually harmless, occasionally not. Decide deliberately.
func Zip[T any, U any](a Seq[T], b Seq[U]) Seq[struct {
	First  T
	Second U
}] {
	panic("TODO")
}

// Enumerate pairs each element with its zero-based index.
func Enumerate[T any](seq Seq[T]) Seq[struct {
	Index int
	Value T
}] {
	panic("TODO")
}

// Chain yields all of a, then all of b.
//
// Once a is exhausted you must remember it, not just fall through. A Seq is
// allowed to be expensive to ask, and re-pulling an ended sequence on every
// call is the kind of thing that turns O(n) into O(n²) without anyone noticing.
func Chain[T any](a, b Seq[T]) Seq[T] {
	panic("TODO")
}

// Scan yields the running accumulation, starting with initial.
//
// It is Reduce that shows its work: where Reduce returns only the final answer,
// Scan emits every intermediate one. Running balances and cumulative totals are
// exactly this. Note that it emits initial *before* consuming anything.
func Scan[T any, U any](seq Seq[T], initial U, fn func(U, T) U) Seq[U] {
	panic("TODO")
}

// ---- Consumers ----
//
// Everything below is eager: it pulls until the sequence ends. Handing any of
// them an infinite sequence hangs the program. That asymmetry — lazy
// combinators, eager consumers — is the shape of every stream library there is.

// ToSlice drains the sequence into a slice.
func ToSlice[T any](seq Seq[T]) []T {
	panic("TODO")
}

// Reduce folds the sequence into a single value.
func Reduce[T any, U any](seq Seq[T], initial U, fn func(U, T) U) U {
	panic("TODO")
}

// ForEach calls fn for every element.
func ForEach[T any](seq Seq[T], fn func(T)) {
	panic("TODO")
}

// Count drains the sequence and reports how many elements it had.
func Count[T any](seq Seq[T]) int {
	panic("TODO")
}

// First returns the next element, which for a fresh sequence is the first.
//
// Already written, and worth reading twice: it is literally just a call. That
// is the clearest statement of what a Seq is.
func First[T any](seq Seq[T]) (T, bool) { panic("TODO") }

// Last drains the sequence and returns its final element, if any.
func Last[T any](seq Seq[T]) (T, bool) {
	panic("TODO")
}

// Any reports whether some element satisfies pred.
//
// Stop at the first match. On an infinite sequence with a match, a
// short-circuiting Any terminates and a draining one does not — so this is not
// only an optimisation.
func Any[T any](seq Seq[T], pred func(T) bool) bool {
	panic("TODO")
}

// All reports whether every element satisfies pred.
//
// Stop at the first failure, and remember the empty case is true — the same
// vacuous truth as specification.All in L3.
func All[T any](seq Seq[T], pred func(T) bool) bool {
	panic("TODO")
}
