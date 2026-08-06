package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rex/go-ddd-tdd/internal/inference/configuration"
	"github.com/rex/go-ddd-tdd/internal/inference/routing"
)

func TestCheckedInConfigLoads(t *testing.T) {
	t.Parallel()

	path := filepath.Join("..", "..", "configs", "inference-lab.json")
	endpoints, err := loadEndpoints(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(endpoints) != 2 {
		t.Fatalf("endpoint count = %d, want 2", len(endpoints))
	}
	for _, endpoint := range endpoints {
		if err := endpoint.Validate(); err != nil {
			t.Fatalf("endpoint %q: %v", endpoint.ID, err)
		}
	}
}

func TestLoadEndpointsRejectsTrailingJSON(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "endpoints.json")
	data := `{"endpoints":[{"id":"one","url":"http://one.example","models":["tiny-llm"],"ready":true}]} {}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := loadEndpoints(path)
	if err == nil || !strings.Contains(err.Error(), "one JSON value") {
		t.Fatalf("error = %v, want trailing JSON error", err)
	}
}

func TestLoadEndpointsRejectsOversizedDocument(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "oversized.json")
	data := strings.Repeat(" ", maxConfigSize) + `{"endpoints":[]}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadEndpoints(path); err == nil {
		t.Fatal("loadEndpoints accepted a document beyond the hard size limit")
	}
}

func TestLoadRuntimeConfigExtractsAndValidatesProductionLimits(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	validPath := writeRuntimeConfig(t, directory, 7, 3, 5)
	got, err := loadRuntimeConfig(validPath)
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != 7 || got.SchedulerName != "balanced" ||
		got.MaxInFlight != 3 || got.MaxQueued != 5 || got.Weights["queue"] != 1 {
		t.Fatalf("runtime config = %+v", got)
	}
	got.Weights["queue"] = 99
	again, err := loadRuntimeConfig(validPath)
	if err != nil {
		t.Fatal(err)
	}
	if again.Weights["queue"] != 1 {
		t.Fatal("loadRuntimeConfig retained mutable state across loads")
	}

	invalidPath := writeRuntimeConfig(t, directory, 8, 0, 5)
	if _, err := loadRuntimeConfig(invalidPath); err == nil {
		t.Fatal("loadRuntimeConfig accepted a non-positive in-flight limit")
	}
	oversizedPath := filepath.Join(directory, "oversized-runtime.yaml")
	oversized := strings.Repeat("#", maxConfigSize+1)
	if err := os.WriteFile(oversizedPath, []byte(oversized), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadRuntimeConfig(oversizedPath); err == nil {
		t.Fatal("loadRuntimeConfig accepted a document beyond the hard size limit")
	}
}

func TestNewServerUsesStreamingSafeBounds(t *testing.T) {
	t.Parallel()
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	server := newServer(":0", handler)
	if server == nil || server.Handler == nil {
		t.Fatal("newServer did not retain the configured handler")
	}
	if server.ReadHeaderTimeout <= 0 || server.IdleTimeout <= 0 || server.MaxHeaderBytes <= 0 {
		t.Fatalf("server bounds are incomplete: %+v", server)
	}
	if server.WriteTimeout != 0 {
		t.Fatalf("WriteTimeout = %v; a total response timeout truncates generation streams", server.WriteTimeout)
	}
}

func TestServeGroupReturnsOneServerFailureAndDrainsEveryServer(t *testing.T) {
	t.Parallel()

	serveFailure := errors.New("listener failed")
	releasePeer := make(chan struct{})
	var shutdowns atomic.Int64
	var closePeer sync.Once
	specs := []serverSpec{
		{
			Name:  "failed",
			Serve: func() error { return serveFailure },
			Shutdown: func(context.Context) error {
				shutdowns.Add(1)
				return nil
			},
		},
		{
			Name: "peer",
			Serve: func() error {
				<-releasePeer
				return http.ErrServerClosed
			},
			Shutdown: func(context.Context) error {
				shutdowns.Add(1)
				closePeer.Do(func() { close(releasePeer) })
				return nil
			},
		},
	}

	err := serveGroup(context.Background(), log.New(io.Discard, "", 0), specs...)
	if !errors.Is(err, serveFailure) {
		t.Fatalf("serveGroup() error = %v, want %v", err, serveFailure)
	}
	if got := shutdowns.Load(); got != int64(len(specs)) {
		t.Fatalf("shutdown calls = %d, want one for each of %d servers", got, len(specs))
	}
}

func TestServeGroupTreatsContextCancellationAsGracefulDrain(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	released := make(chan struct{})
	var closeReleased sync.Once
	spec := serverSpec{
		Name: "router",
		Serve: func() error {
			<-released
			return http.ErrServerClosed
		},
		Shutdown: func(context.Context) error {
			closeReleased.Do(func() { close(released) })
			return nil
		},
	}
	cancel()
	if err := serveGroup(ctx, log.New(io.Discard, "", 0), spec); err != nil {
		t.Fatalf("graceful cancellation returned %v", err)
	}
}

func TestServeGroupUsesBoundedShutdownAndWaitsForServeExit(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	serveExited := make(chan struct{})
	shutdownCalled := make(chan struct{})
	spec := serverSpec{
		Name: "router",
		Serve: func() error {
			<-shutdownCalled
			time.Sleep(20 * time.Millisecond)
			close(serveExited)
			return http.ErrServerClosed
		},
		Shutdown: func(shutdownContext context.Context) error {
			if _, ok := shutdownContext.Deadline(); !ok {
				return errors.New("shutdown context has no deadline")
			}
			close(shutdownCalled)
			return nil
		},
	}
	cancel()
	if err := serveGroup(ctx, log.New(io.Discard, "", 0), spec); err != nil {
		t.Fatalf("serveGroup() = %v", err)
	}
	select {
	case <-serveExited:
	default:
		t.Fatal("serveGroup returned before the Serve loop exited")
	}
}

func TestRunModelServesUntilCancellation(t *testing.T) {
	address := availableAddress(t)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- runModel(ctx, address, "model-a", "tiny-llm", 0, log.New(io.Discard, "", 0))
	}()

	waitForStatus(t, done, "http://"+address+"/readyz", http.StatusOK)
	cancel()
	assertGracefulStop(t, done)
}

