// Package datalayer owns shared, time-sensitive Endpoint state.
package datalayer

import (
	"errors"
	"time"

	"github.com/rex/go-ddd-tdd/internal/inference/routing"
)

var ErrInvalidUpdate = errors.New("invalid endpoint update")

type Clock interface {
	Now() time.Time
}

type Update struct {
	EndpointID string
	Endpoint   routing.Endpoint
	Version    uint64
	ObservedAt time.Time
	Deleted    bool
}

type Snapshot struct {
	Version   uint64
	Fresh     []routing.Endpoint
	StaleIDs  []string
	UpdatedAt time.Time
}

// Store is intentionally opaque. Choose its concurrency strategy only after
// writing down which state each lock protects.
type Store struct{}

func NewStore(maxAge time.Duration, clock Clock) (*Store, error) {
	panic("TODO(exercise): require a positive freshness window and non-nil clock")
}

// Apply returns true only when the update advances effective state.
func (s *Store) Apply(update Update) (bool, error) {
	panic("TODO(exercise): validate, reject out-of-order state, and preserve delete tombstones")
}

// Snapshot separates fresh routing input from stale evidence.
func (s *Store) Snapshot() Snapshot {
	panic("TODO(exercise): return deterministic defensive copies without holding a lock in callers")
}
