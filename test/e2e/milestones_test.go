// Package e2e holds the milestone tests.
//
// The package tests scattered through internal/ ask "is this function right?".
// These ask a different question: "can the system now do a thing it could not
// do before?" Each milestone corresponds to a gate in docs/LEARNING_PATH.md, and the
// milestone going green is the real signal that the level is finished.
//
// Run one at a time, in order:
//
//	go test ./test/e2e -run TestM1 -v
//	go test ./test/e2e -run TestM2 -v
//	...
//
// They will all panic until the levels they depend on are implemented. That is
// expected — the milestone is the target you are aiming at, and reading the
// failure is how you find out what is still missing.
//
// Never edit a milestone to make it pass. It describes the system you are
// building; if it disagrees with your code, your code is what changes.
package e2e

import (
	"context"
	"encoding/json"
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rex/go-ddd-tdd/internal/commerce/application"
	"github.com/rex/go-ddd-tdd/internal/commerce/eventbus"
	"github.com/rex/go-ddd-tdd/internal/commerce/persistence"

	httpadapter "github.com/rex/go-ddd-tdd/internal/commerce/http"
)

// ---- M1 (L1): value objects hold the line ----

// The smallest useful thing the system can do: price a line of an order without
// losing a cent and without mixing currencies.
func TestM1_ValueObjectsPriceALine(t *testing.T) {
	price, err := domain.NewMoney(19.99, domain.AUD)
	if err != nil {
		t.Fatalf("NewMoney() = %v", err)
	}
	qty, err := domain.NewQuantity(3)
	if err != nil {
		t.Fatalf("NewQuantity() = %v", err)
	}

	total := price.Multiply(qty.Value())
	if got, want := total.Cents(), int64(5997); got != want {
		t.Errorf("3 x A$19.99 = %d cents, want %d", got, want)
	}
	if got, want := total.String(), "A$59.97"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}

	// Mixing currencies must not be possible, even by accident.
	usd, _ := domain.NewMoney(1, domain.USD)
	if _, err := total.Add(usd); err == nil {
		t.Error("adding USD to AUD succeeded, want ErrCurrencyMismatch")
	}
}

// Splitting money across sub-orders must not create or destroy any.
func TestM1_MoneySplitsWithoutLoss(t *testing.T) {
	total, _ := domain.NewMoney(100, domain.AUD)

	parts, err := total.Allocate(3)
	if err != nil {
		t.Fatalf("Allocate() = %v", err)
	}

	var sum int64
	for _, p := range parts {
		sum += p.Cents()
	}
	if sum != total.Cents() {
		t.Errorf("parts sum to %d cents, want %d: the split lost money", sum, total.Cents())
	}
}

// ---- M2 (L2): the aggregate protects itself ----

func newTestAddress(t *testing.T) domain.Address {
	t.Helper()
	addr, err := domain.NewAddress("1 George St", "Sydney", "NSW", "2000", "AU")
	if err != nil {
		t.Fatalf("NewAddress() = %v", err)
	}
	return addr
}

func addLine(t *testing.T, o *domain.Order, sku, name string, price float64, qty int) {
	t.Helper()
	p, err := domain.NewMoney(price, domain.AUD)
	if err != nil {
		t.Fatalf("NewMoney() = %v", err)
	}
	q, err := domain.NewQuantity(qty)
	if err != nil {
		t.Fatalf("NewQuantity() = %v", err)
	}
	if err := o.AddLine(domain.ProductID(sku), name, p, q); err != nil {
		t.Fatalf("AddLine() = %v", err)
	}
}

// A whole order lifecycle, and the invariants that hold throughout it.
func TestM2_OrderLifecycle(t *testing.T) {
	order, err := domain.NewOrder("ord-1", "cust-1", newTestAddress(t))
	if err != nil {
		t.Fatalf("NewOrder() = %v", err)
	}
	if order.Status() != domain.OrderPending {
		t.Fatalf("new order status = %v, want pending", order.Status())
	}

	addLine(t, order, "sku-1", "Widget", 19.99, 2)
	addLine(t, order, "sku-2", "Gadget", 5.00, 1)

	total, err := order.TotalAmount()
	if err != nil {
		t.Fatalf("TotalAmount() = %v", err)
	}
	if got, want := total.Cents(), int64(4498); got != want {
		t.Errorf("TotalAmount() = %d cents, want %d", got, want)
	}

	for _, step := range []struct {
		name string
		fn   func() error
	}{
		{"Confirm", order.Confirm},
		{"Ship", order.Ship},
		{"Deliver", order.Deliver},
	} {
		if err := step.fn(); err != nil {
			t.Fatalf("%s() = %v", step.name, err)
		}
	}
	if order.Status() != domain.OrderDelivered {
		t.Errorf("status = %v, want delivered", order.Status())
	}

	// A delivered order is history. It cannot be cancelled.
	if err := order.Cancel(); err == nil {
		t.Error("Cancel() on a delivered order = nil, want an error")
	}
}

