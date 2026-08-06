package pricing

import (
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
)

// ============================================================
// Domain Services
//
// Rules that span aggregates, expressed as pricing strategies.
// ============================================================

// PricingStrategy computes a discount for an order.
//
// This is the Gang of Four Strategy pattern in its natural habitat. Compare it
// with the textbook version in internal/foundation/patterns — same shape, but here it
// exists because the business genuinely has interchangeable pricing rules, not
// because a catalogue said so. That difference is the thing to notice: patterns
// are named solutions to recurring problems, and the problem has to come first.
//
// Why a domain *service* rather than a method on Order? Because a discount can
// depend on the customer's tier, the current campaign, and the catalogue —
// things outside the order domain. A rule that needs data from several
// aggregates does not belong inside any one of them.
type PricingStrategy interface {
	CalculateDiscount(order *domain.Order) (domain.Money, error)
	Name() string
}

// NoDiscount is the null object: a strategy that is always applicable and
// always does nothing.
//
// Having one means callers never need a nil check, which is the entire value of
// the pattern. Note that it hardcodes AUD — a real limitation, and a good
// question to sit with: where should a zero amount get its currency from?
type NoDiscount struct{}

func (d *NoDiscount) CalculateDiscount(order *domain.Order) (domain.Money, error) { panic("TODO") }

func (d *NoDiscount) Name() string { panic("TODO") }

// PercentageOff takes a fixed percentage off the order total.
type PercentageOff struct {
	Percent float64
}

// CalculateDiscount returns the discount *amount*, not the discounted price.
//
// Getting that backwards is the classic bug in pricing code, and it is silent:
// both are Money, both are plausible, and the tests are the only thing that
// will tell you. Read the test before you write the body.
func (d *PercentageOff) CalculateDiscount(order *domain.Order) (domain.Money, error) {
	panic("TODO")
}

func (d *PercentageOff) Name() string { panic("TODO") }

// BulkDiscount gives a flat Discount off orders worth at least MinAmount.
type BulkDiscount struct {
	MinAmount domain.Money
	Discount  domain.Money
}

// CalculateDiscount returns Discount for qualifying orders and zero otherwise.
//
// Comparing the total against MinAmount can fail on a currency mismatch. Decide
// whether a mismatched currency means "does not qualify" or is an error worth
// surfacing — quietly returning zero hides a misconfiguration that someone will
// spend an afternoon tracking down.
func (d *BulkDiscount) CalculateDiscount(order *domain.Order) (domain.Money, error) {
	panic("TODO")
}

func (d *BulkDiscount) Name() string { panic("TODO") }

// PricingService picks the best of the strategies it was configured with.
type PricingService struct {
	strategies []PricingStrategy
}

// PricingOption configures a PricingService at construction.
//
// This is the functional options pattern. It is worth understanding why Go
// reaches for it so often: it keeps the constructor's signature stable as
// options accumulate, makes every option self-documenting at the call site, and
// avoids the config-struct problem where the zero value has to be meaningful
// for every field.
type PricingOption func(*PricingService)

func WithStrategy(s PricingStrategy) PricingOption { panic("TODO") }

// NewPricingService builds a service from options.
//
// Decide what a service with no strategies should do. Returning a zero discount
// is reasonable; so is refusing to be constructed. What you must not do is
// leave it to crash later at the call site.
func NewPricingService(opts ...PricingOption) *PricingService {
	panic("TODO")
}

// BestDiscount returns the largest discount any strategy offers, along with the
// name of the winning strategy.
//
// Strategies compete rather than stack — the same choice policy.BestOfPolicies
// makes, expressed with interfaces instead of funcs. Having now written both,
// you have the material to answer a question worth answering: when is a
// single-method interface better than a function type, and when is it just
// ceremony?
//
// If a strategy errors, decide whether that disqualifies only that strategy or
// fails the whole calculation.
func (ps *PricingService) BestDiscount(order *domain.Order) (domain.Money, string, error) {
	panic("TODO")
}

// FinalPrice is the order total minus the best discount.
//
// Guard the subtraction. A discount larger than the total would drive the price
// negative, and domain.Money will not allow that — which is L1's validation doing
// exactly the job it was built for. Decide what the customer pays in that case.
func (ps *PricingService) FinalPrice(order *domain.Order) (domain.Money, error) {
	panic("TODO")
}
