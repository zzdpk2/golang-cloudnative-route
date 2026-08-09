// Package language contains compact foundation exercises for the Router track.
//
// Each required entry point combines several language rules in one realistic
// contract. Elementary algorithms are intentionally not repeated here when a
// later concurrency, collection, error, or domain exercise already owns them.
package language

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
	panic("TODO")
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
	panic("TODO")
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
	panic("TODO")
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
	panic("TODO optional")
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
