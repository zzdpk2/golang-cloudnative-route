package domain

import "time"

// ============================================================
// Domain Events
//
// Event definitions and the dispatch contract.
// ============================================================

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
	AggregateID() string
}

type BaseEvent struct {
	name        string
	occurredAt  time.Time
	aggregateID string
}

func (e BaseEvent) EventName() string     { panic("TODO") }
func (e BaseEvent) OccurredAt() time.Time { panic("TODO") }
func (e BaseEvent) AggregateID() string   { panic("TODO") }

type OrderCreated struct {
	BaseEvent  // See the corresponding tests for the intended behavior.
	CustomerID string
}

func NewOrderCreated(orderID, customerID string) OrderCreated { panic("TODO") }

type OrderConfirmedEvent struct {
	BaseEvent
	TotalCents int64
}

func NewOrderConfirmed(orderID string, totalCents int64) OrderConfirmedEvent { panic("TODO") }

type OrderCancelledEvent struct {
	BaseEvent
	Reason string
}

func NewOrderCancelled(orderID, reason string) OrderCancelledEvent { panic("TODO") }
