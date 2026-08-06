package observability

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rex/go-ddd-tdd/internal/inference/routing"
)

func TestMetricsExposeBoundedRouterSignals(t *testing.T) {
	t.Parallel()

	metrics := NewMetrics()
	handler := metrics.Instrument("chat_completions", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil))
	metrics.ObserveDecision("model-a", 0.875)
	metrics.SetActiveConfigVersion(12)
	metrics.UpdateEndpoints([]routing.Endpoint{{
		ID: "model-a", QueueDepth: 2, KVCacheUtilization: 0.55,
	}})

	response := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := response.Body.String()
	for _, want := range []string{
		`inference_router_requests_total{code="204",route="chat_completions"} 1`,
		`inference_router_decisions_total{endpoint="model-a"} 1`,
		`inference_router_endpoint_queue_depth{endpoint="model-a"} 2`,
		`inference_router_endpoint_kv_cache_utilization_ratio{endpoint="model-a"} 0.55`,
		`inference_router_active_config_version 12`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("metrics output does not contain %q:\n%s", want, body)
		}
	}
}

func TestMetricsExposeFlowFreshnessAndStreamingSignals(t *testing.T) {
	t.Parallel()
	metrics := NewMetrics()
	if err := metrics.ObserveAdmission("admitted", 25*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	metrics.ObserveTTFT(80 * time.Millisecond)
	if err := metrics.ObserveUpstreamFailure("stream"); err != nil {
		t.Fatal(err)
	}
	metrics.UpdateEndpointFreshness(map[string]time.Duration{"model-a": 2 * time.Second})
	if err := metrics.ObserveAdmission("tenant-123", 0); !errors.Is(err, ErrInvalidMetricDimension) {
		t.Fatalf("unbounded admission result error = %v", err)
	}
	if err := metrics.ObserveUpstreamFailure("http://dynamic.example"); !errors.Is(err, ErrInvalidMetricDimension) {
		t.Fatalf("unbounded upstream stage error = %v", err)
	}

	response := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := response.Body.String()
	for _, name := range []string{
		"inference_router_admissions_total",
		"inference_router_queue_wait_seconds",
		"inference_router_time_to_first_token_seconds",
		"inference_router_upstream_failures_total",
		"inference_router_endpoint_freshness_age_seconds",
	} {
		if !strings.Contains(body, name) {
			t.Errorf("metrics output does not contain %s", name)
		}
	}
}

func TestInstrumentPreservesStreamingFlush(t *testing.T) {
	t.Parallel()

	metrics := NewMetrics()
	response := &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
	handler := metrics.Instrument("chat_completions", func(w http.ResponseWriter, _ *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("instrumented response writer lost http.Flusher")
		}
		flusher.Flush()
	})

	handler(response, httptest.NewRequest(http.MethodPost, "/", nil))

	if !response.flushed {
		t.Fatal("instrumented response writer did not flush the underlying writer")
	}
}

func TestUpdateEndpointsReclaimsRemovedEndpointLabels(t *testing.T) {
	t.Parallel()

	metrics := NewMetrics()
	metrics.ObserveDecision("removed-pod", 0.5)
	metrics.UpdateEndpoints([]routing.Endpoint{{ID: "stable-pool"}})

	response := httptest.NewRecorder()
	metrics.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if strings.Contains(response.Body.String(), `endpoint="removed-pod"`) {
		t.Fatalf("removed endpoint label was retained:\n%s", response.Body.String())
	}
}

type flushRecorder struct {
	*httptest.ResponseRecorder
	flushed bool
}

func (r *flushRecorder) Flush() {
	r.flushed = true
	r.ResponseRecorder.Flush()
}
