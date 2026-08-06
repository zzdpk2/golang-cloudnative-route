// Package configuration owns validated, versioned runtime configuration.
package configuration

import "errors"

var (
	ErrInvalidConfig  = errors.New("invalid Router configuration")
	ErrStaleVersion   = errors.New("configuration version is not newer")
	ErrUnknownVersion = errors.New("unknown configuration version")
)

type Config struct {
	Version       uint64
	SchedulerName string
	Weights       map[string]float64
	MaxInFlight   int
	MaxQueued     int
}

// Manager is intentionally opaque. Readers must observe one complete Config
// version; they must never observe fields assembled from two publications.
type Manager struct{}

func NewManager(initial Config) (*Manager, error) {
	panic("TODO(exercise): validate, defensively copy, and publish the initial version")
}

func (m *Manager) Active() Config {
	panic("TODO(exercise): return a defensive snapshot of one atomically published version")
}

func (m *Manager) Publish(candidate Config) error {
	panic("TODO(exercise): validate completely before replacing active state and retaining rollback history")
}

func (m *Manager) Rollback(version uint64) error {
	panic("TODO(exercise): atomically republish a known historical configuration without mutating history")
}
