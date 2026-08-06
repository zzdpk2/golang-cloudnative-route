// Package semantics exercises Go rules that are easy to use correctly by
// accident and difficult to explain precisely.
//
// The exercises are based on the topic coverage of Go (Fundamentals) 101, but
// the scenarios and tests are original to this repository.
package language

import (
	"reflect"
	"strconv"
)

// ConstantQuotients contrasts integer constant division with a division that
// has a floating-point constant operand.
//
// The result must be (1, 1.5). Do not calculate either value at run time.
func ConstantQuotients() (integer int, fractional float64) {
	panic("TODO")
}

// ShiftMask returns a uint64 with bit set. A run-time shift count greater than
// or equal to 64 produces zero because the left operand is already uint64.
func ShiftMask(bit uint) uint64 {
	panic("TODO")
}

// Swap exchanges two values with one parallel assignment.
func Swap[T any](left, right T) (T, T) {
	panic("TODO")
}

// AddAt adds delta to one element and returns the new value.
//
// index must be called exactly once. A compound assignment evaluates its
// addressable left side once; spelling the operation as a separate read and
// write can accidentally call index twice.
func AddAt(values []int, index func() int, delta int) int {
	panic("TODO")
}

// DeferredValues demonstrates both defer rules in one function.
//
// Register one deferred call whose argument captures start, increment current,
// then register a deferred closure that reads current when it executes. Deferred
// calls execute last-in-first-out, so the result is [start+1, start].
func DeferredValues(start int) (trace []int) {
	panic("TODO")
}

// AdjustNamedReturn returns value+1 by changing a named result in a deferred
// function. The return expression is assigned before deferred calls execute.
func AdjustNamedReturn(value int) (result int) {
	panic("TODO")
}

// PanicValue executes fn and returns its recovered panic value. It returns nil
// when fn completes normally.
//
// recover must be called directly by a deferred function. Hiding it in a
// normal helper called by that function does not stop the panic.
func PanicValue(fn func()) (recovered any) {
	panic("TODO")
}

// ParseOrFallback parses input. On failure it returns fallback and the parsing
// error.
//
// Be careful with := inside the error branch: a new variable can shadow the
// result you meant to return.
func ParseOrFallback(input string, fallback int) (value int, err error) {
	panic("TODO")
}

// MutateArray increments the first element of its array copy and returns that
// copy. Passing an array copies all of its elements.
func MutateArray(value [3]int) [3]int {
	panic("TODO")
}

// MutateSlice increments the first element. Passing a slice copies only its
// header, so the caller and callee still refer to the same backing array.
func MutateSlice(value []int) {
	panic("TODO")
}

// AppendDetached returns src plus value without sharing writable backing storage
// with src. The caller may have spare capacity, so append(src, value) is not
// sufficient.
func AppendDetached[T any](src []T, value T) []T {
	panic("TODO")
}

// CloneMap returns an independent shallow copy. Assigning a map copies a small
// descriptor and continues to share the same run-time map.
func CloneMap[K comparable, V any](src map[K]V) map[K]V {
	panic("TODO")
}

// DeletePointerAt removes one element while preserving order. It must clear the
// now-unused tail slot so the backing array does not retain the removed object.
func DeletePointerAt(values []*int, index int) []*int {
	panic("TODO")
}

// DoubleVariadic doubles every argument in place. When the caller passes
// slice..., the variadic parameter shares that slice's backing array.
func DoubleVariadic(values ...int) {
	panic("TODO")
}

type Counter struct {
	Value int
}

// IncrementMapCounter increments one counter in a map.
//
// Map elements are not addressable, so a pointer-receiver mutation cannot be
// called directly on counters[key]. Copy, modify, and store it back.
func IncrementMapCounter(counters map[string]Counter, key string) {
	panic("TODO")
}

// FirstReady performs a non-blocking receive from either channel. A nil channel
// disables its select case. Return false when neither channel is ready.
func FirstReady(left, right <-chan int) (int, bool) {
	panic("TODO")
}

// EnsureMap stores key/value, allocating the map when it is nil, and returns the
// usable map. Reading a nil map is safe; writing one panics.
func EnsureMap(values map[string]int, key string, value int) map[string]int {
	panic("TODO")
}

// Placeholders keep imports available while the exercises are unfinished.
var (
	_ = reflect.ValueOf
	_ = strconv.Atoi
)
