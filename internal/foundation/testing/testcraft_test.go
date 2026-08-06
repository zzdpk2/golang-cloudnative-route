package testing

import (
	"errors"
	"testing"
)

// ---- The meta-test: how good is your suite? ----

// Your suite must pass against the correct implementation. If this fails, your
// suite is asserting something that is not in the contract.
func TestSplitSuite_PassesTheCorrectImplementation(t *testing.T) {
	rec, panicked := runSuite(SplitSuite, SplitEvenly)

	if panicked != nil {
		t.Fatalf("your suite panicked on the correct implementation: %v", panicked)
	}
	if rec.failed() {
		t.Errorf("your suite rejected SplitEvenly, which is correct.\n"+
			"Either an assertion is wrong, or you are asserting something the "+
			"contract does not promise.\nFailures: %s", rec.summary())
	}
}

// Your suite must reject every mutant. Each one it lets through is a bug it
// would let through in production.
func TestSplitSuite_CatchesEveryMutant(t *testing.T) {
	// A suite that panics or rejects the correct implementation would "catch"
	// every mutant for the wrong reason, and this test would go green while
	// telling you nothing. Establish the baseline first.
	if rec, panicked := runSuite(SplitSuite, SplitEvenly); panicked != nil || rec.failed() {
		t.Skip("skipping: your suite does not pass SplitEvenly yet — " +
			"see TestSplitSuite_PassesTheCorrectImplementation")
	}

	var missed []Mutant

	for _, m := range Mutants {
		rec, panicked := runSuite(SplitSuite, m.Fn)
		caught := rec.failed() || panicked != nil
		if !caught {
			missed = append(missed, m)
			continue
		}
		t.Logf("caught %-38s %s", m.Name, firstLine(rec.summary()))
	}

	if len(missed) == 0 {
		return
	}

	t.Errorf("your suite passed %d of %d mutants — these bugs would have shipped:",
		len(missed), len(Mutants))
	for _, m := range missed {
		t.Errorf("\n  %s\n      %s", m.Name, m.Bug)
	}
	t.Error("\nAdd cases that distinguish these from the correct implementation. " +
		"If a mutant looks impossible to catch with examples alone, that is the " +
		"hint: it needs a property.")
}

// The helpers are used by both the suite and the fuzz target, so they are
// checked directly too.
func TestCheckSumsToTotal(t *testing.T) {
	if rec, _ := runSuite(func(tt T, _ SplitFunc) {
		CheckSumsToTotal(tt, []int64{334, 333, 333}, 1000)
	}, SplitEvenly); rec.failed() {
		t.Errorf("a correct split was rejected: %s", rec.summary())
	}

	if rec, _ := runSuite(func(tt T, _ SplitFunc) {
		CheckSumsToTotal(tt, []int64{333, 333, 333}, 1000)
	}, SplitEvenly); !rec.failed() {
		t.Error("a split losing one cent was accepted")
	}
}

func TestCheckNearlyEqual(t *testing.T) {
	if rec, _ := runSuite(func(tt T, _ SplitFunc) {
		CheckNearlyEqual(tt, []int64{334, 333, 333})
	}, SplitEvenly); rec.failed() {
		t.Errorf("a spread of 1 was rejected: %s", rec.summary())
	}

	if rec, _ := runSuite(func(tt T, _ SplitFunc) {
		CheckNearlyEqual(tt, []int64{4, 2, 2, 2})
	}, SplitEvenly); !rec.failed() {
		t.Error("a spread of 2 was accepted")
	}
}

func TestCheckRemainderGoesFirst(t *testing.T) {
	if rec, _ := runSuite(func(tt T, _ SplitFunc) {
		CheckRemainderGoesFirst(tt, []int64{334, 333, 333})
	}, SplitEvenly); rec.failed() {
		t.Errorf("a correctly ordered split was rejected: %s", rec.summary())
	}

	if rec, _ := runSuite(func(tt T, _ SplitFunc) {
		CheckRemainderGoesFirst(tt, []int64{333, 333, 334})
	}, SplitEvenly); !rec.failed() {
		t.Error("a split with the remainder at the end was accepted")
	}
}

