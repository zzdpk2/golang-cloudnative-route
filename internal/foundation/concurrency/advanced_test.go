package concurrency

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ============================================================
//
// 2. channel：semaphore / or-done / bridge / tee
//
//   go test ./pkg/concurrency -run TestAdvanced -v
//   go test -race ./pkg/concurrency -run TestAdvanced -v
// ============================================================

// ------------------------------------------------------------
// ------------------------------------------------------------

func TestAdvanced_RWMutexCacheConcurrentAccess(t *testing.T) {
	cache := NewRWMutexCache[string, int]()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			cache.Set("shared", n)
			cache.Set("stable", 42)
			_, _ = cache.Get("shared")
			_, _ = cache.Get("stable")
		}(i)
	}
	wg.Wait()

	if got := cache.Len(); got != 2 {
		t.Fatalf("Len() = %d, want 2", got)
	}

	value, ok := cache.Get("stable")
	if !ok {
		t.Fatal("Get(stable) ok = false, want true")
	}
	if value != 42 {
		t.Fatalf("Get(stable) = %d, want 42", value)
	}

	snapshot := cache.Snapshot()
	snapshot["stable"] = 100

	value, _ = cache.Get("stable")
	if value != 42 {
		t.Fatalf("Snapshot() should be defensive copy, cache value = %d, want 42", value)
	}

	cache.Delete("stable")
	if _, ok := cache.Get("stable"); ok {
		t.Fatal("Get(stable) ok = true after Delete, want false")
	}
}

func TestAdvanced_TryMutex(t *testing.T) {
	mu := NewTryMutex()

	if !mu.TryLock() {
		t.Fatal("first TryLock() = false, want true")
	}

	if mu.TryLock() {
		t.Fatal("second TryLock() = true while locked, want false")
	}

	locked := make(chan struct{})
	go func() {
		mu.Lock()
		close(locked)
	}()

	select {
	case <-locked:
		t.Fatal("Lock() should block while mutex is held")
	case <-time.After(20 * time.Millisecond):
	}

	mu.Unlock()

	select {
	case <-locked:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Lock() did not unblock after Unlock()")
	}

	mu.Unlock()

	defer func() {
		if recover() == nil {
			t.Fatal("double Unlock() should panic")
		}
	}()
	mu.Unlock()
}

func TestAdvanced_StripedCounterConcurrentAdd(t *testing.T) {
	counter := NewStripedCounter(8)

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			counter.Add("user-"+string(rune('A'+n%10)), 1)
		}(i)
	}
	wg.Wait()

	if got := counter.Value(); got != 1000 {
		t.Fatalf("Value() = %d, want 1000", got)
	}

	counter.Add("negative", -100)
	if got := counter.Value(); got != 900 {
		t.Fatalf("Value() after negative add = %d, want 900", got)
	}
}

// ------------------------------------------------------------
// ------------------------------------------------------------

func TestAdvanced_SemaphoreLimitsConcurrency(t *testing.T) {
	sem, err := NewSemaphore(3)
	if err != nil {
		t.Fatalf("NewSemaphore() error = %v", err)
	}

	var running atomic.Int64
	var maxRunning atomic.Int64

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := sem.Acquire(context.Background()); err != nil {
				t.Errorf("Acquire() error = %v", err)
				return
			}
			defer sem.Release()

			current := running.Add(1)
			for {
				max := maxRunning.Load()
				if current <= max || maxRunning.CompareAndSwap(max, current) {
					break
				}
			}

			time.Sleep(5 * time.Millisecond)
			running.Add(-1)
		}()
	}
	wg.Wait()

	if got := maxRunning.Load(); got > 3 {
		t.Fatalf("max concurrent workers = %d, want <= 3", got)
	}
	if got := sem.InUse(); got != 0 {
		t.Fatalf("InUse() = %d, want 0", got)
	}
}

func TestAdvanced_SemaphoreTryAcquireAndCancel(t *testing.T) {
	sem, err := NewSemaphore(1)
	if err != nil {
		t.Fatalf("NewSemaphore() error = %v", err)
	}

	if !sem.TryAcquire() {
		t.Fatal("TryAcquire() = false, want true")
	}
	if sem.TryAcquire() {
		t.Fatal("TryAcquire() = true when full, want false")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	if err := sem.Acquire(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Acquire(timeout) error = %v, want context deadline exceeded", err)
	}

	sem.Release()
	if got := sem.InUse(); got != 0 {
		t.Fatalf("InUse() = %d, want 0", got)
	}

	defer func() {
		if recover() == nil {
			t.Fatal("extra Release() should panic")
		}
	}()
	sem.Release()
}

func TestAdvanced_AnyDone(t *testing.T) {
	a := make(chan struct{})
	b := make(chan struct{})
	done := AnyDone(a, b)

	select {
	case <-done:
		t.Fatal("done should not close before any signal")
	case <-time.After(20 * time.Millisecond):
	}

	close(b)

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("done did not close after one signal closed")
	}
}

