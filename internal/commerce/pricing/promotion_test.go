package pricing

import (
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"reflect"
	"sort"
	"testing"
	"time"
)

// ---- fixtures ----

var (
	evalAt = time.Date(2026, time.March, 15, 12, 0, 0, 0, time.UTC)
	always = Window{
		From: evalAt.Add(-30 * 24 * time.Hour),
		To:   evalAt.Add(30 * 24 * time.Hour),
	}
)

func aud(t *testing.T, v float64) domain.Money {
	t.Helper()
	m, err := domain.NewMoney(v, domain.AUD)
	if err != nil {
		t.Fatalf("NewMoney(%v) = %v", v, err)
	}
	return m
}

// cartOf builds a single-line cart worth unit*qty.
func cartOf(t *testing.T, unit float64, qty int) Cart {
	t.Helper()
	return Cart{
		CustomerID: "cust-1",
		Tier:       "gold",
		Lines:      []Line{{SKU: "sku-1", Category: "general", Unit: aud(t, unit), Quantity: qty}},
		At:         evalAt,
	}
}

func engine(t *testing.T, rules []Rule, ex []Excludes, floor float64) *Engine {
	t.Helper()
	e := NewEngine(rules, ex, floor)
	if e == nil {
		t.Fatal("NewEngine returned nil")
	}
	return e
}

func best(t *testing.T, e *Engine, cart Cart) Outcome {
	t.Helper()
	out, err := e.Best(cart)
	if err != nil {
		t.Fatalf("Best() = %v", err)
	}
	return out
}

// ---- the basics ----

func TestCart_Subtotal(t *testing.T) {
	cart := cartOf(t, 100, 5)
	got, err := cart.Subtotal()
	if err != nil {
		t.Fatalf("Subtotal() = %v", err)
	}
	if want := int64(50000); got.Cents() != want {
		t.Errorf("Subtotal() = %d cents, want %d", got.Cents(), want)
	}
}

func TestThresholdDiscount_Applies(t *testing.T) {
	e := engine(t, []Rule{
		ThresholdDiscount("spend500", aud(t, 500), aud(t, 50), always),
	}, nil, 0)

	out := best(t, e, cartOf(t, 100, 5)) // $500

	if out.Discount.Cents() != 5000 {
		t.Errorf("Discount = %d cents, want 5000", out.Discount.Cents())
	}
	if out.Final.Cents() != 45000 {
		t.Errorf("Final = %d cents, want 45000", out.Final.Cents())
	}
	if !reflect.DeepEqual(out.AppliedIDs, []string{"spend500"}) {
		t.Errorf("AppliedIDs = %v, want [spend500]", out.AppliedIDs)
	}
}

func TestThresholdDiscount_BelowThreshold(t *testing.T) {
	e := engine(t, []Rule{
		ThresholdDiscount("spend500", aud(t, 500), aud(t, 50), always),
	}, nil, 0)

	out := best(t, e, cartOf(t, 100, 4)) // $400

	if !out.Discount.IsZero() {
		t.Errorf("Discount = %s, want zero", out.Discount)
	}
	if len(out.AppliedIDs) != 0 {
		t.Errorf("AppliedIDs = %v, want none", out.AppliedIDs)
	}
}

// The threshold is inclusive: spending exactly $500 qualifies.
func TestThresholdDiscount_ExactlyAtThreshold(t *testing.T) {
	e := engine(t, []Rule{
		ThresholdDiscount("spend500", aud(t, 500), aud(t, 50), always),
	}, nil, 0)

	out := best(t, e, cartOf(t, 500, 1))

	if out.Discount.Cents() != 5000 {
		t.Errorf("Discount = %d cents at exactly the threshold, want 5000", out.Discount.Cents())
	}
}

// ---- stacking ----

func TestStackableRulesCombine(t *testing.T) {
	e := engine(t, []Rule{
		ThresholdDiscount("spend500", aud(t, 500), aud(t, 50), always),
		Coupon("welcome", "WELCOME20", aud(t, 20), always),
	}, nil, 0)

	cart := cartOf(t, 100, 5)
	cart.CouponCode = "WELCOME20"

	out := best(t, e, cart)

	if out.Discount.Cents() != 7000 {
		t.Errorf("Discount = %d cents, want 7000 ($50 + $20)", out.Discount.Cents())
	}
	if len(out.AppliedIDs) != 2 {
		t.Errorf("AppliedIDs = %v, want both rules", out.AppliedIDs)
	}
}

