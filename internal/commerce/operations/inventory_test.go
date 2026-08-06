package operations

import (
	"errors"
	"sync"
	"testing"
	"time"
)

const ttl = 30 * time.Minute

// clock is a hand-wound clock, so expiry can be tested without waiting.
type clock struct {
	mu sync.Mutex
	t  time.Time
}

func newClock() *clock {
	return &clock{t: time.Date(2026, time.March, 15, 12, 0, 0, 0, time.UTC)}
}

func (c *clock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *clock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

func stocked(t *testing.T, w WarehouseID, sku SKU, qty int) (*Ledger, *clock) {
	t.Helper()
	c := newClock()
	l := NewLedger(c.now)
	if l == nil {
		t.Fatal("NewLedger returned nil")
	}
	if err := l.Receive(w, sku, qty); err != nil {
		t.Fatalf("Receive: %v", err)
	}
	return l, c
}

func mustStock(t *testing.T, l *Ledger, w WarehouseID, sku SKU) Stock {
	t.Helper()
	s, err := l.Stock(w, sku)
	if err != nil {
		t.Fatalf("Stock: %v", err)
	}
	return s
}

// ---- The three numbers ----

func TestReceive_AddsOnHand(t *testing.T) {
	l, _ := stocked(t, "syd", "sku-1", 10)

	got := mustStock(t, l, "syd", "sku-1")
	if got.OnHand != 10 || got.Reserved != 0 || got.Available() != 10 {
		t.Errorf("stock = %+v, available %d; want OnHand 10, Reserved 0, Available 10",
			got, got.Available())
	}

	if err := l.Receive("syd", "sku-1", 5); err != nil {
		t.Fatalf("second Receive: %v", err)
	}
	if got := mustStock(t, l, "syd", "sku-1"); got.OnHand != 15 {
		t.Errorf("OnHand = %d, want 15", got.OnHand)
	}
}

func TestReceive_RejectsNonPositive(t *testing.T) {
	l, _ := stocked(t, "syd", "sku-1", 10)
	if err := l.Receive("syd", "sku-1", 0); !errors.Is(err, ErrNonPositiveQuantity) {
		t.Errorf("Receive(0) = %v, want ErrNonPositiveQuantity", err)
	}
}

func TestStock_UnknownSKU(t *testing.T) {
	l, _ := stocked(t, "syd", "sku-1", 10)
	if _, err := l.Stock("syd", "nope"); !errors.Is(err, ErrUnknownSKU) {
		t.Errorf("Stock(unknown) = %v, want ErrUnknownSKU", err)
	}
}

// The heart of it: reserving holds stock without removing it.
func TestReserve_HoldsWithoutDeducting(t *testing.T) {
	l, _ := stocked(t, "syd", "sku-1", 10)

	res, err := l.Reserve("ord-1", "syd", "sku-1", 2, ttl)
	if err != nil {
		t.Fatalf("Reserve: %v", err)
	}
	if res.ID == "" {
		t.Error("reservation has no id")
	}
	if res.Quantity != 2 {
		t.Errorf("Quantity = %d, want 2", res.Quantity)
	}

	got := mustStock(t, l, "syd", "sku-1")
	if got.OnHand != 10 {
		t.Errorf("OnHand = %d, want 10: the goods are still in the building", got.OnHand)
	}
	if got.Reserved != 2 {
		t.Errorf("Reserved = %d, want 2", got.Reserved)
	}
	if got.Available() != 8 {
		t.Errorf("Available = %d, want 8", got.Available())
	}
}

func TestReserve_RefusesMoreThanAvailable(t *testing.T) {
	l, _ := stocked(t, "syd", "sku-1", 3)

	if _, err := l.Reserve("ord-1", "syd", "sku-1", 2, ttl); err != nil {
		t.Fatalf("first Reserve: %v", err)
	}
	_, err := l.Reserve("ord-2", "syd", "sku-1", 2, ttl)
	if !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("second Reserve = %v, want ErrInsufficientStock", err)
	}

	if got := mustStock(t, l, "syd", "sku-1"); got.Reserved != 2 {
		t.Errorf("Reserved = %d, want 2: a failed reservation must change nothing",
			got.Reserved)
	}
}

// ---- Commit and release ----

