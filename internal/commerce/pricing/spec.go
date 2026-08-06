package pricing

import (
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
)

// ============================================================
// Specifications
//
// Composable predicates combined with AND, OR, and NOT.
// ============================================================

// Specification answers one yes/no question about a candidate.
//
// The reason this is an interface rather than a plain func(T) bool is
// composition: because And, Or, and Not are themselves Specifications, an
// arbitrarily nested rule is still just a Specification, and everything that
// consumes one keeps working.
//
//	And(IsVIP, Not(IsFreshCategory), AmountOver(500))
//
// That expression is the whole point of the pattern — a business rule you can
// read aloud, build at run time, and test in isolation.
type Specification[T any] interface {
	IsSatisfiedBy(candidate T) bool
}

type AndSpec[T any] struct {
	specs []Specification[T]
}

func And[T any](specs ...Specification[T]) Specification[T] { panic("TODO") }

// IsSatisfiedBy reports whether every child spec is satisfied.
//
// Two things to settle before writing the loop. What does And() with no
// children mean? And should evaluation short-circuit on the first failure?
// Short-circuiting is free performance, but it also means later specs are never
// evaluated — which matters the day somebody writes a spec with a side effect.
// (They should not. They will.)
func (s *AndSpec[T]) IsSatisfiedBy(candidate T) bool {
	panic("TODO")
}

type OrSpec[T any] struct {
	specs []Specification[T]
}

func Or[T any](specs ...Specification[T]) Specification[T] { panic("TODO") }

// IsSatisfiedBy reports whether any child spec is satisfied.
//
// The empty case is the mirror of And's, and the two answers are opposites.
// Work out why from first principles rather than memorising: what is the
// identity element of "and", and what is it for "or"?
func (s *OrSpec[T]) IsSatisfiedBy(candidate T) bool {
	panic("TODO")
}

type NotSpec[T any] struct {
	spec Specification[T]
}

func Not[T any](spec Specification[T]) Specification[T] { panic("TODO") }

// IsSatisfiedBy inverts the wrapped specification.
func (s *NotSpec[T]) IsSatisfiedBy(candidate T) bool {
	panic("TODO")
}

// SpecFunc adapts a plain function into a Specification, so a one-off rule does
// not need its own type.
//
// This is the same trick as http.HandlerFunc: a named function type with a
// method on it, letting a func satisfy an interface. Worth recognising — it
// shows up all over the standard library.
type SpecFunc[T any] func(T) bool

func (f SpecFunc[T]) IsSatisfiedBy(candidate T) bool { panic("TODO") }

// OrderMinAmountSpec is satisfied by orders worth at least a threshold.
//
// It stores cents rather than a Money, which quietly drops the currency. That
// is a real modelling weakness: this spec will happily compare a USD order
// against an AUD threshold. Decide whether that is acceptable here, and what it
// would take to fix.
type OrderMinAmountSpec struct {
	minCents int64
}

func NewOrderMinAmountSpec(minAmount domain.Money) *OrderMinAmountSpec { panic("TODO") }

// IsSatisfiedBy reports whether the order meets the minimum.
//
// TotalAmount can fail. IsSatisfiedBy cannot report that. So you have to decide
// what an order whose total cannot be computed means here — satisfied or not —
// and live with the fact that the interface gave you nowhere to say "I don't
// know". That limitation is the price of the pattern's simplicity, and it is
// worth feeling it once.
func (s *OrderMinAmountSpec) IsSatisfiedBy(order *domain.Order) bool {
	panic("TODO")
}

// OrderStatusSpec is satisfied by orders in one particular status.
type OrderStatusSpec struct {
	status domain.OrderStatus
}

func NewOrderStatusSpec(status domain.OrderStatus) *OrderStatusSpec { panic("TODO") }

// IsSatisfiedBy reports whether the order is in the configured status.
func (s *OrderStatusSpec) IsSatisfiedBy(order *domain.Order) bool {
	panic("TODO")
}

// OrderHasProductSpec is satisfied by orders containing a given product.
type OrderHasProductSpec struct {
	productID string
}

func NewOrderHasProductSpec(productID string) *OrderHasProductSpec { panic("TODO") }

// IsSatisfiedBy reports whether the order carries the configured product.
func (s *OrderHasProductSpec) IsSatisfiedBy(order *domain.Order) bool {
	panic("TODO")
}

// Filter returns the items satisfying spec.
//
// These four helpers are why the Specification interface earns its keep: write
// a rule once, and it works as a filter, a counter, and a quantifier without
// any of them knowing what the rule is.
//
// Decide what an empty result should be — nil or an empty slice — and be
// consistent with the choice you made in Order.FilterLines.
func Filter[T any](items []T, spec Specification[T]) []T {
	panic("TODO")
}

// Count returns how many items satisfy spec.
func Count[T any](items []T, spec Specification[T]) int {
	panic("TODO")
}

// Any reports whether at least one item satisfies spec.
func Any[T any](items []T, spec Specification[T]) bool {
	panic("TODO")
}

// All reports whether every item satisfies spec.
//
// All over an empty slice is true — "vacuous truth", and it surprises people
// every time. It is the same identity question as AndSpec with no children, and
// the consistency is not a coincidence.
func All[T any](items []T, spec Specification[T]) bool {
	panic("TODO")
}
