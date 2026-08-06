package domain

import (
	"sync"
	"time"
)

// ============================================================
// Order Aggregate Root
//
// Protect invariants, record domain events, and hand out defensive copies.
// ============================================================

// ---- Order Status ----

type OrderStatus int

const (
	OrderPending OrderStatus = iota
	OrderConfirmed
	OrderShipped
	OrderDelivered
	OrderCancelled
)

func (s OrderStatus) String() string { panic("TODO") }

// ---- Order Line ----

type OrderLine struct {
	ProductID ProductID
	Name      string
	Price     Money
	Quantity  Quantity
}

// LineTotal is the price of this line: the unit price scaled by the quantity.
// Money already knows how to scale itself; do not reach for the raw amount.
func (l OrderLine) LineTotal() Money {
	panic("TODO")
}

// ---- Order Aggregate Root ----

type OrderID string

type Order struct {
	mu sync.Mutex // See the corresponding tests for the intended behavior.

	id         OrderID
	customerID CustomerID
	lines      []OrderLine
	status     OrderStatus
	shipping   Address
	createdAt  time.Time
	updatedAt  time.Time

	events []DomainEvent
}

// NewOrder is the only way to obtain a valid Order. It rejects an empty id or
// customer id, starts the order in OrderPending, stamps the creation time, and
// records the first domain
//
// Every field of Order is unexported: whatever this constructor fails to set
// can never be repaired from outside the package.
func NewOrder(id OrderID, customerID CustomerID, shipping Address) (*Order, error) {
	panic("TODO")
}

func (o *Order) ID() OrderID              { panic("TODO") }
func (o *Order) CustomerID() CustomerID   { panic("TODO") }
func (o *Order) Status() OrderStatus      { panic("TODO") }
func (o *Order) ShippingAddress() Address { panic("TODO") }
func (o *Order) CreatedAt() time.Time     { panic("TODO") }

// Lines exposes the order lines to callers outside the
//
// Returning the internal slice would let any caller append to it or rewrite an
// element without going through AddLine, which defeats every invariant below.
func (o *Order) Lines() []OrderLine {
	panic("TODO")
}

// AddLine appends a line and fails when the order has left OrderPending, when
// the product is already on the order, or when the line would mix currencies
// with the lines already present.
//
// It also records a domain event and advances updatedAt.
func (o *Order) AddLine(productID ProductID, name string, price Money, qty Quantity) error {
	panic("TODO")
}

// RemoveLine drops the line for productID. Removing a product that is not on
// the order is an error, not a silent no-op, and only a pending order may be
// edited.
func (o *Order) RemoveLine(productID ProductID) error {
	panic("TODO")
}

// TotalAmount sums every line total.
//
// Money addition can fail on mismatched currencies, so decide what an order
// with no lines should return before you write the loop.
func (o *Order) TotalAmount() (Money, error) {
	panic("TODO")
}

// Confirm moves a pending order to OrderConfirmed. An order with no lines has
// nothing to confirm. Any other starting status is a rejected transition.
//
// The four transition methods below share one shape: check the current status,
// apply the new one, stamp updatedAt, record an  Notice how much of the
// state machine is expressible as "which status may precede which".
func (o *Order) Confirm() error {
	panic("TODO")
}

// Ship moves a confirmed order to OrderShipped.
func (o *Order) Ship() error {
	panic("TODO")
}

// Deliver moves a shipped order to OrderDelivered.
func (o *Order) Deliver() error {
	panic("TODO")
}

// Cancel moves an order to OrderCancelled. Decide which statuses are still
// cancellable: an order already delivered or already cancelled is not.
func (o *Order) Cancel() error {
	panic("TODO")
}

// ---- Domain Events ----

// CollectEvents hands the recorded events to the caller and clears the
// aggregate's buffer, so that publishing twice cannot deliver the same event
// twice. The tests call it more than once; the second call must come back empty.
func (o *Order) CollectEvents() []DomainEvent {
	panic("TODO")
}

func (o *Order) addEvent(e DomainEvent) { panic("TODO") }

// SortLinesByPrice orders the lines by unit price, cheapest first. Sorting in
// place is fine here: the caller never sees the internal slice.
func (o *Order) SortLinesByPrice() {
	panic("TODO")
}

// SortLinesByName orders the lines by product name.
func (o *Order) SortLinesByName() {
	panic("TODO")
}

// LineProcessors returns one deferred calculation per line, each of which
// reports that line's total when invoked.
//
// The tests call the returned functions after this method has returned, so the
// question is what each closure actually captured. Write it the way that reads
// most naturally first, run the test, and only then reason about why the result
// is or is not what you expected. What Go version the module targets matters
// here — check the go directive in go.mod before concluding.
func (o *Order) LineProcessors() []func() Money {
	panic("TODO")
}

// FilterLines returns the lines satisfying predicate.
//
// Decide what to return when nothing matches, and whether the caller may
// safely mutate what comes back. Compare your answer with Lines above.
func (o *Order) FilterLines(predicate func(OrderLine) bool) []OrderLine {
	panic("TODO")
}

func (o *Order) LineCount() int { panic("TODO") }

// HasProduct reports whether productID is already on the order.
func (o *Order) HasProduct(productID ProductID) bool {
	panic("TODO")
}

// ---- Stringer ----

func (o *Order) String() string { panic("TODO") }