func TestRunRouterComposesProductionRuntime(t *testing.T) {
	var backendCalls atomic.Int64
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		backendCalls.Add(1)
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	t.Cleanup(backend.Close)

	address := availableAddress(t)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	endpoint := routing.Endpoint{
		ID: "model-a", URL: backend.URL, Models: []string{"tiny-llm"}, Ready: true,
	}
	runtimeConfig := configuration.Config{
		Version: 1, SchedulerName: "default", Weights: map[string]float64{"queue": 1},
		MaxInFlight: 2, MaxQueued: 2,
	}
	go func() {
		done <- runRouter(
			ctx, address, []routing.Endpoint{endpoint}, runtimeConfig,
			log.New(io.Discard, "", 0),
		)
	}()

	waitForStatus(t, done, "http://"+address+"/readyz", http.StatusOK)
	request, err := http.NewRequestWithContext(
		context.Background(), http.MethodPost, "http://"+address+"/v1/chat/completions",
		bytes.NewBufferString(`{"model":"tiny-llm","messages":[{"role":"user","content":"hello"}]}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusOK || backendCalls.Load() != 1 {
		t.Fatalf("router status=%d backend_calls=%d", response.StatusCode, backendCalls.Load())
	}

	cancel()
	assertGracefulStop(t, done)
}

func TestRunAllComposesTwoModelsAndRouterUnderOneCancellationBoundary(t *testing.T) {
	modelA := availableAddress(t)
	modelB := availableAddress(t)
	routerAddress := availableAddress(t)
	directory := t.TempDir()
	endpointPath := filepath.Join(directory, "endpoints.json")
	endpointDocument, err := json.Marshal(config{Endpoints: []routing.Endpoint{
		{ID: "model-a", URL: "http://" + modelA, Models: []string{"tiny-llm"}, Ready: true},
		{ID: "model-b", URL: "http://" + modelB, Models: []string{"tiny-llm"}, Ready: true},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(endpointPath, endpointDocument, 0o600); err != nil {
		t.Fatal(err)
	}
	runtimePath := writeRuntimeConfig(t, directory, 1, 2, 4)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- runAll(ctx, allSettings{
			RouterAddress:      routerAddress,
			EndpointConfigPath: endpointPath,
			RuntimeConfigPath:  runtimePath,
		}, log.New(io.Discard, "", 0))
	}()

	waitForStatus(t, done, "http://"+modelA+"/readyz", http.StatusOK)
	waitForStatus(t, done, "http://"+modelB+"/readyz", http.StatusOK)
	waitForStatus(t, done, "http://"+routerAddress+"/readyz", http.StatusOK)
	request, err := http.NewRequestWithContext(
		context.Background(), http.MethodPost,
		"http://"+routerAddress+"/v1/chat/completions",
		bytes.NewBufferString(`{"model":"tiny-llm","messages":[{"role":"user","content":"hello"}]}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	body, readErr := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if readErr != nil {
		t.Fatal(readErr)
	}
	if response.StatusCode != http.StatusOK ||
		(!bytes.Contains(body, []byte(`"endpoint":"model-a"`)) &&
			!bytes.Contains(body, []byte(`"endpoint":"model-b"`))) {
		t.Fatalf("runAll route status=%d body=%s", response.StatusCode, body)
	}

	cancel()
	assertGracefulStop(t, done)
}

func writeRuntimeConfig(t *testing.T, directory string, version uint64, inFlight, queued int) string {
	t.Helper()
	path := filepath.Join(directory, fmt.Sprintf("runtime-%d-%d.yaml", version, inFlight))
	document := fmt.Sprintf(`apiVersion: learning.llmd.local/v1alpha1
kind: EndpointPickerConfig
metadata:
  name: inference-lab
spec:
  version: %d
  requestHandling: {}
  dataLayer:
    sources:
      - type: file
  flowControl:
    maxInFlight: %d
    maxQueued: %d
  schedulingProfiles:
    - name: balanced
      weights:
        queue: 1
  failurePolicy: {}
`, version, inFlight, queued)
	if err := os.WriteFile(path, []byte(document), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func availableAddress(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	return address
}

func waitForStatus(t *testing.T, done <-chan error, url string, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err := <-done:
			t.Fatalf("server returned before becoming ready: %v", err)
		default:
		}
		response, err := http.Get(url)
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == want {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("%s did not return status %d", url, want)
}

func assertGracefulStop(t *testing.T, done <-chan error) {
	t.Helper()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("graceful shutdown returned %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not drain after cancellation")
	}
}
