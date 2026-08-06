// Package observability owns Router Prometheus instruments.
package observability

import (
	"errors"
	"net/http"
	"time"

	"github.com/rex/go-ddd-tdd/internal/inference/routing"
)

type Metrics struct{}

var ErrInvalidMetricDimension = errors.New("invalid bounded metric dimension")

func NewMetrics() *Metrics {
	panic("TODO(exercise): register bounded Router metrics in an isolated Registry")
}

func (m *Metrics) Handler() http.Handler {
	panic("TODO(exercise): serve only this module's Registry")
}

func (m *Metrics) Instrument(route string, next http.HandlerFunc) http.HandlerFunc {
	panic("TODO(exercise): record status, duration, and active requests while preserving streaming")
}

func (m *Metrics) ObserveDecision(endpoint string, score float64) {
	panic("TODO(exercise): observe a decision and track labels for later lifecycle cleanup")
}

func (m *Metrics) SetActiveConfigVersion(version uint64) {
	panic("TODO(exercise): expose the complete config version used by the routing decision")
}

func (m *Metrics) ObserveAdmission(result string, queueWait time.Duration) error {
	panic("TODO(exercise): validate a bounded result enum and observe admission count plus queue wait")
}

func (m *Metrics) ObserveTTFT(duration time.Duration) {
	panic("TODO(exercise): observe time to first downstream token separately from total duration")
}

func (m *Metrics) ObserveUpstreamFailure(stage string) error {
	panic("TODO(exercise): validate a bounded connect/headers/stream stage and classify upstream failure")
}

func (m *Metrics) UpdateEndpointFreshness(age map[string]time.Duration) {
	panic("TODO(exercise): publish freshness age and reclaim labels for disappeared Endpoints")
}

func (m *Metrics) UpdateEndpoints(endpoints []routing.Endpoint) {
	panic("TODO(exercise): update gauges and delete metric labels for removed Endpoints without races")
}

type statusWriter struct {
	http.ResponseWriter
}

func (w *statusWriter) WriteHeader(status int) {
	panic("TODO(exercise): capture only the first status code")
}

func (w *statusWriter) Write(body []byte) (int, error) {
	panic("TODO(exercise): imply HTTP 200 before the first body write")
}

func (w *statusWriter) Unwrap() http.ResponseWriter {
	panic("TODO(exercise): allow net/http ResponseController to reach the original writer")
}

func (w *statusWriter) Flush() {
	panic("TODO(exercise): preserve streaming flush behavior")
}
