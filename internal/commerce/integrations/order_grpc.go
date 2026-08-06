package integrations

import (
	"context"
)

// ============================================================
// gRPC-Style Contract
//
// A hand-rolled stand-in for generated stubs; there is no protobuf codegen here.
// ============================================================

type CreateOrderRequest struct {
	CustomerID string      `json:"customer_id"`
	Shipping   Address     `json:"shipping"`
	Lines      []OrderLine `json:"lines"`
}

type Address struct {
	Street   string `json:"street"`
	City     string `json:"city"`
	State    string `json:"state"`
	Postcode string `json:"postcode"`
	Country  string `json:"country"`
}

type OrderLine struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`
	Quantity  int32   `json:"quantity"`
}

type GetOrderRequest struct {
	OrderID string `json:"order_id"`
}

type ConfirmOrderRequest struct {
	OrderID string `json:"order_id"`
}

type WatchOrdersRequest struct {
	CustomerID string `json:"customer_id"`
}

type OrderReply struct {
	OrderID    string `json:"order_id"`
	CustomerID string `json:"customer_id"`
	Status     string `json:"status"`
	LineCount  int32  `json:"line_count"`
}

type BulkCreateReply struct {
	Created int32        `json:"created"`
	Orders  []OrderReply `json:"orders"`
}

type OrderEvent struct {
	OrderID     string `json:"order_id"`
	Type        string `json:"type"`
	PayloadJSON string `json:"payload_json"`
}

type OrderServiceServer interface {
	CreateOrder(context.Context, *CreateOrderRequest) (*OrderReply, error)
	GetOrder(context.Context, *GetOrderRequest) (*OrderReply, error)
	ConfirmOrder(context.Context, *ConfirmOrderRequest) (*OrderReply, error)
	WatchOrders(*WatchOrdersRequest, OrderService_WatchOrdersServer) error
	BulkCreate(OrderService_BulkCreateServer) error
	Chat(OrderService_ChatServer) error
}

type ServerStream[T any] interface {
	Send(*T) error
	Context() context.Context
}

type ClientStream[T any] interface {
	Recv() (*T, error)
	Context() context.Context
}

type BidiStream[In any, Out any] interface {
	Send(*Out) error
	Recv() (*In, error)
	Context() context.Context
}

type OrderService_WatchOrdersServer interface {
	ServerStream[OrderEvent]
}

type OrderService_BulkCreateServer interface {
	ClientStream[CreateOrderRequest]
	SendAndClose(*BulkCreateReply) error
}

type OrderService_ChatServer interface {
	BidiStream[OrderEvent, OrderEvent]
}

// MarshalProtoLike stands in for protobuf serialisation.
//
//	MarshalProtoLike(CreateOrderRequest{CustomerID: "cust-1"})
//	  → bytes that UnmarshalProtoLike can turn back into the same value
//
// **It is JSON underneath, and that is a lie worth understanding.** Real
// protobuf is a binary format with a schema: fields are identified by *number*,
// not by name, which is what makes it compact and what makes renaming a field a
// non-breaking change. JSON has neither property.
//
// The consequence for you: the struct tags on the request types above are json
// tags doing a protobuf tag's job. In a real service that file would be
// generated from a .proto and you would never edit it.
func MarshalProtoLike(v any) ([]byte, error) {
	panic("TODO")
}

// UnmarshalProtoLike is the inverse.
//
// Decide what an unknown field should do. Protobuf ignores fields it does not
// recognise — that is deliberate, and it is what lets an old client talk to a
// new server. Rejecting them would break exactly the compatibility the format
// exists to provide.
func UnmarshalProtoLike(data []byte, v any) error {
	panic("TODO")
}

type StatusCode int

const (
	CodeOK StatusCode = iota
	CodeCanceled
	CodeUnknown
	CodeInvalidArgument
	CodeDeadlineExceeded
	CodeNotFound
	CodeAlreadyExists
	CodeInternal
)

// String renders the code in gRPC's SCREAMING_SNAKE convention.
//
//	CodeOK              → "OK"
//	CodeNotFound        → "NOT_FOUND"
//	CodeInvalidArgument → "INVALID_ARGUMENT"
//	StatusCode(99)      → "UNKNOWN"
//
// The out-of-range guard again — third enum in this codebase after
// OrderStatus and Weekday, and the rule has not changed.
func (c StatusCode) String() string {
	panic("TODO")
}

// StatusError carries a code alongside a message.
//
// The same split as platform/errors.DomainError: a machine-readable code and a
// human-readable message. gRPC made it part of the protocol rather than leaving
// it to each service to invent, which is one of the genuinely good decisions in
// its design — every client library can map a code to a retry policy without
// parsing prose.
type StatusError struct {
	Code    StatusCode
	Message string
}

// Error renders "CODE: message".
//
//	NewStatusError(CodeNotFound, "order missing").Error()
//	  → "NOT_FOUND: order missing"
func (e *StatusError) Error() string {
	panic("TODO")
}

// NewStatusError builds a StatusError as an error.
//
// Note the return type is `error`, not `*StatusError`. That is the nil-interface
// trap again — and here it is *safe*, because this constructor never returns
// nil. Convince yourself of the difference between this and entity.CheckProduct
// before moving on; the signature alone does not tell you which you are looking
// at.
func NewStatusError(code StatusCode, msg string) error {
	panic("TODO")
}

// CodeFromError extracts the status code from any error.
//
//	CodeFromError(nil)                              → CodeOK
//	CodeFromError(NewStatusError(CodeNotFound, "")) → CodeNotFound
//	CodeFromError(errors.New("plain"))              → CodeUnknown
//	CodeFromError(fmt.Errorf("wrapped: %w", se))    → the wrapped code
//
// The last line is why this uses errors.As rather than a type assertion: by the
// time an error reaches the transport boundary it has usually been wrapped two
// or three times, and a type assertion sees only the outermost layer.
//
// A plain error becoming UNKNOWN rather than INTERNAL is deliberate: UNKNOWN
// means "this service did not tell me what went wrong", which is exactly true.
func CodeFromError(err error) StatusCode {
	panic("TODO")
}

type Metadata map[string][]string

type metadataKey struct{}

// NewOutgoingContext attaches metadata to a context.
//
//	ctx := NewOutgoingContext(context.Background(), Metadata{"request-id": {"req-1"}})
//	MetadataFromContext(ctx) → that Metadata, true
//
// Metadata is gRPC's header equivalent: request ids, auth tokens, tracing spans
// — everything that travels alongside a call without being part of the message.
//
// metadataKey is an unexported empty struct for the same reason
// application.contextKey is an unexported named string: an unexported type
// cannot collide with any other package's key. Third time this pattern has
// appeared, which is how you know it is the idiom rather than a preference.
func NewOutgoingContext(ctx context.Context, md Metadata) context.Context {
	panic("TODO")
}

// MetadataFromContext reads metadata back out.
//
//	a context with metadata → that Metadata, true
//	a bare context          → nil (or empty), false
//
// Comma-ok, and use it: ctx.Value returns `any`, so an absent value and a value
// of the wrong type both have to come back as "not found". A bare type
// assertion panics on the second case, and that case is reachable whenever two
// packages disagree about what a key holds.
//
// Consider whether to copy the map before returning it. Metadata is a
// map[string][]string, so the caller gets a reference to something other
// goroutines may be reading — the same aliasing question as everywhere else,
// arriving one last time.
func MetadataFromContext(ctx context.Context) (Metadata, bool) {
	panic("TODO")
}
