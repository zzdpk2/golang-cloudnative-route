package concurrency

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestOnceCellPublishesOneValue(t *testing.T) {
	var cell OnceCell[int]
	var calls atomic.Int64

	const workers = 100
	results := make(chan int, workers)
	var wait sync.WaitGroup
	wait.Add(workers)
	for i := 0; i < workers; i++ {
		go func(value int) {
			defer wait.Done()
			results <- cell.Get(func() int {
				calls.Add(1)
				return value
			})
		}(i)
	}
	wait.Wait()
	close(results)

	var first *int
	for result := range results {
		if first == nil {
			value := result
			first = &value
		}
		if result != *first {
			t.Fatalf("callers observed different values: %d and %d", *first, result)
		}
	}
	if calls.Load() != 1 {
		t.Errorf("initializer calls = %d", calls.Load())
	}
}

func TestChannelClosePublishesEarlierWrites(t *testing.T) {
	publication := NewPublication[map[string]int]()
	go publication.Publish(map[string]int{"orders": 42})

	got := publication.Wait()
	if got["orders"] != 42 {
		t.Errorf("published value = %v", got)
	}
}

func TestAtomicSnapshot(t *testing.T) {
	var snapshot Snapshot[map[string]int]
	if _, ok := snapshot.Load(); ok {
		t.Fatal("empty snapshot reported a value")
	}
	snapshot.Store(map[string]int{"version": 2})
	got, ok := snapshot.Load()
	if !ok || got["version"] != 2 {
		t.Errorf("Load() = (%v, %t)", got, ok)
	}
}

func TestCondQueueWaitsWakesAndDrains(t *testing.T) {
	queue := NewQueue[int]()
	result := make(chan int, 1)
	go func() {
		value, ok := queue.Get()
		if ok {
			result <- value
		}
	}()

	select {
	case <-result:
		t.Fatal("Get returned before data was available")
	case <-time.After(10 * time.Millisecond):
	}

	if !queue.Put(42) {
		t.Fatal("Put rejected an open queue")
	}
	select {
	case got := <-result:
		if got != 42 {
			t.Errorf("Get() = %d", got)
		}
	case <-time.After(time.Second):
		t.Fatal("Get did not wake")
	}

	if !queue.Put(1) || !queue.Put(2) {
		t.Fatal("Put rejected an open queue")
	}
	queue.Close()
	for _, want := range []int{1, 2} {
		if got, ok := queue.Get(); !ok || got != want {
			t.Fatalf("Get() = (%d, %t), want %d", got, ok, want)
		}
	}
	if _, ok := queue.Get(); ok {
		t.Fatal("closed drained queue returned a value")
	}
	if queue.Put(3) {
		t.Fatal("Put accepted a closed queue")
	}
	queue.Close()
}
