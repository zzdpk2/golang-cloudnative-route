// Package settlement is L15.3 and L15.4: splitting an order, and unwinding it.
//
// A ● exercise. Two problems that look like arithmetic and are not.
//
// # L15.3 — splitting
//
// An order ships from three warehouses, so it becomes three sub-orders. The
// customer paid one total, with one discount applied to the whole basket. Each
// sub-order now needs its own share of that money, for invoicing, for tax, and
// for the accounts.
//
// The money must still add up. A$100.00 split three ways is not three lots of
// A$33.33 — that loses a cent, and a cent lost per order is a reconciliation
// failure at the end of the month that somebody has to find by hand.
//
// # L15.4 — partial refund
//
// The customer returns one of three items. Now:
//
//   - the discount they received was for spending A$500. After the return they
//     spent A$380. **Do you claw the discount back?**
//   - the coupon was single-use. Is it now spent, or returned to them?
//   - the shipping was free above A$300. It no longer is.
//
// This is *inverse calculation*, and the forward calculation was lossy. You
// cannot simply run the pricing backwards, because several different baskets
// produce the same total. That is the whole difficulty, and it is why refund
// logic in real systems is longer than the pricing logic it undoes.
//
// # Answer these first
//
//  1. Is a sub-order an aggregate of its own, or part of the order aggregate?
//     What changes if a warehouse can confirm its own shipment independently?
//  2. Allocation by line value is the obvious rule. Name a case where it is
//     visibly unfair to somebody, and say who complains.
//  3. Is the extra cent given to the largest sub-order, the first, or the one
//     the customer sees first? Whichever you pick, it must be *deterministic* —
//     the same order must split the same way every time it is recomputed.
//  4. For refunds: is a refund a new transaction, or a mutation of the
//     original? Only one of those answers survives an audit.
//  5. Can the sum of all refunds exceed the original payment? What in your
//     model makes that impossible rather than merely unlikely?
package operations

import (
	"errors"
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
)

var (
	// ErrNoParts means the caller asked to split into nothing.
	ErrNoParts = errors.New("no parts to allocate across")
	// ErrNegativeWeight means a weight was negative, which has no meaning here.
	ErrNegativeWeight = errors.New("weights must not be negative")
	// ErrZeroTotalWeight means every weight was zero, so there is no basis to
	// allocate on.
	ErrZeroTotalWeight = errors.New("total weight is zero")
	// ErrRefundExceedsPaid means a refund would return more than was charged.
	ErrRefundExceedsPaid = errors.New("refund exceeds the amount paid")
	// ErrUnknownLine means the refund names a line the order does not have.
	ErrUnknownLine = errors.New("unknown line")
)

// ---- L15.3: allocation ----

// AllocateByWeight splits an amount in proportion to weights.
//
//	Allocate(10000, [1,1,1])    → [3334 3333 3333]   the cent goes to the first
//	Allocate(10000, [5000,3000,2000]) → [5000 3000 2000]
//	Allocate(10000, [1,0,1])    → [5000 0 5000]      a zero weight gets nothing
//	Allocate(0, [1,1])          → [0 0]
//	Allocate(10000, [])         → ErrNoParts
//	Allocate(10000, [0,0])      → ErrZeroTotalWeight
//	Allocate(10000, [1,-1])     → ErrNegativeWeight
//
// **The parts must sum to exactly the total, always.** That is the one property
// that cannot bend, and it is what makes this harder than `total*w/sum`:
// rounding each share independently loses money.
//
// The technique that works is to compute each share by truncation, then hand
// the leftover units out one at a time. Which parts receive them is question 3
// — decide, and make it deterministic. The test runs the same split repeatedly
// and requires the same answer.
//
// This is domain.Money.Allocate from L1 generalised to unequal shares. If you wrote
// that one well, most of this is already familiar.
func AllocateByWeight(totalCents int64, weights []int64) ([]int64, error) {
	panic("TODO")
}

// SubOrder is one shipment's worth of an order.
type SubOrder struct {
	Warehouse string
	// LineIDs are the order lines shipping from this warehouse.
	LineIDs []string
	// Gross is the value of those lines before any discount.
	Gross domain.Money
	// DiscountShare is this sub-order's share of the order-level discount.
	DiscountShare domain.Money
	// Net is Gross − DiscountShare, and what this sub-order is invoiced for.
	Net domain.Money
}

