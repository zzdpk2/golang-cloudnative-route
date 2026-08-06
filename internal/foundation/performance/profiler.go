// Package profiling is the measurement laboratory.
//
// Two halves, and they answer different questions:
//
//	the pprof plumbing   how do I get a profile out of a running service?
//	the order pipeline   what does the order path actually spend its time on?
//
// The second half is why this package uses real OrderLines instead of a
// BusyWork loop. Profiling a synthetic workload teaches you the tool; profiling
// your own code teaches you your code.
//
// # The rule that makes profiling worth doing
//
// **Measure, do not guess.** Everyone believes this and almost nobody does it —
// the pull toward "obviously the slow part is X" is very strong, and it is
// wrong often enough to waste whole afternoons. The pipeline benchmarks below
// exist so you can practise being wrong cheaply: predict the hot function
// first, write the prediction down, then look.
//
// # Getting a profile
//
//	go test ./internal/foundation/performance -bench=BenchmarkOrderPipeline -cpuprofile=cpu.out
//	go tool pprof -http=:8080 cpu.out
//
// In the web UI, start with Top, then Flame Graph, then Source on whatever came
// top. Learn to read `flat` versus `cum`: flat is time *in* the function, cum
// is time in it plus everything it called. A function with high cum and near-zero
// flat is just a caller — chase its children, not it.
//
// # Which profile answers which question
//
//	cpu         where the time goes while the CPU is busy
//	heap        what is allocated and still live — memory *growth*
//	allocs      everything ever allocated — allocation *rate*, i.e. GC pressure
//	goroutine   every goroutine and its stack; this is how you find leaks
//	block       time spent blocked on channels and mutexes
//	mutex       lock contention specifically
//
// The trap is reaching for cpu when the symptom is latency. If the service is
// slow but the CPU is idle, cpu shows you nothing and block or goroutine shows
// you everything.
package performance

import (
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"io"
	"net/http"

	rpprof "runtime/pprof"
	"time"
)

// ---- The plumbing ----

// RegisterPprofHandlers mounts the pprof endpoints on a mux.
//
// Register these paths:
//
//	/debug/pprof/              the index page
//	/debug/pprof/cmdline
//	/debug/pprof/profile       a CPU profile; blocks for `seconds`
//	/debug/pprof/symbol
//	/debug/pprof/trace
//	/debug/pprof/goroutine     these five come from pprof.Handler(name)
//	/debug/pprof/heap
//	/debug/pprof/threadcreate
//	/debug/pprof/block
//	/debug/pprof/mutex
//
// The test asserts that a GET of /debug/pprof/ returns 200 with a non-empty
// body, and that /debug/pprof/heap works too.
//
// **Never mount these on your public listener.** Importing net/http/pprof for
// its side effect registers them on http.DefaultServeMux, which is how they end
// up exposed by accident — the profile endpoint alone lets anyone hold your CPU
// for thirty seconds, repeatedly. This function takes an explicit mux precisely
// so you have to choose where they go. In L17 that choice becomes a separate
// admin port.
func RegisterPprofHandlers(mux *http.ServeMux) {
	panic("TODO")
}

// NewProfilingServer returns a server serving only the pprof endpoints.
//
//	NewProfilingServer(":6060") → a *http.Server with Addr ":6060" and a
//	                              handler carrying the pprof routes
//
// Separate from the main server on purpose: different port, different exposure,
// different firewall rule.
func NewProfilingServer(addr string) *http.Server {
	panic("TODO")
}

// CaptureCPUProfile records a CPU profile while fn runs.
//
//	var buf bytes.Buffer
//	CaptureCPUProfile(&buf, func() { heavyWork() })
//	// buf now holds a parseable profile
//
// Stop the profile even if fn panics, and return the error from starting it —
// only one CPU profile can be active at a time in a process, so a second
// concurrent call must fail rather than silently corrupt the first.
func CaptureCPUProfile(w io.Writer, fn func()) error {
	panic("TODO")
}

// WriteHeapProfile writes a snapshot of the heap.
//
// A heap profile reports memory that is *still live*, and it is sampled — one
// object in every 512KB by default. So it is excellent for "what is holding
// memory" and misleading for "what allocated once at startup".
//
// Consider whether to runtime.GC() first. Without it the profile includes
// garbage that has not been collected yet, which makes a leak hunt harder;
// with it, the numbers are cleaner but you have perturbed the program you are
// measuring.
func WriteHeapProfile(w io.Writer) error {
	panic("TODO")
}

