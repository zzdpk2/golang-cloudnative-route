package iter

import (
	"fmt"
	"strings"
	"testing"
)

func TestFromSlice(t *testing.T) {
	seq := FromSlice([]int{10, 20, 30})
	got := ToSlice(seq)
	assertSlice(t, got, []int{10, 20, 30})
}

func TestFromSlice_Empty(t *testing.T) {
	seq := FromSlice([]int{})
	got := ToSlice(seq)
	if len(got) != 0 {
		t.Errorf("got %v", got)
	}
}

func TestRange(t *testing.T) {
	assertSlice(t, ToSlice(Range(0, 5)), []int{0, 1, 2, 3, 4})
	assertSlice(t, ToSlice(Range(3, 3)), []int{}) // empty
	assertSlice(t, ToSlice(Range(-2, 2)), []int{-2, -1, 0, 1})
}

func TestRangeStep(t *testing.T) {
	assertSlice(t, ToSlice(RangeStep(0, 10, 3)), []int{0, 3, 6, 9})
	assertSlice(t, ToSlice(RangeStep(10, 0, -2)), []int{10, 8, 6, 4, 2})
}

func TestRepeat(t *testing.T) {
	got := ToSlice(Take(Repeat("x"), 4))
	assertSlice(t, got, []string{"x", "x", "x", "x"})
}

func TestRepeatN(t *testing.T) {
	assertSlice(t, ToSlice(RepeatN(42, 3)), []int{42, 42, 42})
}

func TestGenerate(t *testing.T) {
	powers := Generate(1, func(n int) int { return n * 2 })
	assertSlice(t, ToSlice(Take(powers, 5)), []int{1, 2, 4, 8, 16})
}

func TestFibonacci(t *testing.T) {
	assertSlice(t, ToSlice(Take(Fibonacci(), 8)), []int{0, 1, 1, 2, 3, 5, 8, 13})
}

func TestMap(t *testing.T) {
	seq := Map(FromSlice([]int{1, 2, 3}), func(n int) string {
		return fmt.Sprintf("item-%d", n)
	})
	assertSlice(t, ToSlice(seq), []string{"item-1", "item-2", "item-3"})
}

func TestMap_Lazy(t *testing.T) {
	callCount := 0
	seq := Map(Range(0, 1000), func(n int) int {
		callCount++
		return n * 2
	})
	ToSlice(Take(seq, 3))
	if callCount != 3 {
		t.Errorf("callCount = %d, want 3 (lazy evaluation)", callCount)
	}
}

func TestFilter(t *testing.T) {
	even := Filter(Range(0, 10), func(n int) bool { return n%2 == 0 })
	assertSlice(t, ToSlice(even), []int{0, 2, 4, 6, 8})
}

func TestTake(t *testing.T) {
	assertSlice(t, ToSlice(Take(Range(0, 100), 3)), []int{0, 1, 2})
}

func TestTakeWhile(t *testing.T) {
	seq := TakeWhile(FromSlice([]int{2, 4, 6, 7, 8, 10}), func(n int) bool {
		return n%2 == 0
	})
	assertSlice(t, ToSlice(seq), []int{2, 4, 6})
}

func TestSkip(t *testing.T) {
	assertSlice(t, ToSlice(Skip(Range(0, 10), 7)), []int{7, 8, 9})
}

func TestFlatMap(t *testing.T) {
	seq := FlatMap(FromSlice([]int{1, 2, 3}), func(n int) Seq[int] {
		return FromSlice([]int{n, n * 10})
	})
	assertSlice(t, ToSlice(seq), []int{1, 10, 2, 20, 3, 30})
}

func TestFlatMap_Words(t *testing.T) {
	sentences := FromSlice([]string{"hello world", "foo bar baz"})
	words := FlatMap(sentences, func(s string) Seq[string] {
		return FromSlice(strings.Fields(s))
	})
	assertSlice(t, ToSlice(words), []string{"hello", "world", "foo", "bar", "baz"})
}