// The aggregate must not hand out anything a caller can use to break it.
func TestM2_AggregateResistsTampering(t *testing.T) {
	order, err := domain.NewOrder("ord-2", "cust-1", newTestAddress(t))
	if err != nil {
		t.Fatalf("NewOrder() = %v", err)
	}
	addLine(t, order, "sku-1", "Widget", 10, 1)

	lines := order.Lines()
	if len(lines) != 1 {
		t.Fatalf("Lines() returned %d lines, want 1", len(lines))
	}
	lines[0].Name = "tampered"
	lines = append(lines, lines[0])
	_ = lines

	if again := order.Lines(); again[0].Name != "Widget" {
		t.Errorf("line name = %q after the caller mutated a previous result, want %q",
			again[0].Name, "Widget")
	}
	if order.LineCount() != 1 {
		t.Errorf("LineCount() = %d, want 1: the caller's append reached internal state",
			order.LineCount())
	}
}

// Events are recorded as the order changes, and collected exactly once.
func TestM2_EventsAreCollectedOnce(t *testing.T) {
	order, err := domain.NewOrder("ord-3", "cust-1", newTestAddress(t))
	if err != nil {
		t.Fatalf("NewOrder() = %v", err)
	}
	addLine(t, order, "sku-1", "Widget", 10, 1)

	first := order.CollectEvents()
	if len(first) == 0 {
		t.Fatal("CollectEvents() returned nothing, want at least the creation event")
	}
	if second := order.CollectEvents(); len(second) != 0 {
		t.Errorf("second CollectEvents() returned %d events, want 0: "+
			"collecting must drain the buffer or every event is published twice", len(second))
	}
}

// ---- M3 (L4): the use case orchestrates ----

// recordingPublisher captures what the application layer publishes.
type recordingPublisher struct {
	mu     sync.Mutex
	events []domain.DomainEvent
}

func (p *recordingPublisher) Publish(evt domain.DomainEvent) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = append(p.events, evt)
}

func (p *recordingPublisher) PublishAll(events []domain.DomainEvent) {
	for _, e := range events {
		p.Publish(e)
	}
}

func (p *recordingPublisher) collected() []domain.DomainEvent {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]domain.DomainEvent(nil), p.events...)
}

func newService(t *testing.T) (*application.OrderService, *recordingPublisher) {
	t.Helper()
	pub := &recordingPublisher{}
	svc := application.NewOrderService(
		persistence.NewInMemoryOrderRepository(),
		persistence.NewInMemoryProductRepository(),
		pub,
	)
	return svc, pub
}

func sampleCommand() application.CreateOrderCommand {
	return application.CreateOrderCommand{
		CustomerID: "cust-1",
		Shipping: application.AddressDTO{
			Street: "1 George St", City: "Sydney", State: "NSW",
			Postcode: "2000", Country: "AU",
		},
		Lines: []application.OrderLineDTO{
			{ProductID: "sku-1", Name: "Widget", Price: 19.99, Currency: "AUD", Quantity: 2},
		},
	}
}

func TestM3_CreateOrderUseCase(t *testing.T) {
	svc, pub := newService(t)
	ctx := application.WithRequestID(context.Background(), "req-1")

	order, err := svc.CreateOrder(ctx, sampleCommand())
	if err != nil {
		t.Fatalf("CreateOrder() = %v", err)
	}
	if order.Status() != domain.OrderPending {
		t.Errorf("status = %v, want pending", order.Status())
	}
	if len(pub.collected()) == 0 {
		t.Error("no events published: the use case must publish what the aggregate recorded")
	}
}

func TestM3_RejectsEmptyOrder(t *testing.T) {
	svc, _ := newService(t)

	cmd := sampleCommand()
	cmd.Lines = nil

	if _, err := svc.CreateOrder(context.Background(), cmd); err == nil {
		t.Error("CreateOrder() with no lines = nil, want an error")
	}
}

// ---- M4 (L6): it survives a round trip through storage ----

func TestM4_OrderRoundTripsThroughRepository(t *testing.T) {
	repo := persistence.NewInMemoryOrderRepository()
	ctx := context.Background()

	order, err := domain.NewOrder("ord-round", "cust-1", newTestAddress(t))
	if err != nil {
		t.Fatalf("NewOrder() = %v", err)
	}
	addLine(t, order, "sku-1", "Widget", 12.50, 4)

	if err := repo.Save(ctx, order); err != nil {
		t.Fatalf("Save() = %v", err)
	}

	loaded, err := repo.FindByID(ctx, "ord-round")
	if err != nil {
		t.Fatalf("FindByID() = %v", err)
	}
	if loaded.LineCount() != 1 {
		t.Errorf("LineCount() = %d, want 1", loaded.LineCount())
	}

	total, err := loaded.TotalAmount()
	if err != nil {
		t.Fatalf("TotalAmount() = %v", err)
	}
	if got, want := total.Cents(), int64(5000); got != want {
		t.Errorf("TotalAmount() = %d cents, want %d", got, want)
	}
}

