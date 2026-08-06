package persistence

import (
	"context"
	"fmt"
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"sync"
)

// ============================================================
// In-Memory Repositories
//
// Repository adapters backed by sync.RWMutex and sync.Map.
//
// Two stores, two different concurrency tools, on purpose. Orders use an
// explicit RWMutex around a plain map; products use sync.Map. Once both are
// green you should be able to say which you would reach for on a new store —
// the honest answer is "RWMutex, almost always", and the interesting part is
// knowing the narrow case where that is wrong.
//
// sync.Map is tuned for two specific shapes: keys written once and read many
// times, and disjoint key sets per goroutine. Outside those it is *slower* than
// a guarded map, and it costs type safety at every call because its API is
// `any`. Reaching for it because it sounds like "the concurrent map" is a
// common mistake.
// ============================================================

var (
	ErrNotFound = fmt.Errorf("not found")
	ErrConflict = fmt.Errorf("already exists")
)

// ---- InMemoryOrderRepository ----

// InMemoryOrderRepository is the adapter behind domain.OrderRepository.
//
// It stores *pointers* to aggregates, which makes it a fake rather than a real
// repository: a caller who mutates a returned order changes what is "in the
// database" without ever calling Save. A real store serialises on write and
// reconstitutes on read, so that cannot happen.
//
// Live with it for now, but know it is there. It is exactly the kind of
// difference that lets a bug pass every test and then fail against Postgres.
type InMemoryOrderRepository struct {
	mu     sync.RWMutex
	orders map[domain.OrderID]*domain.Order
}

func NewInMemoryOrderRepository() *InMemoryOrderRepository { panic("TODO") }

// Save inserts or replaces an order.
//
// Decide whether saving an id that already exists is an update or an
// ErrConflict. Both are defensible — upsert is convenient, and an explicit
// conflict is what optimistic locking is built on. The interface gives you one
// method, so you have to pick.
func (r *InMemoryOrderRepository) Save(ctx context.Context, order *domain.Order) error {
	panic("TODO")
}

// FindByID returns the order, or ErrNotFound.
//
// Wrap ErrNotFound rather than returning the bare sentinel, so the caller
// learns *which* id was missing while errors.Is still matches. That is L4's
// platform/errors pattern, used in anger.
func (r *InMemoryOrderRepository) FindByID(ctx context.Context, id domain.OrderID) (*domain.Order, error) {
	panic("TODO")
}

// FindByCustomerID returns every order belonging to a customer.
//
// This scans the whole map — fine for a fake, catastrophic for a real store. It
// is the in-memory equivalent of a query with no index, and it is worth
// noticing that the repository *interface* gives no hint that one method is
// O(1) and another O(n).
//
// Two things to settle: what an empty result is (nil or empty slice), and
// whether order matters. Map iteration is randomised, so a test asserting a
// particular sequence will flake rather than fail.
func (r *InMemoryOrderRepository) FindByCustomerID(ctx context.Context, customerID domain.CustomerID) ([]*domain.Order, error) {
	panic("TODO")
}

// Delete removes an order, or reports ErrNotFound.
//
// Go's delete on a missing key is a silent no-op, so "did it exist?" has to be
// a separate check — and that check and the delete must happen under the same
// lock, or two callers can both believe they were the one who removed it.
func (r *InMemoryOrderRepository) Delete(ctx context.Context, id domain.OrderID) error {
	panic("TODO")
}

// Compile-time proof that this adapter satisfies the domain's port. It costs
// nothing at run time and breaks the build the moment the interface changes.
var _ domain.OrderRepository = (*InMemoryOrderRepository)(nil)

// ---- InMemoryProductRepository ----

// InMemoryProductRepository uses sync.Map, which carries its own locking.
//
// The price shows up in every method below: Load returns an `any`, so each read
// ends in a type assertion the compiler cannot check. Use the comma-ok form — a
// value of the wrong type means something else wrote to this map, and panicking
// inside a repository is not the right response.
type InMemoryProductRepository struct {
	products sync.Map // key: ProductID, value: *domain.Product
}

func NewInMemoryProductRepository() *InMemoryProductRepository { panic("TODO") }

// Save stores a product.
//
// Look at what sync.Map does not offer: no way to say "store only if absent,
// and tell me which happened" except LoadOrStore. If you chose conflict-on-
// duplicate for orders, work out whether you can even implement that decision
// here — and if not, the asymmetry between the two repositories is worth a note
// of its own.
func (r *InMemoryProductRepository) Save(ctx context.Context, product *domain.Product) error {
	panic("TODO")
}

// FindByID returns the product, or ErrNotFound.
func (r *InMemoryProductRepository) FindByID(ctx context.Context, id domain.ProductID) (*domain.Product, error) {
	panic("TODO")
}

// FindByCategory returns every product in a category.
//
// sync.Map has no length and no key list; the only way through it is Range with
// a callback. Returning false from that callback stops the iteration early —
// not what you want here, but exactly what you will want the first time you
// write a "find one" over a sync.Map.
func (r *InMemoryProductRepository) FindByCategory(ctx context.Context, category string) ([]*domain.Product, error) {
	panic("TODO")
}

// FindAll returns every product.
//
// Range makes no consistency guarantee: entries added or removed while it runs
// may or may not appear. So this is a rough snapshot rather than a
// point-in-time one — a distinction that starts to matter the moment anything
// pages through the results.
func (r *InMemoryProductRepository) FindAll(ctx context.Context) ([]*domain.Product, error) {
	panic("TODO")
}

func (r *InMemoryProductRepository) Delete(ctx context.Context, id domain.ProductID) { panic("TODO") }

var _ domain.ProductRepository = (*InMemoryProductRepository)(nil)
