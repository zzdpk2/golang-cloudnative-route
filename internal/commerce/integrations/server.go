package integrations

import (
	"context"

	"github.com/rex/go-ddd-tdd/internal/commerce/application"
)

// OrderGRPCServer is the gRPC face of the same OrderService the HTTP layer
// wraps.
//
// Two transports over one application service is the point of the whole
// layering. Compare this file with adapter/http side by side: the translation
// differs, the orchestration is identical, and neither transport contains a
// business rule. If you find one here that is not in the HTTP handler, one of
// the two is wrong.
type OrderGRPCServer struct {
	service *application.OrderService
}

func NewOrderGRPCServer(service *application.OrderService) *OrderGRPCServer { panic("TODO") }

// CreateOrder is a **unary** call: one request, one response.
//
//	valid request        → an OrderReply carrying the new order id
//	empty customer id    → a StatusError with CodeInvalidArgument
//	the service fails    → a StatusError carrying the mapped code
//
// Same shape as handleCreateOrder in adapter/http: translate the request into a
// command, call the service, translate the result back. The only real
// difference is that gRPC returns a status code instead of an HTTP status.
//
// Map domain errors to codes the way L7 mapped them to statuses — NotFound to
// CodeNotFound, a conflict to CodeAlreadyExists, a validation failure to
// CodeInvalidArgument. Returning CodeInternal for everything throws away the
// information the client needs to decide whether to retry.
func (s *OrderGRPCServer) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*OrderReply, error) {
	panic("TODO")
}

// GetOrder fetches one order.
//
//	known id   → the reply
//	unknown id → CodeNotFound
//	empty id   → CodeInvalidArgument
//
// The distinction between those last two matters: "you asked wrongly" and "what
// you asked for is not here" lead a client to do different things.
func (s *OrderGRPCServer) GetOrder(ctx context.Context, req *GetOrderRequest) (*OrderReply, error) {
	panic("TODO")
}

// ConfirmOrder moves an order to confirmed.
//
//	pending order        → the updated reply
//	already confirmed    → CodeAlreadyExists, or CodeFailedPrecondition in real
//	                       gRPC — this package has no such code, so pick from
//	                       what exists and note the gap
//	unknown id           → CodeNotFound
func (s *OrderGRPCServer) ConfirmOrder(ctx context.Context, req *ConfirmOrderRequest) (*OrderReply, error) {
	panic("TODO")
}

// WatchOrders is **server streaming**: one request, many responses.
//
//	→ send an OrderEvent for each event, then return nil to close the stream
//
// Returning from this method ends the stream. There is no explicit "close" —
// that is the convention, and it means an early return silently truncates the
// client's data rather than erroring.
//
// Watch the context. A client that disconnects cancels it, and a server that
// keeps sending into a dead stream leaks a goroutine per abandoned client.
// Check stream.Context().Done() between sends — this is the single most common
// streaming bug.
func (s *OrderGRPCServer) WatchOrders(req *WatchOrdersRequest, stream OrderService_WatchOrdersServer) error {
	panic("TODO")
}

// BulkCreate is **client streaming**: many requests, one response.
//
//	→ Recv in a loop until it returns io.EOF, creating an order each time
//	→ then SendAndClose with a BulkCreateReply carrying the count
//
// io.EOF from Recv means "the client has finished sending" and is **not an
// error** — treating it as one is the mistake this API invites. Any other error
// is real.
//
// Decide what happens when order 3 of 5 fails: abort the whole batch, or skip
// it and report a partial count? The reply has a Created field, which hints at
// an answer, but it is your call and it is a genuine business decision.
func (s *OrderGRPCServer) BulkCreate(stream OrderService_BulkCreateServer) error {
	panic("TODO")
}

// Chat is **bidirectional streaming**: many requests, many responses, in any
// order.
//
//	→ Recv an event, Send it straight back, until Recv returns io.EOF
//
// An echo, deliberately: the interesting part is not what it does but that
// sending and receiving are independent. A real bidi handler usually reads in
// one goroutine and writes in another, which raises the question of who closes
// what — and gRPC's answer is that the *client* closes its send direction and
// the server signals completion by returning.
func (s *OrderGRPCServer) Chat(stream OrderService_ChatServer) error {
	panic("TODO")
}

// UnaryHandler is a call, reduced to its essentials so an interceptor can wrap
// it.
//
// `any` in and `any` out, which is what lets one interceptor apply to every
// method — and also what costs you all type safety inside the chain. Real gRPC
// makes exactly this trade, for exactly this reason.
type UnaryHandler func(ctx context.Context, req any) (any, error)

// UnaryInterceptor is gRPC's middleware.
//
// Note the signature: it receives the handler as an *argument* rather than
// returning a wrapped one, so calling `handler` is what continues the chain —
// closer to minigin's Context.Next than to net/http's wrapper style. Fifth
// middleware system in this codebase, and the third distinct shape.
type UnaryInterceptor func(ctx context.Context, req any, handler UnaryHandler) (any, error)

// ChainUnaryInterceptors composes interceptors so the first listed is the
// outermost.
//
//	ChainUnaryInterceptors(handler, logging, auth)
//	  → logging sees the request first and the response last
//
// The same ordering trap as the other four chains, and the same quiet failure:
// nothing errors, the chain simply runs backwards.
//
// Build it by wrapping from the last interceptor to the first, and note that
// each wrap needs its own captured copy of the handler — writing the loop the
// obvious way can produce a chain that calls itself.
func ChainUnaryInterceptors(final UnaryHandler, interceptors ...UnaryInterceptor) UnaryHandler {
	panic("TODO")
}

// DeadlineInterceptor rejects a call whose context is already done.
//
//	live context      → calls the handler
//	cancelled context → CodeDeadlineExceeded or CodeCanceled, handler not called
//
// The test accepts either code, because a context that is done cannot always
// tell you *why* — ctx.Err() distinguishes them, and mapping it properly is
// worth doing even though the test does not force it.
//
// Checking before calling is cheap insurance: a request whose deadline has
// already passed has nobody waiting for the answer, and doing the work anyway
// is load you are inflicting on yourself.
func DeadlineInterceptor(ctx context.Context, req any, handler UnaryHandler) (any, error) {
	panic("TODO")
}

// MetadataInterceptor requires metadata to be present.
//
//	context with metadata → calls the handler
//	context without       → CodeInvalidArgument, handler not called
//
// This is where auth tokens and request ids are validated in a real service,
// which is why it belongs near the outside of the chain — everything after it
// gets to assume the caller identified itself.
func MetadataInterceptor(ctx context.Context, req any, handler UnaryHandler) (any, error) {
	panic("TODO")
}

// Compile-time proof that the server implements the full service contract.
var _ OrderServiceServer = (*OrderGRPCServer)(nil)