// Concurrent writers must not corrupt the store. Run this milestone with -race.
func TestM4_RepositoryIsConcurrencySafe(t *testing.T) {
	repo := persistence.NewInMemoryOrderRepository()
	ctx := context.Background()

	const writers = 50
	var wg sync.WaitGroup
	for i := range writers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := domain.OrderID("ord-" + string(rune('a'+i%26)) + string(rune('0'+i/26)))
			order, err := domain.NewOrder(id, "cust-1", newTestAddress(t))
			if err != nil {
				return
			}
			_ = repo.Save(ctx, order)
			_, _ = repo.FindByID(ctx, id)
		}(i)
	}
	wg.Wait()
}

// ---- M5 (L8): events reach their subscribers ----

func TestM5_EventBusDeliversToAllSubscribers(t *testing.T) {
	bus := eventbus.NewInMemoryEventBus(16)

	var mu sync.Mutex
	seen := map[string]int{}
	record := func(name string) eventbus.EventHandler {
		return func(domain.DomainEvent) {
			mu.Lock()
			defer mu.Unlock()
			seen[name]++
		}
	}

	bus.Subscribe("order.created", record("inventory"))
	bus.Subscribe("order.created", record("notifications"))

	order, err := domain.NewOrder("ord-evt", "cust-1", newTestAddress(t))
	if err != nil {
		t.Fatalf("NewOrder() = %v", err)
	}
	for _, e := range order.CollectEvents() {
		bus.Publish(e)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		done := seen["inventory"] > 0 && seen["notifications"] > 0
		mu.Unlock()
		if done {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	t.Errorf("subscribers saw %v, want both to receive order.created", seen)
}

// ---- M6 (L7): the service is alive ----

// This is the milestone that matters most. When it goes green you can start the
// binary and place an order over HTTP — the system stops being an exercise and
// becomes a service.
func TestM6_PlaceAnOrderOverHTTP(t *testing.T) {
	svc, _ := newService(t)
	server := httpadapter.NewServer(":0", svc)
	handler := server.Handler()

	body, err := json.Marshal(sampleCommand())
	if err != nil {
		t.Fatalf("marshal command: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /orders = %d, want %d. Body: %s", rec.Code, http.StatusCreated, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q, want JSON", ct)
	}

	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode response: %v (body %s)", err, rec.Body)
	}
	if created.ID == "" {
		t.Fatal("response carried no order id")
	}

	// And it can be read back.
	getReq := httptest.NewRequest(http.MethodGet, "/orders/"+created.ID, nil)
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("GET /orders/%s = %d, want %d", created.ID, getRec.Code, http.StatusOK)
	}
}

func TestM6_ConfirmOrderOverHTTP(t *testing.T) {
	svc, _ := newService(t)
	handler := httpadapter.NewServer(":0", svc).Handler()

	body, _ := json.Marshal(sampleCommand())
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(string(body))))
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /orders = %d, want 201", rec.Code)
	}
	var created struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &created)

	confirmRec := httptest.NewRecorder()
	handler.ServeHTTP(confirmRec,
		httptest.NewRequest(http.MethodPost, "/orders/"+created.ID+"/confirm", nil))

	if confirmRec.Code != http.StatusOK {
		t.Errorf("POST /orders/%s/confirm = %d, want %d",
			created.ID, confirmRec.Code, http.StatusOK)
	}
}

// ---- M7 (L4 + L7): failures arrive as the right status code ----

// A domain error must not surface as a 500. The mapping from error to status is
// the seam between the domain and the transport, and getting it right is what
// makes the API usable by anyone else.
func TestM7_ErrorsMapToStatusCodes(t *testing.T) {
	svc, _ := newService(t)
	handler := httpadapter.NewServer(":0", svc).Handler()

	tests := []struct {
		name string
		req  func() *http.Request
		want int
	}{
		{
			name: "unknown order is 404",
			req: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/orders/does-not-exist", nil)
			},
			want: http.StatusNotFound,
		},
		{
			name: "malformed JSON is 400",
			req: func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader("{not json"))
				r.Header.Set("Content-Type", "application/json")
				return r
			},
			want: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, tt.req())
			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d. Body: %s", rec.Code, tt.want, rec.Body)
			}
		})
	}
}
