package collection

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestChunkAndFrequency(t *testing.T) {
	if got, want := Chunk([]int{1, 2, 3, 4, 5}, 2), [][]int{{1, 2}, {3, 4}, {5}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Chunk() = %v, want %v", got, want)
	}
	if got := Chunk([]int{1}, 0); got != nil {
		t.Fatalf("Chunk() with invalid size = %v, want nil", got)
	}
	if got, want := Frequency([]string{"go", "tdd", "go"}), map[string]int{"go": 2, "tdd": 1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Frequency() = %v, want %v", got, want)
	}
}

func TestMap(t *testing.T) {
	t.Run("int to string", func(t *testing.T) {
		nums := []int{1, 2, 3}
		result := Map(nums, strconv.Itoa)
		want := []string{"1", "2", "3"}
		assertSliceEqual(t, result, want)
	})

	t.Run("string to length", func(t *testing.T) {
		words := []string{"go", "rust", "python"}
		result := Map(words, func(s string) int { return len(s) })
		want := []int{2, 4, 6}
		assertSliceEqual(t, result, want)
	})

	t.Run("empty slice", func(t *testing.T) {
		result := Map([]int{}, func(n int) string { return "" })
		if len(result) != 0 {
			t.Error("should return empty slice")
		}
	})
}

func TestReduce(t *testing.T) {
	t.Run("sum", func(t *testing.T) {
		nums := []int{1, 2, 3, 4, 5}
		sum := Reduce(nums, 0, func(acc, n int) int { return acc + n })
		if sum != 15 {
			t.Errorf("sum = %d, want 15", sum)
		}
	})

	t.Run("concat strings", func(t *testing.T) {
		words := []string{"hello", " ", "world"}
		result := Reduce(words, "", func(acc, s string) string { return acc + s })
		if result != "hello world" {
			t.Errorf("result = %q", result)
		}
	})

	t.Run("count chars", func(t *testing.T) {
		words := []string{"go", "rust", "c"}
		total := Reduce(words, 0, func(acc int, s string) int { return acc + len(s) })
		if total != 6 {
			t.Errorf("total = %d, want 6", total)
		}
	})
}

func TestFilter(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5, 6}
	even := Filter(nums, func(n int) bool { return n%2 == 0 })
	want := []int{2, 4, 6}
	assertSliceEqual(t, even, want)
}

func TestFind(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}

	t.Run("found", func(t *testing.T) {
		v, ok := Find(nums, func(n int) bool { return n > 3 })
		if !ok || v != 4 {
			t.Errorf("Find = %d, %v", v, ok)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, ok := Find(nums, func(n int) bool { return n > 10 })
		if ok {
			t.Error("should not find")
		}
	})
}

func TestContains(t *testing.T) {
	nums := []int{1, 2, 3}
	if !Contains(nums, 2) {
		t.Error("should contain 2")
	}
	if Contains(nums, 5) {
		t.Error("should not contain 5")
	}

	words := []string{"go", "rust"}
	if !Contains(words, "go") {
		t.Error("should contain 'go'")
	}
}

func TestUnique(t *testing.T) {
	t.Run("ints", func(t *testing.T) {
		nums := []int{1, 2, 2, 3, 3, 3}
		result := Unique(nums)
		if len(result) != 3 {
			t.Errorf("len = %d, want 3", len(result))
		}
		for _, v := range []int{1, 2, 3} {
			if !Contains(result, v) {
				t.Errorf("should contain %d", v)
			}
		}
	})

	t.Run("strings", func(t *testing.T) {
		words := []string{"a", "b", "a", "c", "b"}
		result := Unique(words)
		if len(result) != 3 {
			t.Errorf("len = %d", len(result))
		}
	})
}

func TestGroupBy(t *testing.T) {
	words := []string{"go", "rust", "c", "java", "js"}
	groups := GroupBy(words, func(s string) int { return len(s) })

	if len(groups[2]) != 3 { // "go", "c", "js"
		t.Errorf("len-2 group = %v", groups[2])
	}
	if len(groups[4]) != 2 { // "rust", "java"
		t.Errorf("len-4 group = %v", groups[4])
	}
}

func TestSortBy(t *testing.T) {
	type person struct {
		Name string
		Age  int
	}
	people := []person{
		{"Charlie", 30},
		{"Alice", 25},
		{"Bob", 28},
	}

	byName := SortBy(people, func(p person) string { return p.Name })
	if byName[0].Name != "Alice" || byName[1].Name != "Bob" || byName[2].Name != "Charlie" {
		t.Errorf("sort by name: %v", byName)
	}

	if people[0].Name != "Charlie" {
		t.Error("original slice should not be modified")
	}

	byAge := SortBy(people, func(p person) int { return p.Age })
	if byAge[0].Age != 25 {
		t.Errorf("sort by age: %v", byAge)
	}
}

