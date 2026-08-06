package testing

import (
	"fmt"
	"strings"
)

// T is the slice of *testing.T that a suite is allowed to use.
//
// Your suites take a T rather than a *testing.T, and that one change is what
// makes this package possible: a suite that only knows about an interface can
// be run against a *recorder* instead of the real thing, so the meta-test can
// observe "did this suite fail?" without failing itself.
//
// *testing.T satisfies T already, so calling a suite from an ordinary test is
// just SplitSuite(t, SplitEvenly).
//
// Taking a narrow interface instead of a concrete type, so the thing can be
// driven by something other than production, is exactly the L5 lesson — here
// applied to the test framework itself.
type T interface {
	Helper()
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	Logf(format string, args ...any)
}

// fatal is what a recorder panics with to emulate Fatalf.
//
// The real testing.T ends the test goroutine on Fatalf, via runtime.Goexit. A
// recorder cannot do that without ending the meta-test too, so it panics with
// this sentinel and run() recovers it. That is the only reason this type exists.
type fatal struct{}

// recorder implements T and remembers what a suite complained about.
type recorder struct {
	failures []string
	logs     []string
}

func (r *recorder) Helper() {}

func (r *recorder) Errorf(format string, args ...any) {
	r.failures = append(r.failures, fmt.Sprintf(format, args...))
}

func (r *recorder) Fatalf(format string, args ...any) {
	r.failures = append(r.failures, fmt.Sprintf(format, args...))
	panic(fatal{})
}

func (r *recorder) Logf(format string, args ...any) {
	r.logs = append(r.logs, fmt.Sprintf(format, args...))
}

func (r *recorder) failed() bool { return len(r.failures) > 0 }

func (r *recorder) summary() string {
	if len(r.failures) == 0 {
		return "(no failures)"
	}
	const max = 3
	shown := r.failures
	suffix := ""
	if len(shown) > max {
		shown, suffix = shown[:max], fmt.Sprintf(" (+%d more)", len(r.failures)-max)
	}
	return strings.Join(shown, "; ") + suffix
}

// runSuite runs a suite against one implementation and reports what happened.
//
// It recovers both the Fatalf sentinel and any genuine panic — a suite that
// panics on a mutant has still detected it, which is worth allowing even though
// a panicking test is a poor way to report a failure.
func runSuite(suite func(T, SplitFunc), impl SplitFunc) (rec *recorder, panicked any) {
	rec = &recorder{}
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(fatal); !ok {
				panicked = r
			}
		}
	}()
	suite(rec, impl)
	return rec, nil
}
