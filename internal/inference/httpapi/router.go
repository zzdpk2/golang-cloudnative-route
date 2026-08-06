package httpapi

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/rex/go-ddd-tdd/internal/inference/epp"
	"github.com/rex/go-ddd-tdd/internal/inference/routing"
)

const maxRequestBody = 1 << 20

type DecisionTrace struct {
	RecordedAt time.Time        `json:"recorded_at"`
	Request    routing.Request  `json:"request"`
	Decision   routing.Decision `json:"decision"`
}

type PickerClient interface {
	Pick(context.Context, http.Header, []byte) (epp.Result, error)
	Endpoints() []routing.Endpoint
	Ready() bool
}

type Proxy struct{}

func NewProxy(picker PickerClient, client *http.Client, logger *log.Logger) (*Proxy, error) {
	panic("TODO(exercise): validate dependencies and configure transport without a total generation timeout")
}

func (p *Proxy) Handler() http.Handler {
	panic("TODO(exercise): register data, health, debug, and metrics routes with stable methods")
}

func (p *Proxy) handleReady(http.ResponseWriter, *http.Request) {
	panic("TODO(exercise): distinguish liveness from routing readiness")
}

func (p *Proxy) handleEndpoints(http.ResponseWriter, *http.Request) {
	panic("TODO(exercise): expose defensive debug snapshots only")
}

func (p *Proxy) handleMetrics(http.ResponseWriter, *http.Request) {
	panic("TODO(exercise): update Endpoint gauges before serving Prometheus metrics")
}

func (p *Proxy) handleLastDecision(http.ResponseWriter, *http.Request) {
	panic("TODO(exercise): read the trace without racing with a scheduling decision")
}

func (p *Proxy) handleChatCompletions(http.ResponseWriter, *http.Request) {
	panic("TODO(exercise): pick, publish config evidence, forward, stream, classify failures, and release exactly once")
}

func copyUpstreamResponse(http.ResponseWriter, io.Reader, bool) error {
	panic("TODO(exercise): preserve SSE chunk delivery; io.Copy alone does not flush each chunk")
}

func (p *Proxy) recordDecision(routing.Request, routing.Decision) {
	panic("TODO(exercise): publish an immutable decision trace under the correct lock")
}

func pickerErrorStatus(error) (int, string) {
	panic("TODO(exercise): map domain errors to stable HTTP status and machine-readable code")
}

func readBody(http.ResponseWriter, *http.Request) ([]byte, error) {
	panic("TODO(exercise): enforce the Proxy body limit, reject empty input, and close the body")
}

func decodeJSON(http.ResponseWriter, *http.Request, any) error {
	panic("TODO(exercise): decode one strict bounded JSON value")
}

func writeJSON(http.ResponseWriter, int, any) {
	panic("TODO(exercise): write one JSON response with the correct content type")
}

func writeError(http.ResponseWriter, int, string, error) {
	panic("TODO(exercise): preserve a stable external error envelope")
}

func hopByHopHeaders(http.Header) map[string]struct{} {
	panic("TODO(exercise): include standard and Connection-declared hop-by-hop headers")
}
