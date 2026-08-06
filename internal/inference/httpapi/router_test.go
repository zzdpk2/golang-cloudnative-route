package httpapi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rex/go-ddd-tdd/internal/inference/epp"
	"github.com/rex/go-ddd-tdd/internal/inference/routing"
)

func TestRouterSchedulesAndForwardsChatCompletion(t *testing.T) {
	t.Parallel()

	cached := newBackend(t, "cached")
	headroom := newBackend(t, "headroom")
	var releases atomic.Int64
	picker := &pickerStub{
		ready: true,
		endpoints: []routing.Endpoint{
			{
				ID: "cached", URL: cached.URL, Models: []string{"tiny-llm"}, Ready: true,
				KVCacheUtilization: 0.80, CachedPrefixes: []string{"shared"},
			},
			{
				ID: "headroom", URL: headroom.URL, Models: []string{"tiny-llm"}, Ready: true,
				KVCacheUtilization: 0.10,
			},
		},
		result: epp.Result{
			Request: routing.Request{Model: "tiny-llm", PrefixKey: "shared"},
			Decision: routing.Decision{
				Selected: routing.ScoredEndpoint{Endpoint: routing.Endpoint{
					ID: "cached", URL: cached.URL, Models: []string{"tiny-llm"}, Ready: true,
				}},
				Candidates: []routing.ScoredEndpoint{
					{Endpoint: routing.Endpoint{ID: "cached"}, Score: 2},
					{Endpoint: routing.Endpoint{ID: "headroom"}, Score: 1},
				},
			},
			Release: func() { releases.Add(1) },
		},
	}
	router, err := NewProxy(picker, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/chat/completions",
		strings.NewReader(`{"model":"tiny-llm","messages":[{"role":"user","content":"hello"}]}`),
	)
	request.Header.Set("X-Prefix-Key", "shared")
	response := httptest.NewRecorder()
	router.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("X-Inference-Endpoint"); got != "cached" {
		t.Fatalf("endpoint header = %q, want cached", got)
	}
	if !strings.Contains(response.Body.String(), `"endpoint":"cached"`) {
		t.Fatalf("body = %s", response.Body.String())
	}
	if got := releases.Load(); got != 1 {
		t.Fatalf("release calls = %d, want exactly one", got)
	}

	traceResponse := httptest.NewRecorder()
	router.Handler().ServeHTTP(
		traceResponse,
		httptest.NewRequest(http.MethodGet, "/debug/last-decision", nil),
	)
	if traceResponse.Code != http.StatusOK {
		t.Fatalf("decision trace status = %d, body = %s", traceResponse.Code, traceResponse.Body.String())
	}
	if !strings.Contains(traceResponse.Body.String(), `"prefix_key":"shared"`) ||
		!strings.Contains(traceResponse.Body.String(), `"candidates"`) {
		t.Fatalf("decision trace does not expose request and candidate scores: %s", traceResponse.Body.String())
	}

	metricsResponse := httptest.NewRecorder()
	router.Handler().ServeHTTP(
		metricsResponse,
		httptest.NewRequest(http.MethodGet, "/metrics", nil),
	)
	if metricsResponse.Code != http.StatusOK ||
		!strings.Contains(
			metricsResponse.Body.String(),
			`inference_router_decisions_total{endpoint="cached"} 1`,
		) {
		t.Fatalf("metrics do not expose the routing decision: %s", metricsResponse.Body.String())
	}
}

func TestRouterReturnsUnavailableForUnsupportedModel(t *testing.T) {
	t.Parallel()

	router, err := NewProxy(&pickerStub{ready: true, err: routing.ErrNoEndpoint}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/chat/completions",
		strings.NewReader(`{"model":"missing","messages":[{"role":"user","content":"hello"}]}`),
	)
	response := httptest.NewRecorder()
	router.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}

