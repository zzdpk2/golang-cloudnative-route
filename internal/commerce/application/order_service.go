package application

import (
	"context"
	"encoding/json"
	"fmt"
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"io"
)

// ============================================================
// Order Use Cases
//
// Orchestrate the domain, carry request scope through context, and decode or encode DTOs.
// ============================================================

// ---- Context Keys ----

// contextKey is unexported, and that is the whole trick.
//
// context.WithValue keys are compared by both type and value, so a key of an
// unexported type in this package cannot collide with a key from any other
// package, even one using the identical string. Using a bare string as a key —
// which is what the vet tool warns about — puts every package in the process
// into one shared namespace.
type contextKey string

const (
	ContextKeyUserID    contextKey = "user_id"
	ContextKeyRequestID contextKey = "request_id"
)

func WithUserID(ctx context.Context, userID string) context.Context { panic("TODO") }

// UserIDFromContext reads the user id, reporting whether it was present.
//
// ctx.Value returns an `any`, so it can be absent *or* present holding
// something that is not a string. Both have to come back as "not found" — a
// type assertion without the comma-ok form would panic on the second case, and
// it is reachable whenever two packages disagree about what a key holds.
//
// A note on the pattern as a whole: values in a context are invisible to the
// compiler, so a missing one fails at run time in a place far from the mistake.
// Use it for request-scoped ambient data like this — ids for tracing — and pass
// anything the function genuinely needs as a parameter instead.
func UserIDFromContext(ctx context.Context) (string, bool) {
	panic("TODO")
}

func WithRequestID(ctx context.Context, reqID string) context.Context { panic("TODO") }

// RequestIDFromContext reads the request id, reporting whether it was present.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	panic("TODO")
}

// EventPublisher is defined *here*, by the consumer, rather than in the
// eventbus package that implements it.
//
// That is deliberate and it is the rule that keeps the dependency arrows
// pointing inward: the application layer states what it needs, and the adapter
// bends to fit. If this interface lived in the eventbus package, the
// application would have to import infrastructure, and the layering would be
// inverted.
//
// It also happens to be why the tests can substitute a recording publisher
// without the eventbus package existing at all.
type EventPublisher interface {
	Publish(evt domain.DomainEvent)
	PublishAll(events []domain.DomainEvent)
}

type CreateOrderCommand struct {
	CustomerID string         `json:"customer_id"`
	Shipping   AddressDTO     `json:"shipping"`
	Lines      []OrderLineDTO `json:"lines"`
}

type AddressDTO struct {
	Street   string `json:"street"`
	City     string `json:"city"`
	State    string `json:"state"`
	Postcode string `json:"postcode"`
	Country  string `json:"country"`
}

type OrderLineDTO struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`
	Quantity  int     `json:"quantity"`
}

// ---- Order Application Service ----

type OrderService struct {
	orderRepo   domain.OrderRepository
	productRepo domain.ProductRepository
	publisher   EventPublisher
}

func NewOrderService(
	orderRepo domain.OrderRepository,
	productRepo domain.ProductRepository,
	publisher EventPublisher,
) *OrderService {
	panic("TODO")
}

// CreateOrder turns a command into a persisted order and publishes whatever
// events the aggregate recorded.
//
// The shape of every use case in this layer:
//
//	translate input → load what you need → let the domain decide →
//	persist → publish
//
// The critical constraint is the third step. This method must contain **no
// business rules**. "Only pending orders may be modified", "lines must share a
// currency", "an order needs at least one line" — every one of those lives in
// the domain. If you find yourself writing an `if` here that a
// businessperson would recognise as a rule, it is in the wrong file.
//
// Translating the DTOs is real work, though: strings become domain.Currency,
// float64 becomes domain.Money, and every one of those conversions can fail.
// Decide whether to stop at the first bad line or report all of them, and
// recall what policy.ValidateAll had to say about reporting one error at a time.
//
// Two ordering questions the tests care about. Should the order be saved before
// or after the events go out — and what happens if publishing fails once the
// save has already committed? And should CollectEvents run before or after Save?
func (s *OrderService) CreateOrder(ctx context.Context, cmd CreateOrderCommand) (*domain.Order, error) {
	panic("TODO")
}

// ConfirmOrder loads an order, confirms it, saves it, and publishes.
//
// Note that this method does not decide whether confirming is allowed. It asks
// the aggregate and reports the answer. That is the entire job description of
// an application service.
func (s *OrderService) ConfirmOrder(ctx context.Context, orderID string) error {
	panic("TODO")
}

// GetOrder loads an order by id.
//
// A read with no rules attached. Worth asking whether it earns its place here
// at all, or whether the transport layer should reach the repository directly
// for pure reads. That question is what CQRS is an answer to, and there are
// honest arguments on both sides.
func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	panic("TODO")
}

// CreateOrderFromJSON decodes a command from a stream and creates the order.
//
// Taking an io.Reader rather than a []byte means this works over an HTTP body,
// a file, or a test's strings.Reader without knowing the difference. That is
// the standard library's favourite abstraction and it is worth internalising.
//
// Decoding untrusted input has sharp edges: a malformed body, a truncated
// stream, unknown fields, and a body large enough to exhaust memory are all
// reachable from the network. Decide which of those this method defends against
// and which belong to the HTTP layer in L7 — but decide, rather than letting
// them fall between the two.
func (s *OrderService) CreateOrderFromJSON(ctx context.Context, r io.Reader) (*domain.Order, error) {
	panic("TODO")
}

// ExportOrderJSON loads an order and writes it to w as JSON.
//
// Encoding straight to the writer rather than building a []byte first is the
// idiomatic choice, and it has a consequence worth knowing: by the time the
// encoder returns an error it may already have written a partial document, and
// the HTTP status line is long gone. That is a general hazard of streaming
// responses, not a quirk of this method.
func (s *OrderService) ExportOrderJSON(ctx context.Context, orderID string, w io.Writer) error {
	panic("TODO")
}

// Placeholders that keep the imports valid while the stubs above are unimplemented.
// Remove each one once the corresponding package is genuinely used.
var (
	_ = fmt.Sprintf
	_ = json.NewEncoder
	_ domain.ProductID
	_ domain.Money
)
