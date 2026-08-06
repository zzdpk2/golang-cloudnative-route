package domain

// ============================================================
// Rich Model
//
// Behavior kept next to the state it protects.
// ============================================================

type RichOrderStatus string

const (
	RichOrderPending   RichOrderStatus = "pending"
	RichOrderConfirmed RichOrderStatus = "confirmed"
	RichOrderCancelled RichOrderStatus = "cancelled"
)

type RichOrderLine struct {
	productID string
	name      string
	unitPrice int
	quantity  int
}

func NewRichOrderLine(productID, name string, unitPrice, quantity int) (RichOrderLine, error) {
	panic("TODO")
}

func (l RichOrderLine) ProductID() string { panic("TODO") }
func (l RichOrderLine) Name() string      { panic("TODO") }
func (l RichOrderLine) UnitPrice() int    { panic("TODO") }
func (l RichOrderLine) Quantity() int     { panic("TODO") }
func (l RichOrderLine) LineTotal() int    { panic("TODO") }

type RichOrder struct {
	id       string
	status   RichOrderStatus
	lines    []RichOrderLine
	discount int
}

func NewRichOrder(id string) (*RichOrder, error) {
	panic("TODO")
}

func (o *RichOrder) ID() string              { panic("TODO") }
func (o *RichOrder) Status() RichOrderStatus { panic("TODO") }
func (o *RichOrder) Discount() int           { panic("TODO") }

func (o *RichOrder) Lines() []RichOrderLine {
	panic("TODO")
}

func (o *RichOrder) AddLine(line RichOrderLine) error {
	panic("TODO")
}

func (o *RichOrder) ApplyDiscount(discount int) error {
	panic("TODO")
}

func (o *RichOrder) Confirm() error {
	panic("TODO")
}

func (o *RichOrder) Cancel() error {
	panic("TODO")
}

func (o *RichOrder) Total() int {
	panic("TODO")
}

func (o *RichOrder) PayableAmount() int {
	panic("TODO")
}
