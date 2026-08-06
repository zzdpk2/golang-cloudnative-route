// Package memory contains exercises about retention rather than allocation.
//
// A value may be tiny and still keep a large object reachable. The garbage
// collector works from reachability, not from which bytes the program intends
// to use.
package concurrency

import (
	"context"
	"errors"
)

var ErrInvalidWindow = errors.New("invalid window")

// DetachWindow returns an independent copy of src[start:end].
//
// Returning src[start:end] would keep the whole backing array reachable. This
// matters when a small field is extracted from a large request buffer.
func DetachWindow[T any](src []T, start, end int) ([]T, error) {
	panic("TODO")
}

// DeleteAt removes one element while preserving order and clears the unused
// tail slot. Clearing matters when T contains pointers, slices, maps, strings,
// interfaces, or functions that keep other memory reachable.
func DeleteAt[T any](values []T, index int) ([]T, error) {
	panic("TODO")
}

// Worker calls work for each received item until input closes or ctx is
// cancelled, then closes the returned done channel.
//
// Every send and receive must have a cancellation path. A goroutine blocked on
// a channel remains a GC root and retains everything it references.
func Worker[T any](
	ctx context.Context,
	input <-chan T,
	work func(T),
) <-chan struct{} {
	panic("TODO")
}
