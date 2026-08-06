// Package promotion is L15.1: the promotion engine.
//
// This is a ● exercise. Unlike every level before it, the tests do not walk you
// through an implementation — they state what the business needs and leave the
// model to you. Only the handful of types the tests must name are fixed below;
// everything else, including whatever types you decide you need, is yours.
//
// # The business
//
// Marketing wants to run several promotions at once:
//
//	threshold discount  spend at least X, get Y off
//	percentage discount a percentage off the whole cart
//	gift with purchase  spend at least X, get a free item
//	coupon              a fixed amount off, redeemed by code
//
// They stack, except when they do not:
//
//   - "spend 500, save 50" and "10 percent off" are mutually exclusive. The customer gets
//     whichever is better *for them*, not whichever the code happens to try
//     first.
//   - a coupon stacks with a membership discount, but the two together must
//     never take the price below the cost floor.
//   - gifts do not reduce the amount owed, but they do consume stock, so they
//     have to be reported.
//   - every promotion has a validity window and may be withdrawn mid-campaign.
//
// # Why this is hard
//
// With rules that are wholly independent, you apply them all. With rules that
// are wholly exclusive, you pick the single best. Real promotions are neither:
// A excludes B, but A combines with C, and B combines with C too. Now "the best
// outcome" is a choice among *combinations*, and the obvious greedy answer —
// take the biggest discount first, then whatever still fits — is wrong. The
// tests contain a case that proves it.
//
// So the first question is not how to write the code. It is: what problem is
// this? Once you can name it, you will know whether an exhaustive search is
// acceptable, and what "acceptable" depends on.
//
// # Before you write anything
//
// Answer these in writing. They are the actual exercise; the code is the
// consequence.
//
//  1. Is a promotion an entity or a value object? Does "spend 500, save 50" have an
//     identity that survives marketing editing the threshold to 400?
//  2. Where does exclusion live — on the rule, or in a separate structure that
//     knows about pairs? What happens to a rule that must exclude a rule
//     introduced next quarter?
//  3. "Best for the customer" — best by what measure? Largest discount is the
//     obvious answer and it is wrong as soon as gifts exist, because a gift has
//     value but no price.
//  4. How many combinations can there be, and at what number of rules does an
//     exhaustive search stop being acceptable? Give an actual number.
//  5. The floor rule is expressed as "never below cost". Whose responsibility
//     is that — each rule, the engine, or something outside both?
//
// # Rules of engagement
//
//   - Do not change the test file. It is the specification.
//   - Add whatever types you need. The ones below are only the vocabulary the
//     tests already speak.
//   - When a test surprises you, work out which of your five answers above was
//     wrong before touching the code.
package pricing

import (
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"time"
)

// ---- Vocabulary the tests use. Everything else is your design. ----

// Line is one item in the cart under evaluation.
type Line struct {
	SKU      string
	Category string
	Unit     domain.Money
	Quantity int
}

// Cart is the input to the engine: what is being bought, by whom, and when.
//
// At is the evaluation time, passed in rather than read from the clock. You
// decided this was the right call back in policy.NewYearDiscount; here it is
// what makes the validity-window tests possible at all.
type Cart struct {
	CustomerID string
	Tier       string
	Lines      []Line
	CouponCode string
	At         time.Time
}

// Subtotal is the cart's value before any promotion.
//
// Implement it first — every test needs it, and it is the one function here
// with a single correct answer. Mixed currencies in one cart is a case you have
// to decide about; domain.Money will not let you ignore it.
func (c Cart) Subtotal() (domain.Money, error) {
	panic("TODO")
}

// Outcome is what the engine decides.
type Outcome struct {
	// Discount is the total amount taken off. Gifts contribute nothing here.
	Discount domain.Money
	// Final is Subtotal minus Discount, never below the floor.
	Final domain.Money
	// AppliedIDs lists the rules that were used, so support can answer
	// "why did I pay this?". Order must be deterministic.
	AppliedIDs []string
	// Gifts lists the SKUs to be added at no charge. These still reserve stock.
	Gifts []string
}

// Rule is one promotion.
//
// This interface is a *suggestion*, and the narrowest one that lets the tests
// name rules. If your model wants something different — a sealed set of structs
// and a type switch, say, or rules that return candidate effects rather than
// applying anything — change it. Just keep the constructors below, because the
// tests call them.
//
// Before accepting it as given, notice what it does not say: nothing here
// expresses "A excludes B". That omission is deliberate, and question 2 above
// is asking you to fix it.
type Rule interface {
	ID() string
}

// ---- Constructors the tests call ----

// ThresholdDiscount applies a fixed discount above an inclusive threshold.
func ThresholdDiscount(id string, min, amount domain.Money, window Window) Rule {
	panic("TODO")
}

// PercentageDiscount takes a percentage off the subtotal.
func PercentageDiscount(id string, percent float64, window Window) Rule {
	panic("TODO")
}

// GiftWithPurchase adds giftSKU after the minimum spend.
func GiftWithPurchase(id string, min domain.Money, giftSKU string, window Window) Rule {
	panic("TODO")
}

// Coupon is a fixed amount off, unlocked by code.
func Coupon(id, code string, amount domain.Money, window Window) Rule {
	panic("TODO")
}

// Window is a promotion's validity period, plus whether it has been withdrawn.
//
// Withdrawn is separate from the dates on purpose: marketing pulling a campaign
// at 3pm is not the same event as it reaching its end date, and the two need
// different audit trails. Whether your model actually needs that distinction is
// worth deciding rather than inheriting.
type Window struct {
	From      time.Time
	To        time.Time
	Withdrawn bool
}

// Active reports whether the window covers t.
func (w Window) Active(t time.Time) bool {
	panic("TODO")
}

// Excludes declares that two rules cannot both apply.
//
// The relation is symmetric — if A excludes B then B excludes A — and the tests
// only ever declare it one way round. Whether you store both directions or
// normalise on lookup is your call; getting it wrong shows up as a result that
// depends on the order rules were registered, which is exactly the kind of bug
// the determinism test is there to catch.
type Excludes struct {
	A, B string
}

// Engine evaluates a cart against a set of rules.
type Engine struct {
	// Yours.
}

// NewEngine builds an engine from rules, exclusions, and a cost floor.
//
// The floor is a fraction of subtotal — 0.6 means "never discount below 60% of
// the cart's value". A real system would use per-item cost; this is the
// simplification that keeps the exercise about combination logic.
func NewEngine(rules []Rule, exclusions []Excludes, floor float64) *Engine {
	panic("TODO")
}

// Best returns the outcome most favourable to the customer.
//
// "Most favourable" is the definition you settled in question 3. Whatever you
// chose, two properties are not negotiable and the tests check both:
//
//   - determinism. The same cart and the same rules produce the same Outcome,
//     including the order of AppliedIDs. Ranging over a map and acting on the
//     order will fail this, sometimes, which is the worst way to fail.
//   - no invalid combination is ever returned, no matter how good it looks.
func (e *Engine) Best(cart Cart) (Outcome, error) {
	panic("TODO")
}