func TestSimulatorHonorsCancellation(t *testing.T) {
	t.Parallel()

	simulator, err := NewSimulator("slow", "tiny-llm", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/chat/completions",
		strings.NewReader(`{"model":"tiny-llm","messages":[{"role":"user","content":"hello"}]}`),
	).WithContext(ctx)
	response := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		simulator.Handler().ServeHTTP(response, request)
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("simulator did not stop after cancellation")
	}
}

func newBackend(t *testing.T, id string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Errorf("decode forwarded request: %v", err)
		}
		writeJSON(w, http.StatusOK, map[string]any{"endpoint": id})
	}))
	t.Cleanup(server.Close)
	return server
}

func TestRouterRejectsUnknownJSONFields(t *testing.T) {
	t.Parallel()

	router, err := NewProxy(&pickerStub{ready: true, err: epp.ErrInvalidRequest}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	payload := bytes.NewBufferString(
		`{"model":"tiny-llm","messages":[{"role":"user","content":"hi"}],"surprise":true}`,
	)
	response := httptest.NewRecorder()
	router.Handler().ServeHTTP(
		response,
		httptest.NewRequest(http.MethodPost, "/v1/chat/completions", payload),
	)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestRouterRemovesConnectionDeclaredHopByHopHeaders(t *testing.T) {
	t.Parallel()

	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type":   []string{"application/json"},
				"Connection":     []string{"X-Internal-Hop"},
				"X-Internal-Hop": []string{"must-not-leak"},
				"X-End-To-End":   []string{"preserved"},
			},
			Body: io.NopCloser(strings.NewReader(`{"ok":true}`)),
		}, nil
	})}
	picker := &pickerStub{ready: true, result: resultForURL("one", "http://one.example")}
	router, err := NewProxy(picker, client, nil)
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/chat/completions",
		strings.NewReader(`{"model":"tiny-llm","messages":[{"role":"user","content":"hello"}]}`),
	)
	response := httptest.NewRecorder()
	router.Handler().ServeHTTP(response, request)
	if got := response.Header().Get("X-Internal-Hop"); got != "" {
		t.Fatalf("dynamic hop-by-hop header leaked: %q", got)
	}
	if got := response.Header().Get("X-End-To-End"); got != "preserved" {
		t.Fatalf("end-to-end header = %q, want preserved", got)
	}
}

func TestRouterFlushesStreamingChunks(t *testing.T) {
	t.Parallel()

	releaseSecondChunk := make(chan struct{})
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: first\n\n")
		w.(http.Flusher).Flush()
		select {
		case <-releaseSecondChunk:
			_, _ = io.WriteString(w, "data: second\n\n")
			w.(http.Flusher).Flush()
		case <-request.Context().Done():
		}
	}))
	t.Cleanup(backend.Close)

	picker := &pickerStub{ready: true, result: resultForURL("streaming", backend.URL)}
	router, err := NewProxy(picker, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(router.Handler())
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		server.URL+"/v1/chat/completions",
		strings.NewReader(
			`{"model":"tiny-llm","stream":true,"messages":[{"role":"user","content":"hello"}]}`,
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	firstLine, err := bufio.NewReader(response.Body).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if firstLine != "data: first\n" {
		t.Fatalf("first streamed line = %q", firstLine)
	}
	close(releaseSecondChunk)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

type pickerStub struct {
	result    epp.Result
	err       error
	endpoints []routing.Endpoint
	ready     bool
}

func (p *pickerStub) Pick(context.Context, http.Header, []byte) (epp.Result, error) {
	return p.result, p.err
}

func (p *pickerStub) Endpoints() []routing.Endpoint { return p.endpoints }

func (p *pickerStub) Ready() bool { return p.ready }

func resultForURL(id, url string) epp.Result {
	endpoint := routing.Endpoint{ID: id, URL: url, Models: []string{"tiny-llm"}, Ready: true}
	return epp.Result{
		Request:  routing.Request{Model: "tiny-llm"},
		Decision: routing.Decision{Selected: routing.ScoredEndpoint{Endpoint: endpoint}},
		Release:  func() {},
	}
}
