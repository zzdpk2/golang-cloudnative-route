package pricing

import (
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"testing"
)

func makeSpecificationOrder(t *testing.T, id string, items ...struct {
	pid   string
	price float64
	qty   int
}) *domain.Order {
	t.Helper()
	addr, _ := domain.NewAddress("1 St", "Syd", "NSW", "2000", "AU")
	o, err := domain.NewOrder(domain.OrderID(id), "cust-1", addr)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range items {
		p := domain.MustNewMoney(item.price, domain.AUD)
		q, _ := domain.NewQuantity(item.qty)
		o.AddLine(domain.ProductID(item.pid), item.pid, p, q)
	}
	return o
}

type item struct {
	pid   string
	price float64
	qty   int
}

func TestOrderMinAmountSpec(t *testing.T) {
	o := makeSpecificationOrder(t, "o1", item{"p1", 50, 2}) // total = 100
	min80 := NewOrderMinAmountSpec(domain.MustNewMoney(80, domain.AUD))
	min120 := NewOrderMinAmountSpec(domain.MustNewMoney(120, domain.AUD))

	if !min80.IsSatisfiedBy(o) {
		t.Error("order 100 should satisfy min 80")
	}
	if min120.IsSatisfiedBy(o) {
		t.Error("order 100 should not satisfy min 120")
	}
}

func TestOrderStatusSpec(t *testing.T) {
	o := makeSpecificationOrder(t, "o1", item{"p1", 10, 1})
	pendingSpec := NewOrderStatusSpec(domain.OrderPending)
	confirmedSpec := NewOrderStatusSpec(domain.OrderConfirmed)

	if !pendingSpec.IsSatisfiedBy(o) {
		t.Error("new order should be pending")
	}
	if confirmedSpec.IsSatisfiedBy(o) {
		t.Error("new order should not be confirmed")
	}

	o.Confirm()
	if pendingSpec.IsSatisfiedBy(o) {
		t.Error("confirmed order should not match pending")
	}
	if !confirmedSpec.IsSatisfiedBy(o) {
		t.Error("confirmed order should match confirmed")
	}
}

func TestAndSpec(t *testing.T) {
	o := makeSpecificationOrder(t, "o1", item{"p1", 50, 2}) // total=100, pending
	spec := And[*domain.Order](
		NewOrderMinAmountSpec(domain.MustNewMoney(80, domain.AUD)),
		NewOrderStatusSpec(domain.OrderPending),
	)
	if !spec.IsSatisfiedBy(o) {
		t.Error("should satisfy both: min 80 and pending")
	}

	o.Confirm()
	if spec.IsSatisfiedBy(o) {
		t.Error("should not satisfy: confirmed != pending")
	}
}

func TestOrSpec(t *testing.T) {
	o := makeSpecificationOrder(t, "o1", item{"p1", 10, 1}) // total=10
	spec := Or[*domain.Order](
		NewOrderMinAmountSpec(domain.MustNewMoney(100, domain.AUD)), // false
		NewOrderStatusSpec(domain.OrderPending),                     // true
	)
	if !spec.IsSatisfiedBy(o) {
		t.Error("should satisfy: at least one is true")
	}
}

func TestNotSpec(t *testing.T) {
	o := makeSpecificationOrder(t, "o1", item{"p1", 10, 1})
	notConfirmed := Not[*domain.Order](NewOrderStatusSpec(domain.OrderConfirmed))
	if !notConfirmed.IsSatisfiedBy(o) {
		t.Error("pending order should satisfy NOT confirmed")
	}

	o.Confirm()
	if notConfirmed.IsSatisfiedBy(o) {
		t.Error("confirmed order should not satisfy NOT confirmed")
	}
}

