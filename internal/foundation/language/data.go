// Package types contains compact foundation exercises for the Router track.
//
// Each required entry point combines several language rules in one realistic
// contract. Elementary algorithms are intentionally not repeated here when a
// later concurrency, collection, error, or domain exercise already owns them.
package language

import (
	"errors"
	"sort"
)

// Point is deliberately a value type so the ownership exercise can make slice
// aliasing and range-variable copies visible.
type Point struct {
	X int
	Y int
}

// MoveMode states whether a transformation may modify the caller's backing
// array. The API makes ownership explicit instead of hiding it in a toy helper.
type MoveMode uint8

const (
	MoveDetached MoveMode = iota
	MoveInPlace
)

// TransformPoints moves every point by dx and dy.
//
// MoveDetached must return independent storage and leave input unchanged.
// MoveInPlace must update input and return a slice sharing its backing array.
// Reject an unknown mode. Hint: ranging by value gives you a copy of each Point.
func TransformPoints(input []Point, dx, dy int, mode MoveMode) ([]Point, error) {
	// panic("TODO")
	var input1 []Point

	switch mode {
	case MoveDetached:
		input1 = make([]Point, len(input))
		copy(input1, input)

	case MoveInPlace:
		input1 = input

	default:
		return input, errors.New("unknown move mode!")
	}

	for i := range input1 {
		input1[i].X += dx
		input1[i].Y += dy
	}

	return input1, nil

}

// TextReport captures the observations that matter when a Go string contains
// UTF-8. ByteOffsets contains the byte index produced by range for each rune.
type TextReport struct {
	RuneCount   int
	Runes       []rune
	ByteOffsets []int
	Reversed    string
	Prefix      string
	Frequency   map[rune]int
}

// AnalyzeText returns a report for input. Prefix contains at most limit runes,
// never a partial UTF-8 encoding. Reject a negative limit.
//
// Hint: convert once to []rune for rune-oriented operations, but use range on
// the original string to observe byte offsets.
func AnalyzeText(input string, limit int) (TextReport, error) {

	if limit < 0 {
		return TextReport{}, errors.New("negative limit!")
	}

	runes := []rune(input)
	byteOffsets := make([]int, 0, len(runes))

	for i := range input {
		byteOffsets = append(byteOffsets, i)
	}

	runesBk := []rune(input)
	for i, j := 0, len(runesBk)-1; i < j; i, j = i+1, j-1 {
		runesBk[i], runesBk[j] = runesBk[j], runesBk[i]
	}

	prefix := []rune(input)
	border := limit
	if len(prefix) < limit {
		border = len(prefix)
	}

	freq := make(map[rune]int)
	for _, v := range runes {
		freq[v] += 1
	}

	return TextReport{
		RuneCount:   len(runes),
		Runes:       runes,
		ByteOffsets: byteOffsets,
		Reversed:    string(runesBk),
		Prefix:      string(runes[0:border]),
		Frequency:   freq,
	}, nil
}

// MapReport makes map iteration deterministic at the package boundary. Values
// correspond positionally to sorted Keys. Inverse must reject duplicate values
// rather than silently choosing a winner based on randomized iteration order.
type MapReport struct {
	Keys    []string
	Values  []int
	Total   int
	Inverse map[int]string
}

// AnalyzeMap returns an allocated, deterministic report. A nil input is valid.
// Return an error when two keys map to the same value.
func AnalyzeMap(input map[string]int) (MapReport, error) {

	if len(input) == 0 {
		return MapReport{
			Keys:    make([]string, 0),
			Values:  make([]int, 0),
			Total:   0,
			Inverse: make(map[int]string),
		}, nil
	}

	var keys []string
	var values []int

	for k := range input {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, v := range keys {
		values = append(values, input[v])
	}

	total := 0
	for _, v := range values {
		total += v
	}

	invInput := make(map[int]string)
	for k, v := range input {
		k1, v1 := v, k
		_, ok := invInput[k1]
		if ok {
			return MapReport{}, errors.New("inverse collision must be rejected")
		}
		invInput[k1] = v1
	}

	return MapReport{
		Keys:    keys,
		Values:  values,
		Total:   total,
		Inverse: invInput,
	}, nil

}

// Position is a row-column coordinate. {-1, -1} means no match.
// The matrix exercise is optional because labelled control flow is not a Router
// prerequisite.
type Position struct {
	Row int
	Col int
}

// MatrixReport groups the useful outcomes of nested traversal without asking
// learners to reimplement several unrelated toy algorithms.
type MatrixReport struct {
	Flat           []int
	Transpose      [][]int
	FirstMatch     Position
	NonNegativeSum int
}

// AnalyzeMatrix validates that matrix is rectangular, flattens it in row-major
// order, transposes it, sums non-negative values, and records the first value
// accepted by match. A nil match means no value can match.
//
// Hint: a labelled break can stop both loops after the first match; continue
// can skip negative values without nesting the sum logic.
func AnalyzeMatrix(matrix [][]int, match func(int) bool) (MatrixReport, error) {
	if matrix == nil || matrix[0] == nil {
		return MatrixReport{
			Flat:           []int{},
			Transpose:      [][]int{},
			NonNegativeSum: 0,
		}, errors.New("matrix must not be empty")
	}

	m := len(matrix)
	n := len(matrix[0])

	flat := make([]int, 0, m*n)
	for i, _ := range matrix {
		for _, v := range matrix[i] {
			flat = append(flat, v)
		}
	}

}

// LabelSet cannot use == because Names is a slice. Equal must define an
// order-sensitive, nil-equals-empty policy explicitly.
type LabelSet struct {
	Owner string
	Names []string
}

func (s LabelSet) Equal(other LabelSet) bool {
	panic("TODO")
}

// Permission is retained as optional syntax practice for named integer types,
// iota, and bit masks. It does not block the Router foundation track.
type Permission uint8

// EndpointID is a defined type and cannot be assigned from string without an
// explicit conversion. ModelName is an alias and remains interchangeable with
// string. These declarations preserve the distinction without another toy
// implementation function.
type EndpointID string
type ModelName = string

const (
	Read Permission = 1 << iota
	Write
	Execute
)

// SyntaxReport is the single optional syntax exercise. It replaces separate
// drills for conversions, formatting, flags, break, continue, and labels.
type SyntaxReport struct {
	Flags            string
	Truncated        int
	Rounded          int
	Parsed           int
	Decimal          string
	FirstEven        int
	SumUntilNegative int
}

// InspectSyntax renders permissions as rwx (using '-' for missing flags),
// truncates and rounds decimal, parses integerText, formats parsed in base 10,
// finds the first even number, and sums values until the first negative value.
// Return parsing errors to the caller.
func InspectSyntax(permissions Permission, decimal float64, integerText string, values []int) (SyntaxReport, error) {
	panic("TODO optional")
}
