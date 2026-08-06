package http

import (
	stdhttp "net/http"
)

// ErrorResponse is the single error shape this API returns.
//
// One shape for every failure is worth more than it looks: a client can write
// one error path instead of guessing per endpoint. Code is the machine-readable
// half — the same split as platform/errors.ErrorCode, carried out to the wire.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Request string `json:"request_id,omitempty"`
}

// OrderResponse is the wire shape of an order.
//
// Note how little it exposes. The aggregate has lines, timestamps, events, and
// a status machine; the API publishes four fields. Every field added here is a
// field a client may come to depend on and you can no longer change — which is
// why the safe direction is to start narrow.
type OrderResponse struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	Status     string `json:"status"`
	LineCount  int    `json:"line_count"`
}

// writeJSON sends a JSON response.
//
// Header, then status, then body — in that order, and it is not negotiable.
// WriteHeader flushes the header block, so a Content-Type set afterwards never
// arrives; and the first Write implies WriteHeader(200), so writing the body
// first discards whatever status you meant to send.
//
// If encoding fails halfway the status and part of the body are already gone.
// There is no recovery, only logging.
func writeJSON(w stdhttp.ResponseWriter, status int, v any) {
	panic("TODO")
}

// writeError sends an ErrorResponse.
//
// Whatever reaches the client here is public. An error from the database layer
// may carry a connection string, a table name, or a query — pass err.Error()
// straight through and you have handed all of it to an anonymous caller. Decide
// what is safe to say and log the rest.
func writeError(w stdhttp.ResponseWriter, status int, code string, err error) {
	panic("TODO")
}

// decodeJSON reads a JSON request body into dst.
//
// Three decisions, and the tests only force the first:
//
//   - unknown fields. Rejecting them turns a client's typo into a 400 rather
//     than a silently ignored value. Good for an internal API, hostile for a
//     public one, since it makes adding a field breaking.
//   - body size. Without http.MaxBytesReader, a large enough body is an
//     out-of-memory kill from an unauthenticated caller.
//   - trailing data. A decoder stops at the end of the first JSON value, so
//     `{"a":1}{"b":2}` decodes happily and ignores the rest.
//
// Close the body when you are done.
func decodeJSON(r *stdhttp.Request, dst any) error {
	panic("TODO")
}

// toOrderResponse converts a domain order into its wire shape.
//
// Look at the parameter type: an inline anonymous interface. It is an unusual
// thing to write, and it means this function accepts anything with that one
// method rather than depending on the aggregate package.
//
// Decide whether that is elegant or over-clever. The narrow dependency is real,
// but an anonymous interface cannot be named, documented, or implemented
// deliberately — and the next field you need forces you to change it here and
// at every call site.
func toOrderResponse(order interface {
	ID() interface{ String() string }
}) OrderResponse {
	panic("TODO")
}

// handleCreateOrder serves POST /orders.
//
// Decode, hand to the application service, respond 201 with the created order.
//
// The layering rule from L4 applies just as strictly in reverse: this handler
// must contain no business rules. It translates HTTP into a command, and a
// result back into HTTP. Any `if` here that a businessperson would recognise
// belongs in the domain.
//
// 201 Created, not 200. And a Location header pointing at the new resource is
// what the spec expects, even though nothing in the tests will make you do it.
func (s *Server) handleCreateOrder(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	panic("TODO")
}

// handleGetOrder serves GET /orders/{id}.
//
// r.PathValue reads the wildcard, which needs Go 1.22's routing patterns — the
// reason this file can avoid a router dependency entirely.
//
// The interesting work is the error mapping, and the e2e milestone M7 checks
// it: a missing order is 404, not 500. That mapping is the seam between the
// domain and the transport. errors.Is against the repository's sentinel is what
// makes it possible, which is why platform/errors bothered with wrapping.
func (s *Server) handleGetOrder(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	panic("TODO")
}

// handleConfirmOrder serves POST /orders/{id}/confirm.
//
// Three different failures, three different statuses, and getting them right is
// the exercise:
//
//	the order does not exist        404
//	the body or id is malformed     400
//	the order cannot be confirmed
//	  from its current status       409 or 422
//
// That last one is a genuine judgement call — 409 Conflict says "the state
// disagrees", 422 says "the request was understood but unprocessable". Pick
// one, apply it everywhere, and write down why.
//
// Note the URL shape: a POST to a sub-path rather than a PATCH with a status
// field. Confirming is a *transition*, not a field assignment, and modelling it
// as one keeps the state machine on the server where it belongs.
func (s *Server) handleConfirmOrder(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	panic("TODO")
}

// optionsHandler answers CORS preflight requests.
//
// A browser sends OPTIONS before any non-simple cross-origin request and will
// not proceed unless this answers correctly. It returns no body — 204 No
// Content is the right status, and the headers are the entire response.
func optionsHandler(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	panic("TODO")
}