func TestAdvanced_BridgeFlattensChannelOfChannels(t *testing.T) {
	ctx := context.Background()
	streams := make(chan (<-chan int))

	go func() {
		defer close(streams)
		streams <- Generator(ctx, 1, 2)
		streams <- Generator(ctx, 3)
		streams <- Generator(ctx, 4, 5)
	}()

	got := Collect(Bridge(ctx, streams))
	want := []int{1, 2, 3, 4, 5}

	if len(got) != len(want) {
		t.Fatalf("Bridge() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Bridge()[%d] = %d, want %d; full result=%v", i, got[i], want[i], got)
		}
	}
}

func TestAdvanced_TeeDuplicatesEachValue(t *testing.T) {
	ctx := context.Background()
	in := Generator(ctx, "a", "b", "c")

	out1, out2 := Tee(ctx, in)

	got1Ch := make(chan []string, 1)
	got2Ch := make(chan []string, 1)

	go func() { got1Ch <- Collect(out1) }()
	go func() { got2Ch <- Collect(out2) }()

	got1 := <-got1Ch
	got2 := <-got2Ch
	want := []string{"a", "b", "c"}

	assertStringSliceEqual(t, got1, want)
	assertStringSliceEqual(t, got2, want)
}

// ------------------------------------------------------------
// ------------------------------------------------------------

func TestAdvanced_AtomicCounter(t *testing.T) {
	var counter AtomicCounter

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Inc()
		}()
	}
	wg.Wait()

	if got := counter.Value(); got != 1000 {
		t.Fatalf("Value() = %d, want 1000", got)
	}

	if got := counter.Add(-10); got != 990 {
		t.Fatalf("Add(-10) = %d, want 990", got)
	}

	counter.Reset()
	if got := counter.Value(); got != 0 {
		t.Fatalf("Value() after Reset() = %d, want 0", got)
	}
}

func TestAdvanced_AtomicFlagOnlyOneWinner(t *testing.T) {
	var flag AtomicFlag
	var winners atomic.Int64

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if flag.TrySet() {
				winners.Add(1)
			}
		}()
	}
	wg.Wait()

	if got := winners.Load(); got != 1 {
		t.Fatalf("winners = %d, want 1", got)
	}
	if !flag.IsSet() {
		t.Fatal("IsSet() = false, want true")
	}

	flag.Reset()
	if flag.IsSet() {
		t.Fatal("IsSet() = true after Reset, want false")
	}
}

func TestAdvanced_AtomicValueConfig(t *testing.T) {
	config := NewAtomicValueConfig(RuntimeConfig{
		Version:       1,
		FeatureEnable: false,
		RateLimit:     100,
	})

	if got := config.Load(); got.Version != 1 || got.RateLimit != 100 {
		t.Fatalf("initial config = %+v", got)
	}

	var wg sync.WaitGroup
	for i := 2; i <= 50; i++ {
		wg.Add(1)
		go func(version int) {
			defer wg.Done()
			config.Store(RuntimeConfig{
				Version:       version,
				FeatureEnable: version%2 == 0,
				RateLimit:     version * 10,
			})
			_ = config.Load()
		}(i)
	}
	wg.Wait()

	got := config.Load()
	if got.Version < 2 || got.Version > 50 {
		t.Fatalf("final config version = %d, want between 2 and 50", got.Version)
	}
}

// ------------------------------------------------------------
// ------------------------------------------------------------

func TestAdvanced_BoundedWorkerPoolRunsJobsAndPreservesOrder(t *testing.T) {
	pool, err := NewBoundedWorkerPool[int, int](3, 2, func(ctx context.Context, n int) (int, error) {
		if n == 3 {
			return 0, errors.New("bad number")
		}
		return n * n, nil
	})
	if err != nil {
		t.Fatalf("NewBoundedWorkerPool() error = %v", err)
	}

	results := pool.Run(context.Background(), []int{1, 2, 3, 4, 5})

	if len(results) != 5 {
		t.Fatalf("len(results) = %d, want 5", len(results))
	}

	if results[0].Value != 1 || results[0].Err != nil {
		t.Fatalf("results[0] = %+v, want value=1 nil error", results[0])
	}
	if results[1].Value != 4 || results[1].Err != nil {
		t.Fatalf("results[1] = %+v, want value=4 nil error", results[1])
	}
	if results[2].Err == nil {
		t.Fatalf("results[2].Err = nil, want error")
	}
	if results[3].Value != 16 || results[3].Err != nil {
		t.Fatalf("results[3] = %+v, want value=16 nil error", results[3])
	}
	if results[4].Value != 25 || results[4].Err != nil {
		t.Fatalf("results[4] = %+v, want value=25 nil error", results[4])
	}
}

