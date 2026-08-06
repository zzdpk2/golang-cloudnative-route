package flowcontrol

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestQueueEnforcesHardBoundAndUniqueRequestID(t *testing.T) {
	t.Parallel()

	queue := mustQueue(t, 2)
	now := time.Unix(100, 0)
	if err := queue.Enqueue(request("one", "tenant-a", 0, now)); err != nil {
		t.Fatal(err)
	}
	if err := queue.Enqueue(request("one", "tenant-a", 0, now)); err == nil {
		t.Fatal("queue accepted a duplicate request ID")
	}
	if err := queue.Enqueue(request("two", "tenant-b", 0, now)); err != nil {
		t.Fatal(err)
	}
	if err := queue.Enqueue(request("three", "tenant-c", 0, now)); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("error = %v, want ErrQueueFull", err)
	}
}

func TestQueuePriorityPrecedesFairness(t *testing.T) {
	t.Parallel()

	queue := mustQueue(t, 4)
	now := time.Unix(200, 0)
	_ = queue.Enqueue(request("low", "tenant-a", 1, now))
	_ = queue.Enqueue(request("high", "tenant-b", 9, now.Add(time.Second)))

	got, err := queue.Dequeue(now.Add(2 * time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "high" {
		t.Fatalf("selected %q, want high priority request", got.ID)
	}
}

func TestQueueRotatesFairlyAcrossTenants(t *testing.T) {
	t.Parallel()

	queue := mustQueue(t, 8)
	now := time.Unix(300, 0)
	for _, item := range []Request{
		request("a-1", "tenant-a", 5, now),
		request("a-2", "tenant-a", 5, now.Add(time.Millisecond)),
		request("b-1", "tenant-b", 5, now.Add(2*time.Millisecond)),
		request("b-2", "tenant-b", 5, now.Add(3*time.Millisecond)),
	} {
		_ = queue.Enqueue(item)
	}

	var order []string
	for range 4 {
		item, err := queue.Dequeue(now.Add(time.Second))
		if err != nil {
			t.Fatal(err)
		}
		order = append(order, item.ID)
	}
	want := []string{"a-1", "b-1", "a-2", "b-2"}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order = %v, want %v", order, want)
		}
	}
}

func TestQueueDropsExpiredRequestsBeforeSelection(t *testing.T) {
	t.Parallel()

	queue := mustQueue(t, 4)
	now := time.Unix(400, 0)
	expired := request("expired", "tenant-a", 99, now)
	expired.ExpiresAt = now.Add(time.Second)
	fresh := request("fresh", "tenant-b", 1, now)
	fresh.ExpiresAt = now.Add(time.Minute)
	_ = queue.Enqueue(expired)
	_ = queue.Enqueue(fresh)

	got, err := queue.Dequeue(now.Add(2 * time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "fresh" || queue.Len() != 0 {
		t.Fatalf("selected=%q len=%d, want fresh and empty", got.ID, queue.Len())
	}
}

func TestQueueCancellationReleasesBoundImmediately(t *testing.T) {
	t.Parallel()

	queue := mustQueue(t, 1)
	now := time.Unix(500, 0)
	_ = queue.Enqueue(request("cancel-me", "tenant-a", 0, now))
	if !queue.Cancel("cancel-me") || queue.Cancel("cancel-me") {
		t.Fatal("Cancel must succeed exactly once")
	}
	if err := queue.Enqueue(request("replacement", "tenant-b", 0, now)); err != nil {
		t.Fatalf("queue bound was not released: %v", err)
	}
}

func TestQueueConcurrentMutationPreservesBoundAndUniqueness(t *testing.T) {
	const (
		workers   = 12
		perWorker = 80
	)
	queue := mustQueue(t, workers*perWorker)
	now := time.Unix(600, 0)
	errs := make(chan error, workers)
	var wait sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wait.Add(1)
		go func(worker int) {
			defer wait.Done()
			for item := 0; item < perWorker; item++ {
				id := fmt.Sprintf("%02d-%03d", worker, item)
				if err := queue.Enqueue(request(id, fmt.Sprintf("tenant-%d", worker%3), 1, now)); err != nil {
					errs <- err
					return
				}
				_ = queue.Len()
			}
		}(worker)
	}
	wait.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	if got, want := queue.Len(), workers*perWorker; got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}

	seen := make(map[string]struct{}, workers*perWorker)
	for queue.Len() > 0 {
		item, err := queue.Dequeue(now.Add(time.Second))
		if err != nil {
			t.Fatal(err)
		}
		if _, duplicate := seen[item.ID]; duplicate {
			t.Fatalf("request %q dequeued twice", item.ID)
		}
		seen[item.ID] = struct{}{}
	}
}

