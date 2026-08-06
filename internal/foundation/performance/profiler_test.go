package performance

import (
	"bytes"
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ---- The plumbing ----

func TestRegisterPprofHandlers_ServesTheIndex(t *testing.T) {
	mux := http.NewServeMux()
	RegisterPprofHandlers(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /debug/pprof/ = %d, want 200", rec.Code)
	}
	if rec.Body.Len() == 0 {
		t.Fatal("the pprof index wrote an empty body")
	}
}

func TestRegisterPprofHandlers_ServesNamedProfiles(t *testing.T) {
	mux := http.NewServeMux()
	RegisterPprofHandlers(mux)

	for _, path := range []string{
		"/debug/pprof/heap",
		"/debug/pprof/goroutine",
		"/debug/pprof/block",
		"/debug/pprof/mutex",
	} {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
			if rec.Code != http.StatusOK {
				t.Errorf("GET %s = %d, want 200", path, rec.Code)
			}
		})
	}
}

func TestNewProfilingServer(t *testing.T) {
	srv := NewProfilingServer(":0")
	if srv == nil {
		t.Fatal("NewProfilingServer returned nil")
	}
	if srv.Addr != ":0" {
		t.Errorf("Addr = %q, want %q", srv.Addr, ":0")
	}
	if srv.Handler == nil {
		t.Fatal("Handler is nil: the server must carry the pprof routes, " +
			"not fall back to http.DefaultServeMux")
	}
}

