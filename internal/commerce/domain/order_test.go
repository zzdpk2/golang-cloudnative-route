package domain

import (
	"fmt"
	"sync"
	"testing"
)

func makeTestOrder(t *testing.T) *Order {
	t.Helper()
	addr, _ := NewAddress("1 Main St", "Sydney", "NSW", "2000", "AU")
	o, err := NewOrder("ord-001", "cust-001", addr)
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func addTestLine(t *testing.T, o *Order, pid string, name string, price float64, qty int) {
	t.Helper()
	p := MustNewMoney(price, AUD)
	q, _ := NewQuantity(qty)
	err := o.AddLine(ProductID(pid), name, p, q)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewOrder(t *testing.T) {
	o := makeTestOrder(t)
	if o.Status() != OrderPending {
		t.Error("new order should be pending")
	}
	if o.CreatedAt().IsZero() {
		t.Error("createdAt should be set")
	}

	events := o.CollectEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventName() != "order.created" {
		t.Errorf("event = %q", events[0].EventName())
	}
}

func TestOrder_AddLine(t *testing.T) {
	o := makeTestOrder(t)
	addTestLine(t, o, "p1", "Widget", 9.99, 2)

	if o.LineCount() != 1 {
		t.Errorf("LineCount = %d", o.LineCount())
	}

	lines := o.Lines()
	if lines[0].Name != "Widget" {
		t.Errorf("Name = %q", lines[0].Name)
	}
}

func TestOrder_AddLine_DuplicateProduct(t *testing.T) {
	o := makeTestOrder(t)
	addTestLine(t, o, "p1", "Widget", 9.99, 2)

	p := MustNewMoney(9.99, AUD)
	q, _ := NewQuantity(1)
	err := o.AddLine("p1", "Widget", p, q)
	if err == nil {
		t.Error("should reject duplicate product")
	}
}

func TestOrder_AddLine_NotPending(t *testing.T) {
	o := makeTestOrder(t)
	addTestLine(t, o, "p1", "Widget", 9.99, 1)
	o.Confirm()

	p := MustNewMoney(5, AUD)
	q, _ := NewQuantity(1)
	err := o.AddLine("p2", "Other", p, q)
	if err == nil {
		t.Error("should not add line to confirmed order")
	}
}

func TestOrder_RemoveLine(t *testing.T) {
	o := makeTestOrder(t)
	addTestLine(t, o, "p1", "A", 10, 1)
	addTestLine(t, o, "p2", "B", 20, 1)

	err := o.RemoveLine("p1")
	if err != nil {
		t.Fatal(err)
	}

	if o.LineCount() != 1 {
		t.Errorf("LineCount = %d", o.LineCount())
	}
	if !o.HasProduct("p2") {
		t.Error("should still have p2")
	}
}

func TestOrder_RemoveLine_NotFound(t *testing.T) {
	o := makeTestOrder(t)
	err := o.RemoveLine("nonexistent")
	if err == nil {
		t.Error("should error for non-existent product")
	}
}

func TestOrder_TotalAmount(t *testing.T) {
	o := makeTestOrder(t)
	addTestLine(t, o, "p1", "A", 10, 2)   // 20
	addTestLine(t, o, "p2", "B", 5.50, 3) // 16.50

	total, err := o.TotalAmount()
	if err != nil {
		t.Fatal(err)
	}
	// 2000 + 1650 = 3650 cents
	if total.Cents() != 3650 {
		t.Errorf("total = %d cents, want 3650", total.Cents())
	}
}

func TestOrder_TotalAmount_Empty(t *testing.T) {
	o := makeTestOrder(t)
	_, err := o.TotalAmount()
	if err == nil {
		t.Error("empty order should return error")
	}
}

func TestOrder_StatusTransitions(t *testing.T) {
	t.Run("full lifecycle", func(t *testing.T) {
		o := makeTestOrder(t)
		addTestLine(t, o, "p1", "A", 10, 1)

		if err := o.Confirm(); err != nil {
			t.Fatal(err)
		}
		if o.Status() != OrderConfirmed {
			t.Error("should be confirmed")
		}

		if err := o.Ship(); err != nil {
			t.Fatal(err)
		}
		if o.Status() != OrderShipped {
			t.Error("should be shipped")
		}

		if err := o.Deliver(); err != nil {
			t.Fatal(err)
		}
		if o.Status() != OrderDelivered {
			t.Error("should be delivered")
		}
	})

	t.Run("cannot confirm empty order", func(t *testing.T) {
		o := makeTestOrder(t)
		err := o.Confirm()
		if err == nil {
			t.Error("should not confirm empty order")
		}
	})

	t.Run("cannot ship pending", func(t *testing.T) {
		o := makeTestOrder(t)
		err := o.Ship()
		if err == nil {
			t.Error("should not ship pending order")
		}
	})

	t.Run("cancel pending", func(t *testing.T) {
		o := makeTestOrder(t)
		err := o.Cancel()
		if err == nil && o.Status() != OrderCancelled {
			t.Error("should cancel pending order")
		}
	})

	t.Run("cancel confirmed", func(t *testing.T) {
		o := makeTestOrder(t)
		addTestLine(t, o, "p1", "A", 10, 1)
		o.Confirm()
		err := o.Cancel()
		if err == nil && o.Status() != OrderCancelled {
			t.Error("should cancel confirmed order")
		}
	})

	t.Run("cannot cancel delivered", func(t *testing.T) {
		o := makeTestOrder(t)
		addTestLine(t, o, "p1", "A", 10, 1)
		o.Confirm()
		o.Ship()
		o.Deliver()
		err := o.Cancel()
		if err == nil {
			t.Error("should not cancel delivered order")
		}
	})
}

func TestOrder_CollectEvents(t *testing.T) {
	o := makeTestOrder(t)
	addTestLine(t, o, "p1", "A", 10, 1)
	o.Confirm()

	events := o.CollectEvents()
	if len(events) < 2 {
		t.Errorf("expected at least 2 events, got %d", len(events))
	}

	events2 := o.CollectEvents()
	if len(events2) != 0 {
		t.Errorf("events should be cleared, got %d", len(events2))
	}
}

func TestOrder_SortLinesByPrice(t *testing.T) {
	o := makeTestOrder(t)
	addTestLine(t, o, "p1", "Expensive", 100, 1)
	addTestLine(t, o, "p2", "Cheap", 5, 1)
	addTestLine(t, o, "p3", "Mid", 50, 1)

	o.SortLinesByPrice()
	lines := o.Lines()
	if lines[0].Name != "Cheap" || lines[1].Name != "Mid" || lines[2].Name != "Expensive" {
		t.Errorf("sort order wrong: %v, %v, %v", lines[0].Name, lines[1].Name, lines[2].Name)
	}
}

func TestOrder_SortLinesByName(t *testing.T) {
	o := makeTestOrder(t)
	addTestLine(t, o, "p1", "Cherry", 10, 1)
	addTestLine(t, o, "p2", "Apple", 20, 1)
	addTestLine(t, o, "p3", "Banana", 15, 1)

	o.SortLinesByName()
	lines := o.Lines()
	if lines[0].Name != "Apple" || lines[1].Name != "Banana" || lines[2].Name != "Cherry" {
		t.Errorf("sort order wrong: %v, %v, %v", lines[0].Name, lines[1].Name, lines[2].Name)
	}
}

func TestOrder_LineProcessors(t *testing.T) {
	o := makeTestOrder(t)
	addTestLine(t, o, "p1", "A", 10, 1) // 1000 cents
	addTestLine(t, o, "p2", "B", 20, 2) // 4000 cents
	addTestLine(t, o, "p3", "C", 30, 3) // 9000 cents

	processors := o.LineProcessors()
	if len(processors) != 3 {
		t.Fatalf("expected 3 processors, got %d", len(processors))
	}

	expected := []int64{1000, 4000, 9000}
	for i, fn := range processors {
		got := fn().Cents()
		if got != expected[i] {
			t.Errorf("processor[%d] = %d cents, want %d (loop variable capture error)", i, got, expected[i])
		}
	}
}

func TestOrder_FilterLines(t *testing.T) {
	o := makeTestOrder(t)
	addTestLine(t, o, "p1", "A", 10, 1)
	addTestLine(t, o, "p2", "B", 50, 1)
	addTestLine(t, o, "p3", "C", 100, 1)

	expensive := o.FilterLines(func(l OrderLine) bool {
		return l.Price.Cents() > 2000
	})

	if len(expensive) != 2 {
		t.Errorf("expected 2 expensive items, got %d", len(expensive))
	}
}

func TestOrder_ConcurrentAccess(t *testing.T) {
	o := makeTestOrder(t)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			pid := ProductID(fmt.Sprintf("p-%d", idx))
			p := MustNewMoney(float64(idx+1), AUD)
			q, _ := NewQuantity(1)
			_ = o.AddLine(pid, fmt.Sprintf("Product %d", idx), p, q)
		}(i)
	}
	wg.Wait()

	if o.LineCount() != 10 {
		t.Errorf("LineCount = %d, want 10 (concurrent safety issue!)", o.LineCount())
	}
}

func BenchmarkOrder_AddLine(b *testing.B) {
	addr, _ := NewAddress("1 Main St", "Sydney", "NSW", "2000", "AU")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		o, _ := NewOrder("bench", "cust", addr)
		for j := 0; j < 100; j++ {
			p := MustNewMoney(float64(j+1), AUD)
			q, _ := NewQuantity(1)
			o.AddLine(ProductID(fmt.Sprintf("p%d", j)), "item", p, q)
		}
	}
}
