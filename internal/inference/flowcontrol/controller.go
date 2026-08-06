package flowcontrol

import (
	"context"
	"errors"
	"time"

	"github.com/rex/go-ddd-tdd/internal/inference/routing"
)

var ErrInvalidCapacity = errors.New("invalid flow-control capacity")

// Controller owns the transition between queued and in-flight requests.
// Queue ordering and permit ownership must remain race-free under cancellation.
type Controller struct{}

func NewController(maxInFlight, maxQueued int, maxWait time.Duration) (*Controller, error) {
	panic("TODO(exercise): validate hard bounds and queue wait, then initialize admission state")
}

// Admit either returns an owned permit or waits in the bounded Queue. The
// returned release function must be safe to call more than once.
func (c *Controller) Admit(ctx context.Context, request routing.Request) (release func(), err error) {
	panic("TODO(exercise): derive queue identity, wait under overload, and transfer permits without leaks")
}

func (c *Controller) InFlight() int {
	panic("TODO(exercise): expose a race-free diagnostic count")
}

func (c *Controller) Queued() int {
	panic("TODO(exercise): expose the Queue length without weakening its bound")
}
