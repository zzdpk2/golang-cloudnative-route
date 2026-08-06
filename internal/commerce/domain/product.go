package domain

import (
	"fmt"
	"time"
)

// These small interfaces exist to be composed. Each names one capability, and
// DisplayableProduct is the union — the Go convention of building wide
// interfaces from narrow ones rather than declaring one wide one.
//
// The payoff shows up at the *consumer*: a function that only formats a price
// should accept Pricer, not DisplayableProduct, so anything priceable can be
// passed to it.
type Pricer interface{ Price() Money }
type Describable interface {
	Name() string
	Description() string
}
type Categorizable interface{ Category() string }

type DisplayableProduct interface {
	Pricer
	Describable
	Categorizable
	fmt.Stringer
}

// ---- Product Entity ----

type ProductID string

func NewProductID(id string) ProductID { panic("TODO") }

type ProductStatus int

const (
	ProductDraft ProductStatus = iota
	ProductActive
	ProductDiscontinued
)

func (s ProductStatus) String() string { panic("TODO") }

// Product is an entity with a small lifecycle: draft, active, discontinued.
//
// metadata is the escape hatch — an untyped bag for things the model does not
// know about yet. Useful, and worth being suspicious of: every field that lives
// in there is a field the compiler cannot check and the domain cannot reason
// about. Notice which parts of this file are harder to work with because of it.
type Product struct {
	id          ProductID
	name        string
	description string
	price       Money
	category    string
	status      ProductStatus
	createdAt   time.Time
	updatedAt   time.Time
	metadata    map[string]any
}

// NewProduct creates a product in ProductDraft, rejecting an empty id or name.
//
// A new product is not immediately sellable — it has to be activated. Ask
// yourself why a lifecycle needs a draft state at all, and what would go wrong
// if products were born active.
func NewProduct(id ProductID, name string, price Money, category string) (*Product, error) {
	panic("TODO")
}

func (p *Product) ID() ProductID         { panic("TODO") }
func (p *Product) Name() string          { panic("TODO") }
func (p *Product) Description() string   { panic("TODO") }
func (p *Product) Price() Money          { panic("TODO") }
func (p *Product) Category() string      { panic("TODO") }
func (p *Product) Status() ProductStatus { panic("TODO") }
func (p *Product) CreatedAt() time.Time  { panic("TODO") }
func (p *Product) UpdatedAt() time.Time  { panic("TODO") }

// UpdatePrice sets a new price and records that the product changed.
//
// Every mutating method here has to touch updatedAt. Forgetting one is easy and
// silent. Note how the aggregate in L2 has the same problem, and think about
// whether there is a way to make it structurally impossible to forget rather
// than a rule to remember.
func (p *Product) UpdatePrice(newPrice Money) {
	panic("TODO")
}

// Activate makes a product sellable. A discontinued product cannot come back.
func (p *Product) Activate() error {
	panic("TODO")
}

// Discontinue retires a product permanently.
func (p *Product) Discontinue() error {
	panic("TODO")
}

// SetMetadata stores an arbitrary value.
//
// The map is nil on a zero Product. Writing to a nil map panics, reading from
// one does not — one of Go's sharper edges. Decide where you handle that:
// here, or in the constructor.
func (p *Product) SetMetadata(key string, value any) {
	panic("TODO")
}

// GetMetadata returns a stored value and whether it was present.
//
// The comma-ok form is the whole point: a missing key and a key holding nil are
// different situations, and a single return value cannot tell them apart.
func (p *Product) GetMetadata(key string) (any, bool) {
	panic("TODO")
}

// GetMetadataAs is the typed read: it succeeds only when the key exists *and*
// the stored value is a T.
//
// This is a free function rather than a method because Go methods cannot take
// type parameters. That restriction is not arbitrary — look up why, and note
// that this same shape appears again in txdsl.GetAs in L9.
func GetMetadataAs[T any](p *Product, key string) (T, bool) {
	panic("TODO")
}

// Equals compares by identity, not by attributes.
//
// This is the opposite of Money.Equals, and the contrast is the lesson: two
// products with identical names and prices are two different products, and a
// product that has been renamed and repriced is still the same product. Only
// the id decides.
//
// Handle nil on either side.
func (p *Product) Equals(other *Product) bool {
	panic("TODO")
}

func (p *Product) String() string { panic("TODO") }

// A compile-time assertion that *Product satisfies the composed interface. It
// costs nothing at run time and fails the build the moment a method is renamed.
var _ DisplayableProduct = (*Product)(nil)

// DescribeItem formats any value, dispatching on its dynamic type.
//
// The cases, in the order the tests expect them to be tried:
//
//	*Product     → "Product: " + name
//	Pricer       → "Priced item: " + price
//	fmt.Stringer → "Item: " + string form
//	string       → "Raw: " + the value
//	anything else→ "Unknown item"
//
// Order matters, and this is the real exercise. *Product satisfies *all* of the
// first three, so whichever case comes first is the one that wins. A type
// switch is not a set of independent conditions; it is a sequence.
func DescribeItem(item any) string {
	panic("TODO")
}

func ValidateProduct(p *Product) error { panic("TODO") }

// CheckProduct returns nil when the product is valid.
//
// Read this one carefully before running its test — it is already written, and
// it is wrong. It is the canonical Go nil-interface trap: a nil *ProductError
// assigned into an error interface produces an interface that is not nil,
// because the interface holds a type even when the value is absent.
//
// Do not fix it by reading the answer. Run the test, look at what
// `err != nil` reports, print `%T`, and work out where the nil went. Then fix
// it, and write down the rule you would apply to avoid it again.
func CheckProduct(p *Product) error { panic("TODO") }

type ProductError struct{ Field, Message string }

func (e *ProductError) Error() string { panic("TODO") }

func validateWithCustomError(p *Product) *ProductError { panic("TODO") }

// ---- DTO ----

// ProductDTO is the wire shape. It exists so that the entity's internals can
// change without breaking every client — the same reason Money has moneyJSON.
type ProductDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	Category    string  `json:"category"`
	Status      string  `json:"status"`
}

// ToDTO flattens the entity for transport.
//
// Note what the DTO does *not* carry: metadata, timestamps, and the unexported
// invariants. Deciding what to leave out is the interesting half of the job —
// every field you expose is a field a client can come to depend on.
func (p *Product) ToDTO() ProductDTO {
	panic("TODO")
}

func (p *Product) MarshalJSON() ([]byte, error) { panic("TODO") }
