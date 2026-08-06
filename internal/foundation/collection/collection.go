package collection

import (
	"cmp"
)

// ============================================================
// Slices, Maps, and Sets
//
// Slice internals and aliasing, append and copy, ordered traversal, grouping,
// and set operations.
// ============================================================

// Number constrains to the types you can meaningfully add.
//
// The ~ matters: ~int means "int, or any named type whose underlying type is
// int". Without it, `type OrderCount int` would not satisfy the constraint, and
// domain code is full of named types exactly like that.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~float32 | ~float64
}

// Map applies fn to every element, producing a new slice.
//
// Preallocate with the length you already know. Growing a slice by repeated
// append reallocates and copies as it doubles, and here there is no excuse —
// the final size is len(items) before you start.
//
// Decide what an empty input returns: nil or an empty slice. Whichever you
// pick, every function in this file should agree, and reflect.DeepEqual in the
// tests can tell the two apart.
func Map[T any, U any](items []T, fn func(T) U) []U {
	panic("TODO")
}

// Reduce folds the slice into a single value.
//
// Map and Filter are both Reduce in disguise. Worth proving to yourself once —
// it is the clearest way to see why this one function keeps reappearing in
// every language.
func Reduce[T any, U any](items []T, initial U, fn func(U, T) U) U {
	panic("TODO")
}

// Filter keeps the elements satisfying predicate.
//
// Preallocating here is a judgement call rather than an obvious win: you do not
// know how many will survive. Allocating len(items) wastes memory when few
// match; allocating nothing reallocates when many do. Pick one and be able to
// say why.
func Filter[T any](items []T, predicate func(T) bool) []T {
	panic("TODO")
}

// Find returns the first matching element and whether there was one.
//
// Stop at the first match. The comma-ok return is what distinguishes "not
// found" from "found the zero value", which a bare T could never do.
func Find[T any](items []T, predicate func(T) bool) (T, bool) {
	panic("TODO")
}

// Contains reports whether target is present.
//
// Linear, and that is correct for a slice. If you find yourself calling this
// inside a loop over another slice, you have written an O(n*m) scan where a map
// would have made it O(n+m) — the single most common performance bug in
// otherwise reasonable Go.
func Contains[T comparable](items []T, target T) bool {
	panic("TODO")
}

// Unique removes duplicates, preserving first-seen order.
//
// A map for the seen set, a slice for the order. Using only the map would be
// simpler and would lose the ordering — and map iteration order in Go is
// deliberately randomised, so the result would differ between runs. That is the
// same determinism trap the promotion engine's test hunts for in L15.
func Unique[T comparable](items []T) []T {
	panic("TODO")
}

// GroupBy buckets items by a computed key.
//
// This is the shape behind "orders by customer" and "lines by warehouse" — the
// split step of L15.3's order splitting. Appending to a map entry works
// directly even when the key is absent, because the zero value of a slice is
// nil and append handles nil.
func GroupBy[T any, K comparable](items []T, keyFn func(T) K) map[K][]T {
	panic("TODO")
}

// SortBy returns items ordered by a computed key.
//
// Copy before sorting unless you intend to reorder the caller's slice.
// sort.Slice works in place, so without a copy this function silently mutates
// its argument — the same trap as functional.SortBy and types.SortPeopleByAge.
// All three should behave identically; check that yours do.
//
// Also: is your sort stable? sort.Slice is not. For equal keys the order is
// unspecified, which means a "sorted" report can shuffle between runs.
func SortBy[T any, K cmp.Ordered](items []T, keyFn func(T) K) []T {
	panic("TODO")
}

// Sum totals the elements. An empty slice sums to zero.
func Sum[T Number](items []T) T {
	panic("TODO")
}

// Max returns the largest element and whether there was one.
//
// The bool is doing real work: there is no sensible maximum of nothing, and the
// zero value would be a lie for a slice of negative numbers.
func Max[T cmp.Ordered](items []T) (T, bool) {
	panic("TODO")
}

// Min returns the smallest element and whether there was one.
func Min[T cmp.Ordered](items []T) (T, bool) {
	panic("TODO")
}

// Chunk divides values into consecutive groups of at most size elements.
// Return nil when size is not positive.
func Chunk(values []int, size int) [][]int {
	panic("TODO")
}

// Frequency counts each string in the input.
func Frequency(values []string) map[string]int {
	panic("TODO")
}

// Pair is a two-field tuple, since Go has no tuple type.
type Pair[K any, V any] struct {
	Key   K
	Value V
}

func NewPair[K any, V any](k K, v V) Pair[K, V] { panic("TODO") }

// Zip pairs elements positionally, stopping at the shorter slice.
func Zip[T any, U any](a []T, b []U) []Pair[T, U] {
	panic("TODO")
}

// Unzip splits pairs back into two slices.
//
// Round-trip property worth testing: Unzip(Zip(a, b)) returns a and b unchanged
// whenever they were the same length. State the precondition and you have found
// the edge case.
func Unzip[T any, U any](pairs []Pair[T, U]) ([]T, []U) {
	panic("TODO")
}

// Optional is a value that may be absent.
//
// Third Option-shaped type in this codebase, after fp.Option and the
// interface-based Result in lab/types. They exist independently because each
// package grew its own — which is itself the lesson. Three near-identical types
// with different names is how a codebase accumulates weight, and consolidating
// them would be a genuine improvement to make once everything is green.
type Optional[T any] struct {
	value *T
}

func Some[T any](v T) Optional[T] { panic("TODO") }

func None[T any]() Optional[T] { panic("TODO") }

func (o Optional[T]) IsPresent() bool { panic("TODO") }

// Get returns the value and whether it was present.
func (o Optional[T]) Get() (T, bool) {
	panic("TODO")
}

// OrElse returns the value, or defaultVal when absent.
func (o Optional[T]) OrElse(defaultVal T) T {
	panic("TODO")
}

// Map transforms a present value and leaves an absent one alone.
func (o Optional[T]) Map(fn func(T) T) Optional[T] {
	panic("TODO")
}