func TestZip(t *testing.T) {
	names := FromSlice([]string{"Alice", "Bob", "Charlie"})
	ages := FromSlice([]int{25, 30, 35, 40}) // See the corresponding tests for the intended behavior.

	pairs := Zip(names, ages)
	result := ToSlice(pairs)
	if len(result) != 3 {
		t.Fatalf("len = %d", len(result))
	}
	if result[0].First != "Alice" || result[0].Second != 25 {
		t.Errorf("[0] = %v", result[0])
	}
}

func TestEnumerate(t *testing.T) {
	seq := Enumerate(FromSlice([]string{"a", "b", "c"}))
	result := ToSlice(seq)
	if result[0].Index != 0 || result[0].Value != "a" {
		t.Error("[0]")
	}
	if result[2].Index != 2 || result[2].Value != "c" {
		t.Error("[2]")
	}
}

func TestChain(t *testing.T) {
	a := FromSlice([]int{1, 2})
	b := FromSlice([]int{3, 4, 5})
	assertSlice(t, ToSlice(Chain(a, b)), []int{1, 2, 3, 4, 5})
}

func TestScan(t *testing.T) {
	seq := Scan(FromSlice([]int{1, 2, 3}), 0, func(acc, v int) int { return acc + v })
	assertSlice(t, ToSlice(seq), []int{0, 1, 3, 6})
}

func TestReduce(t *testing.T) {
	sum := Reduce(FromSlice([]int{1, 2, 3, 4, 5}), 0, func(acc, v int) int { return acc + v })
	if sum != 15 {
		t.Errorf("sum = %d", sum)
	}
}

func TestForEach(t *testing.T) {
	var collected []int
	ForEach(FromSlice([]int{10, 20, 30}), func(n int) { collected = append(collected, n) })
	assertSlice(t, collected, []int{10, 20, 30})
}

func TestCount(t *testing.T) {
	if Count(Range(0, 10)) != 10 {
		t.Error("count")
	}
	if Count(FromSlice([]int{})) != 0 {
		t.Error("empty")
	}
}

func TestLast(t *testing.T) {
	v, ok := Last(FromSlice([]int{1, 2, 3}))
	if !ok || v != 3 {
		t.Errorf("Last = %d, %v", v, ok)
	}

	_, ok = Last(FromSlice([]int{}))
	if ok {
		t.Error("empty should return false")
	}
}

func TestAny(t *testing.T) {
	if !Any(Range(0, 100), func(n int) bool { return n == 50 }) {
		t.Error("should find 50")
	}
	if Any(Range(0, 10), func(n int) bool { return n > 20 }) {
		t.Error("no > 20 in 0..10")
	}
}

func TestAll(t *testing.T) {
	if !All(Range(0, 5), func(n int) bool { return n < 10 }) {
		t.Error("all < 10")
	}
	if All(Range(0, 10), func(n int) bool { return n < 5 }) {
		t.Error("not all < 5")
	}
}

func TestCombinedPipeline(t *testing.T) {
	result := ToSlice(
		Take(
			Map(
				Filter(
					Generate(1, func(n int) int { return n + 1 }),
					func(n int) bool { return n%3 == 0 },
				),
				func(n int) int { return n * n },
			),
			5,
		),
	)
	// 3,6,9,12,15 → 9,36,81,144,225
	assertSlice(t, result, []int{9, 36, 81, 144, 225})
}

func TestLazyVsEager(t *testing.T) {
	eagerCallCount := 0
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}
	var filtered []int
	for _, v := range data {
		eagerCallCount++
		if v%2 == 0 {
			filtered = append(filtered, v)
		}
	}
	_ = filtered[:3]

	lazyCallCount := 0
	result := ToSlice(
		Take(
			Filter(
				FromSlice(data),
				func(n int) bool { lazyCallCount++; return n%2 == 0 },
			),
			3,
		),
	)

	assertSlice(t, result, []int{0, 2, 4})
	if lazyCallCount > 10 {
		t.Errorf("lazy processed %d items (should be ~5)", lazyCallCount)
	}
	t.Logf("Eager: %d calls, Lazy: %d calls", eagerCallCount, lazyCallCount)
}

func assertSlice[T comparable](t *testing.T, got, want []T) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len=%d want %d; got=%v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("[%d] got %v, want %v", i, got[i], want[i])
		}
	}
}
