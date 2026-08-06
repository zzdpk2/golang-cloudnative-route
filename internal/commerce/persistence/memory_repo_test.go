package persistence

import (
	"context"
	"errors"
	"fmt"
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"sync"
	"testing"
	"time"
)

func makeOrder(t *testing.T, id string, custID string) *domain.Order {
	t.Helper()
	addr, _ := domain.NewAddress("1 St", "Syd", "NSW", "2000", "AU")
	o, err := domain.NewOrder(domain.OrderID(id), domain.CustomerID(custID), addr)
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func TestOrderRepo_SaveAndFind(t *testing.T) {
	repo := NewInMemoryOrderRepository()
	ctx := context.Background()

	o := makeOrder(t, "o1", "c1")
	err := repo.Save(ctx, o)
	if err != nil {
		t.Fatal(err)
	}

	found, err := repo.FindByID(ctx, "o1")
	if err != nil {
		t.Fatal(err)
	}
	if found.ID() != "o1" {
		t.Errorf("ID = %s", found.ID())
	}
}

func TestOrderRepo_FindNotFound(t *testing.T) {
	repo := NewInMemoryOrderRepository()
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestOrderRepo_FindByCustomerID(t *testing.T) {
	repo := NewInMemoryOrderRepository()
	ctx := context.Background()

	repo.Save(ctx, makeOrder(t, "o1", "c1"))
	repo.Save(ctx, makeOrder(t, "o2", "c1"))
	repo.Save(ctx, makeOrder(t, "o3", "c2"))

	orders, err := repo.FindByCustomerID(ctx, "c1")
	if err != nil {
		t.Fatal(err)
	}
	if len(orders) != 2 {
		t.Errorf("expected 2 orders for c1, got %d", len(orders))
	}
}

func TestOrderRepo_Delete(t *testing.T) {
	repo := NewInMemoryOrderRepository()
	ctx := context.Background()

	repo.Save(ctx, makeOrder(t, "o1", "c1"))
	err := repo.Delete(ctx, "o1")
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.FindByID(ctx, "o1")
	if !errors.Is(err, ErrNotFound) {
		t.Error("should be deleted")
	}
}

func TestOrderRepo_DeleteNotFound(t *testing.T) {
	repo := NewInMemoryOrderRepository()
	ctx := context.Background()

	err := repo.Delete(ctx, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestOrderRepo_ContextCancelled(t *testing.T) {
	repo := NewInMemoryOrderRepository()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // See the corresponding tests for the intended behavior.

	_, err := repo.FindByID(ctx, "o1")
	if err == nil {
		t.Error("should error on cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Logf("error: %v (may wrap context.Canceled)", err)
	}
}

func TestOrderRepo_ContextTimeout(t *testing.T) {
	repo := NewInMemoryOrderRepository()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(1 * time.Millisecond) // See the corresponding tests for the intended behavior.

	err := repo.Save(ctx, makeOrder(t, "o1", "c1"))
	if err == nil {
		t.Error("should error on timed out context")
	}
}

func TestOrderRepo_ConcurrentAccess(t *testing.T) {
	repo := NewInMemoryOrderRepository()
	ctx := context.Background()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			id := domain.OrderID(fmt.Sprintf("o-%d", idx))
			o := makeOrder(t, string(id), "c1")
			repo.Save(ctx, o)
		}(i)
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			id := domain.OrderID(fmt.Sprintf("o-%d", idx))
			repo.FindByID(ctx, id) // may or may not find
		}(i)
	}

	wg.Wait()
}

func makeProduct(t *testing.T, id, name string) *domain.Product {
	t.Helper()
	price := domain.MustNewMoney(10, domain.AUD)
	p, err := domain.NewProduct(domain.ProductID(id), name, price, "test")
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestProductRepo_SaveAndFind(t *testing.T) {
	repo := NewInMemoryProductRepository()
	ctx := context.Background()

	p := makeProduct(t, "p1", "Widget")
	repo.Save(ctx, p)

	found, err := repo.FindByID(ctx, "p1")
	if err != nil {
		t.Fatal(err)
	}
	if found.Name() != "Widget" {
		t.Errorf("Name = %q", found.Name())
	}
}

func TestProductRepo_FindByCategory(t *testing.T) {
	repo := NewInMemoryProductRepository()
	ctx := context.Background()

	p1 := makeProduct(t, "p1", "A")
	p2 := makeProduct(t, "p2", "B")

	repo.Save(ctx, p1)
	repo.Save(ctx, p2)

	products, err := repo.FindByCategory(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 2 {
		t.Errorf("expected 2, got %d", len(products))
	}
}

func TestProductRepo_ConcurrentAccess(t *testing.T) {
	repo := NewInMemoryProductRepository()
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			p := makeProduct(t, fmt.Sprintf("p-%d", idx), fmt.Sprintf("Product %d", idx))
			repo.Save(ctx, p)
		}(i)
	}
	wg.Wait()

	all, _ := repo.FindAll(ctx)
	if len(all) != 100 {
		t.Errorf("expected 100 products, got %d", len(all))
	}
}

func init() {}
