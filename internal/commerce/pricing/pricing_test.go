package pricing

import (
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"testing"
)

func makeTestOrder(t *testing.T, totalPrice float64) *domain.Order {
	t.Helper()
	addr, _ := domain.NewAddress("1 St", "Syd", "NSW", "2000", "AU")
	o, _ := domain.NewOrder("o1", "c1", addr)
	p := domain.MustNewMoney(totalPrice, domain.AUD)
	q, _ := domain.NewQuantity(1)
	o.AddLine(domain.ProductID("p1"), "item", p, q)
	return o
}

func TestPricingService_NoDiscount(t *testing.T) {
	svc := NewPricingService() // See the corresponding tests for the intended behavior.
	o := makeTestOrder(t, 100)

	discount, name, err := svc.BestDiscount(o)
	if err != nil {
		t.Fatal(err)
	}
	if !discount.IsZero() {
		t.Errorf("discount = %s, want 0", discount)
	}
	if name != "no_discount" {
		t.Errorf("name = %q", name)
	}
}

func TestPricingService_PercentageOff(t *testing.T) {
	svc := NewPricingService(
		WithStrategy(&PercentageOff{Percent: 10}),
	)
	o := makeTestOrder(t, 100) // 10000 cents

	discount, _, err := svc.BestDiscount(o)
	if err != nil {
		t.Fatal(err)
	}
	// 10% of 10000 = 1000 cents
	if discount.Cents() != 1000 {
		t.Errorf("discount = %d cents, want 1000", discount.Cents())
	}
}

func TestPricingService_BulkDiscount(t *testing.T) {
	svc := NewPricingService(
		WithStrategy(&BulkDiscount{
			MinAmount: domain.MustNewMoney(80, domain.AUD),
			Discount:  domain.MustNewMoney(15, domain.AUD),
		}),
	)

	t.Run("qualifies", func(t *testing.T) {
		o := makeTestOrder(t, 100)
		discount, _, err := svc.BestDiscount(o)
		if err != nil {
			t.Fatal(err)
		}
		if discount.Cents() != 1500 {
			t.Errorf("discount = %d cents, want 1500", discount.Cents())
		}
	})

	t.Run("does not qualify", func(t *testing.T) {
		o := makeTestOrder(t, 50)
		discount, _, err := svc.BestDiscount(o)
		if err != nil {
			t.Fatal(err)
		}
		if !discount.IsZero() {
			t.Errorf("discount = %s, should be zero", discount)
		}
	})
}

func TestPricingService_BestOfMultiple(t *testing.T) {
	svc := NewPricingService(
		WithStrategy(&PercentageOff{Percent: 10}), // 10% of 200 = 20
		WithStrategy(&BulkDiscount{ // See the corresponding tests for the intended behavior.
			MinAmount: domain.MustNewMoney(100, domain.AUD),
			Discount:  domain.MustNewMoney(15, domain.AUD),
		}),
	)

	o := makeTestOrder(t, 200)
	discount, name, err := svc.BestDiscount(o)
	if err != nil {
		t.Fatal(err)
	}

	if discount.Cents() != 2000 {
		t.Errorf("discount = %d cents, want 2000", discount.Cents())
	}
	if name != "10%_off" {
		t.Errorf("strategy = %q, want '10%%_off'", name)
	}
}

func TestPricingService_FinalPrice(t *testing.T) {
	svc := NewPricingService(
		WithStrategy(&PercentageOff{Percent: 20}),
	)
	o := makeTestOrder(t, 100) // 10000 cents, 20% off = 2000 discount

	final, err := svc.FinalPrice(o)
	if err != nil {
		t.Fatal(err)
	}
	// 10000 - 2000 = 8000
	if final.Cents() != 8000 {
		t.Errorf("final = %d cents, want 8000", final.Cents())
	}
}

func TestPricingService_FunctionalOptions(t *testing.T) {
	strategies := []PricingOption{
		WithStrategy(&PercentageOff{5}),
		WithStrategy(&PercentageOff{10}),
		WithStrategy(&PercentageOff{15}),
	}

	svc := NewPricingService(strategies...)
	o := makeTestOrder(t, 100)

	discount, name, err := svc.BestDiscount(o)
	if err != nil {
		t.Fatal(err)
	}
	// 15% is best
	if discount.Cents() != 1500 {
		t.Errorf("discount = %d", discount.Cents())
	}
	if name != "15%_off" {
		t.Errorf("name = %q", name)
	}
}
