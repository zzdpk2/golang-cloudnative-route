package domain

import (
	"fmt"
	"time"
)

var ErrInvalidQuantity = fmt.Errorf("invalid quantity")

// Quantity is a count of items that is guaranteed positive.
//
// The whole type exists so that "how many" cannot be zero or negative anywhere
// downstream. Once you hold a Quantity, you never have to check it again — that
// is the payoff for validating at the boundary.
type Quantity struct{ value int }

// NewQuantity rejects anything that is not a positive count.
//
// Zero is rejected too. Decide for yourself whether that is right before you
// accept it: is "an order line for 0 widgets" a valid thing that later becomes
// invalid, or is it nonsense from the start?
func NewQuantity(v int) (Quantity, error) {
	panic("TODO")
}

func (q Quantity) Value() int { panic("TODO") }

// Add combines two quantities. Two positive numbers cannot sum to something
// invalid, so this cannot fail.
func (q Quantity) Add(other Quantity) Quantity {
	panic("TODO")
}

// Subtract removes other from q, and fails when the result would not be a valid
// Quantity.
//
// This is the asymmetry worth noticing: Add is total, Subtract is partial. The
// type's invariant is what makes the difference, and it shows up in the
// signatures.
func (q Quantity) Subtract(other Quantity) (Quantity, error) {
	panic("TODO")
}

func (q Quantity) Equals(other Quantity) bool { panic("TODO") }

// ---- DateRange ----

var (
	ErrInvalidDateRange = fmt.Errorf("invalid date range")
	ErrDateRangeOverlap = fmt.Errorf("date ranges overlap")
)

// DateRange is a span of time used for promotion windows and reservation
// holds. Like every value object here, it validates once and is then immutable.
type DateRange struct {
	start time.Time
	end   time.Time
}

// NewDateRange rejects a range whose end does not come after its start.
func NewDateRange(start, end time.Time) (DateRange, error) {
	panic("TODO")
}

func (d DateRange) Start() time.Time { panic("TODO") }
func (d DateRange) End() time.Time   { panic("TODO") }

// Duration is how long the range lasts.
func (d DateRange) Duration() time.Duration {
	panic("TODO")
}

// Days is the duration expressed in whole days.
//
// Careful with the units: a Duration is nanoseconds, and converting to days by
// dividing hours is only correct while every day has 24 hours. Across a
// daylight-saving boundary it is not. The tests use UTC so you will not be
// caught here, but know that you are making an assumption.
func (d DateRange) Days() int {
	panic("TODO")
}

// Contains reports whether t falls inside the range, endpoints included.
//
// Inclusive or exclusive endpoints is a decision, not a detail — get it wrong
// and a promotion silently expires an instant early. time.Time has Before and
// After but no "not after", so watch the negations.
func (d DateRange) Contains(t time.Time) bool {
	panic("TODO")
}

// Overlaps reports whether two ranges share any time at all.
//
// The tests pin down the boundary case: two ranges that merely touch — one ends
// exactly when the next begins — do not overlap. Write the condition for
// "definitely disjoint" first; the answer is its negation, and that version is
// much harder to get wrong than reasoning about overlap directly.
func (d DateRange) Overlaps(other DateRange) bool {
	panic("TODO")
}

// Extend returns a new range with a later end. The receiver is unchanged.
func (d DateRange) Extend(duration time.Duration) DateRange {
	panic("TODO")
}
