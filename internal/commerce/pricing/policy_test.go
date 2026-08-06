package pricing

import (
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"testing"
)

func makePolicyOrder(t *testing.T, lines int, priceEach float64) *domain.Order {
	t.Helper()
	addr, _ := domain.NewAddress("1 St", "Syd", "NSW", "2000", "AU")
	o, _ := domain.NewOrder("o1", "c1", addr)
	for i := 0; i < lines; i++ {
		p := domain.MustNewMoney(priceEach, domain.AUD)
		q, _ := domain.NewQuantity(1)
		o.AddLine(domain.ProductID(string(rune('A'+i))), "Item", p, q)
	}
	return o
}

func TestChainPolicies(t *testing.T) {
	order := makePolicyOrder(t, 5, 100) // 5 items × $100 = $500
	total, _ := order.TotalAmount()

	chain := ChainPolicies(
		BulkOrderDiscount(3, 10), // 5 >= 3 → 10% off: $500 → $450
		MinimumPrice(domain.MustNewMoney(100, domain.AUD)),
	)

	result, desc, err := chain(order, total)
	if err != nil {
		t.Fatal(err)
	}
	if result.Cents() != 45000 {
		t.Errorf("result = %d cents, want 45000", result.Cents())
	}
	if desc == "" {
		t.Error("should have description")
	}
	t.Logf("Chain result: %s (%s)", result, desc)
}

func TestBestOfPolicies(t *testing.T) {
	order := makePolicyOrder(t, 5, 100) // $500
	total, _ := order.TotalAmount()

	best := BestOfPolicies(
		BulkOrderDiscount(3, 10),  // 10% → $450
		LoyaltyDiscount("gold"),   // 15% → $425
		BulkOrderDiscount(10, 20), // See the corresponding tests for the intended behavior.
	)

	result, desc, err := best(order, total)
	if err != nil {
		t.Fatal(err)
	}
	if result.Cents() != 42500 {
		t.Errorf("result = %d cents, want 42500", result.Cents())
	}
	t.Logf("Best result: %s (%s)", result, desc)
}

func TestBulkOrderDiscount(t *testing.T) {
	t.Run("qualifies", func(t *testing.T) {
		order := makePolicyOrder(t, 5, 100)
		total, _ := order.TotalAmount()
		result, desc, _ := BulkOrderDiscount(3, 10)(order, total)
		if result.Cents() != 45000 {
			t.Errorf("got %d", result.Cents())
		}
		if desc == "" {
			t.Error("should describe discount")
		}
	})
	t.Run("does not qualify", func(t *testing.T) {
		order := makePolicyOrder(t, 2, 100)
		total, _ := order.TotalAmount()
		result, desc, _ := BulkOrderDiscount(3, 10)(order, total)
		if result.Cents() != total.Cents() {
			t.Error("should be unchanged")
		}
		if desc != "" {
			t.Error("should have empty description")
		}
	})
}

func TestLoyaltyDiscount(t *testing.T) {
	tests := []struct {
		tier    string
		wantPct float64
	}{
		{"gold", 15}, {"silver", 10}, {"bronze", 5}, {"none", 0},
	}
	for _, tt := range tests {
		t.Run(tt.tier, func(t *testing.T) {
			price := domain.MustNewMoney(100, domain.AUD) // 10000 cents
			order := makePolicyOrder(t, 1, 100)
			result, _, _ := LoyaltyDiscount(tt.tier)(order, price)
			wantCents := int64(10000 * (1 - tt.wantPct/100))
			if result.Cents() != wantCents {
				t.Errorf("got %d, want %d", result.Cents(), wantCents)
			}
		})
	}
}

func TestMinimumPrice(t *testing.T) {
	order := makePolicyOrder(t, 1, 10)          // $10
	price := domain.MustNewMoney(5, domain.AUD) // See the corresponding tests for the intended behavior.
	min := domain.MustNewMoney(8, domain.AUD)   // See the corresponding tests for the intended behavior.

	result, desc, _ := MinimumPrice(min)(order, price)
	if result.Cents() != 800 {
		t.Errorf("got %d", result.Cents())
	}
	if desc == "" {
		t.Error("should indicate minimum applied")
	}

	result2, desc2, _ := MinimumPrice(min)(order, domain.MustNewMoney(20, domain.AUD))
	if result2.Cents() != 2000 {
		t.Error("should keep higher price")
	}
	if desc2 != "" {
		t.Error("should be empty when not applied")
	}
}

// ---- Validation Rules ----

func TestValidateAll(t *testing.T) {
	t.Run("all pass", func(t *testing.T) {
		order := makePolicyOrder(t, 3, 50) // 3 lines, $150
		errs := ValidateAll(order,
			MaxLinesRule(10),
			MaxAmountRule(domain.MustNewMoney(500, domain.AUD)),
		)
		if len(errs) != 0 {
			t.Errorf("errors = %v", errs)
		}
	})

	t.Run("multiple failures", func(t *testing.T) {
		order := makePolicyOrder(t, 15, 100) // 15 lines, $1500
		errs := ValidateAll(order,
			MaxLinesRule(10),
			MaxAmountRule(domain.MustNewMoney(500, domain.AUD)),
		)
		if len(errs) != 2 {
			t.Errorf("expected 2 errors, got %d", len(errs))
		}
	})
}
