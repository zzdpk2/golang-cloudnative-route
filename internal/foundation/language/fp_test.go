package language

import (
	"fmt"
	"strings"
	"testing"
)

func TestMakeCounters_Wrong(t *testing.T) {
	counters := MakeCounters_Wrong(3)
	for idx, fn := range counters {
		got := fn()
		if got != 3 {
			t.Logf("counter[%d] = %d (should be 3 due to closure bug)", idx, got)
		}
	}
	if counters[0]() != counters[2]() {
		t.Error("all counters should return same value (the closure bug)")
	}
}

func TestMakeCounters_Correct(t *testing.T) {
	counters := MakeCounters_Correct(3)
	for i, fn := range counters {
		got := fn()
		if got != i {
			t.Errorf("counter[%d] = %d, want %d", i, got, i)
		}
	}
}

func TestAccumulator(t *testing.T) {
	acc := Accumulator(10)
	if acc(5) != 15 {
		t.Error("10+5=15")
	}
	if acc(3) != 18 {
		t.Error("15+3=18")
	}
	if acc(-8) != 10 {
		t.Error("18-8=10")
	}

	acc2 := Accumulator(0)
	if acc2(1) != 1 {
		t.Error("separate state")
	}
	if acc(0) != 10 {
		t.Error("acc should be unaffected")
	}
}

// ---- Defer ----

func TestDeferOrder(t *testing.T) {
	result := DeferOrder()
	// LIFO: main body, then third, second, first
	want := []string{"main body", "third defer", "second defer", "first defer"}
	if len(result) != len(want) {
		t.Fatalf("got %v, want %v", result, want)
	}
	for i, w := range want {
		if result[i] != w {
			t.Errorf("[%d] = %q, want %q", i, result[i], w)
		}
	}
}

func TestDeferArgEval(t *testing.T) {
	got := DeferArgEval()
	if got != "before" {
		t.Errorf("got %q, want 'before' (args evaluated at defer time)", got)
	}
}

func TestDeferClosureVsArg(t *testing.T) {
	closure, arg := DeferClosureVsArg()
	if closure != "modified" {
		t.Errorf("closure = %q, want 'modified' (reads latest)", closure)
	}
	if arg != "initial" {
		t.Errorf("arg = %q, want 'initial' (fixed at defer time)", arg)
	}
}

func TestDeferModifyReturn(t *testing.T) {
	got := DeferModifyReturn()
	if got != 2 {
		t.Errorf("got %d, want 2 (return 1, then defer n++)", got)
	}
}

// ---- Panic / Recover ----

func TestSafeDiv(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		result, err := SafeDiv(10, 3)
		if err != nil {
			t.Fatal(err)
		}
		if result != 3 {
			t.Errorf("got %d", result)
		}
	})
	t.Run("divide by zero", func(t *testing.T) {
		_, err := SafeDiv(10, 0)
		if err == nil {
			t.Fatal("should recover from panic")
		}
		t.Logf("recovered error: %v", err)
	})
}

func TestMustParse(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		if MustParse("42") != 42 {
			t.Error("should parse")
		}
	})
	t.Run("panics on invalid", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("should panic")
			}
		}()
		MustParse("abc")
	})
}

func TestSafeCall(t *testing.T) {
	t.Run("no panic", func(t *testing.T) {
		err := SafeCall(func() { /* nothing */ })
		if err != nil {
			t.Error("should not error")
		}
	})
	t.Run("catches panic string", func(t *testing.T) {
		err := SafeCall(func() { panic("boom") })
		if err == nil {
			t.Fatal("should catch")
		}
		if !strings.Contains(err.Error(), "boom") {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("catches panic error", func(t *testing.T) {
		err := SafeCall(func() { panic(fmt.Errorf("error boom")) })
		if err == nil {
			t.Fatal("should catch")
		}
	})
}

func TestApply(t *testing.T) {
	nums := []int{1, 2, 3}
	doubled := Apply(nums, func(n int) int { return n * 2 })
	if doubled[0] != 2 || doubled[1] != 4 || doubled[2] != 6 {
		t.Errorf("got %v", doubled)
	}
	if nums[0] != 1 {
		t.Error("original modified")
	}
}

func TestNegate(t *testing.T) {
	isEven := Pred[int](func(n int) bool { return n%2 == 0 })
	isOdd := Negate(isEven)
	if !isOdd(3) {
		t.Error("3 is odd")
	}
	if isOdd(4) {
		t.Error("4 is not odd")
	}
}

func TestComposeAll(t *testing.T) {
	addOne := func(n int) int { return n + 1 }
	double := func(n int) int { return n * 2 }
	square := func(n int) int { return n * n }

	// 3 → +1 → 4 → *2 → 8 → ^2 → 64
	f := ComposeAll(addOne, double, square)
	if f(3) != 64 {
		t.Errorf("got %d", f(3))
	}
}

func TestSortBy(t *testing.T) {
	words := []string{"banana", "apple", "cherry"}
	sorted := SortBy(words, func(a, b string) bool { return a < b })
	if sorted[0] != "apple" || sorted[1] != "banana" || sorted[2] != "cherry" {
		t.Errorf("got %v", sorted)
	}
	if words[0] != "banana" {
		t.Error("original modified")
	}
}

func TestMethodValueDemo(t *testing.T) {
	if MethodValueDemo() != 2 {
		t.Error("method value should work")
	}
}

func TestMethodExprDemo(t *testing.T) {
	if MethodExprDemo() != 2 {
		t.Error("method expression should work")
	}
}
