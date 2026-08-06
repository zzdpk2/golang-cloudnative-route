package domain

import "testing"

func TestAnemicModelCanBypassBusinessRules(t *testing.T) {
	order := &AnemicOrder{
		ID:     "order-1",
		Status: AnemicOrderPending,
	}

	service := AnemicOrderService{}

	line := AnemicOrderLine{
		ProductID: "product-1",
		Name:      "Keyboard",
		UnitPrice: 100,
		Quantity:  2,
	}
	if err := service.AddLine(order, line); err != nil {
		t.Fatalf("AddLine() error = %v", err)
	}

	if err := service.ApplyDiscount(order, 50); err != nil {
		t.Fatalf("ApplyDiscount() error = %v", err)
	}

	if got, want := service.PayableAmount(order), 150; got != want {
		t.Fatalf("PayableAmount() = %d, want %d", got, want)
	}

	order.Discount = 10_000
	order.Status = AnemicOrderConfirmed

	if got := service.PayableAmount(order); got >= 0 {
		t.Fatalf("anemic model should allow invalid payable amount after direct mutation, got %d", got)
	}
}

func TestAnemicModelServiceProtectsRulesOnlyWhenUsed(t *testing.T) {
	order := &AnemicOrder{
		ID:     "order-1",
		Status: AnemicOrderPending,
	}

	service := AnemicOrderService{}

	if err := service.Confirm(order); err == nil {
		t.Fatal("Confirm() expected error for empty order")
	}

	if err := service.ApplyDiscount(order, -1); err == nil {
		t.Fatal("ApplyDiscount() expected error for negative discount")
	}
}

func TestRichModelProtectsInvariants(t *testing.T) {
	order, err := NewRichOrder("order-1")
	if err != nil {
		t.Fatalf("NewRichOrder() error = %v", err)
	}

	line, err := NewRichOrderLine("product-1", "Keyboard", 100, 2)
	if err != nil {
		t.Fatalf("NewRichOrderLine() error = %v", err)
	}

	if err := order.AddLine(line); err != nil {
		t.Fatalf("AddLine() error = %v", err)
	}

	if err := order.ApplyDiscount(50); err != nil {
		t.Fatalf("ApplyDiscount() error = %v", err)
	}

	if got, want := order.Total(), 200; got != want {
		t.Fatalf("Total() = %d, want %d", got, want)
	}
	if got, want := order.PayableAmount(), 150; got != want {
		t.Fatalf("PayableAmount() = %d, want %d", got, want)
	}

	if err := order.ApplyDiscount(10_000); err == nil {
		t.Fatal("ApplyDiscount() expected error when discount exceeds total")
	}

	if err := order.Confirm(); err != nil {
		t.Fatalf("Confirm() error = %v", err)
	}

	if order.Status() != RichOrderConfirmed {
		t.Fatalf("Status() = %s, want %s", order.Status(), RichOrderConfirmed)
	}

	if err := order.AddLine(line); err == nil {
		t.Fatal("AddLine() expected error after order confirmed")
	}
}

func TestRichModelReturnsDefensiveCopy(t *testing.T) {
	order, err := NewRichOrder("order-1")
	if err != nil {
		t.Fatalf("NewRichOrder() error = %v", err)
	}

	line, err := NewRichOrderLine("product-1", "Keyboard", 100, 2)
	if err != nil {
		t.Fatalf("NewRichOrderLine() error = %v", err)
	}

	if err := order.AddLine(line); err != nil {
		t.Fatalf("AddLine() error = %v", err)
	}

	lines := order.Lines()
	lines[0], err = NewRichOrderLine("product-2", "Mouse", 1, 1)
	if err != nil {
		t.Fatalf("NewRichOrderLine() error = %v", err)
	}

	if got, want := order.Lines()[0].ProductID(), "product-1"; got != want {
		t.Fatalf("defensive copy failed: ProductID() = %s, want %s", got, want)
	}
}

func TestRichModelRejectsInvalidLine(t *testing.T) {
	if _, err := NewRichOrderLine("", "Keyboard", 100, 1); err == nil {
		t.Fatal("NewRichOrderLine() expected error for empty product id")
	}

	if _, err := NewRichOrderLine("product-1", "Keyboard", 0, 1); err == nil {
		t.Fatal("NewRichOrderLine() expected error for non-positive price")
	}

	if _, err := NewRichOrderLine("product-1", "Keyboard", 100, 0); err == nil {
		t.Fatal("NewRichOrderLine() expected error for non-positive quantity")
	}
}
