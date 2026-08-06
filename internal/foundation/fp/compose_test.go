package fp

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestCompose(t *testing.T) {
	double := func(n int) int { return n * 2 }
	addOne := func(n int) int { return n + 1 }

	// compose(double, addOne)(3) = double(addOne(3)) = double(4) = 8
	f := Compose(double, addOne)
	if f(3) != 8 {
		t.Errorf("got %d", f(3))
	}
}

func TestCompose3(t *testing.T) {
	toString := func(n int) string { return strconv.Itoa(n) }
	double := func(n int) int { return n * 2 }
	addOne := func(n int) int { return n + 1 }

	// toString(double(addOne(3))) = "8"
	f := Compose3(toString, double, addOne)
	if f(3) != "8" {
		t.Errorf("got %q", f(3))
	}
}

func TestPipe2(t *testing.T) {
	got := Pipe2(3,
		func(n int) int { return n + 1 },
		func(n int) int { return n * 2 },
	)
	if got != 8 {
		t.Errorf("got %d", got)
	}
}

func TestPipe3(t *testing.T) {
	// "hello" → upper → add "!" → length
	got := Pipe3("hello",
		strings.ToUpper,
		func(s string) string { return s + "!" },
		func(s string) int { return len(s) },
	)
	if got != 6 {
		t.Errorf("got %d", got)
	} // "HELLO!" = 6
}

func TestCurry2(t *testing.T) {
	add := func(a, b int) int { return a + b }
	curriedAdd := Curry2(add)

	add5 := curriedAdd(5)
	if add5(3) != 8 {
		t.Error("curry")
	}
	if add5(10) != 15 {
		t.Error("curry reuse")
	}
}

func TestUncurry2(t *testing.T) {
	curriedAdd := func(a int) func(int) int {
		return func(b int) int { return a + b }
	}
	add := Uncurry2(curriedAdd)
	if add(3, 5) != 8 {
		t.Error("uncurry")
	}
}

func TestPartial(t *testing.T) {
	multiply := func(a, b int) int { return a * b }
	triple := Partial(multiply, 3)
	if triple(5) != 15 {
		t.Error("partial")
	}
}

func TestPartialRight(t *testing.T) {
	greet := func(name, greeting string) string {
		return greeting + ", " + name + "!"
	}
	helloTo := PartialRight(greet, "Hello")
	if helloTo("Rex") != "Hello, Rex!" {
		t.Errorf("got %q", helloTo("Rex"))
	}
}

func TestMemoize(t *testing.T) {
	var callCount int32
	expensive := func(n int) int {
		atomic.AddInt32(&callCount, 1)
		return n * n
	}

	memo := Memoize(expensive)

	if memo(5) != 25 {
		t.Error("first call")
	}
	if memo(5) != 25 {
		t.Error("cached call")
	}
	if memo(3) != 9 {
		t.Error("different key")
	}

	if atomic.LoadInt32(&callCount) > 3 {
		t.Errorf("callCount = %d, expected <= 3", callCount)
	}
}

func TestMemoize_Concurrent(t *testing.T) {
	var callCount int32
	slow := func(n int) int {
		atomic.AddInt32(&callCount, 1)
		return n * 2
	}

	memo := Memoize(slow)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			memo(42) // See the corresponding tests for the intended behavior.
		}()
	}
	wg.Wait()

	if memo(42) != 84 {
		t.Error("wrong result")
	}
	t.Logf("Memoize callCount = %d (may be > 1)", atomic.LoadInt32(&callCount))
}

func TestMemoizeStrict(t *testing.T) {
	var callCount int32
	slow := func(n int) int {
		atomic.AddInt32(&callCount, 1)
		return n * 2
	}

	memo := MemoizeStrict(slow)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			memo(42)
		}()
	}
	wg.Wait()

	if memo(42) != 84 {
		t.Error("wrong result")
	}
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("strict memoize should compute exactly once, got %d", atomic.LoadInt32(&callCount))
	}
}

func TestPredicate(t *testing.T) {
	isEven := Predicate[int](func(n int) bool { return n%2 == 0 })
	isPositive := Predicate[int](func(n int) bool { return n > 0 })

	evenAndPositive := isEven.And(isPositive)
	if !evenAndPositive(4) {
		t.Error("4 is even and positive")
	}
	if evenAndPositive(-2) {
		t.Error("-2 is even but not positive")
	}
	if evenAndPositive(3) {
		t.Error("3 is positive but not even")
	}

	evenOrPositive := isEven.Or(isPositive)
	if !evenOrPositive(-2) {
		t.Error("-2 is even")
	}
	if !evenOrPositive(3) {
		t.Error("3 is positive")
	}
	if evenOrPositive(-3) {
		t.Error("-3 is neither")
	}

	notEven := isEven.Not()
	if notEven(4) {
		t.Error("4 is even")
	}
	if !notEven(3) {
		t.Error("3 is not even")
	}
}

func TestLazy(t *testing.T) {
	var callCount int
	lazy := NewLazy(func() int {
		callCount++
		return 42
	})

	if callCount != 0 {
		t.Error("should not evaluate yet")
	}

	v := lazy.Force()
	if v != 42 {
		t.Error("wrong value")
	}
	if callCount != 1 {
		t.Error("should evaluate once")
	}

	v2 := lazy.Force()
	if v2 != 42 {
		t.Error("wrong value second time")
	}
	if callCount != 1 {
		t.Error("should not re-evaluate")
	}
}

func TestRealWorldComposition(t *testing.T) {
	trim := func(s string) string { return strings.TrimSpace(s) }
	lower := strings.ToLower
	validate := func(s string) Result[string] {
		if len(s) < 3 {
			return Err[string](fmt.Errorf("too short: %q", s))
		}
		return Ok(s)
	}
	format := func(s string) string { return "[" + s + "]" }

	// Compose: right-to-left
	clean := Compose(lower, trim)

	process := func(input string) Result[string] {
		return validate(clean(input)).Map(format)
	}

	got := process("  HELLO  ")
	if got.Unwrap() != "[hello]" {
		t.Errorf("got %q", got.Unwrap())
	}

	got2 := process("  AB  ")
	if got2.IsOk() {
		t.Error("should fail validation")
	}
}
