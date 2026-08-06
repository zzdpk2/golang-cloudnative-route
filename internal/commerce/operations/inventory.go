// Package inventory is L15.2: reservation, expiry, and oversell.
//
// A ● exercise. The tests state what the business needs; the model is yours.
//
// # The business
//
// Placing an order does not take stock away — it *holds* it. The customer has
// thirty minutes to pay, and if they do not, the hold expires and the stock
// goes back on sale. Only payment turns a hold into a real deduction.
//
// So there are three numbers per SKU per warehouse, and confusing any two of
// them is how a shop oversells:
//
//	OnHand     physically in the building
//	Reserved   held for orders that have not paid yet
//	Available  OnHand − Reserved, the only number a customer may be shown
//
// # Why this is hard
//
//   - **Expiry has no natural trigger.** Nothing happens at minute thirty-one.
//     Something has to go looking, and whatever that something is, it can be
//     late, can run twice, or can be down.
//   - **A restart loses nothing and forgets everything.** The holds are still
//     in the ledger; the timer that would have expired them is not.
//   - **Concurrency is the whole point.** Two hundred people want the last
//     unit. Exactly one may get it, and the other 199 must be told no rather
//     than being told yes and disappointed later.
//   - **Multi-warehouse turns "is there stock?" into "where from?"**, and a
//     naive answer splits an order that could have shipped in one box.
//
// # Answer these before writing code
//
//  1. Is a Reservation an entity or a value object? It has an id and a
//     lifecycle — but does anything about it ever *change*, or does it only
//     ever end?
//  2. Where does expiry live? A field on the reservation, a sweeper, a
//     priority queue, a delayed message? Name the failure mode of each.
//  3. Committing an expired reservation: error, or silently re-reserve? Both
//     ship in real systems. Which would you defend to a customer?
//  4. Is Available a stored number or a computed one? Storing it is faster and
//     is a second source of truth. Computing it is slower and cannot drift.
//  5. If the process dies between Reserve and Commit, what is true when it
//     comes back? Answer for both a persisted ledger and this in-memory one.
//
// Those five are the exercise. The code is the consequence.
package operations

import (
	"errors"
	"sync"
	"time"
)

var (
	// ErrInsufficientStock means the available quantity cannot cover the request.
	ErrInsufficientStock = errors.New("insufficient stock")
	// ErrUnknownSKU means the warehouse does not carry the SKU at all.
	ErrUnknownSKU = errors.New("unknown sku")
	// ErrUnknownReservation means the id does not name a live reservation.
	ErrUnknownReservation = errors.New("unknown reservation")
	// ErrReservationExpired means the hold ran out before it was committed.
	ErrReservationExpired = errors.New("reservation expired")
	// ErrNonPositiveQuantity means the caller asked for zero or fewer units.
	ErrNonPositiveQuantity = errors.New("quantity must be positive")
)

type (
	SKU         string
	WarehouseID string
)

// Stock is one SKU in one warehouse.
//
//	OnHand 10, Reserved 3 → Available 7
//
// Available is derived, not stored — see question 4 in the package comment. If
// you decide otherwise, be able to say what keeps the two in step.
type Stock struct {
	OnHand   int
	Reserved int
}

// Available is what may still be sold.
func (s Stock) Available() int {
	panic("TODO")
}

// Reservation is a hold on stock.
type Reservation struct {
	ID        string
	OrderID   string
	Warehouse WarehouseID
	SKU       SKU
	Quantity  int
	ExpiresAt time.Time
}

// Ledger is the stock book for every warehouse.
//
// The clock is injected because every expiry test would otherwise have to
// actually wait. Same reasoning as promotion.Cart.At and MySQLReconciler.now.
type Ledger struct {
	mu sync.Mutex
	// Yours.
	now func() time.Time
}

// NewLedger returns an empty ledger reading time from now.
//
//	NewLedger(time.Now)                    → production
//	NewLedger(func() time.Time { return t }) → a test with a frozen clock
func NewLedger(now func() time.Time) *Ledger {
	panic("TODO")
}

