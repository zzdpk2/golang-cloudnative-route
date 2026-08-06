package workflows

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"go.temporal.io/sdk/testsuite"
)

// mockAny matches the activity context argument, which the test never cares about.
var mockAny = mock.Anything

// newEnv builds an in-process workflow environment with all activities
// registered. No Temporal server is involved and timers fire instantly.
func newEnv(t *testing.T) *testsuite.TestWorkflowEnvironment {
	t.Helper()
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.RegisterActivity(ChargePayment)
	env.RegisterActivity(RefundPayment)
	env.RegisterActivity(ReserveStock)
	env.RegisterActivity(ReleaseStock)
	env.RegisterActivity(ShipOrder)
	return env
}

func sampleRequest() FulfillmentRequest {
	return FulfillmentRequest{
		OrderID:    "ord-1",
		CustomerID: "cust-1",
		SKU:        "sku-1",
		Quantity:   2,
		AmountCent: 4999,
	}
}

func TestFulfillmentWorkflow_HappyPath(t *testing.T) {
	env := newEnv(t)
	req := sampleRequest()

	env.OnActivity(ChargePayment, mockAny, req).Return("pay-1", nil)
	env.OnActivity(ReserveStock, mockAny, req).Return("res-1", nil)
	env.OnActivity(ShipOrder, mockAny, req).Return("trk-1", nil)

	env.ExecuteWorkflow(FulfillmentWorkflow, req)

	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if err := env.GetWorkflowError(); err != nil {
		t.Fatalf("workflow error = %v", err)
	}

	var result FulfillmentResult
	if err := env.GetWorkflowResult(&result); err != nil {
		t.Fatalf("GetWorkflowResult() = %v", err)
	}
	if result.PaymentID != "pay-1" || result.ReservationID != "res-1" || result.TrackingID != "trk-1" {
		t.Errorf("result = %+v", result)
	}
	if len(result.Compensated) != 0 {
		t.Errorf("Compensated = %v, want none on the happy path", result.Compensated)
	}
	env.AssertExpectations(t)
}

// Shipping fails last, so both earlier steps have to be undone, in reverse.
func TestFulfillmentWorkflow_ShipFailureCompensatesBoth(t *testing.T) {
	env := newEnv(t)
	req := sampleRequest()

	env.OnActivity(ChargePayment, mockAny, req).Return("pay-1", nil)
	env.OnActivity(ReserveStock, mockAny, req).Return("res-1", nil)
	env.OnActivity(ShipOrder, mockAny, req).Return("", ErrCarrierRejected)
	env.OnActivity(ReleaseStock, mockAny, "res-1").Return(nil)
	env.OnActivity(RefundPayment, mockAny, "pay-1").Return(nil)

	env.ExecuteWorkflow(FulfillmentWorkflow, req)

	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if env.GetWorkflowError() == nil {
		t.Fatal("workflow error = nil, want the carrier failure to surface")
	}
	env.AssertExpectations(t)
}

// A declined card must not reserve stock at all.
func TestFulfillmentWorkflow_PaymentFailureStopsEarly(t *testing.T) {
	env := newEnv(t)
	req := sampleRequest()

	env.OnActivity(ChargePayment, mockAny, req).Return("", ErrPaymentDeclined)

	env.ExecuteWorkflow(FulfillmentWorkflow, req)

	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if env.GetWorkflowError() == nil {
		t.Fatal("workflow error = nil, want the decline to surface")
	}
	// ReserveStock and ShipOrder were never mocked; calling them fails the test.
	env.AssertExpectations(t)
}

// Out of stock after a successful charge: refund, do not ship.
func TestFulfillmentWorkflow_StockFailureRefunds(t *testing.T) {
	env := newEnv(t)
	req := sampleRequest()

	env.OnActivity(ChargePayment, mockAny, req).Return("pay-1", nil)
	env.OnActivity(ReserveStock, mockAny, req).Return("", ErrOutOfStock)
	env.OnActivity(RefundPayment, mockAny, "pay-1").Return(nil)

	env.ExecuteWorkflow(FulfillmentWorkflow, req)

	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if env.GetWorkflowError() == nil {
		t.Fatal("workflow error = nil, want the stock failure to surface")
	}
	env.AssertExpectations(t)
}

// The query handler answers while the workflow is still running. This is the
// capability the hand-rolled saga in L10 has no way to provide.
func TestFulfillmentWorkflow_StatusQuery(t *testing.T) {
	env := newEnv(t)
	req := sampleRequest()

	env.OnActivity(ChargePayment, mockAny, req).Return("pay-1", nil)
	env.OnActivity(ReserveStock, mockAny, req).Return("res-1", nil)
	env.OnActivity(ShipOrder, mockAny, req).Return("trk-1", nil)

	env.ExecuteWorkflow(FulfillmentWorkflow, req)

	value, err := env.QueryWorkflow("status")
	if err != nil {
		t.Fatalf("QueryWorkflow() = %v", err)
	}
	var status string
	if err := value.Get(&status); err != nil {
		t.Fatalf("decode status = %v", err)
	}
	if status != StatusShipped {
		t.Errorf("status = %q, want %q", status, StatusShipped)
	}
}
