package pricing

import (
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"time"
)

// ============================================================
// Business Policies
//
// Named rules and their composition.
// ============================================================

// DiscountPolicy takes an order and the price so far, and returns the new
// price plus a human-readable description of what it did.
//
// The shape is the design. Because a policy takes the *current* price rather
// than the original, policies compose by chaining — each one sees what the
// previous ones already did. And because it returns a description, the customer
// can be told why they paid what they paid, which is a hard requirement in
// retail and an easy thing to forget until support asks for it.
//
// A policy that does not apply returns the price unchanged and an empty
// description. That convention matters: it is what lets a chain contain
// policies that are mostly no-ops without special-casing them.
//
// Compare this with specification.Specification. A spec answers "does this
// qualify?"; a policy answers "what does this cost?". Keep the two apart —
// the moment a policy starts making qualification decisions, both get harder
// to test.
type DiscountPolicy func(order *domain.Order, currentPrice domain.Money) (domain.Money, string, error)

// ChainPolicies applies every policy in order, feeding each one the price the
// previous produced, and joins the non-empty descriptions with " + ".
//
// Discounts stack here. Whether they *should* stack is a business decision, not
// a technical one — and BestOfPolicies below is the other answer.
func ChainPolicies(policies ...DiscountPolicy) DiscountPolicy {
	panic("TODO")
}

// BestOfPolicies applies every policy to the *same* starting price and keeps
// whichever single result is cheapest for the customer.
//
// Note the difference from ChainPolicies: these do not compose, they compete.
// Each policy sees the original price, and only one wins.
//
// This is the mutually-exclusive-promotions rule from L15.1 in its simplest
// form. Once you have written it, look at how badly it scales: with rules that
// are *partially* exclusive — A excludes B but combines with C — picking the
// best combination is no longer a single pass. That is the real problem waiting
// in L15, and this function is the warm-up.
func BestOfPolicies(policies ...DiscountPolicy) DiscountPolicy {
	panic("TODO")
}

// NewYearDiscount takes 10% off orders created in January.
//
// It reads the month off the order rather than the clock, which is what makes
// it testable. A policy that calls time.Now itself can only be tested in
// January.
func NewYearDiscount() DiscountPolicy {
	panic("TODO")
}

// BulkOrderDiscount takes discountPercent off orders with at least minLines
// lines, and leaves smaller orders alone.
//
// The description is expected in the form "Bulk 3+ items 10% off".
//
// You are multiplying an integer cent amount by a float percentage. The same
// rounding decision you met in Money.PercentageDiscount comes back here — make
// it the same way, or the two will disagree on some orders and the bug will be
// found by a customer.
func BulkOrderDiscount(minLines int, discountPercent float64) DiscountPolicy {
	panic("TODO")
}

// LoyaltyDiscount applies a percentage based on the customer tier:
// gold 15%, silver 10%, bronze 5%, and anything else nothing.
//
// An unknown tier returning the price untouched is deliberate. Ask whether a
// typo in a tier name should really be silent, and what it would cost to make
// it loud instead.
func LoyaltyDiscount(tier string) DiscountPolicy {
	panic("TODO")
}

// MinimumPrice raises the price back up to min when discounts have driven it
// below.
//
// This is a floor, not a discount, which makes it the odd one out: it is the
// only policy here that can *increase* the price. That is exactly why it has to
// be last in a chain, and nothing in the type system enforces that. Note the
// fragility — a rule whose correctness depends on call order is a rule someone
// will eventually break.
func MinimumPrice(min domain.Money) DiscountPolicy {
	panic("TODO")
}

// ValidationRule reports whether an order violates one specific constraint.
type ValidationRule func(order *domain.Order) error

// ValidateAll runs every rule and collects all the failures.
//
// It does not stop at the first one, and that is the point. A form that reports
// one error, then the next after you fix it, then a third, is a form people
// hate. Gather everything, report it once.
//
// Decide what an order with no violations returns — nil or an empty slice — and
// check what the test expects before you assume.
func ValidateAll(order *domain.Order, rules ...ValidationRule) []error {
	panic("TODO")
}

// MaxLinesRule rejects orders with more than max lines.
func MaxLinesRule(max int) ValidationRule {
	panic("TODO")
}

// MaxAmountRule rejects orders totalling more than max.
//
// Careful: getting the total can itself fail. A ValidationRule returns one
// error, so you have to decide whether "could not compute the total" is a
// validation failure or something else entirely. There is a defensible argument
// both ways; make the choice consciously.
func MaxAmountRule(max domain.Money) ValidationRule {
	panic("TODO")
}

// NotExpiredRule rejects orders older than maxAge.
//
// This one does read the clock, unlike NewYearDiscount. That makes it awkward
// to test — the test has to construct an order with a creation time in the
// past, or wait. Notice the cost, and remember it when you decide whether the
// domain should ever call time.Now directly.
func NotExpiredRule(maxAge time.Duration) ValidationRule {
	panic("TODO")
}
