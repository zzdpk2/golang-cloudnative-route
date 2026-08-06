package runtime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rex/go-ddd-tdd/internal/inference/configuration"
	"github.com/rex/go-ddd-tdd/internal/inference/datalayer"
	"github.com/rex/go-ddd-tdd/internal/inference/epp"
	"github.com/rex/go-ddd-tdd/internal/inference/httpapi"
	"github.com/rex/go-ddd-tdd/internal/inference/routing"
)

func TestRuntimeComposesRealStateIntoHTTPRouting(t *testing.T) {
	var backendCalls atomic.Int64
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		backendCalls.Add(1)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(backend.Close)
	clock := &manualClock{now: time.Unix(100, 0)}
	endpoint := routing.Endpoint{
		ID: "model-a", URL: backend.URL, Models: []string{"tiny-llm"}, Ready: true,
	}
	resolver := &resolverStub{scheduler: fixedScheduler{endpoint: endpoint}}
	runtime, err := New(
		Settings{
			Freshness: time.Minute, QueueMaxWait: time.Second,
			Initial: validConfig(1),
		},
		clock,
		[]datalayer.Update{{
			EndpointID: "model-a", Endpoint: endpoint, Version: 1, ObservedAt: clock.Now(),
		}},
		resolver,
	)
	if err != nil {
		t.Fatal(err)
	}
	proxy, err := httpapi.NewProxy(runtime.Picker(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	response := serve(proxy.Handler())
	if response.Code != http.StatusOK || backendCalls.Load() != 1 {
		t.Fatalf("fresh route status=%d backend_calls=%d body=%s",
			response.Code, backendCalls.Load(), response.Body.String())
	}
	if resolver.version.Load() != 1 {
		t.Fatalf("resolver config version=%d, want 1", resolver.version.Load())
	}

	clock.Advance(2 * time.Minute)
	response = serve(proxy.Handler())
	if response.Code != http.StatusServiceUnavailable || backendCalls.Load() != 1 {
		t.Fatalf("stale route status=%d backend_calls=%d", response.Code, backendCalls.Load())
	}
}

func TestRuntimePublishesConfigUsedByNextDecision(t *testing.T) {
	clock := &manualClock{now: time.Unix(200, 0)}
	endpoint := routing.Endpoint{
		ID: "model-a", URL: "http://model-a.example", Models: []string{"tiny-llm"}, Ready: true,
	}
	resolver := &resolverStub{scheduler: fixedScheduler{endpoint: endpoint}}
	runtime, err := New(
		Settings{Freshness: time.Minute, QueueMaxWait: time.Second, Initial: validConfig(1)},
		clock,
		[]datalayer.Update{{
			EndpointID: "model-a", Endpoint: endpoint, Version: 1, ObservedAt: clock.Now(),
		}},
		resolver,
	)
	if err != nil {
		t.Fatal(err)
	}
	next := validConfig(2)
	next.MaxInFlight = 2
	next.MaxQueued = 3
	if err := runtime.Publish(next); err != nil {
		t.Fatal(err)
	}
	result, err := runtime.Picker().Pick(context.Background(), http.Header{}, []byte(
		`{"model":"tiny-llm","messages":[{"role":"user","content":"hello"}]}`,
	))
	if err != nil {
		t.Fatal(err)
	}
	defer result.Release()
	if result.ConfigVersion != 2 || runtime.Active().Version != 2 {
		t.Fatalf("result version=%d active=%d, want 2", result.ConfigVersion, runtime.Active().Version)
	}
}

func TestRuntimeConfigChangePreservesOwnedPermitAcrossGenerations(t *testing.T) {
	clock := &manualClock{now: time.Unix(300, 0)}
	endpoint := routing.Endpoint{
		ID: "model-a", URL: "http://model-a.example", Models: []string{"tiny-llm"}, Ready: true,
	}
	initial := validConfig(1)
	initial.MaxInFlight = 2
	runtime, err := New(
		Settings{Freshness: time.Minute, QueueMaxWait: time.Second, Initial: initial},
		clock,
		[]datalayer.Update{{
			EndpointID: "model-a", Endpoint: endpoint, Version: 1, ObservedAt: clock.Now(),
		}},
		&resolverStub{scheduler: fixedScheduler{endpoint: endpoint}},
	)
	if err != nil {
		t.Fatal(err)
	}

	first, err := runtime.Picker().Pick(context.Background(), http.Header{}, runtimeValidBody())
	if err != nil {
		t.Fatal(err)
	}
	latest := validConfig(2)
	latest.MaxInFlight = 1
	if err := runtime.Publish(latest); err != nil {
		t.Fatal(err)
	}

	second := make(chan pickResult, 1)
	go func() {
		result, pickErr := runtime.Picker().Pick(context.Background(), http.Header{}, runtimeValidBody())
		second <- pickResult{result: result, err: pickErr}
	}()
	assertPickBlocked(t, second, "new capacity ignored the permit owned by the previous config")

	first.Release()
	first.Release()
	secondResult := awaitPick(t, second)
	if secondResult.ConfigVersion != 2 {
		t.Fatalf("second config version = %d, want 2", secondResult.ConfigVersion)
	}

	third := make(chan pickResult, 1)
	go func() {
		result, pickErr := runtime.Picker().Pick(context.Background(), http.Header{}, runtimeValidBody())
		third <- pickResult{result: result, err: pickErr}
	}()
	first.Release()
	assertPickBlocked(t, third, "old release changed the active capacity generation")
	secondResult.Release()
	thirdResult := awaitPick(t, third)
	thirdResult.Release()
}

func runtimeValidBody() []byte {
	return []byte(`{"model":"tiny-llm","messages":[{"role":"user","content":"hello"}]}`)
}

type pickResult struct {
	result epp.Result
	err    error
}

func assertPickBlocked(t *testing.T, result <-chan pickResult, message string) {
	t.Helper()
	select {
	case got := <-result:
		if got.err != nil {
			t.Fatalf("Pick returned early with error: %v", got.err)
		}
		got.result.Release()
		t.Fatal(message)
	case <-time.After(20 * time.Millisecond):
	}
}

func awaitPick(t *testing.T, result <-chan pickResult) epp.Result {
	t.Helper()
	select {
	case got := <-result:
		if got.err != nil {
			t.Fatal(got.err)
		}
		return got.result
	case <-time.After(time.Second):
		t.Fatal("queued Pick did not receive released capacity")
		return epp.Result{}
	}
}

func serve(handler http.Handler) *httptest.ResponseRecorder {
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/chat/completions",
		strings.NewReader(`{"model":"tiny-llm","messages":[{"role":"user","content":"hello"}]}`),
	)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func validConfig(version uint64) configuration.Config {
	return configuration.Config{
		Version: version, SchedulerName: "default",
		Weights:     map[string]float64{"queue": 1},
		MaxInFlight: 1, MaxQueued: 2,
	}
}

type resolverStub struct {
	version   atomic.Uint64
	scheduler epp.SchedulingProfile
}

func (r *resolverStub) Resolve(config configuration.Config) (epp.SchedulingProfile, error) {
	r.version.Store(config.Version)
	return r.scheduler, nil
}

type fixedScheduler struct {
	endpoint routing.Endpoint
}

func (s fixedScheduler) Schedule(
	context.Context,
	routing.Request,
	[]routing.Endpoint,
) (routing.Decision, error) {
	selected := routing.ScoredEndpoint{Endpoint: s.endpoint}
	return routing.Decision{Selected: selected}, nil
}

type manualClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *manualClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *manualClock) Advance(duration time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(duration)
	c.mu.Unlock()
}