// WriteGoroutineProfile writes every goroutine and its stack.
//
//	WriteGoroutineProfile(w, 0) → binary protobuf, for go tool pprof
//	WriteGoroutineProfile(w, 1) → human-readable stacks
//	WriteGoroutineProfile(w, 2) → the format a panic prints
//
// The first tool to reach for on a suspected goroutine leak: take one now, take
// another in five minutes, and see which stack grew. Every leak you were warned
// about in platform/channel shows up here as a pile of goroutines parked on the
// same line.
//
// Return an error if the named profile does not exist.
func WriteGoroutineProfile(w io.Writer, debug int) error {
	panic("TODO")
}

// WithProfileFile creates a file and hands it to capture.
//
//	WithProfileFile("heap.out", WriteHeapProfile)
//
// Close the file even when capture fails — and note that a deferred Close whose
// error is dropped can silently lose the last buffered write. Decide whether
// this helper should report that.
func WithProfileFile(path string, capture func(io.Writer) error) error {
	panic("TODO")
}

// ---- A custom profile ----

// customProfile tracks a resource this program cares about.
//
// rpprof.NewProfile is how you get pprof to count something the runtime knows
// nothing about: open database connections, in-flight orders, leased inventory
// reservations. Each Add records the caller's stack, so the profile tells you
// not just how many are outstanding but *where they were created* — which is
// the only question that matters when they are leaking.
var customProfile = rpprof.NewProfile("go-ddd-tdd/custom-resource")

// TrackCustomResource records that a resource is now outstanding.
//
// skip is how many stack frames to hide, so the profile blames the caller
// rather than this helper — the same idea as t.Helper() in tests. Passing 0
// records this function; the test expects the caller.
//
// The value is used as a map key, so it must be comparable and unique. A
// pointer is the usual choice.
func TrackCustomResource(value any, skip int) {
	panic("TODO")
}

// UntrackCustomResource records that a resource has been released.
//
// Every Track needs exactly one Untrack, and a defer next to the Track is how
// you get that. A missing Untrack is precisely the leak this profile is for, so
// the imbalance is the signal, not a bug in the tracking.
func UntrackCustomResource(value any) {
	panic("TODO")
}

// CustomProfileCount reports how many resources are outstanding.
//
//	Track(a); Track(b); Count() → 2
//	Untrack(a);          Count() → 1
func CustomProfileCount() int {
	panic("TODO")
}

// ---- The order pipeline: something real to profile ----
//
// These three functions do the kind of work the order path does — build lines,
// order them, total them. They are the subject of BenchmarkOrderPipeline.
//
// Before you profile it, write down which of the three you expect to dominate.
// Then look. The gap between the prediction and the profile is the lesson; the
// profile alone is just a picture.

// BuildOrderLines produces n lines with descending SKUs, so that SortLinesByName
// has real work to do rather than confirming an already-sorted slice.
//
//	BuildOrderLines(3) → lines named "sku-2", "sku-1", "sku-0"
//
// Price each with domain.Zero(domain.AUD) — the pipeline measures time and allocation,
// neither of which depends on the amount.
func BuildOrderLines(n int) []domain.OrderLine {
	panic("TODO")
}

// SortLinesByName orders lines by Name, ascending, in place.
//
//	["sku-2" "sku-1" "sku-0"] → ["sku-0" "sku-1" "sku-2"]
//
// Sorting is usually where the time is in a pipeline like this. Usually. Check.
func SortLinesByName(lines []domain.OrderLine) {
	panic("TODO")
}

// TotalCents sums the line prices.
//
//	BuildOrderLines(100) → 0   (every fixture price is zero)
//
// The cheapest of the three, and worth profiling anyway: a function that costs
// nothing per call can still dominate a profile if it is called enough times.
func TotalCents(lines []domain.OrderLine) int64 {
	panic("TODO")
}

// SleepWork blocks for d.
//
// Here to make a point the CPU profile cannot: sleeping burns no CPU, so a
// service that spends all its time here shows an *empty* CPU profile while
// being unusably slow. That is when you reach for the block profile or a trace
// instead. Run the pipeline benchmark with and without it and compare the two
// profiles.
func SleepWork(d time.Duration) {
	panic("TODO")
}
