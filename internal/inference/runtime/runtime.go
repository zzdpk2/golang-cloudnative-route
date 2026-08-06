// Package runtime composes the production Data Layer, Flow Control,
// configuration Manager, and Production Picker.
package runtime

import (
	"time"

	"github.com/rex/go-ddd-tdd/internal/inference/configuration"
	"github.com/rex/go-ddd-tdd/internal/inference/datalayer"
	"github.com/rex/go-ddd-tdd/internal/inference/epp"
)

type Settings struct {
	Freshness    time.Duration
	QueueMaxWait time.Duration
	Initial      configuration.Config
}

type Runtime struct{}

func New(
	settings Settings,
	clock datalayer.Clock,
	initial []datalayer.Update,
	resolver epp.SchedulerResolver,
) (*Runtime, error) {
	panic("TODO(exercise): compose the real Store, Controller, Manager, Request Handler, and Production Picker")
}

func (r *Runtime) Picker() *epp.ProductionPicker {
	panic("TODO(exercise): expose the composed Picker without exposing mutable runtime ownership")
}

func (r *Runtime) Apply(update datalayer.Update) (bool, error) {
	panic("TODO(exercise): route source updates through the owned Data Layer")
}

func (r *Runtime) Publish(config configuration.Config) error {
	panic("TODO(exercise): publish config atomically and reconcile capacity changes without dropping owned permits")
}

func (r *Runtime) Active() configuration.Config {
	panic("TODO(exercise): expose one defensive active configuration snapshot")
}