// Line is one line of the original order.
type Line struct {
	ID        string
	Warehouse string
	Amount    domain.Money
}

// SplitOrder groups lines by warehouse and shares the order-level discount out
// across the resulting sub-orders.
//
// Worked example — three lines, A$100 discount on an A$500 order:
//
//	line-1  syd  A$300     →  syd: Gross 300, DiscountShare 60, Net 240
//	line-2  syd  A$100        (syd is 400/500 of the order, so 80% of 100)
//	line-3  mel  A$100     →  mel: Gross 100, DiscountShare 20, Net 80
//
// Rules:
//
//   - Sub-orders come back in the order their warehouse first appears in lines,
//     so the result is deterministic.
//   - The discount shares must sum to exactly the discount.
//   - The nets must sum to exactly gross minus discount.
//   - A discount larger than the order total is a caller error.
//   - No lines is not an error; it is an empty result.
//
// Allocate by *line value*, which is question 2. It is the defensible default
// and it is not neutral: a sub-order of cheap items gets a small share of the
// discount, so if the customer returns exactly that shipment they get less back
// than they might expect. Know that before you ship it.
func SplitOrder(lines []Line, discount domain.Money) ([]SubOrder, error) {
	panic("TODO")
}

// ---- L15.4: partial refund ----

// Payment is what the customer actually paid.
type Payment struct {
	// Gross is the value of all lines before discounts.
	Gross domain.Money
	// Discount is what came off the whole order.
	Discount domain.Money
	// Shipping is what they paid to have it delivered.
	Shipping domain.Money
	// Paid is Gross − Discount + Shipping.
	Paid domain.Money
}

// RefundRequest names the lines being returned.
type RefundRequest struct {
	LineIDs []string
	// RefundShipping says whether the delivery charge comes back too. Usually
	// it does not, unless the return is the seller's fault.
	RefundShipping bool
}

// RefundResult is what will actually be paid back, and why.
type RefundResult struct {
	// LineRefund is the returned lines' gross value.
	LineRefund domain.Money
	// DiscountClawback is the share of the discount attached to those lines,
	// which the customer no longer qualifies for and does not get back.
	DiscountClawback domain.Money
	// ShippingRefund is the delivery charge, when it is being returned.
	ShippingRefund domain.Money
	// Total is what hits the customer's card: LineRefund − DiscountClawback
	// + ShippingRefund.
	Total domain.Money
	// Reasons explains each component, because a refund a customer cannot
	// understand becomes a support ticket.
	Reasons []string
}

// CalculateRefund works out what to return for a partial return.
//
// Worked example — an A$500 order with an A$100 discount and A$20 shipping,
// so the customer paid A$420. They return a line worth A$100:
//
//	LineRefund       A$100
//	DiscountClawback A$20      the returned line carried 100/500 of the discount
//	ShippingRefund   A$0       not the seller's fault
//	Total            A$80
//
// The customer paid A$420 and gets A$80 back. **They do not get A$100**, and
// explaining why is exactly what Reasons is for.
//
// Rules the tests pin down:
//
//   - Returning every line refunds everything paid, to the cent. That is the
//     property to build around: full return must reconcile exactly, or the
//     partial cases are wrong too.
//   - Refunds cannot exceed what was paid → ErrRefundExceedsPaid.
//   - An unknown line id → ErrUnknownLine.
//   - Returning nothing is a zero refund, not an error.
//
// # The trap
//
// Refunding the line's gross value and forgetting the clawback means every
// partial return leaks the discount — the customer keeps a discount for
// spending money they no longer spent. It is a real and expensive bug, and it
// looks correct in every single-item test.
//
// # The part with no clean answer
//
// If the discount was "spend A$500, get A$100 off" and the return drops them to
// A$400, they no longer qualify *at all*. Clawing back proportionally
// under-recovers; clawing back the whole discount can make a refund negative.
// The tests use proportional because it is defensible and simple. Write down
// what you would do about the threshold case — it is question 5, and it is why
// this file is a ● rather than a ◐.
func CalculateRefund(payment Payment, lines []Line, req RefundRequest) (RefundResult, error) {
	panic("TODO")
}
