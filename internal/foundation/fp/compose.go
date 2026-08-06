package fp

import (
	"sync"
)

// ============================================================
// Function Composition
//
// Compose, pipe, curry, and partial application over generic functions.
// ============================================================

// ---- Compose ----

// Compose returns the function that applies g and then f, so
// Compose(f, g)(x) is f(g(x)).
//
// Note the argument order: the function applied *last* is written first. That
// is the mathematical convention (f ∘ g) and it reads backwards to most
// programmers, which is exactly why Pipe below also exists.
func Compose[A, B, C any](f func(B) C, g func(A) B) func(A) C {
	panic("TODO")
}

// Compose3 chains three functions right to left.
//
// Build it out of Compose rather than writing the nesting again. If your
// two-argument version has the right shape, this one is nearly free — and if it
// is not, that is a signal worth listening to.
func Compose3[A, B, C, D any](f func(C) D, g func(B) C, h func(A) B) func(A) D {
	panic("TODO")
}

// ---- Pipe ----

// Pipe2 applies f then g to a value, so Pipe2(x, f, g) is g(f(x)).
//
// The same operation as Compose, written in the order things actually happen,
// and taking the value immediately rather than returning a function.
//
// Notice the ceiling you are about to hit: Pipe2, Pipe3, then Pipe4 when
// somebody needs it. Go's generics cannot express "a variadic list of functions
// whose types line up end to end", so this library grows by hand. Worth knowing
// before planning a codebase around the style.
func Pipe2[A, B, C any](a A, f func(A) B, g func(B) C) C {
	panic("TODO")
}

// Pipe3 applies three functions in order.
func Pipe3[A, B, C, D any](a A, f func(A) B, g func(B) C, h func(C) D) D {
	panic("TODO")
}

// ---- Curry / Partial ----

// Curry2 turns a two-argument function into a chain of one-argument functions,
// so f(a, b) becomes f(a)(b).
//
// The payoff is that f(a) becomes a *value* — a function specialised to a, which
// you can store and pass around. Everything below is a variation on that idea.
func Curry2[A, B, C any](fn func(A, B) C) func(A) func(B) C {
	panic("TODO")
}

// Uncurry2 is the inverse of Curry2.
//
// A good check on your Curry2: Uncurry2(Curry2(f)) must behave exactly like f
// for every input. Write that as a test if it is not already there — round-trip
// properties catch mistakes that example-based tests walk straight past.
func Uncurry2[A, B, C any](fn func(A) func(B) C) func(A, B) C {
	panic("TODO")
}

// Partial fixes the first argument, returning a function of the rest.
//
// This is the practical face of currying, and you have already used it without
// naming it: policy.BulkOrderDiscount(3, 10) returns a DiscountPolicy with those
// values baked in.
func Partial[A, B, C any](fn func(A, B) C, a A) func(B) C {
	panic("TODO")
}

// PartialRight fixes the last argument instead.
func PartialRight[A, B, C any](fn func(A, B) C, b B) func(A) C {
	panic("TODO")
}

// ---- Memoize ----

// Memoize caches results by argument, so fn runs at most once per distinct key.
//
// Only sound when fn is pure. Memoizing something that reads a database or the
// clock produces a cache that is confidently wrong — and wrong forever, because
// nothing here ever evicts.
//
// Two things this deliberately does not do, both of which matter in production:
// there is no eviction, so it is an unbounded memory leak keyed on whatever the
// caller passes in; and there is no TTL. Compare it with the LRU cache in L6.
// The difference is not sophistication, it is knowing which the situation needs.
func Memoize[K comparable, V any](fn func(K) V) func(K) V {
	panic("TODO")
}

// MemoizeStrict guarantees fn runs exactly once per key, even under
// concurrency.
//
// Memoize above has a gap: two goroutines can both miss and both call fn.
// Usually harmless, occasionally not — if fn charges a card, "at most once" and
// "exactly once" are very different promises.
//
// Closing the gap means holding something *per key* that later arrivals can
// wait on. Work out why simply holding one lock across the call to fn is not an
// acceptable answer, then look at what sync.Once gives you.
func MemoizeStrict[K comparable, V any](fn func(K) V) func(K) V {
	panic("TODO")
}

// ---- Predicate combinators ----

// Predicate is a named function type, which is what lets it carry methods.
//
// This is specification.Specification from L3 again, in function form rather
// than interface form. Having built both, you can answer which you would rather
// maintain — and notice that the answer may hinge on whether the rules need
// names for logging and error messages.
type Predicate[T any] func(T) bool

// And returns a predicate satisfied when both are.
func (p Predicate[T]) And(other Predicate[T]) Predicate[T] {
	panic("TODO")
}

// Or returns a predicate satisfied when either is.
func (p Predicate[T]) Or(other Predicate[T]) Predicate[T] {
	panic("TODO")
}

// Not inverts a predicate.
func (p Predicate[T]) Not() Predicate[T] {
	panic("TODO")
}

// ---- Lazy evaluation ----

// Lazy defers a computation until something needs the value, then remembers it.
type Lazy[T any] struct {
	fn    func() T
	value T
	once  sync.Once
}

func NewLazy[T any](fn func() T) *Lazy[T] { panic("TODO") }

// Force evaluates the computation on the first call and returns the cached
// value afterwards.
//
// sync.Once does more here than a boolean flag would: a second caller arriving
// mid-computation *blocks* until the first finishes, rather than observing a
// half-built value. That memory-ordering guarantee is the reason to reach for
// it over `if !done`.
//
// A Lazy must never be copied after first use, because sync.Once must not be.
// That is why the method is on *Lazy and NewLazy returns a pointer — `go vet`
// will catch you if you forget.
func (l *Lazy[T]) Force() T {
	panic("TODO")
}

// ---- Identity / Const ----

// Identity returns its argument. It looks useless and is not: it is what you
// pass when an API demands a transformation and you want none.
func Identity[T any](v T) T { panic("TODO") }

// Const ignores its argument and always returns v.
func Const[T any, U any](v T) func(U) T { panic("TODO") }

var _ sync.Mutex
