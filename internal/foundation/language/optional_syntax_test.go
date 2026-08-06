//go:build optional

package language

import (
	"reflect"
	"testing"
)

func TestOptionalSyntaxFoundations(t *testing.T) {
	t.Run("defined types, aliases, flags, conversions, and control flow", func(t *testing.T) {
		raw := "model-a"
		var alias ModelName = raw
		defined := EndpointID(raw)
		if alias != string(defined) {
			t.Fatal("defined types require conversion while aliases do not")
		}

		got, err := InspectSyntax(Read|Execute, -3.6, "42", []int{1, 3, 4, -1, 8})
		if err != nil {
			t.Fatal(err)
		}
		want := SyntaxReport{
			Flags:            "r-x",
			Truncated:        -3,
			Rounded:          -4,
			Parsed:           42,
			Decimal:          "42",
			FirstEven:        4,
			SumUntilNegative: 8,
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("syntax report = %#v, want %#v", got, want)
		}
		if _, err := InspectSyntax(0, 0, "not-an-int", nil); err == nil {
			t.Fatal("invalid integer text must be rejected")
		}
	})

	t.Run("matrix traversal, labels, and continue", func(t *testing.T) {
		got, err := AnalyzeMatrix([][]int{{1, -2, 3}, {4, 5, 6}}, func(value int) bool {
			return value > 3
		})
		if err != nil {
			t.Fatal(err)
		}
		want := MatrixReport{
			Flat:           []int{1, -2, 3, 4, 5, 6},
			Transpose:      [][]int{{1, 4}, {-2, 5}, {3, 6}},
			FirstMatch:     Position{Row: 1, Col: 0},
			NonNegativeSum: 19,
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("report = %#v, want %#v", got, want)
		}
		if _, err := AnalyzeMatrix([][]int{{1}, {2, 3}}, nil); err == nil {
			t.Fatal("ragged matrix must be rejected")
		}
		withoutMatch, err := AnalyzeMatrix([][]int{{1, 2}}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if withoutMatch.FirstMatch != (Position{Row: -1, Col: -1}) {
			t.Fatalf("nil predicate match = %#v", withoutMatch.FirstMatch)
		}
	})
}
