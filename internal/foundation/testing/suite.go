package testing

// This file is your work. Everything else in the package is scaffolding.

// SplitSuite is a test suite for any SplitFunc.
//
// Write it so that it passes against SplitEvenly and fails against every mutant
// in Mutants. Run the meta-test to find out how you did:
//
//	go test ./internal/foundation/testing -run TestSplitSuite -v
//
// # Write it as a table
//
// A table-driven test separates the *cases* from the *checking*, so adding a
// case is one line rather than one more copied block. It is the default shape
// for a pure function in Go, and it is what you should reach for here:
//
//	tests := []struct {
//	    name  string
//	    total int64
//	    parts int
//	    want  []int64
//	    wantErr error
//	}{
//	    {name: "divides evenly", total: 1000, parts: 4, want: []int64{250, 250, 250, 250}},
//	    ...
//	}
//
// Then loop over it. Name every case — a failure reporting `case 3` sends you
// counting rows; one reporting `remainder goes to the earliest parts` does not.
//
// # Two kinds of case, and you need both
//
// **Examples** pin down specific inputs and outputs. They are precise and they
// only cover what you thought of.
//
// **Properties** hold for every input, and they catch what you did not think
// of. For this function the properties are:
//
//	the result has exactly `parts` elements
//	the elements sum to exactly the total
//	no two elements differ by more than 1
//	the earlier parts are the larger ones
//
// A suite built only from examples will let at least one mutant through. So
// will a suite built only from the sum property. Work out which, and why.
//
// # Reporting failures
//
// Say what you got, what you wanted, and for which input:
//
//	t.Errorf("Split(%d, %d) = %v, want %v", tt.total, tt.parts, got, tt.want)
//
// Not `t.Errorf("wrong")`. You are writing for the person who sees this fail at
// 3am with no context, and that person is usually you.
//
// # Errorf or Fatalf
//
// Errorf records the failure and keeps going; Fatalf stops the suite. Use
// Fatalf only when continuing is meaningless — a nil slice you are about to
// index. Reaching for Fatalf everywhere means you fix one failure per run
// instead of seeing all five at once.
func SplitSuite(t T, split SplitFunc) {
	panic("TODO")
}

// ---- Property helpers ----
//
// Write these first. They are what turns "I tested three inputs" into "I
// checked the rules on every input", and FuzzSplit reuses them.

// CheckSumsToTotal reports an error unless the parts sum to exactly total.
//
//	CheckSumsToTotal(t, []int64{334, 333, 333}, 1000)  → passes
//	CheckSumsToTotal(t, []int64{333, 333, 333}, 1000)  → fails, one cent lost
//
// Call t.Helper() first, so a failure is reported at the caller's line instead
// of inside this function. Without it, every failure in your table points here.
func CheckSumsToTotal(t T, parts []int64, total int64) {
	panic("TODO")
}

// CheckNearlyEqual reports an error unless no two parts differ by more than 1.
//
//	CheckNearlyEqual(t, []int64{334, 333, 333})  → passes
//	CheckNearlyEqual(t, []int64{4, 2, 2, 2})     → fails, spread of 2
//
// This is the property that catches the mutant which sums correctly but
// distributes badly.
func CheckNearlyEqual(t T, parts []int64) {
	panic("TODO")
}

// CheckRemainderGoesFirst reports an error unless the parts are non-increasing.
//
//	CheckRemainderGoesFirst(t, []int64{334, 333, 333})  → passes
//	CheckRemainderGoesFirst(t, []int64{333, 333, 334})  → fails
//
// The weakest-looking property here, and the one that catches a mutant no
// amount of sum-checking will. In a real system it decides which sub-order is
// charged the extra cent — a question finance departments genuinely care about.
func CheckRemainderGoesFirst(t T, parts []int64) {
	panic("TODO")
}
