package concurrency

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

// ---- Pipeline Tests ----

func TestPipeline_GenerateTransformFilter(t *testing.T) {
	ctx := context.Background()

	nums := Generator(ctx, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	doubled := Transform(ctx, nums, func(n int) int { return n * 2 })
	big := FilterChan(ctx, doubled, func(n int) bool { return n > 10 })
	result := Collect(big)

	// 6*2=12, 7*2=14, 8*2=16, 9*2=18, 10*2=20
	want := []int{12, 14, 16, 18, 20}
	if len(result) != len(want) {
		t.Fatalf("got %v, want %v", result, want)
	}
	for i, v := range result {
		if v != want[i] {
			t.Errorf("[%d] = %d, want %d", i, v, want[i])
		}
	}
}

func TestPipeline_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	gen := func() <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			i := 0
			for {
				select {
				case out <- i:
					i++
				case <-ctx.Done():
					return
				}
			}
		}()
		return out
	}

	ch := gen()
	for i := 0; i < 5; i++ {
		<-ch
	}
	cancel()

	time.Sleep(10 * time.Millisecond)
	_, ok := <-ch
	for ok {
		_, ok = <-ch
	}
}

// ---- Fan-Out / Fan-In ----

func TestFanOutFanIn(t *testing.T) {
	ctx := context.Background()

	nums := Generator(ctx, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20)

	outputs := FanOut(ctx, nums, func(n int) int { return n * 10 }, 4)

	merged := FanIn(ctx, outputs...)
	result := Collect(merged)

	if len(result) != 20 {
		t.Errorf("expected 20 results, got %d", len(result))
	}

	for _, v := range result {
		if v%10 != 0 {
			t.Errorf("unexpected value: %d", v)
		}
	}
}

// ---- Timeout ----

func TestWithTimeout_Success(t *testing.T) {
	result, err := WithTimeout(1*time.Second, func() int {
		return 42
	})
	if err != nil {
		t.Fatal(err)
	}
	if result != 42 {
		t.Errorf("result = %d", result)
	}
}

func TestWithTimeout_Exceeded(t *testing.T) {
	_, err := WithTimeout(10*time.Millisecond, func() int {
		time.Sleep(1 * time.Second)
		return 42
	})
	if err == nil {
		t.Error("should timeout")
	}
}

// ---- Debounce ----

func TestDebounce(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan int)
	out := Debounce(ctx, in, 50*time.Millisecond)

	go func() {
		for i := 0; i < 5; i++ {
			in <- i
			time.Sleep(10 * time.Millisecond) // See the corresponding tests for the intended behavior.
		}
		time.Sleep(100 * time.Millisecond)
		close(in)
	}()

	results := Collect(out)

	if len(results) > 2 {
		t.Errorf("debounce should reduce values, got %d results: %v", len(results), results)
	}
	if len(results) > 0 && results[len(results)-1] != 4 {
		t.Errorf("last value should be 4, got %v", results)
	}
}

// ---- RunAll ----

func TestRunAll(t *testing.T) {
	ctx := context.Background()
	results := RunAll(ctx,
		func(ctx context.Context) (string, error) {
			return "first", nil
		},
		func(ctx context.Context) (string, error) {
			return "", fmt.Errorf("failed")
		},
		func(ctx context.Context) (string, error) {
			return "third", nil
		},
	)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	if results[0].Value != "first" || results[0].Err != nil {
		t.Errorf("results[0] = %v", results[0])
	}
	if results[1].Err == nil {
		t.Error("results[1] should have error")
	}
	if results[2].Value != "third" || results[2].Err != nil {
		t.Errorf("results[2] = %v", results[2])
	}
}

// ---- RunFirst ----

func TestRunFirst(t *testing.T) {
	ctx := context.Background()

	t.Run("fastest wins", func(t *testing.T) {
		result, err := RunFirst(ctx,
			func(ctx context.Context) (string, error) {
				time.Sleep(100 * time.Millisecond)
				return "slow", nil
			},
			func(ctx context.Context) (string, error) {
				return "fast", nil // See the corresponding tests for the intended behavior.
			},
			func(ctx context.Context) (string, error) {
				time.Sleep(200 * time.Millisecond)
				return "slowest", nil
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if result != "fast" {
			t.Errorf("result = %q, want 'fast'", result)
		}
	})

	t.Run("all fail", func(t *testing.T) {
		_, err := RunFirst(ctx,
			func(ctx context.Context) (string, error) {
				return "", errors.New("fail1")
			},
			func(ctx context.Context) (string, error) {
				return "", errors.New("fail2")
			},
		)
		if err == nil {
			t.Error("should error when all fail")
		}
	})
}

// ---- SafeCounter vs UnsafeCounter ----

func TestSafeCounter(t *testing.T) {
	c := &SafeCounter{}
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Increment()
		}()
	}
	wg.Wait()

	if c.Value() != 1000 {
		t.Errorf("count = %d, want 1000", c.Value())
	}
}

// go test -race -run TestUnsafeCounter_ShowRace ./pkg/concurrency/
func TestUnsafeCounter_ShowRace(t *testing.T) {
	t.Skip("Unskip this and run with -race to see the data race warning")
	c := &UnsafeCounter{}
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Increment()
		}()
	}
	wg.Wait()

	t.Logf("UnsafeCounter result: %d (should be 1000 but probably isn't)", c.Value())
}

func TestNewConcurrentAccumulator_ClosureState(t *testing.T) {
	add := NewConcurrentAccumulator(10)

	if got := add(5); got != 15 {
		t.Errorf("first call = %d, want 15", got)
	}
	if got := add(7); got != 22 {
		t.Errorf("second call = %d, want 22", got)
	}
	if got := add(-2); got != 20 {
		t.Errorf("third call = %d, want 20", got)
	}
}

func TestNewConcurrentAccumulator_ConcurrentCalls(t *testing.T) {
	add := NewConcurrentAccumulator(0)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			add(1)
		}()
	}
	wg.Wait()

	if got := add(0); got != 100 {
		t.Errorf("final total = %d, want 100", got)
	}
}
