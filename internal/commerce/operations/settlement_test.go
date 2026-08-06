package operations

import (
	"errors"
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"reflect"
	"testing"
)

func aud(t *testing.T, v float64) domain.Money {
	t.Helper()
	m, err := domain.NewMoney(v, domain.AUD)
	if err != nil {
		t.Fatalf("NewMoney(%v) = %v", v, err)
	}
	return m
}

func sum(parts []int64) int64 {
	var total int64
	for _, p := range parts {
		total += p
	}
	return total
}

// ---- AllocateByWeight ----

func TestAllocateByWeight_Examples(t *testing.T) {
	tests := []struct {
		name    string
		total   int64
		weights []int64
		want    []int64
	}{
		{"equal thirds, one cent over", 10000, []int64{1, 1, 1}, []int64{3334, 3333, 3333}},
		{"proportional and exact", 10000, []int64{5000, 3000, 2000}, []int64{5000, 3000, 2000}},
		{"a zero weight gets nothing", 10000, []int64{1, 0, 1}, []int64{5000, 0, 5000}},
		{"nothing to split", 0, []int64{1, 1}, []int64{0, 0}},
		{"one part takes it all", 999, []int64{7}, []int64{999}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AllocateByWeight(tt.total, tt.weights)
			if err != nil {
				t.Fatalf("AllocateByWeight = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// The property that cannot bend.
func TestAllocateByWeight_AlwaysSumsToTheTotal(t *testing.T) {
	cases := []struct {
		total   int64
		weights []int64
	}{
		{10000, []int64{1, 1, 1}},
		{10000, []int64{1, 1, 1, 1, 1, 1, 1}},
		{1, []int64{1, 1, 1}},
		{7, []int64{3, 3, 3}},
		{100, []int64{99, 1}},
		{123456789, []int64{7, 11, 13, 17}},
	}

	for _, c := range cases {
		got, err := AllocateByWeight(c.total, c.weights)
		if err != nil {
			t.Fatalf("AllocateByWeight(%d, %v) = %v", c.total, c.weights, err)
		}
		if s := sum(got); s != c.total {
			t.Errorf("AllocateByWeight(%d, %v) = %v, which sums to %d — "+
				"money was created or lost", c.total, c.weights, got, s)
		}
	}
}

// Recomputing a split must give the same answer, or two systems disagree about
// which sub-order owes the extra cent.
func TestAllocateByWeight_IsDeterministic(t *testing.T) {
	first, err := AllocateByWeight(10000, []int64{1, 1, 1})
	if err != nil {
		t.Fatal(err)
	}
	for i := range 50 {
		again, err := AllocateByWeight(10000, []int64{1, 1, 1})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(first, again) {
			t.Fatalf("run %d gave %v, first run gave %v", i, again, first)
		}
	}
}

func TestAllocateByWeight_Errors(t *testing.T) {
	if _, err := AllocateByWeight(100, nil); !errors.Is(err, ErrNoParts) {
		t.Errorf("no weights = %v, want ErrNoParts", err)
	}
	if _, err := AllocateByWeight(100, []int64{0, 0}); !errors.Is(err, ErrZeroTotalWeight) {
		t.Errorf("all-zero weights = %v, want ErrZeroTotalWeight", err)
	}
	if _, err := AllocateByWeight(100, []int64{1, -1}); !errors.Is(err, ErrNegativeWeight) {
		t.Errorf("negative weight = %v, want ErrNegativeWeight", err)
	}
}

// ---- SplitOrder ----

func testLines(t *testing.T) []Line {
	t.Helper()
	return []Line{
		{ID: "line-1", Warehouse: "syd", Amount: aud(t, 300)},
		{ID: "line-2", Warehouse: "syd", Amount: aud(t, 100)},
		{ID: "line-3", Warehouse: "mel", Amount: aud(t, 100)},
	}
}

func TestSplitOrder_GroupsAndSharesTheDiscount(t *testing.T) {
	subs, err := SplitOrder(testLines(t), aud(t, 100))
	if err != nil {
		t.Fatalf("SplitOrder = %v", err)
	}
	if len(subs) != 2 {
		t.Fatalf("got %d sub-orders, want 2", len(subs))
	}

	syd, mel := subs[0], subs[1]
	if syd.Warehouse != "syd" {
		t.Fatalf("first sub-order is %q, want syd (first appearance wins)", syd.Warehouse)
	}
	if got, want := syd.Gross.Cents(), int64(40000); got != want {
		t.Errorf("syd Gross = %d, want %d", got, want)
	}
	if got, want := syd.DiscountShare.Cents(), int64(8000); got != want {
		t.Errorf("syd DiscountShare = %d, want %d (400/500 of the discount)", got, want)
	}
	if got, want := syd.Net.Cents(), int64(32000); got != want {
		t.Errorf("syd Net = %d, want %d", got, want)
	}
	if got, want := mel.DiscountShare.Cents(), int64(2000); got != want {
		t.Errorf("mel DiscountShare = %d, want %d", got, want)
	}
	if want := []string{"line-1", "line-2"}; !reflect.DeepEqual(syd.LineIDs, want) {
		t.Errorf("syd LineIDs = %v, want %v", syd.LineIDs, want)
	}
}

func TestSplitOrder_DiscountSharesSumToTheDiscount(t *testing.T) {
	// 100 split across three equal warehouses: 33.34 / 33.33 / 33.33.
	lines := []Line{
		{ID: "a", Warehouse: "w1", Amount: aud(t, 100)},
		{ID: "b", Warehouse: "w2", Amount: aud(t, 100)},
		{ID: "c", Warehouse: "w3", Amount: aud(t, 100)},
	}
	subs, err := SplitOrder(lines, aud(t, 100))
	if err != nil {
		t.Fatalf("SplitOrder = %v", err)
	}

	var shares, nets int64
	for _, s := range subs {
		shares += s.DiscountShare.Cents()
		nets += s.Net.Cents()
	}
	if shares != 10000 {
		t.Errorf("discount shares sum to %d, want 10000 — a cent went missing", shares)
	}
	if nets != 20000 {
		t.Errorf("nets sum to %d, want 20000 (30000 gross − 10000 discount)", nets)
	}
}

func TestSplitOrder_EmptyAndUndiscounted(t *testing.T) {
	if subs, err := SplitOrder(nil, aud(t, 0)); err != nil || len(subs) != 0 {
		t.Errorf("SplitOrder(nil) = %v, %v; want an empty result and no error", subs, err)
	}

	subs, err := SplitOrder(testLines(t), aud(t, 0))
	if err != nil {
		t.Fatalf("SplitOrder with no discount = %v", err)
	}
	for _, s := range subs {
		if s.Net.Cents() != s.Gross.Cents() {
			t.Errorf("%s: Net %d != Gross %d with no discount",
				s.Warehouse, s.Net.Cents(), s.Gross.Cents())
		}
	}
}

func TestSplitOrder_IsDeterministic(t *testing.T) {
	first, err := SplitOrder(testLines(t), aud(t, 100))
	if err != nil {
		t.Fatal(err)
	}
	for i := range 30 {
		again, err := SplitOrder(testLines(t), aud(t, 100))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(first, again) {
			t.Fatalf("run %d differed from the first", i)
		}
	}
}

// ---- CalculateRefund ----

func testPayment(t *testing.T) Payment {
	t.Helper()
	// 500 gross − 100 discount + 20 shipping = 420 paid.
	return Payment{
		Gross:    aud(t, 500),
		Discount: aud(t, 100),
		Shipping: aud(t, 20),
		Paid:     aud(t, 420),
	}
}

func TestCalculateRefund_ClawsBackTheDiscountShare(t *testing.T) {
	got, err := CalculateRefund(testPayment(t), testLines(t), RefundRequest{LineIDs: []string{"line-3"}})
	if err != nil {
		t.Fatalf("CalculateRefund = %v", err)
	}

	if c := got.LineRefund.Cents(); c != 10000 {
		t.Errorf("LineRefund = %d, want 10000", c)
	}
	if c := got.DiscountClawback.Cents(); c != 2000 {
		t.Errorf("DiscountClawback = %d, want 2000 (100/500 of the discount)", c)
	}
	if c := got.ShippingRefund.Cents(); c != 0 {
		t.Errorf("ShippingRefund = %d, want 0", c)
	}
	if c := got.Total.Cents(); c != 8000 {
		t.Errorf("Total = %d, want 8000. Refunding the line's gross value and "+
			"forgetting the clawback lets the customer keep a discount for money "+
			"they no longer spent.", c)
	}
	if len(got.Reasons) == 0 {
		t.Error("Reasons is empty: a refund the customer cannot understand is a " +
			"support ticket")
	}
}

// The property everything else has to be consistent with.
func TestCalculateRefund_FullReturnReconcilesExactly(t *testing.T) {
	payment := testPayment(t)
	lines := testLines(t)

	got, err := CalculateRefund(payment, lines, RefundRequest{
		LineIDs:        []string{"line-1", "line-2", "line-3"},
		RefundShipping: true,
	})
	if err != nil {
		t.Fatalf("CalculateRefund = %v", err)
	}

	if got.Total.Cents() != payment.Paid.Cents() {
		t.Errorf("returning everything refunds %d, but the customer paid %d. "+
			"If the full case does not reconcile, no partial case is right either.",
			got.Total.Cents(), payment.Paid.Cents())
	}
}

func TestCalculateRefund_ShippingOnlyWhenAsked(t *testing.T) {
	withShipping, err := CalculateRefund(testPayment(t), testLines(t), RefundRequest{
		LineIDs:        []string{"line-3"},
		RefundShipping: true,
	})
	if err != nil {
		t.Fatalf("CalculateRefund = %v", err)
	}
	if withShipping.ShippingRefund.Cents() != 2000 {
		t.Errorf("ShippingRefund = %d, want 2000", withShipping.ShippingRefund.Cents())
	}
	if withShipping.Total.Cents() != 10000 {
		t.Errorf("Total = %d, want 10000 (8000 + 2000 shipping)",
			withShipping.Total.Cents())
	}
}

func TestCalculateRefund_NothingReturned(t *testing.T) {
	got, err := CalculateRefund(testPayment(t), testLines(t), RefundRequest{})
	if err != nil {
		t.Fatalf("refunding nothing = %v, want no error", err)
	}
	if got.Total.Cents() != 0 {
		t.Errorf("Total = %d, want 0", got.Total.Cents())
	}
}

func TestCalculateRefund_UnknownLine(t *testing.T) {
	_, err := CalculateRefund(testPayment(t), testLines(t), RefundRequest{LineIDs: []string{"nope"}})
	if !errors.Is(err, ErrUnknownLine) {
		t.Errorf("unknown line = %v, want ErrUnknownLine", err)
	}
}

// Nothing in the model may allow refunding more than was charged.
func TestCalculateRefund_CannotExceedWhatWasPaid(t *testing.T) {
	payment := testPayment(t)
	payment.Paid = aud(t, 50) // deliberately inconsistent with the lines

	_, err := CalculateRefund(payment, testLines(t), RefundRequest{
		LineIDs:        []string{"line-1", "line-2", "line-3"},
		RefundShipping: true,
	})
	if !errors.Is(err, ErrRefundExceedsPaid) {
		t.Errorf("over-refund = %v, want ErrRefundExceedsPaid", err)
	}
}
