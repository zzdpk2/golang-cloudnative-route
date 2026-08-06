// Package unsafe101 isolates exercises that bypass Go's type and memory safety.
//
// Run this package with pointer checking enabled:
//
//	go test -gcflags=all=-d=checkptr ./internal/foundation/performance
package performance

import (
	"errors"
)

var ErrIndexOutOfRange = errors.New("index out of range")

type Pair struct {
	Flag bool
	Size int64
}

// SizePointer returns a pointer to Pair.Size using unsafe.Add and
// unsafe.Offsetof. A normal &pair.Size is always preferable; this exists to
// expose how reflection, serializers, and low-level libraries reach fields.
func SizePointer(pair *Pair) *int64 {
	panic("TODO")
}

// ElementPointer returns a pointer to values[index] using unsafe.SliceData,
// unsafe.Add, and unsafe.Sizeof. All pointer arithmetic must stay in pointer
// form; storing an intermediate uintptr across operations is invalid.
func ElementPointer(values []int, index int) (*int, error) {
	panic("TODO")
}

// StringView returns a zero-copy string view of bytes.
//
// The caller must not mutate bytes while the string is in use. The safe
// conversion string(bytes) should be the default everywhere outside this lab.
func StringView(bytes []byte) string {
	panic("TODO")
}

// SafeString returns a string that is independent of later byte-slice changes.
func SafeString(bytes []byte) string {
	panic("TODO")
}
