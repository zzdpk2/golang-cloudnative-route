package language

import (
	"reflect"
	"testing"
)

func TestDataFoundations(t *testing.T) {
	t.Run("slice ownership and range copies", func(t *testing.T) {
		original := []Point{{X: 1, Y: 2}, {X: 3, Y: 4}}
		detached, err := TransformPoints(original, 10, -1, MoveDetached)
		if err != nil {
			t.Fatal(err)
		}
		if want := []Point{{X: 11, Y: 1}, {X: 13, Y: 3}}; !reflect.DeepEqual(detached, want) {
			t.Fatalf("detached = %#v, want %#v", detached, want)
		}
		if want := []Point{{X: 1, Y: 2}, {X: 3, Y: 4}}; !reflect.DeepEqual(original, want) {
			t.Fatalf("detached mode mutated input: %#v", original)
		}
		detached[0].X = 99
		if original[0].X == 99 {
			t.Fatal("detached result still aliases input")
		}

		inPlace, err := TransformPoints(original, 1, 1, MoveInPlace)
		if err != nil {
			t.Fatal(err)
		}
		if want := []Point{{X: 2, Y: 3}, {X: 4, Y: 5}}; !reflect.DeepEqual(inPlace, want) {
			t.Fatalf("in-place result = %#v, want %#v", inPlace, want)
		}
		inPlace[0].X = 77
		if original[0].X != 77 {
			t.Fatal("in-place result must share the caller's backing array")
		}
		if _, err := TransformPoints(nil, 0, 0, MoveMode(99)); err == nil {
			t.Fatal("unknown move mode must be rejected")
		}
	})

	t.Run("UTF-8 uses runes without losing byte offsets", func(t *testing.T) {
		got, err := AnalyzeText("aé🙂a", 3)
		if err != nil {
			t.Fatal(err)
		}
		want := TextReport{
			RuneCount:   4,
			Runes:       []rune{'a', 'é', '🙂', 'a'},
			ByteOffsets: []int{0, 1, 3, 7},
			Reversed:    "a🙂éa",
			Prefix:      "aé🙂",
			Frequency:   map[rune]int{'a': 2, 'é': 1, '🙂': 1},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("report = %#v, want %#v", got, want)
		}
		if _, err := AnalyzeText("x", -1); err == nil {
			t.Fatal("negative rune limit must be rejected")
		}
		zero, err := AnalyzeText("é", 0)
		if err != nil || zero.Prefix != "" {
			t.Fatalf("zero limit = (%#v, %v)", zero, err)
		}
		large, err := AnalyzeText("é", 10)
		if err != nil || large.Prefix != "é" {
			t.Fatalf("large limit = (%#v, %v)", large, err)
		}
	})

	t.Run("maps become deterministic and lossless", func(t *testing.T) {
		got, err := AnalyzeMap(map[string]int{"gamma": 3, "alpha": 1, "beta": 2})
		if err != nil {
			t.Fatal(err)
		}
		want := MapReport{
			Keys:    []string{"alpha", "beta", "gamma"},
			Values:  []int{1, 2, 3},
			Total:   6,
			Inverse: map[int]string{1: "alpha", 2: "beta", 3: "gamma"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("report = %#v, want %#v", got, want)
		}
		if _, err := AnalyzeMap(map[string]int{"a": 1, "b": 1}); err == nil {
			t.Fatal("inverse collision must be rejected")
		}
		empty, err := AnalyzeMap(nil)
		if err != nil {
			t.Fatal(err)
		}
		if empty.Keys == nil || empty.Values == nil || empty.Inverse == nil || empty.Total != 0 {
			t.Fatalf("nil map report must contain allocated empty collections: %#v", empty)
		}
	})

	t.Run("non-comparable values need an equality policy", func(t *testing.T) {
		if !(LabelSet{Owner: "router"}).Equal(LabelSet{Owner: "router", Names: []string{}}) {
			t.Fatal("nil and empty names should be equal")
		}
		if (LabelSet{Owner: "router", Names: []string{"a", "b"}}).Equal(LabelSet{Owner: "router", Names: []string{"b", "a"}}) {
			t.Fatal("name order is significant")
		}
		if !(LabelSet{Owner: "router", Names: []string{"a", "b"}}).Equal(LabelSet{Owner: "router", Names: []string{"a", "b"}}) {
			t.Fatal("equal non-empty label sets must compare equal")
		}
		if (LabelSet{Owner: "router"}).Equal(LabelSet{Owner: "worker"}) {
			t.Fatal("owner is part of equality")
		}
	})
}
