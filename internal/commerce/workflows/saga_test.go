package workflows

import (
	"context"
	"fmt"
	"testing"
)

func TestSaga_AllSuccess(t *testing.T) {
	var executed []string

	s := NewSaga().
		AddStep(SagaStep{
			Name:       "step1",
			Execute:    func(ctx context.Context) error { executed = append(executed, "exec1"); return nil },
			Compensate: func(ctx context.Context) error { executed = append(executed, "comp1"); return nil },
		}).
		AddStep(SagaStep{
			Name:       "step2",
			Execute:    func(ctx context.Context) error { executed = append(executed, "exec2"); return nil },
			Compensate: func(ctx context.Context) error { executed = append(executed, "comp2"); return nil },
		}).
		AddStep(SagaStep{
			Name:       "step3",
			Execute:    func(ctx context.Context) error { executed = append(executed, "exec3"); return nil },
			Compensate: func(ctx context.Context) error { executed = append(executed, "comp3"); return nil },
		})

	result := s.Execute(context.Background())
	if !result.IsSuccess() {
		t.Fatal(result.Error())
	}
	if len(result.CompletedSteps) != 3 {
		t.Errorf("completed = %v", result.CompletedSteps)
	}
	want := []string{"exec1", "exec2", "exec3"}
	if len(executed) != len(want) {
		t.Errorf("executed = %v", executed)
	}
	for i, w := range want {
		if executed[i] != w {
			t.Errorf("[%d] = %q, want %q", i, executed[i], w)
		}
	}
}

func TestSaga_FailAtStep2_CompensatesStep1(t *testing.T) {
	var executed []string

	s := NewSaga().
		AddStep(SagaStep{
			Name:       "reserve_inventory",
			Execute:    func(ctx context.Context) error { executed = append(executed, "reserve"); return nil },
			Compensate: func(ctx context.Context) error { executed = append(executed, "release"); return nil },
		}).
		AddStep(SagaStep{
			Name:       "charge_payment",
			Execute:    func(ctx context.Context) error { return fmt.Errorf("payment declined") },
			Compensate: func(ctx context.Context) error { executed = append(executed, "refund"); return nil },
		}).
		AddStep(SagaStep{
			Name:       "ship_order",
			Execute:    func(ctx context.Context) error { executed = append(executed, "ship"); return nil },
			Compensate: func(ctx context.Context) error { executed = append(executed, "cancel_ship"); return nil },
		})

	result := s.Execute(context.Background())
	if result.IsSuccess() {
		t.Fatal("should fail")
	}
	if result.FailedStep != "charge_payment" {
		t.Errorf("failed step = %q", result.FailedStep)
	}

	// step1 executed then compensated, step2 failed (no compensation needed for it),
	// step3 never executed
	if len(result.CompletedSteps) != 1 {
		t.Errorf("completed = %v", result.CompletedSteps)
	}

	// executed should be: reserve, release (compensation in reverse)
	// "ship" should NOT appear
	found := false
	for _, e := range executed {
		if e == "ship" {
			t.Error("step3 should not have executed")
		}
		if e == "release" {
			found = true
		}
	}
	if !found {
		t.Errorf("step1 should be compensated, executed = %v", executed)
	}
}

func TestSaga_FailAtStep3_CompensatesStep2ThenStep1(t *testing.T) {
	var compensations []string

	s := NewSaga().
		AddStep(SagaStep{
			Name:       "step1",
			Execute:    func(ctx context.Context) error { return nil },
			Compensate: func(ctx context.Context) error { compensations = append(compensations, "comp1"); return nil },
		}).
		AddStep(SagaStep{
			Name:       "step2",
			Execute:    func(ctx context.Context) error { return nil },
			Compensate: func(ctx context.Context) error { compensations = append(compensations, "comp2"); return nil },
		}).
		AddStep(SagaStep{
			Name:       "step3",
			Execute:    func(ctx context.Context) error { return fmt.Errorf("fail") },
			Compensate: func(ctx context.Context) error { compensations = append(compensations, "comp3"); return nil },
		})

	result := s.Execute(context.Background())

	if result.IsSuccess() {
		t.Fatal("saga should fail when step3 fails")
	}
	if result.FailedStep != "step3" {
		t.Errorf("FailedStep = %q, want %q", result.FailedStep, "step3")
	}
	if len(compensations) != 2 {
		t.Fatalf("expected 2 compensations, got %v", compensations)
	}
	if compensations[0] != "comp2" || compensations[1] != "comp1" {
		t.Errorf("wrong compensation order: %v", compensations)
	}
}

