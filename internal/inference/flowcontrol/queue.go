// Package flowcontrol owns queueing, priority, fairness, expiry, and capacity.
package flowcontrol

import (
	"errors"
	"time"
)

var (
	ErrQueueFull  = errors.New("flow-control queue is full")
	ErrQueueEmpty = errors.New("flow-control queue is empty")
)

type Request struct {
	ID         string
	FairnessID string
	Priority   int
	EnqueuedAt time.Time
	ExpiresAt  time.Time
}

type Queue struct{}

func NewQueue(maxQueued int) (*Queue, error) {
	panic("TODO(exercise): require a positive hard queue bound")
}

func (q *Queue) Enqueue(request Request) error {
	panic("TODO(exercise): validate identity, reject duplicates, and enforce the hard bound")
}

// Dequeue chooses the highest priority first, then rotates fairly across
// FairnessIDs while preserving FIFO within one FairnessID.
func (q *Queue) Dequeue(now time.Time) (Request, error) {
	panic("TODO(exercise): discard expired requests and schedule without tenant starvation")
}

func (q *Queue) Cancel(id string) bool {
	panic("TODO(exercise): remove queued cancellation without corrupting fairness rotation")
}

func (q *Queue) Len() int {
	panic("TODO(exercise): report the current bounded queue depth safely")
}
