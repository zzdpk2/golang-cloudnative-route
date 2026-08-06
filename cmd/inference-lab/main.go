// Command inference-lab is a GPU-free Router exercise harness.
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/rex/go-ddd-tdd/internal/inference/configuration"
	"github.com/rex/go-ddd-tdd/internal/inference/routing"
)

type config struct {
	Endpoints []routing.Endpoint `json:"endpoints"`
}

// allSettings keeps process addresses and learner-authored configuration
// explicit so the full local topology can be tested without fixed ports.
type allSettings struct {
	RouterAddress      string
	EndpointConfigPath string
	RuntimeConfigPath  string
}

const maxConfigSize = 1 << 20

func main() {
	panic("TODO(exercise): parse modes, own signal cancellation, compose the selected runtime, and report terminal errors")
}

func runAll(context.Context, allSettings, *log.Logger) error {
	panic("TODO(exercise): load both configs, compose two Model Servers and one Router, and share one shutdown boundary")
}

func runRouter(
	context.Context,
	string,
	[]routing.Endpoint,
	configuration.Config,
	*log.Logger,
) error {
	panic("TODO(exercise): compose and serve the production Runtime with explicit Router and Model roles")
}

func runModel(context.Context, string, string, string, time.Duration, *log.Logger) error {
	panic("TODO(exercise): construct and serve one Model Server simulator")
}

func loadEndpoints(string) ([]routing.Endpoint, error) {
	panic("TODO(exercise): read one bounded strict JSON document and require at least one Endpoint")
}

func loadRuntimeConfig(string) (configuration.Config, error) {
	panic("TODO(exercise): decode and validate the learning config into one versioned runtime candidate")
}

// serverSpec is the lifecycle seam used to test coordinated serving without
// binding every contract test to an operating-system listener.
type serverSpec struct {
	Name     string
	Serve    func() error
	Shutdown func(context.Context) error
}

func newServer(string, http.Handler) *http.Server {
	panic("TODO(exercise): set explicit HTTP server timeouts suitable for streaming")
}

func serveGroup(context.Context, *log.Logger, ...serverSpec) error {
	panic("TODO(exercise): fail the group on one server error and gracefully drain every server")
}