// ---- The reference behaviour ----
//
// These describe SplitEvenly directly. They are deliberately *not* a model
// answer for your suite: they show the shape of a table test, and they are
// nowhere near strong enough to catch every mutant. Use them as a starting
// point, not a destination.

func TestSplitEvenly_Examples(t *testing.T) {
	tests := []struct {
		name  string
		total int64
		parts int
		want  []int64
	}{
		{name: "divides evenly", total: 1000, parts: 4, want: []int64{250, 250, 250, 250}},
		{name: "one cent left over", total: 1000, parts: 3, want: []int64{334, 333, 333}},
		{name: "two cents left over", total: 1001, parts: 3, want: []int64{334, 334, 333}},
		{name: "more parts than cents", total: 2, parts: 5, want: []int64{1, 1, 0, 0, 0}},
		{name: "nothing to split", total: 0, parts: 3, want: []int64{0, 0, 0}},
		{name: "single part takes everything", total: 999, parts: 1, want: []int64{999}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SplitEvenly(tt.total, tt.parts)
			if err != nil {
				t.Fatalf("SplitEvenly(%d, %d) returned %v", tt.total, tt.parts, err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("SplitEvenly(%d, %d) = %v, want %v", tt.total, tt.parts, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("SplitEvenly(%d, %d) = %v, want %v",
						tt.total, tt.parts, got, tt.want)
					break
				}
			}
		})
	}
}

func TestSplitEvenly_Errors(t *testing.T) {
	tests := []struct {
		name    string
		total   int64
		parts   int
		wantErr error
	}{
		{name: "zero parts", total: 100, parts: 0, wantErr: ErrNonPositiveParts},
		{name: "negative parts", total: 100, parts: -2, wantErr: ErrNonPositiveParts},
		{name: "negative total", total: -1, parts: 3, wantErr: ErrNegativeTotal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SplitEvenly(tt.total, tt.parts)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("SplitEvenly(%d, %d) error = %v, want %v",
					tt.total, tt.parts, err, tt.wantErr)
			}
		})
	}
}

// ---- Fuzzing ----
//
//	go test ./internal/foundation/testing -run=Fuzz -fuzz=FuzzSplit -fuzztime=30s
//
// A fuzzer generates inputs you would never have written down and keeps any
// that reach new code paths. It cannot check "the answer is 334" — it has no
// idea what the answer should be. It can only check **properties**, which is
// why you wrote the Check* helpers first.
//
// When it finds a failure it writes the input to testdata/fuzz/ and that input
// becomes a permanent regression test. That corpus is the real output of
// fuzzing; the run itself is temporary.
func FuzzSplit(f *testing.F) {
	f.Add(int64(1000), 3)
	f.Add(int64(0), 1)
	f.Add(int64(2), 5)
	f.Add(int64(1), 1)

	f.Fuzz(func(t *testing.T, total int64, parts int) {
		// TODO: call SplitEvenly and assert the properties on any input that
		// should succeed.
		//
		// Start by deciding which inputs are *out of contract* and should be
		// skipped with t.Skip — negative totals and non-positive parts return
		// errors by design, and a fuzzer will hand you both immediately.
		//
		// Then guard the input range. A fuzzer will try parts = 2_000_000_000,
		// and allocating that slice will kill the test process rather than fail
		// it. Cap it, and note that knowing where to cap requires understanding
		// the function — fuzzing is not a substitute for thinking.
		//
		// Then apply CheckSumsToTotal, CheckNearlyEqual, and
		// CheckRemainderGoesFirst. They take a T, and *testing.T satisfies it.
		t.Fatal("TODO(exercise): implement the fuzz target")
	})
}

func firstLine(s string) string {
	for i := range len(s) {
		if s[i] == '\n' {
			return s[:i]
		}
	}
	return s
}