// Receive adds stock to a warehouse — a delivery arriving.
//
//	Receive("syd", "sku-1", 10) → OnHand 10, Available 10
//	Receive("syd", "sku-1", 5)  → OnHand 15, Available 15
//	Receive("syd", "sku-1", 0)  → ErrNonPositiveQuantity
//
// Receiving a SKU the warehouse has never seen creates it, rather than
// erroring: a first delivery is exactly how a SKU comes to exist.
func (l *Ledger) Receive(w WarehouseID, sku SKU, qty int) error {
	panic("TODO")
}

// Stock reports the numbers for one SKU in one warehouse.
//
//	a SKU that was never received → Stock{}, ErrUnknownSKU
//
// Note that "carried but sold out" and "not carried" are different answers, the
// same distinction testkit.FakeInventory made in L5.
func (l *Ledger) Stock(w WarehouseID, sku SKU) (Stock, error) {
	panic("TODO")
}

// Reserve holds stock for an order.
//
//	10 on hand, Reserve(2, ttl) → a Reservation; Available 8, OnHand still 10
//	reserving more than Available → ErrInsufficientStock, nothing changes
//	qty <= 0                      → ErrNonPositiveQuantity
//
// **OnHand does not move.** The goods are still in the building; they are just
// spoken for. Getting this wrong is how a warehouse report stops matching the
// shelves.
//
// The id must be unique. How you generate it is your call, but note that the
// test reserves concurrently, so a counter needs the same protection the stock
// does.
func (l *Ledger) Reserve(orderID string, w WarehouseID, sku SKU, qty int, ttl time.Duration) (Reservation, error) {
	panic("TODO")
}

// Commit turns a hold into a real deduction — the customer paid.
//
//	after Reserve(2) from 10: Commit → OnHand 8, Reserved 0, Available 8
//	committing twice          → ErrUnknownReservation the second time
//	committing after expiry   → ErrReservationExpired
//
// The double-commit case matters more than it looks: a payment webhook that
// arrives twice is not unusual, and charging one order against stock twice is a
// real incident. Making the second call fail is the cheap form of idempotence —
// think about whether *succeeding* silently would be better, and what each
// choice tells the caller.
func (l *Ledger) Commit(reservationID string) error {
	panic("TODO")
}

// Release cancels a hold — the customer abandoned the order.
//
//	after Reserve(2) from 10: Release → OnHand 10, Reserved 0, Available 10
//	releasing twice                    → ErrUnknownReservation
//
// Nothing is deducted. Compare with Commit and make sure the two cannot be
// confused at a call site.
func (l *Ledger) Release(reservationID string) error {
	panic("TODO")
}

// ExpireDue releases every reservation whose deadline has passed, and reports
// how many it released.
//
//	nothing due  → 0
//	two overdue  → 2, and their stock is available again
//	run twice    → the second call returns 0
//
// **This is the answer to question 2, and it is deliberately crude.** Nothing
// calls it on your behalf: a real system runs it on a timer, or uses a delay
// queue, or checks lazily on read. Each has a different failure mode, and this
// signature forces you to pick one at the call site instead of hiding it.
//
// It must be idempotent and safe to run concurrently with Reserve. A sweeper
// that races the thing it sweeps is worse than no sweeper.
func (l *Ledger) ExpireDue() int {
	panic("TODO")
}

// Allocate picks warehouses to fill a quantity, preferring those earlier in
// the list.
//
//	syd 3, mel 10, want 2  → [{syd 2}]              — one warehouse suffices
//	syd 3, mel 10, want 5  → [{syd 3} {mel 2}]      — split, in order
//	syd 3, mel 10, want 20 → ErrInsufficientStock, nothing reserved
//
// Two things to sit with.
//
// **Preferring the first warehouse is a placeholder for a real rule** — nearest
// to the customer, cheapest to ship, most likely to have the rest of the order.
// The signature takes an ordered list so the *caller* owns that policy, which
// is a deliberate boundary: the ledger knows about stock, not about shipping.
//
// **Splitting has a cost the ledger cannot see.** Two boxes means two shipping
// charges and two chances to lose one. A rule that avoids splitting where
// possible is often worth more than one that minimises distance — and that is
// L15.3's problem, which this function hands off to.
//
// Reserve nothing unless the whole quantity can be covered. A partial
// allocation that leaves holds scattered across warehouses is the worst of both.
func (l *Ledger) Allocate(orderID string, sku SKU, qty int, ttl time.Duration, preference []WarehouseID) ([]Reservation, error) {
	panic("TODO")
}
