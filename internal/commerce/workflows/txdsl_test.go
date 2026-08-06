package workflows

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

// noop is a step that succeeds and does nothing.
func noop(context.Context, *State) error { return nil }

// record returns a step that appends label to log.
func record(log *[]string, label string) StepFunc {
	return func(context.Context, *State) error {
		*log = append(*log, label)
		return nil
	}
}

func failWith(err error) StepFunc {
	return func(context.Context, *State) error { return err }
}

func TestRun_AllStepsSucceed(t *testing.T) {
	var log []string
	res := New().
		Step("one", record(&log, "do1"), record(&log, "undo1")).
		Step("two", record(&log, "do2"), record(&log, "undo2")).
		Run(context.Background())

	if !res.OK() {
		t.Fatalf("OK() = false, Err = %v", res.Err)
	}
	if want := []string{"do1", "do2"}; !reflect.DeepEqual(log, want) {
		t.Errorf("execution order = %v, want %v", log, want)
	}
	if want := []string{"one", "two"}; !reflect.DeepEqual(res.Completed, want) {
		t.Errorf("Completed = %v, want %v", res.Completed, want)
	}
	if len(res.Compensated) != 0 {
		t.Errorf("Compensated = %v, want none", res.Compensated)
	}
}

func TestRun_CompensatesInReverseOrder(t *testing.T) {
	var log []string
	boom := errors.New("payment declined")

	res := New().
		Step("stock", record(&log, "do:stock"), record(&log, "undo:stock")).
		Step("coupon", record(&log, "do:coupon"), record(&log, "undo:coupon")).
		Step("payment", failWith(boom), record(&log, "undo:payment")).
		Run(context.Background())

	if res.OK() {
		t.Fatal("OK() = true, want false")
	}
	if !errors.Is(res.Err, boom) {
		t.Errorf("Err = %v, want %v", res.Err, boom)
	}
	if res.FailedStep != "payment" {
		t.Errorf("FailedStep = %q, want %q", res.FailedStep, "payment")
	}

	// The failing step never completed, so it must not be rolled back.
	want := []string{"do:stock", "do:coupon", "undo:coupon", "undo:stock"}
	if !reflect.DeepEqual(log, want) {
		t.Errorf("log = %v, want %v", log, want)
	}
	if wantC := []string{"coupon", "stock"}; !reflect.DeepEqual(res.Compensated, wantC) {
		t.Errorf("Compensated = %v, want %v", res.Compensated, wantC)
	}
}

func TestRun_ContinuesCompensatingAfterACompensationFails(t *testing.T) {
	var log []string
	compBoom := errors.New("release failed")

	res := New().
		Step("a", record(&log, "do:a"), record(&log, "undo:a")).
		Step("b", record(&log, "do:b"), failWith(compBoom)).
		Step("c", failWith(errors.New("boom")), noop).
		Run(context.Background())

	// b's compensation blew up, but a's still has to run.
	if want := []string{"do:a", "do:b", "undo:a"}; !reflect.DeepEqual(log, want) {
		t.Errorf("log = %v, want %v", log, want)
	}
	if len(res.CompensateErrs) != 1 || !errors.Is(res.CompensateErrs[0], compBoom) {
		t.Errorf("CompensateErrs = %v, want exactly [%v]", res.CompensateErrs, compBoom)
	}
}

func TestRun_NilCompensationIsSkipped(t *testing.T) {
	var log []string
	res := New().
		Step("read", record(&log, "do:read"), nil).
		Step("write", failWith(errors.New("boom")), nil).
		Run(context.Background())

	if res.OK() {
		t.Fatal("OK() = true, want false")
	}
	if want := []string{"do:read"}; !reflect.DeepEqual(log, want) {
		t.Errorf("log = %v, want %v", log, want)
	}
	if len(res.CompensateErrs) != 0 {
		t.Errorf("CompensateErrs = %v, want none", res.CompensateErrs)
	}
}

