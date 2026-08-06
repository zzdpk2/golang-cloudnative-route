// Package bce is the bounds-check-elimination laboratory.
//
// Correct results are only the first gate. Inspect compiler diagnostics too:
//
//	go test -gcflags="-d=ssa/check_bce/debug=1" ./internal/foundation/performance
package performance

// SumIndex totals values by index. Structure the loop so the compiler can prove
// every index is within bounds.
func SumIndex(values []int) int {
	panic("TODO")
}

// SumPairs totals adjacent pairs:
//
//	[a,b,c,d] -> (a+b) + (b+c) + (c+d)
//
// Establish the len(values) >= 2 precondition once, then write a loop whose
// relationship to len is visible to the compiler.
func SumPairs(values []int) int {
	panic("TODO")
}

// SumFour totals exactly the first four elements and reports false when fewer
// than four exist. One bounds hint can eliminate checks for all four accesses.
func SumFour(values []int) (int, bool) {
	panic("TODO")
}
