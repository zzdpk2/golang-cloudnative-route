package application

import (
	"bytes"
	"context"
	"encoding/json"
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"sync"
	"testing"

	"github.com/rex/go-ddd-tdd/internal/commerce/persistence"
)

// ---- Mock Event Publisher ----

type MockPublisher struct {
	mu     sync.Mutex
	Events []domain.DomainEvent
}

func (m *MockPublisher) Publish(evt domain.DomainEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Events = append(m.Events, evt)
}

func (m *MockPublisher) PublishAll(events []domain.DomainEvent) {
	for _, evt := range events {
		m.Publish(evt)
	}
}

func (m *MockPublisher) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Events)
}

func setupService(t *testing.T) (*OrderService, *MockPublisher) {
	t.Helper()
	orderRepo := persistence.NewInMemoryOrderRepository()
	productRepo := persistence.NewInMemoryProductRepository()
	publisher := &MockPublisher{}
	svc := NewOrderService(orderRepo, productRepo, publisher)
	return svc, publisher
}

func TestContextValues(t *testing.T) {
	ctx := context.Background()
	ctx = WithUserID(ctx, "user-123")
	ctx = WithRequestID(ctx, "req-abc")

	uid, ok := UserIDFromContext(ctx)
	if !ok || uid != "user-123" {
		t.Errorf("UserID = %q, %v", uid, ok)
	}

	rid, ok := RequestIDFromContext(ctx)
	if !ok || rid != "req-abc" {
		t.Errorf("RequestID = %q, %v", rid, ok)
	}

	emptyCtx := context.Background()
	_, ok = UserIDFromContext(emptyCtx)
	if ok {
		t.Error("should return false for missing key")
	}
}

func TestOrderService_CreateOrder(t *testing.T) {
	svc, publisher := setupService(t)
	ctx := WithRequestID(context.Background(), "req-001")

	cmd := CreateOrderCommand{
		CustomerID: "cust-001",
		Shipping: AddressDTO{
			Street:   "1 George St",
			City:     "Sydney",
			State:    "NSW",
			Postcode: "2000",
			Country:  "AU",
		},
		Lines: []OrderLineDTO{
			{ProductID: "p1", Name: "Widget", Price: 29.99, Currency: "AUD", Quantity: 2},
			{ProductID: "p2", Name: "Gadget", Price: 49.99, Currency: "AUD", Quantity: 1},
		},
	}

	order, err := svc.CreateOrder(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}

	if order.LineCount() != 2 {
		t.Errorf("LineCount = %d, want 2", order.LineCount())
	}

	if publisher.Count() < 1 {
		t.Error("should have published events")
	}
}

func TestOrderService_ConfirmOrder(t *testing.T) {
	svc, _ := setupService(t)
	ctx := context.Background()

	cmd := CreateOrderCommand{
		CustomerID: "cust-001",
		Shipping:   AddressDTO{Street: "1 St", City: "Syd", Postcode: "2000", Country: "AU"},
		Lines: []OrderLineDTO{
			{ProductID: "p1", Name: "Widget", Price: 10, Currency: "AUD", Quantity: 1},
		},
	}

	order, err := svc.CreateOrder(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}

	err = svc.ConfirmOrder(ctx, string(order.ID()))
	if err != nil {
		t.Fatal(err)
	}

	found, err := svc.GetOrder(ctx, string(order.ID()))
	if err != nil {
		t.Fatal(err)
	}
	if found.Status().String() != "confirmed" {
		t.Errorf("status = %s", found.Status())
	}
}

func TestOrderService_CreateOrderFromJSON(t *testing.T) {
	svc, _ := setupService(t)
	ctx := context.Background()

	jsonData := `{
		"customer_id": "cust-json",
		"shipping": {
			"street": "1 JSON St",
			"city": "Sydney",
			"state": "NSW",
			"postcode": "2000",
			"country": "AU"
		},
		"lines": [
			{"product_id": "p1", "name": "From JSON", "price": 15.50, "currency": "AUD", "quantity": 3}
		]
	}`
	reader := bytes.NewBufferString(jsonData)

	order, err := svc.CreateOrderFromJSON(ctx, reader)
	if err != nil {
		t.Fatal(err)
	}

	if order.LineCount() != 1 {
		t.Errorf("LineCount = %d", order.LineCount())
	}
}

func TestOrderService_ExportOrderJSON(t *testing.T) {
	svc, _ := setupService(t)
	ctx := context.Background()

	cmd := CreateOrderCommand{
		CustomerID: "cust-001",
		Shipping:   AddressDTO{Street: "1 St", City: "Syd", Postcode: "2000", Country: "AU"},
		Lines: []OrderLineDTO{
			{ProductID: "p1", Name: "Widget", Price: 10, Currency: "AUD", Quantity: 1},
		},
	}
	order, _ := svc.CreateOrder(ctx, cmd)

	var buf bytes.Buffer
	err := svc.ExportOrderJSON(ctx, string(order.ID()), &buf)
	if err != nil {
		t.Fatal(err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Errorf("output is not valid JSON: %s", buf.String())
	}
	t.Logf("Exported JSON: %s", buf.String())
}
