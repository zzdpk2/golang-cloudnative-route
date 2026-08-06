package performance

import "testing"

func TestBoundsCheckExercises(t *testing.T) {
	if got := SumIndex([]int{1, 2, 3, 4}); got != 10 {
		t.Errorf("SumIndex() = %d", got)
	}
	if got := SumPairs([]int{1, 2, 3, 4}); got != 15 {
		t.Errorf("SumPairs() = %d", got)
	}
	if got := SumPairs(nil); got != 0 {
		t.Errorf("SumPairs(nil) = %d", got)
	}
	if got, ok := SumFour([]int{1, 2, 3, 4, 100}); !ok || got != 10 {
		t.Errorf("SumFour() = (%d, %t)", got, ok)
	}
	if got, ok := SumFour([]int{1, 2, 3}); ok || got != 0 {
		t.Errorf("SumFour(short) = (%d, %t)", got, ok)
	}
}
