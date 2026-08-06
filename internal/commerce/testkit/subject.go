// Package testkit is the L5 exercise: test doubles.
//
// Everything in this file is already implemented. It is the *subject under
// test*, not your work. Your job lives in doubles.go: build a fake, a stub, a
// spy, and a generated mock for the two ports below, then make the tests pass.
//
// Read subject.go carefully first. You cannot write a useful double for a
// collaborator whose contract you have not read.
package testkit

import (
	"context"
	"errors"
	"fmt"
)

var (
	// ErrOutOfStock means the SKU exists but cannot cover the requested quantity.
	ErrOutOfStock = errors.New("out of stock")
	// ErrUnknownSKU means the SKU is not carried at all.
	ErrUnknownSKU = errors.New("unknown sku")
)

// InventoryPort is the warehouse the reservation service talks to.
//
// Available reports the sellable quantity for a SKU, and returns ErrUnknownSKU
// for a SKU the warehouse does not carry. Reserve holds stock and returns
// ErrOutOfStock when the quantity is no longer available.
type InventoryPort interface {
	Available(ctx context.Context, sku string) (int, error)
	Reserve(ctx context.Context, sku string, qty int) error
}

// Notifier delivers a message to a customer. Delivery is best effort: a failure
// here must not fail the reservation that already succeeded.
type Notifier interface {
	Notify(ctx context.Context, customerID, message string) error
}

// ReservationService is the code under test.
type ReservationService struct {
	inventory InventoryPort
	notifier  Notifier
}

func NewReservationService(inventory InventoryPort, notifier Notifier) *ReservationService {
	return &ReservationService{inventory: inventory, notifier: notifier}
}

// Reserve holds qty units of sku for a customer and notifies them.
//
// The ordering matters and the tests pin it down:
//   - a non-positive qty is rejected before any port is touched
//   - availability is checked before reserving
//   - the customer is notified only after the reservation succeeded
//   - a notification failure is swallowed: the stock is already held, so the
//     reservation stands
func (s *ReservationService) Reserve(ctx context.Context, customerID, sku string, qty int) error {
	if qty <= 0 {
		return fmt.Errorf("quantity must be positive, got %d", qty)
	}

	available, err := s.inventory.Available(ctx, sku)
	if err != nil {
		return fmt.Errorf("check availability: %w", err)
	}
	if available < qty {
		return fmt.Errorf("reserve %d of %s: %w", qty, sku, ErrOutOfStock)
	}

	if err := s.inventory.Reserve(ctx, sku, qty); err != nil {
		return fmt.Errorf("reserve: %w", err)
	}

	// Best effort on purpose: the stock is already held.
	_ = s.notifier.Notify(ctx, customerID, fmt.Sprintf("reserved %d x %s", qty, sku))
	return nil
}
