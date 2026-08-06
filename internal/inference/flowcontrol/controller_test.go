package flowcontrol

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rex/go-ddd-tdd/internal/inference/routing"
)

func TestControllerRejectsInvalidCapacity(t *testing.T) {
	t.Parallel()
	for _, limits := range [][2]int{{0, 1}, {1, 0}, {-1, 1}, {1, -1}} {
		if _, err := NewController(limits[0], limits[1], time.Second); !errors.Is(err, ErrInvalidCapacity) {
			t.Fatalf("NewController(%d, %d) error = %v, want ErrInvalidCapacity", limits[0], limits[1], err)
		}
	}
	if _, err := NewController(1, 1, 0); !errors.Is(err, ErrInvalidCapacity) {
		t.Fatalf("zero max wait error = %v, want ErrInvalidCapacity", err)
	}
}

func TestControllerQueuesThenTransfersReleasedPermit(t *testing.T) {
	t.Parallel()
	controller := mustController(t, 1, 2)
	firstRelease, err := controller.Admit(context.Background(), liveRequest("first"))
	if err != nil {
		t.Fatal(err)
	}

	type result struct {
		release func()
		err     error
	}
	second := make(chan result, 1)
	go func() {
		release, admitErr := controller.Admit(context.Background(), liveRequest("second"))
		second <- result{release: release, err: admitErr}
	}()
	waitForCount(t, controller.Queued, 1)
	firstRelease()
	firstRelease() // Release is idempotent.

	select {
	case got := <-second:
		if got.err != nil {
			t.Fatal(got.err)
		}
		if controller.InFlight() != 1 || controller.Queued() != 0 {
			t.Fatalf("in_flight=%d queued=%d, want 1/0", controller.InFlight(), controller.Queued())
		}
		got.release()
	case <-time.After(time.Second):
		t.Fatal("queued request did not receive the released permit")
	}
}

func TestControllerCancellationRemovesQueuedRequest(t *testing.T) {
	t.Parallel()
	controller := mustController(t, 1, 1)
	release, err := controller.Admit(context.Background(), liveRequest("holder"))
	if err != nil {
		t.Fatal(err)
	}
	defer release()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, admitErr := controller.Admit(ctx, liveRequest("cancelled"))
		done <- admitErr
	}()
	waitForCount(t, controller.Queued, 1)
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Admit error = %v, want context.Canceled", err)
	}
	waitForCount(t, controller.Queued, 0)
}

func TestControllerQueueFullAndMaxWaitReleaseTheBound(t *testing.T) {
	t.Parallel()
	controller, err := NewController(1, 1, 20*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	holder, err := controller.Admit(context.Background(), liveRequest("holder"))
	if err != nil {
		t.Fatal(err)
	}
	defer holder()
	waiting := make(chan error, 1)
	go func() {
		_, admitErr := controller.Admit(context.Background(), liveRequest("waiting"))
		waiting <- admitErr
	}()
	waitForCount(t, controller.Queued, 1)
	if _, err := controller.Admit(context.Background(), liveRequest("overflow")); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("overflow error = %v, want ErrQueueFull", err)
	}
	if err := <-waiting; !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("max-wait error = %v, want context.DeadlineExceeded", err)
	}
	waitForCount(t, controller.Queued, 0)
}

func TestControllerPromotesHigherPriorityQueuedRequestFirst(t *testing.T) {
	t.Parallel()
	controller := mustController(t, 1, 4)
	holderRelease, err := controller.Admit(context.Background(), liveRequest("holder"))
	if err != nil {
		t.Fatal(err)
	}
	type admitted struct {
		id      string
		release func()
		err     error
	}
	results := make(chan admitted, 2)
	for _, request := range []routing.Request{
		{ID: "low", Model: "tiny-llm", FairnessID: "tenant-a", Priority: 1},
		{ID: "high", Model: "tiny-llm", FairnessID: "tenant-b", Priority: 9},
	} {
		request := request
		go func() {
			release, admitErr := controller.Admit(context.Background(), request)
			results <- admitted{id: request.ID, release: release, err: admitErr}
		}()
	}
	waitForCount(t, controller.Queued, 2)
	holderRelease()

	first := <-results
	if first.err != nil {
		t.Fatal(first.err)
	}
	if first.id != "high" {
		t.Fatalf("first promoted request = %q, want high", first.id)
	}
	first.release()
	second := <-results
	if second.err != nil {
		t.Fatal(second.err)
	}
	second.release()
}

func TestControllerNeverExceedsInFlightLimitConcurrently(t *testing.T) {
	controller := mustController(t, 4, 64)
	var active atomic.Int64
	var peak atomic.Int64
	var wait sync.WaitGroup
	errs := make(chan error, 32)
	for worker := 0; worker < 32; worker++ {
		wait.Add(1)
		go func(worker int) {
			defer wait.Done()
			release, err := controller.Admit(context.Background(), liveRequest(fmt.Sprintf("worker-%d", worker)))
			if err != nil {
				errs <- err
				return
			}
			current := active.Add(1)
			for {
				observed := peak.Load()
				if current <= observed || peak.CompareAndSwap(observed, current) {
					break
				}
			}
			time.Sleep(time.Millisecond)
			active.Add(-1)
			release()
		}(worker)
	}
	wait.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	if got := peak.Load(); got > 4 {
		t.Fatalf("peak in-flight = %d, hard limit = 4", got)
	}
	if controller.InFlight() != 0 || controller.Queued() != 0 {
		t.Fatalf("leaked state: in_flight=%d queued=%d", controller.InFlight(), controller.Queued())
	}
}

func mustController(t *testing.T, inFlight, queued int) *Controller {
	t.Helper()
	controller, err := NewController(inFlight, queued, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	return controller
}

func liveRequest(id string) routing.Request {
	return routing.Request{ID: id, Model: "tiny-llm", FairnessID: "tenant", Priority: 1}
}

func waitForCount(t *testing.T, count func() int, want int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if count() == want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("count = %d, want %d", count(), want)
}