func TestCommit_DeductsOnHand(t *testing.T) {
	l, _ := stocked(t, "syd", "sku-1", 10)
	res, _ := l.Reserve("ord-1", "syd", "sku-1", 2, ttl)

	if err := l.Commit(res.ID); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	got := mustStock(t, l, "syd", "sku-1")
	if got.OnHand != 8 || got.Reserved != 0 || got.Available() != 8 {
		t.Errorf("after commit: %+v available %d; want OnHand 8, Reserved 0, Available 8",
			got, got.Available())
	}
}

// A payment webhook arriving twice must not deduct twice.
func TestCommit_IsNotRepeatable(t *testing.T) {
	l, _ := stocked(t, "syd", "sku-1", 10)
	res, _ := l.Reserve("ord-1", "syd", "sku-1", 2, ttl)

	if err := l.Commit(res.ID); err != nil {
		t.Fatalf("first Commit: %v", err)
	}
	if err := l.Commit(res.ID); !errors.Is(err, ErrUnknownReservation) {
		t.Errorf("second Commit = %v, want ErrUnknownReservation", err)
	}
	if got := mustStock(t, l, "syd", "sku-1"); got.OnHand != 8 {
		t.Errorf("OnHand = %d, want 8: the second commit deducted again", got.OnHand)
	}
}

func TestRelease_ReturnsStock(t *testing.T) {
	l, _ := stocked(t, "syd", "sku-1", 10)
	res, _ := l.Reserve("ord-1", "syd", "sku-1", 2, ttl)

	if err := l.Release(res.ID); err != nil {
		t.Fatalf("Release: %v", err)
	}

	got := mustStock(t, l, "syd", "sku-1")
	if got.OnHand != 10 || got.Reserved != 0 || got.Available() != 10 {
		t.Errorf("after release: %+v available %d; want everything back",
			got, got.Available())
	}
	if err := l.Release(res.ID); !errors.Is(err, ErrUnknownReservation) {
		t.Errorf("second Release = %v, want ErrUnknownReservation", err)
	}
}

// ---- Expiry ----

func TestExpireDue_ReleasesOverdueHolds(t *testing.T) {
	l, c := stocked(t, "syd", "sku-1", 10)
	if _, err := l.Reserve("ord-1", "syd", "sku-1", 2, ttl); err != nil {
		t.Fatalf("Reserve: %v", err)
	}

	if n := l.ExpireDue(); n != 0 {
		t.Errorf("ExpireDue before the deadline released %d, want 0", n)
	}

	c.advance(ttl + time.Minute)

	if n := l.ExpireDue(); n != 1 {
		t.Errorf("ExpireDue = %d, want 1", n)
	}
	if got := mustStock(t, l, "syd", "sku-1"); got.Available() != 10 {
		t.Errorf("Available = %d, want 10: expiry must return the stock", got.Available())
	}
	if n := l.ExpireDue(); n != 0 {
		t.Errorf("a second ExpireDue released %d, want 0: it must be idempotent", n)
	}
}

func TestCommit_AfterExpiryIsRejected(t *testing.T) {
	l, c := stocked(t, "syd", "sku-1", 10)
	res, _ := l.Reserve("ord-1", "syd", "sku-1", 2, ttl)

	c.advance(ttl + time.Minute)

	err := l.Commit(res.ID)
	if !errors.Is(err, ErrReservationExpired) && !errors.Is(err, ErrUnknownReservation) {
		t.Errorf("Commit after expiry = %v, want ErrReservationExpired "+
			"(or ErrUnknownReservation if the sweeper got there first)", err)
	}
	if got := mustStock(t, l, "syd", "sku-1"); got.OnHand != 10 {
		t.Errorf("OnHand = %d, want 10: an expired hold must not deduct", got.OnHand)
	}
}

// Expiry must be observable without the sweeper having run — otherwise the
// answer depends on timing.
func TestReserve_SucceedsOnceAHoldHasLapsed(t *testing.T) {
	l, c := stocked(t, "syd", "sku-1", 2)
	if _, err := l.Reserve("ord-1", "syd", "sku-1", 2, ttl); err != nil {
		t.Fatalf("Reserve: %v", err)
	}

	c.advance(ttl + time.Minute)

	if _, err := l.Reserve("ord-2", "syd", "sku-1", 2, ttl); err != nil {
		t.Errorf("Reserve after the first hold lapsed = %v, want success. "+
			"Either Reserve considers expiry itself, or it sweeps first — "+
			"but the caller must not have to remember to.", err)
	}
}

