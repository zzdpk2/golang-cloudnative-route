package domain

import (
	"context"
)

// ============================================================
// Repository Contracts
//
// The ports the domain depends on; adapters live in infrastructure.
// ============================================================

type OrderRepository interface {
	Save(ctx context.Context, order *Order) error
	FindByID(ctx context.Context, id OrderID) (*Order, error)
	FindByCustomerID(ctx context.Context, customerID CustomerID) ([]*Order, error)
	Delete(ctx context.Context, id OrderID) error
}

type CustomerRepository interface {
	Save(ctx context.Context, customer *Customer) error
	FindByID(ctx context.Context, id CustomerID) (*Customer, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type ProductRepository interface {
	Save(ctx context.Context, product *Product) error
	FindByID(ctx context.Context, id ProductID) (*Product, error)
	FindByCategory(ctx context.Context, category string) ([]*Product, error)
	FindAll(ctx context.Context) ([]*Product, error)
}
