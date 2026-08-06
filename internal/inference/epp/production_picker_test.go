package epp

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/rex/go-ddd-tdd/internal/inference/configuration"
	"github.com/rex/go-ddd-tdd/internal/inference/datalayer"
	"github.com/rex/go-ddd-tdd/internal/inference/routing"
)

func TestProductionPickerRejectsStaleOnlyState(t *testing.T) {
	t.Parallel()
	source := &snapshotStub{snapshot: datalayer.Snapshot{StaleIDs: []string{"model-a"}}}
	admission := &admissionStub{}
	resolver := &resolverStub{scheduler: schedulerStub{}}
	picker := mustProductionPicker(t, source, admission, resolver)

	_, err := picker.Pick(context.Background(), http.Header{}, validBody())
	if !errors.Is(err, routing.ErrNoEndpoint) {
		t.Fatalf("Pick error = %v, want ErrNoEndpoint", err)
	}
	if resolver.calls.Load() != 0 {
		t.Fatal("scheduler ran with no fresh Endpoint state")
	}
	if admission.acquired.Load() != admission.released.Load() {
		t.Fatalf("admission leak: acquired=%d released=%d", admission.acquired.Load(), admission.released.Load())
	}
}

func TestProductionPickerUsesOneActiveConfigAndFreshSnapshot(t *testing.T) {
	t.Parallel()
	endpoint := routing.Endpoint{
		ID: "model-a", URL: "http://model-a.example", Models: []string{"tiny-llm"}, Ready: true,
	}
	source := &snapshotStub{snapshot: datalayer.Snapshot{Version: 9, Fresh: []routing.Endpoint{endpoint}}}
	admission := &admissionStub{}
	resolver := &resolverStub{scheduler: schedulerStub{endpoint: endpoint}}
	config := configStub{config: configuration.Config{
		Version: 7, SchedulerName: "balanced", Weights: map[string]float64{"queue": 1},
		MaxInFlight: 2, MaxQueued: 4,
	}}
	handler := NewDefaultRequestHandler()
	picker, err := NewProductionPicker(handler, source, admission, config, resolver)
	if err != nil {
		t.Fatal(err)
	}

	result, err := picker.Pick(context.Background(), http.Header{}, validBody())
	if err != nil {
		t.Fatal(err)
	}
	if result.Decision.Selected.Endpoint.ID != "model-a" {
		t.Fatalf("selected = %+v", result.Decision.Selected.Endpoint)
	}
	if result.ConfigVersion != 7 {
		t.Fatalf("result config version = %d, want 7", result.ConfigVersion)
	}
	if resolver.version.Load() != 7 {
		t.Fatalf("resolved config version = %d, want 7", resolver.version.Load())
	}
	if admission.lastModel != "tiny-llm" {
		t.Fatalf("admission model = %q", admission.lastModel)
	}
	result.Release()
	result.Release()
	if admission.acquired.Load() != 1 || admission.released.Load() != 1 {
		t.Fatalf("permit ownership acquired=%d released=%d", admission.acquired.Load(), admission.released.Load())
	}
}

func TestProductionPickerCancellationReleasesEveryAcquiredLayer(t *testing.T) {
	t.Parallel()
	endpoint := routing.Endpoint{
		ID: "model-a", URL: "http://model-a.example", Models: []string{"tiny-llm"}, Ready: true,
	}
	source := &snapshotStub{snapshot: datalayer.Snapshot{Fresh: []routing.Endpoint{endpoint}}}
	admission := &admissionStub{}
	resolver := &resolverStub{err: context.Canceled}
	picker := mustProductionPicker(t, source, admission, resolver)

	_, err := picker.Pick(context.Background(), http.Header{}, validBody())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Pick error = %v, want context.Canceled", err)
	}
	if admission.acquired.Load() != admission.released.Load() {
		t.Fatalf("permit leak: acquired=%d released=%d", admission.acquired.Load(), admission.released.Load())
	}
}

func mustProductionPicker(
	t *testing.T,
	source EndpointSnapshotSource,
	admission RuntimeAdmitter,
	resolver SchedulerResolver,
) *ProductionPicker {
	t.Helper()
	picker, err := NewProductionPicker(
		NewDefaultRequestHandler(),
		source,
		admission,
		configStub{config: configuration.Config{
			Version: 1, SchedulerName: "default", Weights: map[string]float64{"queue": 1},
			MaxInFlight: 1, MaxQueued: 1,
		}},
		resolver,
	)
	if err != nil {
		t.Fatal(err)
	}
	return picker
}

func validBody() []byte {
	return []byte(`{"model":"tiny-llm","messages":[{"role":"user","content":"hello"}]}`)
}

type snapshotStub struct {
	snapshot datalayer.Snapshot
}

func (s *snapshotStub) Snapshot() datalayer.Snapshot { return s.snapshot }

type admissionStub struct {
	acquired  atomic.Int64
	released  atomic.Int64
	lastModel string
}

func (a *admissionStub) Admit(_ context.Context, request routing.Request) (func(), error) {
	a.lastModel = request.Model
	a.acquired.Add(1)
	var released atomic.Bool
	return func() {
		if released.CompareAndSwap(false, true) {
			a.released.Add(1)
		}
	}, nil
}

type configStub struct {
	config configuration.Config
}

func (c configStub) Active() configuration.Config { return c.config }

type resolverStub struct {
	calls     atomic.Int64
	version   atomic.Uint64
	scheduler SchedulingProfile
	err       error
}

func (r *resolverStub) Resolve(config configuration.Config) (SchedulingProfile, error) {
	r.calls.Add(1)
	r.version.Store(config.Version)
	return r.scheduler, r.err
}

type schedulerStub struct {
	endpoint routing.Endpoint
	err      error
}

func (s schedulerStub) Schedule(
	_ context.Context,
	_ routing.Request,
	endpoints []routing.Endpoint,
) (routing.Decision, error) {
	if s.err != nil {
		return routing.Decision{}, s.err
	}
	selected := s.endpoint
	if selected.ID == "" && len(endpoints) > 0 {
		selected = endpoints[0]
	}
	return routing.Decision{Selected: routing.ScoredEndpoint{Endpoint: selected}}, nil
}