func TestCaptureCPUProfile_ProducesAProfile(t *testing.T) {
	var buf bytes.Buffer

	err := CaptureCPUProfile(&buf, func() {
		lines := BuildOrderLines(2000)
		SortLinesByName(lines)
		_ = TotalCents(lines)
	})
	if err != nil {
		t.Fatalf("CaptureCPUProfile returned %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("the CPU profile is empty")
	}
}

// Only one CPU profile can be active per process. A second concurrent start
// must fail rather than quietly corrupting the first.
func TestCaptureCPUProfile_RejectsANestedCapture(t *testing.T) {
	var outer, inner bytes.Buffer
	var innerErr error

	err := CaptureCPUProfile(&outer, func() {
		innerErr = CaptureCPUProfile(&inner, func() {})
	})
	if err != nil {
		t.Fatalf("outer capture returned %v", err)
	}
	if innerErr == nil {
		t.Error("a nested CaptureCPUProfile returned nil; only one CPU profile " +
			"may be active at a time, so the second must report that")
	}
}

// A panic inside fn must not leave the process profiling forever.
func TestCaptureCPUProfile_StopsOnPanic(t *testing.T) {
	var buf bytes.Buffer

	func() {
		defer func() { _ = recover() }()
		_ = CaptureCPUProfile(&buf, func() { panic("boom") })
	}()

	// If the profile were still running, this one could not start.
	var second bytes.Buffer
	if err := CaptureCPUProfile(&second, func() {}); err != nil {
		t.Errorf("a later capture returned %v: the panicking call left the "+
			"profile running. Stop it from a defer.", err)
	}
}

func TestWriteHeapProfile(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteHeapProfile(&buf); err != nil {
		t.Fatalf("WriteHeapProfile returned %v", err)
	}
	if buf.Len() == 0 {
		t.Error("the heap profile is empty")
	}
}

func TestWriteGoroutineProfile_Debug1IsReadable(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteGoroutineProfile(&buf, 1); err != nil {
		t.Fatalf("WriteGoroutineProfile returned %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("goroutine")) {
		t.Error("debug=1 should produce human-readable stacks mentioning 'goroutine'")
	}
}

func TestWithProfileFile_WritesToDisk(t *testing.T) {
	path := filepath.Join(t.TempDir(), "heap.out")

	if err := WithProfileFile(path, WriteHeapProfile); err != nil {
		t.Fatalf("WithProfileFile returned %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if info.Size() == 0 {
		t.Error("the profile file is empty; a deferred Close whose error is " +
			"dropped can lose the last buffered write")
	}
}

// ---- The custom profile ----

func TestCustomProfile_TracksAndReleases(t *testing.T) {
	type resource struct{ id int }

	a, b := &resource{1}, &resource{2}
	before := CustomProfileCount()

	TrackCustomResource(a, 1)
	TrackCustomResource(b, 1)
	if got, want := CustomProfileCount(), before+2; got != want {
		t.Fatalf("count after two Tracks = %d, want %d", got, want)
	}

	UntrackCustomResource(a)
	if got, want := CustomProfileCount(), before+1; got != want {
		t.Errorf("count after one Untrack = %d, want %d", got, want)
	}

	UntrackCustomResource(b)
	if got := CustomProfileCount(); got != before {
		t.Errorf("count back at the start = %d, want %d: every Track needs "+
			"exactly one Untrack, and the imbalance is the leak signal", got, before)
	}
}

// ---- The order pipeline ----

func TestBuildOrderLines(t *testing.T) {
	lines := BuildOrderLines(3)
	if len(lines) != 3 {
		t.Fatalf("len = %d, want 3", len(lines))
	}
	// Descending, so the sort has real work to do.
	if lines[0].Name != "sku-2" || lines[2].Name != "sku-0" {
		t.Errorf("names = %q..%q, want sku-2..sku-0 (descending)",
			lines[0].Name, lines[2].Name)
	}
	if got := BuildOrderLines(0); len(got) != 0 {
		t.Errorf("BuildOrderLines(0) returned %d lines, want 0", len(got))
	}
}

func TestSortLinesByName(t *testing.T) {
	lines := BuildOrderLines(5)
	SortLinesByName(lines)

	for i := 1; i < len(lines); i++ {
		if lines[i-1].Name > lines[i].Name {
			t.Fatalf("not sorted at index %d: %q then %q",
				i, lines[i-1].Name, lines[i].Name)
		}
	}
	if lines[0].Name != "sku-0" {
		t.Errorf("first = %q, want %q", lines[0].Name, "sku-0")
	}
}

func TestTotalCents(t *testing.T) {
	if got := TotalCents(BuildOrderLines(100)); got != 0 {
		t.Errorf("TotalCents = %d, want 0 (every fixture price is domain.Zero)", got)
	}
	if got := TotalCents(nil); got != 0 {
		t.Errorf("TotalCents(nil) = %d, want 0", got)
	}
}

func TestSleepWork(t *testing.T) {
	start := time.Now()
	SleepWork(20 * time.Millisecond)
	if elapsed := time.Since(start); elapsed < 15*time.Millisecond {
		t.Errorf("SleepWork returned after %v, want at least ~20ms", elapsed)
	}
}

// ---- Benchmarks: something real to profile ----
//
//	go test ./internal/foundation/performance -bench=BenchmarkOrderPipeline -cpuprofile=cpu.out
//	go tool pprof -http=:8080 cpu.out
//
// Write down which of build / sort / total you think dominates, *before*
// looking. Then run the three separate benchmarks below and check.
//
// Then run the pipeline with -memprofile as well and ask a second question:
// is the time going into work, or into the garbage collector cleaning up after
// the work?

var (
	sinkLines []domain.OrderLine
	sinkCents int64
)

func BenchmarkOrderPipeline(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		lines := BuildOrderLines(1000)
		SortLinesByName(lines)
		sinkCents = TotalCents(lines)
	}
}

func BenchmarkPipelineBuild(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		sinkLines = BuildOrderLines(1000)
	}
}

func BenchmarkPipelineSort(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		b.StopTimer()
		lines := BuildOrderLines(1000)
		b.StartTimer()
		SortLinesByName(lines)
	}
}

func BenchmarkPipelineTotal(b *testing.B) {
	lines := BuildOrderLines(1000)
	SortLinesByName(lines)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		sinkCents = TotalCents(lines)
	}
}