func TestRun_NilActionIsAProgrammingError(t *testing.T) {
	res := New().
		Step("fine", noop, nil).
		Step("broken", nil, nil).
		Run(context.Background())

	if res.OK() {
		t.Fatal("OK() = true, want false")
	}
	if !errors.Is(res.Err, ErrNilStep) {
		t.Errorf("Err = %v, want ErrNilStep", res.Err)
	}
}

// A cancelled context stops forward progress, and the compensations still run.
func TestRun_CancelledContextStillCompensates(t *testing.T) {
	var log []string
	ctx, cancel := context.WithCancel(context.Background())

	res := New().
		Step("a", record(&log, "do:a"), record(&log, "undo:a")).
		Step("cancel here", func(context.Context, *State) error {
			cancel()
			return nil
		}, record(&log, "undo:cancel")).
		Step("never runs", record(&log, "do:never"), record(&log, "undo:never")).
		Run(ctx)

	if res.OK() {
		t.Fatal("OK() = true, want false after cancellation")
	}
	for _, entry := range log {
		if entry == "do:never" {
			t.Error("a step ran after the context was cancelled")
		}
	}
	// Both completed steps must be rolled back even though ctx is done.
	if want := []string{"undo:cancel", "undo:a"}; !reflect.DeepEqual(res.Compensated, []string{"cancel here", "a"}) {
		t.Errorf("Compensated = %v, want [cancel here a] (log tail wanted %v)", res.Compensated, want)
	}
}

func TestState_PassesValuesBetweenSteps(t *testing.T) {
	res := New().
		Step("create payment", func(_ context.Context, s *State) error {
			s.Set("paymentID", "pay-42")
			return nil
		}, nil).
		Step("use payment", func(_ context.Context, s *State) error {
			id, ok := GetAs[string](s, "paymentID")
			if !ok {
				return errors.New("paymentID missing")
			}
			if id != "pay-42" {
				return errors.New("paymentID = " + id)
			}
			return nil
		}, nil).
		Run(context.Background())

	if !res.OK() {
		t.Fatalf("OK() = false, Err = %v", res.Err)
	}
}

func TestState_GetAsRejectsWrongType(t *testing.T) {
	var s State
	s.Set("qty", 7)

	if _, ok := GetAs[string](&s, "qty"); ok {
		t.Error("GetAs[string] on an int value = ok, want not ok")
	}
	if got, ok := GetAs[int](&s, "qty"); !ok || got != 7 {
		t.Errorf("GetAs[int] = (%v, %v), want (7, true)", got, ok)
	}
	if _, ok := GetAs[int](&s, "absent"); ok {
		t.Error("GetAs on a missing key = ok, want not ok")
	}
}

// The compensation must be able to see what its forward action recorded.
func TestState_CompensationSeesForwardState(t *testing.T) {
	var seen string
	res := New().
		Step("reserve", func(_ context.Context, s *State) error {
			s.Set("reservationID", "res-9")
			return nil
		}, func(_ context.Context, s *State) error {
			id, _ := GetAs[string](s, "reservationID")
			seen = id
			return nil
		}).
		Step("charge", failWith(errors.New("declined")), nil).
		Run(context.Background())

	if res.OK() {
		t.Fatal("OK() = true, want false")
	}
	if seen != "res-9" {
		t.Errorf("compensation saw reservationID = %q, want %q", seen, "res-9")
	}
}

func TestRun_EmptyTransactionSucceeds(t *testing.T) {
	res := New().Run(context.Background())
	if !res.OK() {
		t.Errorf("an empty transaction should succeed, Err = %v", res.Err)
	}
}

// Compensations must not inherit a deadline that has already expired, or they
// all fail at once and the rollback is useless.
func TestRun_CompensationNotBlockedByExpiredDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	compensated := false
	res := New().
		Step("a", noop, func(ctx context.Context, _ *State) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			compensated = true
			return nil
		}).
		Step("slow", func(context.Context, *State) error {
			time.Sleep(30 * time.Millisecond)
			return errors.New("too slow")
		}, nil).
		Run(ctx)

	if res.OK() {
		t.Fatal("OK() = true, want false")
	}
	if !compensated {
		t.Error("the compensation was handed an already-expired context and could not run")
	}
}
