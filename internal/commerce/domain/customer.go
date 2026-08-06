package domain

import (
	"errors"
)

// ValidationError says which field of which entity was wrong. It is the
// error type you reach for when the *caller* needs to know the detail —
// a form that has to highlight one input, for instance.
type ValidationError struct {
	Entity, Field, Reason string
}

func (e *ValidationError) Error() string { panic("TODO") }

// TemporaryError marks failures worth retrying. The retry logic in L8 keys off
// exactly this kind of interface rather than off concrete types, which is what
// lets it retry errors it has never heard of.
type TemporaryError interface {
	error
	IsTemporary() bool
}

type RetryableError struct {
	Cause      error
	RetryAfter int
}

func (e *RetryableError) Error() string { panic("TODO") }

func (e *RetryableError) IsTemporary() bool { panic("TODO") }
func (e *RetryableError) Unwrap() error     { panic("TODO") }

// Sentinel errors. Callers compare against these with errors.Is, never with ==,
// because by the time an error reaches them it has usually been wrapped.
var (
	ErrCustomerNotFound    = errors.New("customer not found")
	ErrCustomerDeactivated = errors.New("customer is deactivated")
	ErrDuplicateEmail      = errors.New("duplicate email")
)

type CustomerID string
type CustomerStatus int

const (
	CustomerActive CustomerStatus = iota
	CustomerDeactivated
)

// Customer is an *entity*, not a value object: it has id Two customers
// with the same name and email are still two different customers, and one
// customer stays the same customer after changing both.
//
// That is the whole distinction from the valueobject package. Money is defined
// entirely by its attributes; a Customer is defined by its id. Everything else
// about it may change over a lifetime.
type Customer struct {
	id        CustomerID
	name      string
	email     Email
	addresses []Address
	status    CustomerStatus
}

// NewCustomer creates an active customer, rejecting an empty id or name.
//
// The email arrives already validated — Email cannot exist in an invalid
// state. Notice how much validation this constructor does *not* have to do,
// and that this is the payoff for L1.
func NewCustomer(id CustomerID, name string, email Email) (*Customer, error) {
	panic("TODO")
}

func (c *Customer) ID() CustomerID         { panic("TODO") }
func (c *Customer) Name() string           { panic("TODO") }
func (c *Customer) Email() Email           { panic("TODO") }
func (c *Customer) Status() CustomerStatus { panic("TODO") }

// Addresses exposes the address book.
//
// Same question you answered for Order.Lines in L2, and the answer had better
// be the same. A slice returned directly is a slice the caller can rewrite.
func (c *Customer) Addresses() []Address {
	panic("TODO")
}

// AddAddress appends to the address book.
func (c *Customer) AddAddress(addr Address) {
	panic("TODO")
}

// RemoveAddress drops the address at index, and errors when the index is out of
// range.
//
// Removing from the middle of a slice is the classic place to leak memory: the
// obvious append-based trick leaves the removed element still referenced by the
// backing array. It does not matter for a value type like Address, but build
// the habit of asking, because it very much matters for a slice of pointers.
func (c *Customer) RemoveAddress(index int) error {
	panic("TODO")
}

func (c *Customer) Deactivate() { panic("TODO") }

// UpdateEmail changes the email, refusing when the customer is deactivated.
//
// The test pins down two things at once: errors.Is must find
// ErrCustomerDeactivated, *and* the message must say more than the bare
// sentinel does. That combination is what %w exists for — returning the
// sentinel unwrapped satisfies the first and fails the second.
func (c *Customer) UpdateEmail(newEmail Email) error {
	panic("TODO")
}

// PlaceOrder checks whether this customer is currently allowed to order: they
// must be active and have at least one address on file.
//
// The two failures are reported in deliberately different ways, and the tests
// check both:
//
//   - deactivated wraps ErrCustomerDeactivated, found with errors.Is, through
//     more than one layer of wrapping
//   - no address returns a *ValidationError with Field "addresses", found with
//     errors.As
//
// That is the distinction worth internalising. errors.Is answers "is this that
// specific problem?"; errors.As answers "is this a kind of problem carrying
// details I need?". Sentinels are for the first, typed errors for the second,
// and the test asserts a deactivated customer is *not* a ValidationError — so
// do not reach for one type to express both.
func (c *Customer) PlaceOrder() error {
	panic("TODO")
}

// AuditLog performs an action and returns an audit line for it. "delete" is
// rejected; anything else succeeds.
//
// The successful line is exactly:
//
//	AUDIT [cust-001] update: SUCCESS
//
// The catch is in the failing case: the test asserts result is non-empty *even
// though an error was returned*. A plain `return "", err` cannot do that.
//
// This is what the named return values in the signature are for. A deferred
// closure can still write to them after the return statement has run, which is
// the one thing anonymous returns cannot do. Write it with a bare `return`
// first and watch what the defer can reach.
func (c *Customer) AuditLog(action string) (result string, err error) {
	panic("TODO")
}
