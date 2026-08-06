// Package memorymodel turns happens-before rules into small synchronization
// types. Passing these tests without synchronization is not acceptable even
// when a local run happens to observe the expected values.
package concurrency

import (
	"sync"
	"sync/atomic"
)

// OnceCell runs initialization once. The return from once.Do synchronizes with
// every later return from once.Do, so all callers observe the initialized value.
type OnceCell[T any] struct {
	once  sync.Once
	value T
}

func (c *OnceCell[T]) Get(initialize func() T) T {
	panic("TODO")
}

// Publication publishes one immutable value by closing ready. A channel close
// happens before a receive that observes the close.
type Publication[T any] struct {
	once  sync.Once
	value T
	ready chan struct{}
}

func NewPublication[T any]() *Publication[T] {
	panic("TODO")
}

func (p *Publication[T]) Publish(value T) {
	panic("TODO")
}

func (p *Publication[T]) Wait() T {
	panic("TODO")
}

// Snapshot atomically replaces a pointer to an immutable snapshot. Do not
// mutate a value after storing its address; atomic publication does not make
// the pointed-to object safe for concurrent mutation.
type Snapshot[T any] struct {
	value atomic.Pointer[T]
}

func (s *Snapshot[T]) Store(value T) {
	panic("TODO")
}

func (s *Snapshot[T]) Load() (T, bool) {
	panic("TODO")
}

// Queue uses sync.Cond for a condition over shared state: data is available or
// the queue is closed.
type Queue[T any] struct {
	mu     sync.Mutex
	ready  *sync.Cond
	items  []T
	closed bool
}

func NewQueue[T any]() *Queue[T] {
	panic("TODO")
}

// Put returns false after Close.
func (q *Queue[T]) Put(value T) bool {
	panic("TODO")
}

// Get blocks while the queue is empty and open. It returns false once the queue
// is closed and drained. Wait must be inside a loop because wakeups only mean
// the condition may have changed.
func (q *Queue[T]) Get() (T, bool) {
	panic("TODO")
}

func (q *Queue[T]) Close() {
	panic("TODO")
}
