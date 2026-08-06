package epp

import (
	"context"
	"net/http"

	"github.com/rex/go-ddd-tdd/internal/inference/configuration"
	"github.com/rex/go-ddd-tdd/internal/inference/datalayer"
	"github.com/rex/go-ddd-tdd/internal/inference/routing"
)

type EndpointSnapshotSource interface {
	Snapshot() datalayer.Snapshot
}

// Result transfers request and capacity ownership from the Picker to the
// transport. Release must be safe to call repeatedly, but takes effect once.
type Result struct {
	Request       routing.Request
	Decision      routing.Decision
	ConfigVersion uint64
	Release       func()
}

type RuntimeAdmitter interface {
	Admit(context.Context, routing.Request) (release func(), err error)
}

type SchedulingProfile interface {
	Schedule(context.Context, routing.Request, []routing.Endpoint) (routing.Decision, error)
}

type SchedulerResolver interface {
	Resolve(configuration.Config) (SchedulingProfile, error)
}

type ActiveConfiguration interface {
	Active() configuration.Config
}

// ProductionPicker is the integration exercise. Build it only after the Data
// Layer, Flow Control, and Config contracts are independently green.
type ProductionPicker struct{}

func NewProductionPicker(
	handler *RequestHandler,
	endpoints EndpointSnapshotSource,
	admission RuntimeAdmitter,
	config ActiveConfiguration,
	schedulers SchedulerResolver,
) (*ProductionPicker, error) {
	panic("TODO(exercise): validate and retain the production runtime dependencies")
}

func (p *ProductionPicker) Pick(
	ctx context.Context,
	headers http.Header,
	body []byte,
) (Result, error) {
	panic("TODO(exercise): compose parsing, active config, bounded admission, fresh state, scheduling, and reverse-order release")
}

func (p *ProductionPicker) Endpoints() []routing.Endpoint {
	panic("TODO(exercise): expose only a defensive fresh Endpoint snapshot")
}

func (p *ProductionPicker) Ready() bool {
	panic("TODO(exercise): stale-only state is not routing readiness")
}