func TestSum(t *testing.T) {
	t.Run("ints", func(t *testing.T) {
		if Sum([]int{1, 2, 3}) != 6 {
			t.Error("sum ints")
		}
	})

	t.Run("float64", func(t *testing.T) {
		result := Sum([]float64{1.5, 2.5, 3.0})
		if result != 7.0 {
			t.Errorf("sum = %f", result)
		}
	})

	t.Run("custom type", func(t *testing.T) {
		type Score int
		scores := []Score{10, 20, 30}
		if Sum(scores) != 60 {
			t.Error("sum custom type")
		}
	})
}

func TestMaxMin(t *testing.T) {
	t.Run("max", func(t *testing.T) {
		v, ok := Max([]int{3, 1, 4, 1, 5})
		if !ok || v != 5 {
			t.Errorf("Max = %d", v)
		}
	})

	t.Run("min", func(t *testing.T) {
		v, ok := Min([]int{3, 1, 4, 1, 5})
		if !ok || v != 1 {
			t.Errorf("Min = %d", v)
		}
	})

	t.Run("empty", func(t *testing.T) {
		_, ok := Max([]int{})
		if ok {
			t.Error("empty should return false")
		}
	})

	t.Run("string max", func(t *testing.T) {
		v, _ := Max([]string{"banana", "apple", "cherry"})
		if v != "cherry" {
			t.Errorf("max string = %q", v)
		}
	})
}

func TestPair(t *testing.T) {
	p := NewPair("name", 42)
	if p.Key != "name" || p.Value != 42 {
		t.Errorf("Pair = %v", p)
	}
}

func TestZipUnzip(t *testing.T) {
	names := []string{"Alice", "Bob", "Charlie"}
	ages := []int{25, 30, 35, 40} // See the corresponding tests for the intended behavior.

	pairs := Zip(names, ages)
	if len(pairs) != 3 {
		t.Fatalf("len = %d, want 3", len(pairs))
	}
	if pairs[0].Key != "Alice" || pairs[0].Value != 25 {
		t.Errorf("pairs[0] = %v", pairs[0])
	}

	ns, as := Unzip(pairs)
	assertSliceEqual(t, ns, names)
	assertSliceEqual(t, as, []int{25, 30, 35})
}

func TestOptional(t *testing.T) {
	t.Run("Some", func(t *testing.T) {
		opt := Some(42)
		if !opt.IsPresent() {
			t.Error("should be present")
		}
		v, ok := opt.Get()
		if !ok || v != 42 {
			t.Errorf("Get = %d, %v", v, ok)
		}
	})

	t.Run("None", func(t *testing.T) {
		opt := None[string]()
		if opt.IsPresent() {
			t.Error("should not be present")
		}
		_, ok := opt.Get()
		if ok {
			t.Error("Get should return false")
		}
	})

	t.Run("OrElse", func(t *testing.T) {
		some := Some(10)
		if some.OrElse(0) != 10 {
			t.Error("should return value")
		}
		none := None[int]()
		if none.OrElse(99) != 99 {
			t.Error("should return default")
		}
	})

	t.Run("Map", func(t *testing.T) {
		doubled := Some(5).Map(func(n int) int { return n * 2 })
		v, _ := doubled.Get()
		if v != 10 {
			t.Errorf("Map result = %d", v)
		}

		none := None[int]().Map(func(n int) int { return n * 2 })
		if none.IsPresent() {
			t.Error("Map on None should be None")
		}
	})
}

func TestFunctionalPipeline(t *testing.T) {
	words := []string{"Hello", "World", "Go", "Is", "Great"}

	result := Reduce(
		Filter(
			Map(words, strings.ToLower),
			func(s string) bool { return len(s) > 2 },
		),
		"",
		func(acc, s string) string {
			if acc == "" {
				return s
			}
			return acc + " " + s
		},
	)

	if result != "hello world great" {
		t.Errorf("pipeline result = %q", result)
	}
}

// ---- helper ----
func assertSliceEqual[T comparable](t *testing.T, got, want []T) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len(got)=%d, len(want)=%d; got=%v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("[%d] got %v, want %v", i, got[i], want[i])
		}
	}
}
