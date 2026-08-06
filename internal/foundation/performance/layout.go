// Package layout is the memory-layout laboratory.
//
// It answers three questions about the real order types:
//
//	How big is an OrderLine, and how much of that is padding?
//	Does passing one by value actually cost anything?
//	Why does a striped counter get slower when the shards are adjacent?
//
// The subject is domain.Order, domain.OrderLine, and domain.Money — not toy
// structs. "Reordering fields saves 8 bytes" is trivia; "**OrderLine wastes 14
// bytes per line, and an order with 40 lines wastes half a kilobyte**" is a
// number you can act on.
//
// # What decides a struct's size
//
// Every type has an *alignment*: an int64 must start at an address divisible by
// 8, an int32 by 4, a bool by 1. The compiler inserts padding before a field
// whose alignment demands it, and pads the end of the struct so that an array
// of them stays aligned.
//
// Go does **not** reorder your fields. Other languages do; Go deliberately does
// not, so the declaration order you write is the layout you get — which means
// the waste is yours to fix.
//
// # Alignment is not the same as size
//
//	unsafe.Sizeof     how many bytes the value occupies, padding included
//	unsafe.Alignof    the boundary it must start on
//	unsafe.Offsetof   where a field sits inside its struct
//
// All three are compile-time constants. None of them reads memory, and none is
// unsafe in the dangerous sense — despite living in that package.
package performance

import (
	"fmt"
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"reflect"
	"unsafe"
)

// BadLayout and GoodLayout hold identical fields in different orders.
//
// Work out both sizes on paper before running the test. bool is 1 byte and
// int64 is 8, so the question is only where the padding lands.
type BadLayout struct {
	A bool
	B int64
	C bool
	D int64
}

type GoodLayout struct {
	B int64
	D int64
	A bool
	C bool
}

// SizeOf reports how many bytes a value of type T occupies.
//
//	SizeOf[bool]()       → 1
//	SizeOf[int64]()      → 8
//	SizeOf[BadLayout]()  → 32
//	SizeOf[GoodLayout]() → 24
//
// Same four fields, eight bytes apart. Sizeof needs a value, and a zero one is
// enough — the size never depends on the contents.
func SizeOf[T any]() uintptr {
	panic("TODO")
}

// AlignOf reports the boundary a value of type T must start on.
//
//	AlignOf[bool]()      → 1
//	AlignOf[int64]()     → 8
//	AlignOf[BadLayout]() → 8   (the widest field wins)
//
// A struct's alignment is the largest alignment among its fields, which is why
// one int64 forces the whole struct onto an 8-byte boundary.
func AlignOf[T any]() uintptr {
	panic("TODO")
}

// PaddingWaste reports how many bytes of a struct are padding rather than data.
//
//	PaddingWaste[BadLayout]()  → 14   (32 bytes total, 18 of actual fields)
//	PaddingWaste[GoodLayout]() → 6
//	PaddingWaste[int64]()      → 0    (not a struct: no padding)
//
// Sum the sizes of the fields with reflect and subtract from Sizeof. A non-struct
// type has no padding — return 0 rather than panicking, because the test passes
// one.
//
// This is the number worth carrying around. Run it on OrderLine, multiply by
// the lines in a large order, and decide whether you care. Often you will not —
// and knowing you do not care *because you measured* is the point.
func PaddingWaste[T any]() uintptr {
	panic("TODO")
}

// FieldOffsets reports the byte offset of every field, in declaration order.
//
//	FieldOffsets[BadLayout]()  → [0 8 16 24]
//	FieldOffsets[GoodLayout]() → [0 8 16 17]
//
// Read the gaps: BadLayout jumps 8 bytes after a 1-byte bool, and that gap is
// the padding. GoodLayout packs the two bools next to each other, so the second
// costs 1 byte instead of 8.
//
// A non-struct type has no fields; return nil.
func FieldOffsets[T any]() []uintptr {
	panic("TODO")
}

// ---- The real types ----

// OrderLineSize reports sizeof(domain.OrderLine).
//
// Predict it first. An OrderLine holds a ProductID (a string), a Name (a
// string), a domain.Money, and a domain.Quantity. A string header is 16 bytes on a
// 64-bit build — a pointer and a length — regardless of how long the text is.
//
// **The text itself is not in the struct.** That is why two OrderLines are the
// same size whether the product is called "Mug" or has a 200-character name,
// and why copying one does not copy the strings.
func OrderLineSize() uintptr {
	panic("TODO")
}

