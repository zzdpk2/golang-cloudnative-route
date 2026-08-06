package fp

// ============================================================
// Result Type
//
// An Ok/Err carrier with Map, FlatMap, and unwrap helpers.
// ============================================================

// ---- Result[T] ----

// Result carries either a value or an error, never both.
//
// Go already has (T, error), so be clear about what this buys and what it
// costs before you fall in love with it. It buys *chaining*: with Result you
// can write a validation pipeline where the first failure short-circuits the
// rest and no `if err != nil` appears in between, which is what
// application.CreateOrder wants for its five-step check.
//
// It costs idiom. Go code returns tuples, so a Result has to be converted at
// every boundary, and a reviewer who has not seen it will be slower reading it
// than reading the ifs. Use it inside a pipeline, convert at the edges — that
// is why FromTuple and ToTuple exist and why they matter more than Map does.
type Result[T any] struct {
	value T
	err   error
	ok    bool
}

func Ok[T any](v T) Result[T] { panic("TODO") }

func Err[T any](err error) Result[T] { panic("TODO") }

// FromTuple lifts an ordinary Go (T, error) return into a Result. It is the
// entry point to a pipeline.
func FromTuple[T any](v T, err error) Result[T] {
	panic("TODO")
}

// IsOk / IsErr
func (r Result[T]) IsOk() bool  { panic("TODO") }
func (r Result[T]) IsErr() bool { panic("TODO") }

// Unwrap returns the value, and panics when there is none.
//
// Only reach for this when a failure means the program itself is wrong — the
// same bar as vo.MustNewMoney. In ordinary code use UnwrapOr or ToTuple. A
// library that panics on data it was handed is a library that takes the
// caller's process down over a typo.
func (r Result[T]) Unwrap() T {
	panic("TODO")
}

// UnwrapOr returns the value, or defaultVal when the Result is an error.
func (r Result[T]) UnwrapOr(defaultVal T) T {
	panic("TODO")
}

// UnwrapOrElse returns the value, or computes a fallback from the error.
//
// The difference from UnwrapOr is that the fallback is only computed when it is
// needed. That matters when producing it is expensive, and it is the same
// reason the standard library has both Value and Func variants all over.
func (r Result[T]) UnwrapOrElse(fn func(error) T) T {
	panic("TODO")
}

// ToTuple converts back to Go's native (T, error). It is the exit from a
// pipeline, and every exported function should be doing this.
func (r Result[T]) ToTuple() (T, error) {
	panic("TODO")
}

func (r Result[T]) Error() error { panic("TODO") }

// Map applies fn to the value, and does nothing to an error.
//
// "Does nothing to an error" is the entire mechanism. Because every combinator
// passes failures through untouched, a chain of ten Maps needs no error check
// between them — the failure travels to the end and comes out of ToTuple.
func (r Result[T]) Map(fn func(T) T) Result[T] {
	panic("TODO")
}

// MapResult is Map that may change the type.
//
// It is a free function rather than a method because Go methods cannot
// introduce new type parameters. This is the same restriction that made
// GetMetadataAs and txdsl.GetAs free functions, and it is why a fluent
// Result API in Go can never be as smooth as it is in Rust or Scala. Worth
// knowing before you build an abstraction that fights the language.
func MapResult[T any, U any](r Result[T], fn func(T) U) Result[U] {
	panic("TODO")
}

// FlatMap applies a function that itself returns a Result, without nesting.
//
// Map with a func(T) Result[T] would give you Result[Result[T]]. FlatMap is
// what keeps the chain flat, and it is the operation that makes a pipeline of
// *fallible* steps possible rather than just a pipeline of transformations.
func (r Result[T]) FlatMap(fn func(T) Result[T]) Result[T] {
	panic("TODO")
}

// FlatMapResult is FlatMap that may change the type.
func FlatMapResult[T any, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	panic("TODO")
}

// MapErr transforms the error and leaves a success untouched — the mirror of
// Map. Useful for adding context without unwrapping.
func (r Result[T]) MapErr(fn func(error) error) Result[T] {
	panic("TODO")
}

// Recover turns a failure back into a success, and leaves a success alone.
//
// This is the fallback step: try the cache, and on a miss go to the database.
func (r Result[T]) Recover(fn func(error) Result[T]) Result[T] {
	panic("TODO")
}