func TestCoupon_RequiresMatchingCode(t *testing.T) {
	e := engine(t, []Rule{
		Coupon("welcome", "WELCOME20", aud(t, 20), always),
	}, nil, 0)

	out := best(t, e, cartOf(t, 100, 5)) // no code supplied

	if !out.Discount.IsZero() {
		t.Errorf("Discount = %s without a coupon code, want zero", out.Discount)
	}
}

// ---- exclusion ----

func TestMutuallyExclusiveRules_PicksBetterForCustomer(t *testing.T) {
	// On a $500 cart both promotions give $50, so the stable tie-break applies.
	// so use $600 where the percentage clearly wins.
	e := engine(t, []Rule{
		ThresholdDiscount("spend500", aud(t, 500), aud(t, 50), always),
		PercentageDiscount("tenoff", 10, always),
	}, []Excludes{{A: "spend500", B: "tenoff"}}, 0)

	out := best(t, e, cartOf(t, 100, 6)) // $600: fixed $50 versus percentage $60

	if out.Discount.Cents() != 6000 {
		t.Errorf("Discount = %d cents, want 6000 (the better of the two)", out.Discount.Cents())
	}
	if !reflect.DeepEqual(out.AppliedIDs, []string{"tenoff"}) {
		t.Errorf("AppliedIDs = %v, want [tenoff]", out.AppliedIDs)
	}
}

// The exclusion is symmetric even though it was declared in one direction.
func TestExclusionIsSymmetric(t *testing.T) {
	rules := []Rule{
		ThresholdDiscount("spend500", aud(t, 500), aud(t, 50), always),
		PercentageDiscount("tenoff", 10, always),
	}

	forward := best(t, engine(t, rules, []Excludes{{A: "spend500", B: "tenoff"}}, 0), cartOf(t, 100, 6))
	backward := best(t, engine(t, rules, []Excludes{{A: "tenoff", B: "spend500"}}, 0), cartOf(t, 100, 6))

	if forward.Discount.Cents() != backward.Discount.Cents() {
		t.Errorf("declaring the exclusion the other way round changed the result: %d vs %d",
			forward.Discount.Cents(), backward.Discount.Cents())
	}
}

// This is the case that breaks the greedy answer.
//
// On a $1000 cart:
//
//	big   $250 off, excludes small
//	small $150 off, stacks with bonus
//	bonus $200 off, excludes big
//
// Greedy takes big first ($250), then cannot take bonus, and small is excluded
// by big — total $250. The best legal combination is small + bonus = $350.
func TestPartialExclusion_GreedyIsWrong(t *testing.T) {
	e := engine(t, []Rule{
		ThresholdDiscount("big", aud(t, 100), aud(t, 250), always),
		ThresholdDiscount("small", aud(t, 100), aud(t, 150), always),
		ThresholdDiscount("bonus", aud(t, 100), aud(t, 200), always),
	}, []Excludes{
		{A: "big", B: "small"},
		{A: "big", B: "bonus"},
	}, 0)

	out := best(t, e, cartOf(t, 100, 10)) // $1000

	if out.Discount.Cents() != 35000 {
		t.Errorf("Discount = %d cents, want 35000 (small + bonus). "+
			"A greedy pass takes big first and lands on 25000.", out.Discount.Cents())
	}

	got := append([]string(nil), out.AppliedIDs...)
	sort.Strings(got)
	if want := []string{"bonus", "small"}; !reflect.DeepEqual(got, want) {
		t.Errorf("AppliedIDs = %v, want %v", got, want)
	}
}

// ---- the floor ----