// ---- No oversell ----

// Two hundred goroutines want the last hundred units. Exactly a hundred may
// win. Run with -race.
func TestReserve_NoOversellUnderConcurrency(t *testing.T) {
	const stock = 100
	const attempts = 200

	l, _ := stocked(t, "syd", "sku-1", stock)

	var mu sync.Mutex
	granted := 0

	var wg sync.WaitGroup
	for i := range attempts {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if _, err := l.Reserve("ord", "syd", "sku-1", 1, ttl); err == nil {
				mu.Lock()
				granted++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	if granted != stock {
		t.Errorf("granted %d reservations against %d units", granted, stock)
	}
	got := mustStock(t, l, "syd", "sku-1")
	if got.Available() != 0 {
		t.Errorf("Available = %d, want 0", got.Available())
	}
	if got.Reserved != stock {
		t.Errorf("Reserved = %d, want %d", got.Reserved, stock)
	}
}

// Reservation ids must be unique even when handed out concurrently.
func TestReserve_IDsAreUnique(t *testing.T) {
	l, _ := stocked(t, "syd", "sku-1", 200)

	var mu sync.Mutex
	seen := map[string]bool{}

	var wg sync.WaitGroup
	for range 200 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := l.Reserve("ord", "syd", "sku-1", 1, ttl)
			if err != nil {
				return
			}
			mu.Lock()
			defer mu.Unlock()
			if seen[res.ID] {
				t.Errorf("duplicate reservation id %q", res.ID)
			}
			seen[res.ID] = true
		}()
	}
	wg.Wait()
}

// ---- Multi-warehouse ----

func TestAllocate_PrefersTheFirstWarehouse(t *testing.T) {
	c := newClock()
	l := NewLedger(c.now)
	_ = l.Receive("syd", "sku-1", 3)
	_ = l.Receive("mel", "sku-1", 10)

	got, err := l.Allocate("ord-1", "sku-1", 2, ttl, []WarehouseID{"syd", "mel"})
	if err != nil {
		t.Fatalf("Allocate: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("allocated across %d warehouses, want 1", len(got))
	}
	if got[0].Warehouse != "syd" || got[0].Quantity != 2 {
		t.Errorf("allocation = %+v, want 2 from syd", got[0])
	}
}

func TestAllocate_SplitsInPreferenceOrder(t *testing.T) {
	c := newClock()
	l := NewLedger(c.now)
	_ = l.Receive("syd", "sku-1", 3)
	_ = l.Receive("mel", "sku-1", 10)

	got, err := l.Allocate("ord-1", "sku-1", 5, ttl, []WarehouseID{"syd", "mel"})
	if err != nil {
		t.Fatalf("Allocate: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("allocated across %d warehouses, want 2", len(got))
	}
	if got[0].Warehouse != "syd" || got[0].Quantity != 3 {
		t.Errorf("first = %+v, want all 3 from syd", got[0])
	}
	if got[1].Warehouse != "mel" || got[1].Quantity != 2 {
		t.Errorf("second = %+v, want the remaining 2 from mel", got[1])
	}
}

// All or nothing: a partial allocation leaves holds scattered for an order that
// can never ship.
func TestAllocate_ReservesNothingWhenItCannotCoverTheWhole(t *testing.T) {
	c := newClock()
	l := NewLedger(c.now)
	_ = l.Receive("syd", "sku-1", 3)
	_ = l.Receive("mel", "sku-1", 10)

	if _, err := l.Allocate("ord-1", "sku-1", 20, ttl, []WarehouseID{"syd", "mel"}); !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("Allocate(20) = %v, want ErrInsufficientStock", err)
	}

	for _, w := range []WarehouseID{"syd", "mel"} {
		if got := mustStock(t, l, w, "sku-1"); got.Reserved != 0 {
			t.Errorf("%s Reserved = %d, want 0: a failed allocation must hold nothing",
				w, got.Reserved)
		}
	}
}