func TestQueueConcurrentEnqueueCancelDequeueAndLen(t *testing.T) {
	const initial = 500
	queue := mustQueue(t, initial*2)
	now := time.Unix(650, 0)
	for id := 0; id < initial; id++ {
		if err := queue.Enqueue(request(fmt.Sprintf("old-%03d", id), "tenant-old", 1, now)); err != nil {
			t.Fatal(err)
		}
	}

	var seenMu sync.Mutex
	seen := map[string]struct{}{}
	errs := make(chan error, 16)
	var wait sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		wait.Add(1)
		go func(worker int) {
			defer wait.Done()
			for id := worker; id < initial; id += 4 {
				err := queue.Enqueue(request(
					fmt.Sprintf("new-%03d", id),
					fmt.Sprintf("tenant-%d", worker),
					1,
					now,
				))
				if err != nil {
					errs <- err
					return
				}
			}
		}(worker)
		wait.Add(1)
		go func(worker int) {
			defer wait.Done()
			for id := worker; id < initial; id += 4 {
				queue.Cancel(fmt.Sprintf("old-%03d", id))
			}
		}(worker)
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 100 {
				item, err := queue.Dequeue(now.Add(time.Second))
				if errors.Is(err, ErrQueueEmpty) {
					continue
				}
				if err != nil {
					errs <- err
					return
				}
				seenMu.Lock()
				if _, duplicate := seen[item.ID]; duplicate {
					errs <- fmt.Errorf("request %q dequeued twice", item.ID)
				}
				seen[item.ID] = struct{}{}
				seenMu.Unlock()
			}
		}()
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 1000 {
				if length := queue.Len(); length < 0 || length > initial*2 {
					errs <- fmt.Errorf("Len() escaped hard bound: %d", length)
					return
				}
			}
		}()
	}
	wait.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
}

func TestQueueSustainedArrivalCannotStarveAnActiveTenant(t *testing.T) {
	queue := mustQueue(t, 128)
	now := time.Unix(700, 0)
	for tenant := 0; tenant < 3; tenant++ {
		for sequence := 0; sequence < 8; sequence++ {
			id := fmt.Sprintf("t%d-%02d", tenant, sequence)
			if err := queue.Enqueue(request(id, fmt.Sprintf("tenant-%d", tenant), 5, now)); err != nil {
				t.Fatal(err)
			}
		}
	}

	// Tenant 0 remains noisy. Equal-priority requests from tenants 1 and 2
	// must still appear in every complete three-dequeue rotation.
	for rotation := 0; rotation < 8; rotation++ {
		served := map[string]bool{}
		for range 3 {
			item, err := queue.Dequeue(now.Add(time.Second))
			if err != nil {
				t.Fatal(err)
			}
			served[item.FairnessID] = true
			_ = queue.Enqueue(request(
				fmt.Sprintf("noise-%02d-%s", rotation, item.ID),
				"tenant-0",
				5,
				now.Add(time.Millisecond),
			))
		}
		for _, tenant := range []string{"tenant-0", "tenant-1", "tenant-2"} {
			if !served[tenant] {
				t.Fatalf("rotation %d served %v; %s was starved", rotation, served, tenant)
			}
		}
	}
}

func mustQueue(t *testing.T, max int) *Queue {
	t.Helper()
	queue, err := NewQueue(max)
	if err != nil {
		t.Fatal(err)
	}
	return queue
}

func request(id, tenant string, priority int, now time.Time) Request {
	return Request{
		ID: id, FairnessID: tenant, Priority: priority,
		EnqueuedAt: now, ExpiresAt: now.Add(time.Minute),
	}
}
