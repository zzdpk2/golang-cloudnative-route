package language

import (
	"reflect"
	"testing"
)

func TestConstantsAndEvaluation(t *testing.T) {
	integer, fractional := ConstantQuotients()
	if integer != 1 || fractional != 1.5 {
		t.Errorf("ConstantQuotients() = (%d, %g)", integer, fractional)
	}
	if got := ShiftMask(5); got != 32 {
		t.Errorf("ShiftMask(5) = %d", got)
	}
	if got := ShiftMask(64); got != 0 {
		t.Errorf("ShiftMask(64) = %d", got)
	}
	if left, right := Swap("left", "right"); left != "right" || right != "left" {
		t.Errorf("Swap() = (%q, %q)", left, right)
	}

	values := []int{10}
	calls := 0
	got := AddAt(values, func() int {
		calls++
		return 0
	}, 5)
	if got != 15 || values[0] != 15 || calls != 1 {
		t.Errorf("AddAt() = %d, values=%v, index calls=%d", got, values, calls)
	}
}

func TestDeferPanicAndScope(t *testing.T) {
	if got, want := DeferredValues(10), []int{11, 10}; !reflect.DeepEqual(got, want) {
		t.Errorf("DeferredValues() = %v, want %v", got, want)
	}
	if got := AdjustNamedReturn(10); got != 11 {
		t.Errorf("AdjustNamedReturn() = %d", got)
	}
	if got := PanicValue(func() { panic("boom") }); got != "boom" {
		t.Errorf("PanicValue() = %#v", got)
	}
	if got := PanicValue(func() {}); got != nil {
		t.Errorf("PanicValue(normal) = %#v", got)
	}

	if got, err := ParseOrFallback("42", 7); got != 42 || err != nil {
		t.Errorf("ParseOrFallback(valid) = (%d, %v)", got, err)
	}
	if got, err := ParseOrFallback("nope", 7); got != 7 || err == nil {
		t.Errorf("ParseOrFallback(invalid) = (%d, %v)", got, err)
	}
}

func TestValuePartsAndAddressability(t *testing.T) {
	array := [3]int{1, 2, 3}
	changed := MutateArray(array)
	if array != [3]int{1, 2, 3} || changed != [3]int{2, 2, 3} {
		t.Errorf("array=%v changed=%v", array, changed)
	}

	slice := []int{1, 2, 3}
	MutateSlice(slice)
	if slice[0] != 2 {
		t.Errorf("slice = %v", slice)
	}

	base := make([]int, 2, 8)
	base[0], base[1] = 1, 2
	detached := AppendDetached(base, 3)
	detached[0] = 99
	if base[0] != 1 || !reflect.DeepEqual(detached, []int{99, 2, 3}) {
		t.Errorf("base=%v detached=%v", base, detached)
	}

	original := map[string]int{"a": 1}
	cloned := CloneMap(original)
	cloned["a"] = 2
	if original["a"] != 1 {
		t.Errorf("CloneMap shared state: original=%v cloned=%v", original, cloned)
	}

	a, b, c := 1, 2, 3
	pointers := make([]*int, 3, 3)
	pointers[0], pointers[1], pointers[2] = &a, &b, &c
	shortened := DeletePointerAt(pointers, 1)
	if len(shortened) != 2 || shortened[0] != &a || shortened[1] != &c {
		t.Errorf("DeletePointerAt() = %v", shortened)
	}
	if backing := shortened[:cap(shortened)]; backing[2] != nil {
		t.Error("deleted pointer is retained in the backing array")
	}

	variadic := []int{1, 2, 3}
	DoubleVariadic(variadic...)
	if !reflect.DeepEqual(variadic, []int{2, 4, 6}) {
		t.Errorf("DoubleVariadic() left %v", variadic)
	}

	counters := map[string]Counter{"orders": {Value: 4}}
	IncrementMapCounter(counters, "orders")
	if counters["orders"].Value != 5 {
		t.Errorf("counter = %+v", counters["orders"])
	}
}

func TestNilMapAndChannelSemantics(t *testing.T) {
	ready := make(chan int, 1)
	ready <- 42
	if got, ok := FirstReady(nil, ready); !ok || got != 42 {
		t.Errorf("FirstReady() = (%d, %t)", got, ok)
	}
	if got, ok := FirstReady(nil, nil); ok || got != 0 {
		t.Errorf("FirstReady(nil,nil) = (%d, %t)", got, ok)
	}

	values := EnsureMap(nil, "answer", 42)
	if values["answer"] != 42 {
		t.Errorf("EnsureMap() = %v", values)
	}
}