func TestSaga_CompensationAlsoFails(t *testing.T) {
	s := NewSaga().
		AddStep(SagaStep{
			Name:       "step1",
			Execute:    func(ctx context.Context) error { return nil },
			Compensate: func(ctx context.Context) error { return fmt.Errorf("comp1 failed") },
		}).
		AddStep(SagaStep{
			Name:       "step2",
			Execute:    func(ctx context.Context) error { return fmt.Errorf("step2 failed") },
			Compensate: func(ctx context.Context) error { return nil },
		})

	result := s.Execute(context.Background())
	if result.IsSuccess() {
		t.Fatal("should fail")
	}
	if len(result.CompensateErrs) != 1 {
		t.Errorf("expected 1 compensation error, got %d", len(result.CompensateErrs))
	}
}

func TestSaga_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s := NewSaga().
		AddStep(SagaStep{
			Name:       "step1",
			Execute:    func(ctx context.Context) error { return nil },
			Compensate: func(ctx context.Context) error { return nil },
		})

	result := s.Execute(ctx)
	if result.IsSuccess() {
		t.Error("cancelled context should prevent execution")
	}
}

// ---- StatefulSaga Tests ----

func TestStatefulSaga(t *testing.T) {
	s := NewStatefulSaga().
		SetState("order_id", "ORD-001").
		AddStep("reserve",
			func(ctx context.Context, state map[string]any) error {
				state["reservation_id"] = "RES-123"
				return nil
			},
			func(ctx context.Context, state map[string]any) error {
				resID, ok := state["reservation_id"]
				if !ok {
					return fmt.Errorf("no reservation_id to cancel")
				}
				t.Logf("cancelled reservation %v", resID)
				return nil
			},
		).
		AddStep("charge",
			func(ctx context.Context, state map[string]any) error {
				_, ok := state["reservation_id"]
				if !ok {
					return fmt.Errorf("no reservation to charge for")
				}
				state["payment_id"] = "PAY-456"
				return nil
			},
			func(ctx context.Context, state map[string]any) error {
				payID := state["payment_id"]
				t.Logf("refunded payment %v", payID)
				return nil
			},
		)

	result := s.Execute(context.Background())
	if !result.IsSuccess() {
		t.Fatal(result.Error())
	}

	payID, ok := s.GetState("payment_id")
	if !ok || payID != "PAY-456" {
		t.Errorf("payment_id = %v", payID)
	}
}

func TestStatefulSaga_FailWithCompensation(t *testing.T) {
	var compensated []string

	s := NewStatefulSaga().
		AddStep("step1",
			func(ctx context.Context, state map[string]any) error {
				state["data"] = "important"
				return nil
			},
			func(ctx context.Context, state map[string]any) error {
				compensated = append(compensated, fmt.Sprintf("undo step1, data=%v", state["data"]))
				return nil
			},
		).
		AddStep("step2",
			func(ctx context.Context, state map[string]any) error {
				return fmt.Errorf("step2 boom")
			},
			func(ctx context.Context, state map[string]any) error {
				compensated = append(compensated, "undo step2")
				return nil
			},
		)

	result := s.Execute(context.Background())
	if result.IsSuccess() {
		t.Fatal("should fail")
	}

	if len(compensated) != 1 { // See the corresponding tests for the intended behavior.
		t.Fatalf("compensated = %v", compensated)
	}
	if compensated[0] != "undo step1, data=important" {
		t.Errorf("compensation = %q", compensated[0])
	}
}