// MoneySize reports sizeof(domain.Money).
//
// Money holds an int64 and a Currency (an int). Work out what that comes to,
// and then ask whether Currency needed to be an int — the type has three
// members and could have been a uint8.
//
// Then ask the more useful question: would shrinking it change anything? Money
// is embedded in every OrderLine, so the answer depends on the alignment of
// whatever sits next to it. Predict, then measure with PaddingWaste.
func MoneySize() uintptr {
	panic("TODO")
}

// OrderLineWastePerOrder reports how many padding bytes an order of n lines
// carries.
//
//	OrderLineWastePerOrder(1)   → the waste in one line
//	OrderLineWastePerOrder(40)  → forty times that
//
// A cheap function whose only job is to turn a per-struct number into a
// per-order one, because that is the unit a decision gets made in.
func OrderLineWastePerOrder(lines int) uintptr {
	panic("TODO")
}

// ---- Zero-size types ----

// ZeroSizeDemo has an empty struct wedged between two bytes.
//
// struct{} occupies zero bytes, which is why `map[string]struct{}` is the Go
// idiom for a set and `chan struct{}` for a signal — neither carries data, and
// neither should pay for any.
//
// The quirk worth knowing: a zero-size field at the *end* of a struct is padded
// anyway, because taking its address must not produce a pointer past the end of
// the allocation. In the middle, as here, it costs nothing. Predict the size of
// this struct before you run the test.
type ZeroSizeDemo struct {
	A     byte
	Empty struct{}
	B     byte
}

// SliceMemory estimates the bytes a slice of n elements occupies, ignoring the
// header.
//
//	SliceMemory[int64](100)             → 800
//	SliceMemory[struct{}](1_000_000)    → 0
//	SliceMemory[domain.OrderLine](40) → 40 × sizeof(OrderLine)
//
// Element size times count. Note what it deliberately leaves out: the slice
// header itself, and anything the elements *point at*. A []string of a million
// long strings reports 24MB and occupies far more, because the text lives
// elsewhere — the same header/content split as in OrderLineSize above.
//
// A negative n has no meaningful answer; return 0.
func SliceMemory[T any](n int) uintptr {
	panic("TODO")
}

// ---- False sharing ----

// CacheLineSize is the granularity at which CPUs move memory between cores.
//
// 64 bytes on essentially every current amd64 and arm64 machine. Two variables
// within the same 64-byte span are, as far as cache coherency is concerned, the
// same variable.
const CacheLineSize = 64

// NaiveShard is one slice of a striped counter, laid out the obvious way.
//
// This is exactly the shard from platform/concurrency.StripedCounter in L6, and
// the hint there promised you would come back to it.
//
// The striping was supposed to remove contention: different keys hash to
// different shards, so goroutines updating different shards should never wait
// for one another. Measure it and they still do.
//
// The reason is that a NaiveShard is small enough that several fit in one cache
// line. When core 1 writes shard 0 and core 2 writes shard 1, the hardware sees
// two cores writing the same line and bounces exclusive ownership back and
// forth between them. The code is correct, the locks are uncontended, and the
// cores are queueing anyway. That is **false sharing**.
type NaiveShard struct {
	Count int64
}

// PaddedShard is the same shard, padded so no two share a cache line.
//
// Add a field that pushes the struct out to CacheLineSize. Work out how big it
// needs to be from what NaiveShard already occupies — and note that this is the
// one place where deliberately *wasting* memory is the optimisation.
type PaddedShard struct {
	Count int64
	// TODO: pad this struct out to exactly CacheLineSize bytes.
}

// ShardsPerCacheLine reports how many T fit in a single cache line.
//
//	ShardsPerCacheLine[NaiveShard]()  → 8
//	ShardsPerCacheLine[PaddedShard]() → 1
//
// Anything above 1 means goroutines touching neighbouring shards are contending
// in hardware. Return 0 for a type larger than a cache line.
func ShardsPerCacheLine[T any]() int {
	panic("TODO")
}

// SumNaive and SumPadded exist so the benchmark has something to run. Both add
// up the counters; the interesting difference is in the concurrent benchmark,
// not in these.
func SumNaive(shards []NaiveShard) int64 {
	panic("TODO")
}

// SumPadded totals the padded shards. Identical work to SumNaive — the
// difference only shows in the concurrent benchmark.
func SumPadded(shards []PaddedShard) int64 {
	panic("TODO")
}

// Placeholders keeping the imports valid while the stubs above are
// unimplemented. Note that unsafe.Sizeof cannot appear here as a value: it is a
// built-in, not a function, so it must be called rather than referenced.
var (
	_ = unsafe.Sizeof(int64(0))
	_ = reflect.TypeOf
	_ = fmt.Sprintf
	_ = domain.OrderLine{}
	_ = domain.Money{}
)
