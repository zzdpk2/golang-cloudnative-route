// Package testcraft teaches writing tests, rather than passing them.
//
// Everywhere else in this repository the tests are handed to you and the
// implementation is yours. Here it is the other way round: **the implementation
// is given, and the tests are your work.**
//
// # How you get graded
//
// Writing tests is easy; writing tests that would *catch a bug* is the skill,
// and it is not the same skill. A suite of twenty assertions that all pass no
// matter what the code does is worse than no suite at all, because it buys
// false confidence.
//
// So this package grades your tests by mutation. Below is one correct
// implementation of a money-splitting function, and a set of **mutants** — the
// same function, each with exactly one deliberate bug. Your suite must:
//
//	pass  against SplitEvenly
//	fail  against every single mutant
//
// A mutant your suite lets through is a bug your suite would let through in
// production. The meta-test in testcraft_test.go names the ones you missed and
// tells you what each one does wrong.
//
// This is real mutation testing, done by hand. Tools like go-mutesting automate
// it; doing a few by hand first is what makes the output meaningful.
//
// # The subject
//
// SplitEvenly divides an amount of money into equal parts. It is the same
// problem as vo.Money.Allocate from L1 and the order-splitting in L15.3, and it
// is a good subject for exactly the reason it is a good exercise: the obvious
// implementation is wrong in a way that only a carefully chosen test reveals.
//
// # Suggested order
//
//  1. Read SplitEvenly and its contract. Do not read the mutants yet.
//  2. Write SplitSuite in suite.go from the contract alone.
//  3. Run the meta-test. See which mutants you missed.
//  4. Only now read what those mutants do — and note that each one is a bug you
//     did not think to test for.
//  5. Add cases until every mutant is caught.
//  6. Then write FuzzSplit, and see whether it finds anything your table missed.
package testing

import "errors"

var (
	// ErrNonPositiveParts means the caller asked for zero or fewer parts.
	ErrNonPositiveParts = errors.New("parts must be positive")
	// ErrNegativeTotal means the caller passed a negative amount.
	ErrNegativeTotal = errors.New("total must not be negative")
)

// SplitFunc divides totalCents into parts, in minor units.
//
// The contract, in full — this is what your suite has to pin down:
//
//   - The result has exactly `parts` elements.
//   - The elements sum to exactly totalCents. Nothing is created or lost.
//   - The parts are as equal as possible: no two differ by more than 1.
//   - Any remainder goes to the *earliest* parts, one unit each.
//   - parts <= 0 returns ErrNonPositiveParts.
//   - totalCents < 0 returns ErrNegativeTotal.
//   - A total of 0 is fine and yields parts of 0.
//
// Worked examples:
//
//	Split(1000, 4) → [250 250 250 250]
//	Split(1000, 3) → [334 333 333]        remainder 1 goes to the first part
//	Split(1001, 3) → [334 334 333]        remainder 2 goes to the first two
//	Split(2, 5)    → [1 1 0 0 0]          more parts than cents is legal
//	Split(0, 3)    → [0 0 0]
//	Split(100, 0)  → error
//	Split(-1, 3)   → error
type SplitFunc func(totalCents int64, parts int) ([]int64, error)

// SplitEvenly is the correct implementation. Your suite must pass against it.
func SplitEvenly(totalCents int64, parts int) ([]int64, error) {
	if parts <= 0 {
		return nil, ErrNonPositiveParts
	}
	if totalCents < 0 {
		return nil, ErrNegativeTotal
	}

	base := totalCents / int64(parts)
	remainder := totalCents % int64(parts)

	out := make([]int64, parts)
	for i := range out {
		out[i] = base
		if int64(i) < remainder {
			out[i]++
		}
	}
	return out, nil
}

// Mutant is a deliberately broken implementation.
//
// Bug describes what it gets wrong. Read these *after* your first run, not
// before — the point is to find out which failures you did not anticipate.
type Mutant struct {
	Name string
	Bug  string
	Fn   SplitFunc
}

// Mutants are the implementations your suite has to reject.
//
// They get progressively less obvious. The last two are the ones that catch
// most people, and both are bugs that really do ship.
var Mutants = []Mutant{
	{
		Name: "drops the remainder",
		Bug:  "divides and ignores the leftover, so the parts sum to less than the total",
		Fn: func(totalCents int64, parts int) ([]int64, error) {
			if parts <= 0 {
				return nil, ErrNonPositiveParts
			}
			if totalCents < 0 {
				return nil, ErrNegativeTotal
			}
			base := totalCents / int64(parts)
			out := make([]int64, parts)
			for i := range out {
				out[i] = base
			}
			return out, nil
		},
	},
	{
		Name: "accepts zero parts",
		Bug:  "no guard on parts <= 0, so it returns an empty slice instead of an error",
		Fn: func(totalCents int64, parts int) ([]int64, error) {
			if totalCents < 0 {
				return nil, ErrNegativeTotal
			}
			if parts == 0 {
				return []int64{}, nil
			}
			return SplitEvenly(totalCents, parts)
		},
	},
	{
		Name: "accepts a negative total",
		Bug:  "no guard on totalCents < 0, so it happily splits a debt",
		Fn: func(totalCents int64, parts int) ([]int64, error) {
			if parts <= 0 {
				return nil, ErrNonPositiveParts
			}
			base := totalCents / int64(parts)
			rem := totalCents % int64(parts)
			out := make([]int64, parts)
			for i := range out {
				out[i] = base
				if int64(i) < rem {
					out[i]++
				}
			}
			return out, nil
		},
	},
	{
		Name: "returns one part too few",
		Bug:  "off-by-one on the slice length",
		Fn: func(totalCents int64, parts int) ([]int64, error) {
			if parts <= 0 {
				return nil, ErrNonPositiveParts
			}
			if totalCents < 0 {
				return nil, ErrNegativeTotal
			}
			full, _ := SplitEvenly(totalCents, parts)
			if len(full) > 1 {
				return full[:len(full)-1], nil
			}
			return full, nil
		},
	},
	{
		Name: "remainder goes to the last parts",
		Bug: "the parts still sum to the total and are still within 1 of each other — " +
			"only the *order* is wrong. A suite that just checks the sum misses this, " +
			"and it is a real bug: it changes which sub-order is charged the extra cent",
		Fn: func(totalCents int64, parts int) ([]int64, error) {
			if parts <= 0 {
				return nil, ErrNonPositiveParts
			}
			if totalCents < 0 {
				return nil, ErrNegativeTotal
			}
			base := totalCents / int64(parts)
			rem := totalCents % int64(parts)
			out := make([]int64, parts)
			for i := range out {
				out[i] = base
				if int64(parts-1-i) < rem {
					out[i]++
				}
			}
			return out, nil
		},
	},
	{
		Name: "dumps the whole remainder on the first part",
		Bug: "sums correctly, but the parts can differ by far more than 1. " +
			"Split(1000,3) gives [334 333 333] correctly, yet Split(10,4) gives " +
			"[4 2 2 2] instead of [3 3 2 2] — so a suite testing only one input " +
			"can pass while the function is wrong",
		Fn: func(totalCents int64, parts int) ([]int64, error) {
			if parts <= 0 {
				return nil, ErrNonPositiveParts
			}
			if totalCents < 0 {
				return nil, ErrNegativeTotal
			}
			base := totalCents / int64(parts)
			rem := totalCents % int64(parts)
			out := make([]int64, parts)
			for i := range out {
				out[i] = base
			}
			out[0] += rem
			return out, nil
		},
	},
}
