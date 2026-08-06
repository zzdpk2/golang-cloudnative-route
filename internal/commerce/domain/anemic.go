package domain

// ============================================================
// Anemic Model
//
// Data-only structs whose rules live in the service layer.
// ============================================================

type AnemicOrderStatus string

const (
	AnemicOrderPending   AnemicOrderStatus = "pending"
	AnemicOrderConfirmed AnemicOrderStatus = "confirmed"
	AnemicOrderCancelled AnemicOrderStatus = "cancelled"
)

type AnemicOrderLine struct {
	ProductID string
	Name      string
	UnitPrice int
	Quantity  int
}

// AnemicOrder exposes every field. Nothing stops a caller from setting Status
// or appending to Lines directly, so the invariants are only as good as the
// discipline of every call site.
type AnemicOrder struct {
	ID       string
	Status   AnemicOrderStatus
	Lines    []AnemicOrderLine
	Discount int
}

// AnemicOrderService owns the rules that AnemicOrder does not protect itself.
// Compare each method here with the equivalent method on RichOrder.
type AnemicOrderService struct{}

func (s AnemicOrderService) AddLine(order *AnemicOrder, line AnemicOrderLine) error {
	panic("TODO")
}

func (s AnemicOrderService) ApplyDiscount(order *AnemicOrder, discount int) error {
	panic("TODO")
}

func (s AnemicOrderService) Confirm(order *AnemicOrder) error {
	panic("TODO")
}

func (s AnemicOrderService) Total(order *AnemicOrder) int {
	panic("TODO")
}

func (s AnemicOrderService) PayableAmount(order *AnemicOrder) int {
	panic("TODO")
}