func TestSpecFunc(t *testing.T) {
	hasMultipleLines := SpecFunc[*domain.Order](func(o *domain.Order) bool {
		return o.LineCount() > 1
	})

	o := makeSpecificationOrder(t, "o1", item{"p1", 10, 1})
	if hasMultipleLines.IsSatisfiedBy(o) {
		t.Error("single line should not match")
	}

	p := domain.MustNewMoney(20, domain.AUD)
	q, _ := domain.NewQuantity(1)
	o.AddLine("p2", "p2", p, q)

	if !hasMultipleLines.IsSatisfiedBy(o) {
		t.Error("two lines should match")
	}
}

func TestFilter(t *testing.T) {
	orders := []*domain.Order{
		makeSpecificationOrder(t, "o1", item{"p1", 10, 1}),  // 10
		makeSpecificationOrder(t, "o2", item{"p1", 50, 2}),  // 100
		makeSpecificationOrder(t, "o3", item{"p1", 200, 1}), // 200
	}

	min50 := NewOrderMinAmountSpec(domain.MustNewMoney(50, domain.AUD))
	filtered := Filter(orders, min50)
	if len(filtered) != 2 {
		t.Errorf("expected 2 orders >= 50, got %d", len(filtered))
	}
}

func TestCount(t *testing.T) {
	orders := []*domain.Order{
		makeSpecificationOrder(t, "o1", item{"p1", 10, 1}),
		makeSpecificationOrder(t, "o2", item{"p1", 50, 1}),
		makeSpecificationOrder(t, "o3", item{"p1", 200, 1}),
	}

	min50 := NewOrderMinAmountSpec(domain.MustNewMoney(50, domain.AUD))
	if Count(orders, min50) != 2 {
		t.Error("count mismatch")
	}
}

func TestAny(t *testing.T) {
	orders := []*domain.Order{
		makeSpecificationOrder(t, "o1", item{"p1", 10, 1}),
		makeSpecificationOrder(t, "o2", item{"p1", 20, 1}),
	}

	min100 := NewOrderMinAmountSpec(domain.MustNewMoney(100, domain.AUD))
	if Any(orders, min100) {
		t.Error("none should be >= 100")
	}

	min5 := NewOrderMinAmountSpec(domain.MustNewMoney(5, domain.AUD))
	if !Any(orders, min5) {
		t.Error("some should be >= 5")
	}
}

func TestAll(t *testing.T) {
	orders := []*domain.Order{
		makeSpecificationOrder(t, "o1", item{"p1", 10, 1}),
		makeSpecificationOrder(t, "o2", item{"p1", 20, 1}),
	}

	min5 := NewOrderMinAmountSpec(domain.MustNewMoney(5, domain.AUD))
	if !All(orders, min5) {
		t.Error("all should be >= 5")
	}

	min15 := NewOrderMinAmountSpec(domain.MustNewMoney(15, domain.AUD))
	if All(orders, min15) {
		t.Error("not all are >= 15")
	}
}

func TestFilter_SimpleTypes(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	even := SpecFunc[int](func(n int) bool { return n%2 == 0 })

	result := Filter(numbers, even)
	if len(result) != 5 {
		t.Errorf("expected 5 even numbers, got %d", len(result))
	}
}

func TestComplexSpecComposition(t *testing.T) {
	orders := []*domain.Order{
		makeSpecificationOrder(t, "o1", item{"widget", 30, 2}),  // 60, pending
		makeSpecificationOrder(t, "o2", item{"widget", 30, 2}),  // 60, will confirm
		makeSpecificationOrder(t, "o3", item{"gadget", 100, 1}), // 100, will confirm but wrong product
	}

	orders[1].Confirm()
	orders[2].Confirm()

	shippable := And[*domain.Order](
		NewOrderStatusSpec(domain.OrderConfirmed),
		NewOrderMinAmountSpec(domain.MustNewMoney(50, domain.AUD)),
		NewOrderHasProductSpec("widget"),
	)

	result := Filter(orders, shippable)
	if len(result) != 1 {
		t.Errorf("expected 1 shippable order, got %d", len(result))
	}
	if result[0].ID() != "o2" {
		t.Errorf("expected o2, got %s", result[0].ID())
	}
}
