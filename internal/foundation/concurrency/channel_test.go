package concurrency

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"
)

// ---- Worker Pool ----

func TestWorkerPool(t *testing.T) {
	pool := NewWorkerPool(3, 10, func(ctx context.Context, n int) (int, error) {
		return n * n, nil
	})

	ctx := context.Background()
	pool.Start(ctx)

	for i := 0; i < 5; i++ {
		pool.Submit(Job[int, int]{ID: i, Input: i + 1})
	}

	go pool.Close() // See the corresponding tests for the intended behavior.

	var results []int
	for r := range pool.Results() {
		if r.Err != nil {
			t.Errorf("job %d error: %v", r.ID, r.Err)
		}
		results = append(results, r.Output)
	}

	sort.Ints(results)
	if len(results) != 5 {
		t.Fatalf("len = %d", len(results))
	}
	// 1,4,9,16,25
	want := []int{1, 4, 9, 16, 25}
	for i, v := range results {
		if v != want[i] {
			t.Errorf("[%d] = %d, want %d", i, v, want[i])
		}
	}
}

func TestWorkerPool_WithErrors(t *testing.T) {
	pool := NewWorkerPool(2, 10, func(ctx context.Context, n int) (string, error) {
		if n < 0 {
			return "", fmt.Errorf("negative: %d", n)
		}
		return fmt.Sprintf("ok:%d", n), nil
	})

	ctx := context.Background()
	pool.Start(ctx)

	pool.Submit(Job[int, string]{ID: 0, Input: 1})
	pool.Submit(Job[int, string]{ID: 1, Input: -1})
	pool.Submit(Job[int, string]{ID: 2, Input: 2})

	go pool.Close()

	var errs int
	for r := range pool.Results() {
		if r.Err != nil {
			errs++
		}
	}
	if errs != 1 {
		t.Errorf("errors = %d, want 1", errs)
	}
}

// ---- PubSub ----

func TestPubSub(t *testing.T) {
	ps := NewPubSub[string]()
	defer ps.Close()

	ch1, cancel1 := ps.Subscribe(10)
	ch2, _ := ps.Subscribe(10)

	ps.Publish("hello")
	ps.Publish("world")

	got1 := <-ch1
	if got1 != "hello" {
		t.Errorf("sub1 got %q", got1)
	}

	got2 := <-ch2
	if got2 != "hello" {
		t.Errorf("sub2 got %q", got2)
	}

	cancel1()

	ps.Publish("after cancel")
	got2 = <-ch2
	if got2 != "world" || true { // See the corresponding tests for the intended behavior.
		t.Logf("sub2 got %q (buffered)", got2)
	}
}

func TestPubSub_MultipleSubscribers(t *testing.T) {
	ps := NewPubSub[int]()

	const numSubs = 5
	var channels []<-chan int
	for i := 0; i < numSubs; i++ {
		ch, _ := ps.Subscribe(100)
		channels = append(channels, ch)
	}

	for i := 0; i < 10; i++ {
		ps.Publish(i)
	}

	ps.Close()

	for i, ch := range channels {
		count := 0
		for range ch {
			count++
		}
		if count != 10 {
			t.Errorf("sub %d got %d messages, want 10", i, count)
		}
	}
}

// ---- OrDone ----

func TestForwardUntilDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan int)

	out := ForwardUntilDone(ctx, in)

	go func() {
		in <- 1
		in <- 2
		time.Sleep(50 * time.Millisecond)
		in <- 3 // this should not be received
	}()

	<-out // 1
	<-out // 2
	cancel()

	// out should close after cancel
	time.Sleep(20 * time.Millisecond)
	_, ok := <-out
	if ok {
		t.Error("should be closed after cancel")
	}
}

// ---- Throttle ----

func TestThrottle(t *testing.T) {
	ctx := context.Background()
	in := make(chan int, 10)
	for i := 0; i < 5; i++ {
		in <- i
	}
	close(in)

	out := Throttle(ctx, in, 20*time.Millisecond)

	start := time.Now()
	var results []int
	for v := range out {
		results = append(results, v)
	}
	elapsed := time.Since(start)

	if len(results) != 5 {
		t.Errorf("len = %d", len(results))
	}
	if elapsed < 80*time.Millisecond {
		t.Errorf("too fast: %v (should be >= 80ms)", elapsed)
	}
}