func TestAdvanced_BoundedWorkerPoolLimitsConcurrency(t *testing.T) {
	pool, err := NewBoundedWorkerPool[int, int](2, 10, func(ctx context.Context, n int) (int, error) {
		time.Sleep(30 * time.Millisecond)
		return n, nil
	})
	if err != nil {
		t.Fatalf("NewBoundedWorkerPool() error = %v", err)
	}

	start := time.Now()
	results := pool.Run(context.Background(), []int{1, 2, 3, 4})
	elapsed := time.Since(start)

	if len(results) != 4 {
		t.Fatalf("len(results) = %d, want 4", len(results))
	}

	if elapsed < 55*time.Millisecond {
		t.Fatalf("pool seems to run more than 2 jobs concurrently, elapsed = %s", elapsed)
	}
}

func TestAdvanced_CircuitBreakerOpensAndRecovers(t *testing.T) {
	breaker, err := NewCircuitBreaker(2, 30*time.Millisecond)
	if err != nil {
		t.Fatalf("NewCircuitBreaker() error = %v", err)
	}

	fail := func(ctx context.Context) error {
		return errors.New("downstream failed")
	}
	success := func(ctx context.Context) error {
		return nil
	}

	if err := breaker.Execute(context.Background(), fail); err == nil {
		t.Fatal("first failure got nil error, want failure")
	}
	if breaker.State() != CircuitClosed {
		t.Fatalf("State() = %s, want closed before threshold", breaker.State())
	}

	if err := breaker.Execute(context.Background(), fail); err == nil {
		t.Fatal("second failure got nil error, want failure")
	}
	if breaker.State() != CircuitOpen {
		t.Fatalf("State() = %s, want open after threshold", breaker.State())
	}

	if err := breaker.Execute(context.Background(), success); !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("Execute() while open error = %v, want ErrCircuitOpen", err)
	}

	time.Sleep(40 * time.Millisecond)
	if breaker.State() != CircuitHalfOpen {
		t.Fatalf("State() = %s, want half-open after timeout", breaker.State())
	}

	if err := breaker.Execute(context.Background(), success); err != nil {
		t.Fatalf("half-open success Execute() error = %v", err)
	}
	if breaker.State() != CircuitClosed {
		t.Fatalf("State() = %s, want closed after successful probe", breaker.State())
	}
}

func TestAdvanced_SingleFlightDeduplicatesSameKey(t *testing.T) {
	group := NewSingleFlight[int]()

	start := make(chan struct{})
	var calls atomic.Int64

	const goroutines = 20

	var wg sync.WaitGroup
	results := make(chan int, goroutines)
	sharedCount := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start

			value, err, shared := group.Do("hot-key", func() (int, error) {
				calls.Add(1)
				time.Sleep(30 * time.Millisecond)
				return 99, nil
			})
			if err != nil {
				t.Errorf("Do() error = %v", err)
				return
			}
			results <- value
			sharedCount <- shared
		}()
	}

	close(start)
	wg.Wait()
	close(results)
	close(sharedCount)

	if got := calls.Load(); got != 1 {
		t.Fatalf("fn call count = %d, want 1", got)
	}

	for value := range results {
		if value != 99 {
			t.Fatalf("result value = %d, want 99", value)
		}
	}

	sharedSeen := false
	for shared := range sharedCount {
		if shared {
			sharedSeen = true
			break
		}
	}
	if !sharedSeen {
		t.Fatal("expected at least one shared result")
	}
}

func TestAdvanced_SingleFlightDifferentKeysRunIndependently(t *testing.T) {
	group := NewSingleFlight[string]()

	var calls atomic.Int64

	var wg sync.WaitGroup
	for _, key := range []string{"a", "b", "c"} {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			value, err, _ := group.Do(key, func() (string, error) {
				calls.Add(1)
				return "value-" + key, nil
			})
			if err != nil {
				t.Errorf("Do(%s) error = %v", key, err)
				return
			}
			if value != "value-"+key {
				t.Errorf("Do(%s) value = %q", key, value)
			}
		}(key)
	}
	wg.Wait()

	if got := calls.Load(); got != 3 {
		t.Fatalf("fn call count = %d, want 3 for different keys", got)
	}
}

func assertStringSliceEqual(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("slice = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("slice[%d] = %q, want %q; full slice=%v", i, got[i], want[i], got)
		}
	}
}
