package testkit

import (
	"context"
	"errors"
	"testing"
)

// ---- The fake carries state across a sequence of calls ----

func TestFakeInventory_ReserveReducesAvailable(t *testing.T) {
	inv := NewFakeInventory(map[string]int{"sku-1": 10})
	ctx := context.Background()

	if err := inv.Reserve(ctx, "sku-1", 4); err != nil {
		t.Fatalf("Reserve() = %v, want nil", err)
	}

	got, err := inv.Available(ctx, "sku-1")
	if err != nil {
		t.Fatalf("Available() error = %v", err)
	}
	if want := 6; got != want {
		t.Errorf("Available() after reserving 4 of 10 = %d, want %d", got, want)
	}
}

func TestFakeInventory_UnknownSKU(t *testing.T) {
	inv := NewFakeInventory(map[string]int{"sku-1": 1})
	ctx := context.Background()

	if _, err := inv.Available(ctx, "nope"); !errors.Is(err, ErrUnknownSKU) {
		t.Errorf("Available(unknown) = %v, want ErrUnknownSKU", err)
	}
	if err := inv.Reserve(ctx, "nope", 1); !errors.Is(err, ErrUnknownSKU) {
		t.Errorf("Reserve(unknown) = %v, want ErrUnknownSKU", err)
	}
}

// A SKU that is carried but exhausted is a different answer from one that is
// not carried at all. Collapsing the two is the most common fake bug.
func TestFakeInventory_CarriedButExhausted(t *testing.T) {
	inv := NewFakeInventory(map[string]int{"sku-1": 2})
	ctx := context.Background()

	if err := inv.Reserve(ctx, "sku-1", 2); err != nil {
		t.Fatalf("Reserve() = %v, want nil", err)
	}

	got, err := inv.Available(ctx, "sku-1")
	if err != nil {
		t.Fatalf("Available() on an exhausted SKU should not error, got %v", err)
	}
	if got != 0 {
		t.Errorf("Available() = %d, want 0", got)
	}
	if err := inv.Reserve(ctx, "sku-1", 1); !errors.Is(err, ErrOutOfStock) {
		t.Errorf("Reserve() past the last unit = %v, want ErrOutOfStock", err)
	}
}

// The fake must not share storage with the map the caller passed in.
func TestFakeInventory_DoesNotAliasSeedMap(t *testing.T) {
	seed := map[string]int{"sku-1": 5}
	inv := NewFakeInventory(seed)

	seed["sku-1"] = 999

	got, err := inv.Available(context.Background(), "sku-1")
	if err != nil {
		t.Fatalf("Available() error = %v", err)
	}
	if got != 5 {
		t.Errorf("Available() = %d, want 5: the fake aliased the caller's map", got)
	}
}

// ---- The stub feeds one canned answer into the subject ----

func TestReserve_OutOfStock(t *testing.T) {
	svc := NewReservationService(
		&StubInventory{AvailableQty: 2},
		&SpyNotifier{},
	)

	err := svc.Reserve(context.Background(), "cust-1", "sku-1", 5)
	if !errors.Is(err, ErrOutOfStock) {
		t.Errorf("Reserve(5 of 2 available) = %v, want ErrOutOfStock", err)
	}
}

func TestReserve_AvailabilityLookupFails(t *testing.T) {
	svc := NewReservationService(
		&StubInventory{AvailableErr: ErrUnknownSKU},
		&SpyNotifier{},
	)

	err := svc.Reserve(context.Background(), "cust-1", "sku-1", 1)
	if !errors.Is(err, ErrUnknownSKU) {
		t.Errorf("Reserve() = %v, want the wrapped ErrUnknownSKU", err)
	}
}

// ---- The spy answers questions about the interaction itself ----

func TestReserve_NotifiesOnSuccess(t *testing.T) {
	spy := &SpyNotifier{}
	svc := NewReservationService(NewFakeInventory(map[string]int{"sku-1": 10}), spy)

	if err := svc.Reserve(context.Background(), "cust-7", "sku-1", 3); err != nil {
		t.Fatalf("Reserve() = %v, want nil", err)
	}

	calls := spy.Calls()
	if len(calls) != 1 {
		t.Fatalf("Notify called %d times, want 1", len(calls))
	}
	if calls[0].CustomerID != "cust-7" {
		t.Errorf("notified %q, want %q", calls[0].CustomerID, "cust-7")
	}
}

// Nothing was reserved, so nobody should be told anything.
func TestReserve_DoesNotNotifyOnFailure(t *testing.T) {
	spy := &SpyNotifier{}
	svc := NewReservationService(&StubInventory{AvailableQty: 0}, spy)

	_ = svc.Reserve(context.Background(), "cust-7", "sku-1", 1)

	if got := len(spy.Calls()); got != 0 {
		t.Errorf("Notify called %d times on a failed reservation, want 0", got)
	}
}

// The stock is already held; a failed notification must not undo that.
func TestReserve_NotificationFailureDoesNotFailReservation(t *testing.T) {
	spy := &SpyNotifier{Err: errors.New("sms gateway down")}
	fake := NewFakeInventory(map[string]int{"sku-1": 10})
	svc := NewReservationService(fake, spy)

	if err := svc.Reserve(context.Background(), "cust-7", "sku-1", 3); err != nil {
		t.Errorf("Reserve() = %v, want nil despite the notifier failing", err)
	}
	if got := len(spy.Calls()); got != 1 {
		t.Errorf("Notify recorded %d calls, want 1: record before returning the error", got)
	}

	left, _ := fake.Available(context.Background(), "sku-1")
	if left != 7 {
		t.Errorf("Available() = %d, want 7: the reservation should stand", left)
	}
}

// Ports must not be touched at all for an obviously invalid request.
func TestReserve_RejectsNonPositiveQuantityBeforeTouchingPorts(t *testing.T) {
	spy := &SpyNotifier{}
	fake := NewFakeInventory(map[string]int{"sku-1": 10})
	svc := NewReservationService(fake, spy)

	if err := svc.Reserve(context.Background(), "cust-1", "sku-1", 0); err == nil {
		t.Error("Reserve(qty=0) = nil, want an error")
	}
	if got := len(spy.Calls()); got != 0 {
		t.Errorf("Notify called %d times, want 0", got)
	}

	left, _ := fake.Available(context.Background(), "sku-1")
	if left != 10 {
		t.Errorf("Available() = %d, want 10: nothing should have been reserved", left)
	}
}

// Calls() must not hand out a slice the caller can use to rewrite history.
func TestSpyNotifier_CallsAreNotAliased(t *testing.T) {
	spy := &SpyNotifier{}
	_ = spy.Notify(context.Background(), "cust-1", "hello")

	calls := spy.Calls()
	if len(calls) != 1 {
		t.Fatalf("Calls() returned %d entries, want 1", len(calls))
	}
	calls[0].CustomerID = "tampered"

	if again := spy.Calls(); again[0].CustomerID != "cust-1" {
		t.Errorf("Calls() = %q after the caller mutated an earlier result, want %q",
			again[0].CustomerID, "cust-1")
	}
}