func TestDiscountNeverBreachesTheFloor(t *testing.T) {
	// Floor 0.6 on a $1000 cart means the customer can never pay less than $600,
	// so the discount is capped at $400 however many rules qualify.
	e := engine(t, []Rule{
		ThresholdDiscount("a", aud(t, 100), aud(t, 300), always),
		ThresholdDiscount("b", aud(t, 100), aud(t, 300), always),
	}, nil, 0.6)

	out := best(t, e, cartOf(t, 100, 10))

	if out.Final.Cents() < 60000 {
		t.Errorf("Final = %d cents, below the 60000 floor", out.Final.Cents())
	}
	if out.Discount.Cents() > 40000 {
		t.Errorf("Discount = %d cents, want at most 40000", out.Discount.Cents())
	}
	if sum := out.Final.Cents() + out.Discount.Cents(); sum != 100000 {
		t.Errorf("Final + Discount = %d, want 100000: the arithmetic must still balance", sum)
	}
}

// ---- gifts ----

func TestGiftDoesNotReduceTheAmountButIsReported(t *testing.T) {
	e := engine(t, []Rule{
		GiftWithPurchase("freemug", aud(t, 300), "sku-mug", always),
	}, nil, 0)

	out := best(t, e, cartOf(t, 100, 5)) // $500

	if !out.Discount.IsZero() {
		t.Errorf("Discount = %s, want zero: a gift costs the customer nothing "+
			"but discounts nothing either", out.Discount)
	}
	if !reflect.DeepEqual(out.Gifts, []string{"sku-mug"}) {
		t.Errorf("Gifts = %v, want [sku-mug]", out.Gifts)
	}
	if len(out.AppliedIDs) != 1 {
		t.Errorf("AppliedIDs = %v, want the gift rule to be recorded as applied", out.AppliedIDs)
	}
}

// ---- validity ----

func TestExpiredPromotionIsIgnored(t *testing.T) {
	expired := Window{
		From: evalAt.Add(-60 * 24 * time.Hour),
		To:   evalAt.Add(-1 * 24 * time.Hour),
	}
	e := engine(t, []Rule{
		ThresholdDiscount("old", aud(t, 100), aud(t, 50), expired),
	}, nil, 0)

	out := best(t, e, cartOf(t, 100, 5))

	if !out.Discount.IsZero() {
		t.Errorf("Discount = %s, want zero for an expired promotion", out.Discount)
	}
}

func TestWithdrawnPromotionIsIgnored(t *testing.T) {
	withdrawn := always
	withdrawn.Withdrawn = true

	e := engine(t, []Rule{
		ThresholdDiscount("pulled", aud(t, 100), aud(t, 50), withdrawn),
	}, nil, 0)

	out := best(t, e, cartOf(t, 100, 5))

	if !out.Discount.IsZero() {
		t.Errorf("Discount = %s, want zero for a withdrawn promotion", out.Discount)
	}
}

// ---- determinism ----

// Same input, same output — including the order of AppliedIDs.
//
// Ranging over a map and acting on the order will fail this test some of the
// time, which is the worst way for it to fail.
func TestResultIsDeterministic(t *testing.T) {
	rules := []Rule{
		ThresholdDiscount("a", aud(t, 100), aud(t, 100), always),
		ThresholdDiscount("b", aud(t, 100), aud(t, 100), always),
		ThresholdDiscount("c", aud(t, 100), aud(t, 100), always),
	}
	e := engine(t, rules, nil, 0)
	cart := cartOf(t, 100, 10)

	first := best(t, e, cart)
	for i := range 50 {
		again := best(t, e, cart)
		if !reflect.DeepEqual(first.AppliedIDs, again.AppliedIDs) {
			t.Fatalf("run %d produced AppliedIDs %v, first run gave %v",
				i, again.AppliedIDs, first.AppliedIDs)
		}
		if first.Discount.Cents() != again.Discount.Cents() {
			t.Fatalf("run %d produced discount %d, first run gave %d",
				i, again.Discount.Cents(), first.Discount.Cents())
		}
	}
}

func TestEmptyCart(t *testing.T) {
	e := engine(t, []Rule{
		ThresholdDiscount("spend500", aud(t, 500), aud(t, 50), always),
	}, nil, 0)

	out, err := e.Best(Cart{CustomerID: "cust-1", At: evalAt})
	if err != nil {
		t.Fatalf("Best() on an empty cart = %v", err)
	}
	if !out.Discount.IsZero() {
		t.Errorf("Discount = %s on an empty cart, want zero", out.Discount)
	}
}
