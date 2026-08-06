// Package workflow is the L11 exercise: the same order-fulfillment saga you
// hand-rolled in L9 and L10, handed to a workflow engine instead.
//
// Before writing a line here, go read your own txdsl and saga code. This
// package is only worth doing as a comparison — the question it answers is
// "what did I get, and what did it cost me?"
//
// What disappears: retry loops, backoff, timeout bookkeeping, persisting where
// the process got to, and the code that resumes it after a crash. Temporal does
// all of that from the event history.
//
// What it costs: the workflow function must be *deterministic*. It gets replayed
// from the beginning every time the worker recovers, and replay has to take the
// same branches it took the first time. So inside a workflow function you may
// not call time.Now, rand, os.Getenv, or do IO; you may not iterate a map and
// act on the order; you may not start a bare goroutine. Everything
// non-deterministic goes in an Activity, or through the workflow.* equivalents
// (workflow.Now, workflow.NewTimer, workflow.SideEffect).
//
// The tests use go.temporal.io/sdk/testsuite, which runs a workflow in-process
// with a mocked clock. You do not need a Temporal server, and timers fire
// instantly rather than actually waiting.
package workflows

import (
	"context"
	"errors"
	"time"

	"go.temporal.io/sdk/workflow"
)

// Business errors the activities report. They are plain errors on purpose: part
// of this exercise is noticing which failures Temporal should retry and which
// it must not.
var (
	ErrPaymentDeclined = errors.New("payment declined")
	ErrOutOfStock      = errors.New("out of stock")
	ErrCarrierRejected = errors.New("carrier rejected the shipment")
)

// FulfillmentRequest is the workflow input.
type FulfillmentRequest struct {
	OrderID    string
	CustomerID string
	SKU        string
	Quantity   int
	AmountCent int64
}

// FulfillmentResult is the workflow output.
type FulfillmentResult struct {
	OrderID       string
	PaymentID     string
	ReservationID string
	TrackingID    string
	// Compensated lists the rollbacks that ran, in the order they ran.
	Compensated []string
}

// Status values reported by the "status" query handler.
const (
	StatusStarted   = "started"
	StatusPaid      = "paid"
	StatusReserved  = "reserved"
	StatusShipped   = "shipped"
	StatusRefunded  = "refunded"
	StatusCancelled = "cancelled"
)

// CancelSignal is the signal name a human uses to abort a fulfillment that is
// waiting for confirmation.
const CancelSignal = "cancel"

// ---- Activities ----
//
// Activities are ordinary functions. They may do IO, they may fail, and
// Temporal retries them according to the policy the workflow attaches. Keep
// them idempotent: a retry must not charge the card twice.

// ChargePayment charges the customer and returns a payment id.
//
// In this exercise the body can be a stub that returns a deterministic id — the
// tests mock the activities anyway. What matters is the signature and that the
// workflow calls it correctly.
func ChargePayment(ctx context.Context, req FulfillmentRequest) (string, error) {
	panic("TODO")
}

// RefundPayment is the compensation for ChargePayment.
func RefundPayment(ctx context.Context, paymentID string) error {
	panic("TODO")
}

// ReserveStock holds inventory and returns a reservation id.
func ReserveStock(ctx context.Context, req FulfillmentRequest) (string, error) {
	panic("TODO")
}

// ReleaseStock is the compensation for ReserveStock.
func ReleaseStock(ctx context.Context, reservationID string) error {
	panic("TODO")
}

// ShipOrder hands the parcel to a carrier and returns a tracking id.
func ShipOrder(ctx context.Context, req FulfillmentRequest) (string, error) {
	panic("TODO")
}

// ---- Workflow ----

// FulfillmentWorkflow charges, reserves, and ships, compensating in reverse on
// failure — the same shape as your txdsl chain, expressed for the engine.
//
// What you have to wire up:
//
//   - activity options. Nothing runs without a StartToCloseTimeout; the SDK
//     rejects the call. Attach them with workflow.WithActivityOptions and think
//     about which activities deserve a RetryPolicy and which must not be
//     retried at all. A declined card will be declined again.
//
//   - the calls themselves, via workflow.ExecuteActivity(ctx, Fn, args) and
//     .Get(ctx, &out). Note that Get blocks the workflow coroutine without
//     blocking a thread.
//
//   - compensation. On failure, undo what succeeded, in reverse, and record the
//     names in FulfillmentResult.Compensated. You already decided the hard part
//     of this in txdsl: which context do the compensations run under when the
//     workflow itself is being cancelled? Temporal's answer is
//     workflow.NewDisconnectedContext — go read why it exists and compare it
//     with the answer you invented.
//
//   - a "status" query handler via workflow.SetQueryHandler, updated as the
//     workflow advances. This is the thing L10's hand-rolled saga could not do
//     at all: ask a running instance where it is.
//
//   - a cancel signal. Take workflow.GetSignalChannel(ctx, CancelSignal) and
//     let a human abort before shipping. Selecting on a signal channel and an
//     activity future at the same time needs workflow.NewSelector.
//
// Start with the happy path and one activity. Get that green, then add the
// second, then compensation, then the query, then the signal. Do not try to
// write the whole thing before running a test.
func FulfillmentWorkflow(ctx workflow.Context, req FulfillmentRequest) (FulfillmentResult, error) {
	panic("TODO")
}

// defaultActivityOptions is a starting point you are meant to argue with.
//
// Is 30 seconds right for a card charge? Should a shipment retry 3 times or
// forever? Which of these belongs on the workflow and which on each activity?
var defaultActivityOptions = workflow.ActivityOptions{
	StartToCloseTimeout: 30 * time.Second,
}