// Match collapses a Result into a single value by handling both cases.
//
// Unlike everything above, Match *forces* you to deal with the error — there is
// no path through it that ignores the failure. That is what pattern matching
// buys in languages that have it, and it is the closest Go gets.
func Match[T any, U any](r Result[T], onOk func(T) U, onErr func(error) U) U {
	panic("TODO")
}

// Catch runs fn and converts a panic into an error Result.
//
// Three things to get right, and the test checks all of them:
//   - recover only works inside a deferred function, called directly by the
//     function that panicked
//   - the recovered value is an `any`. It is commonly an error, often a string,
//     and can be anything at all. Handle the non-error case rather than
//     asserting and hoping.
//   - the result has to be assigned to the *named* return value, because the
//     return statement has already run by the time the defer executes. Same
//     mechanism as Customer.AuditLog.
//
// Then the judgement call: this converts panics into values, which is exactly
// what you want at a goroutine boundary or an HTTP handler, and exactly what
// you do not want sprinkled through ordinary code. A panic usually means an
// invariant broke, and swallowing it hides the bug rather than handling it.
func Catch[T any](fn func() T) (result Result[T]) {
	panic("TODO")
}

// ---- Option[T] ----

// Option is a value that may be absent.
//
// Where Result says "this failed, and here is why", Option says "there is
// nothing here" with no explanation. Use Option when absence is normal — a
// cache miss, an optional field — and Result when it is a failure.
//
// The pointer is how absence is represented, which means Some(nil) on a pointer
// type is a Some containing nil. Convince yourself that is coherent.
type Option[T any] struct {
	value *T
}

func Some[T any](v T) Option[T] { panic("TODO") }
func None[T any]() Option[T]    { panic("TODO") }

func (o Option[T]) IsSome() bool { panic("TODO") }
func (o Option[T]) IsNone() bool { panic("TODO") }

// Unwrap returns the value and panics on None. Same warning as Result.Unwrap.
func (o Option[T]) Unwrap() T {
	panic("TODO")
}

// UnwrapOr returns the value, or def when there is none.
func (o Option[T]) UnwrapOr(def T) T {
	panic("TODO")
}

// Map transforms a present value and leaves None alone.
func (o Option[T]) Map(fn func(T) T) Option[T] {
	panic("TODO")
}

// FlatMap chains a function that itself returns an Option.
func (o Option[T]) FlatMap(fn func(T) Option[T]) Option[T] {
	panic("TODO")
}

// ToResult attaches a reason to an absence, turning Option into Result.
//
// This is the bridge between "nothing here" and "this failed", and it is where
// a repository turns a cache miss into a NotFound.
func (o Option[T]) ToResult(err error) Result[T] {
	panic("TODO")
}

// FromPtr lifts a possibly-nil pointer into an Option.
//
// This is the practical reason Option earns its place in Go: it turns a nil
// pointer — which the compiler will happily let you dereference — into
// something you have to open before you can use.
func FromPtr[T any](ptr *T) Option[T] {
	panic("TODO")
}

// ---- Either[L, R] ----

// Either holds one of two *different* types.
//
// Result is Either specialised to "error or value". The general version shows
// up when both outcomes are legitimate: a payment that either settled or needs
// 3-D Secure, say, where neither is a failure.
//
// Note the type can represent nonsense — both sides nil, or in principle both
// set. Ask what it would take to make those unrepresentable in Go, and why this
// implementation did not bother.
type Either[L any, R any] struct {
	left  *L
	right *R
}

func Left[L any, R any](v L) Either[L, R] { panic("TODO") }

func Right[L any, R any](v R) Either[L, R] { panic("TODO") }

func (e Either[L, R]) IsLeft() bool  { panic("TODO") }
func (e Either[L, R]) IsRight() bool { panic("TODO") }

// Left returns the left value and whether it was present.
func (e Either[L, R]) Left() (L, bool) {
	panic("TODO")
}

// Right returns the right value and whether it was present.
func (e Either[L, R]) Right() (R, bool) {
	panic("TODO")
}

// Fold collapses an Either into one value by handling both sides — Match,
// generalised. Decide what it should do with an Either that is neither.
func Fold[L any, R any, U any](e Either[L, R], onLeft func(L) U, onRight func(R) U) U {
	panic("TODO")
}
