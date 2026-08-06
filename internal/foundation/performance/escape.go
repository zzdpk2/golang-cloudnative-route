// Package escape is the allocation laboratory.
//
// Every experiment here operates on the real order types — domain.OrderLine,
// domain.Money, domain.ProductID. "Some toy struct escaped" is a fact; "**the order
// path allocates three extra times per line**" is a bug you can go and fix.
//
// # How to read this package
//
// Go decides *where* a value lives by asking one question: can a reference to
// it outlive the function that created it? If yes, it goes on the heap. That
// analysis happens at compile time and does not care whether you wrote `new`,
// `&x`, or a plain literal.
//
// Ask the compiler directly, any time:
//
//	go build -gcflags="-m -m" ./internal/foundation/performance
//
// # Suggested order
//
//  1. NewLineValue / NewLinePointer   the base case
//  2. BoxLine                         allocation with no pointer in sight
//  3. Format* and Join*               where string handling costs you
//  4. Build* and Sum*                 slices, and where the cost is really paid
//  5. MakeLineCounter / SendLine      closures and channels
//  6. the pool                        when reuse beats allocation
//
// # Two things about the fixtures
//
// Every price is domain.Zero. That is deliberate, not laziness: **allocation and
// copy cost depend on a type's size and lifetime, never on its contents.** An
// OrderLine holding A$0.00 costs exactly what one holding A$19.99 costs. It
// also means this package works before L1 is finished.
//
// And the tests assert *allocation counts*, not return values. Before this
// rewrite they only checked the numbers came out right, so you could get a
// green bar having learned nothing about escape analysis. Now the red/green
// light hangs on the thing the package is about.
package performance

import (
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"sync"
)

// ---- Value versus pointer ----

// NewLineValue builds a line and returns it by value.
//
//	NewLineValue("sku-1", "Widget")
//	  → OrderLine{ProductID: "sku-1", Name: "Widget", Price: A$0.00 AUD}
//
// Price must be a zero AUD amount — domain.Zero is the only Money constructor that
// works without L1. Quantity stays at its zero value.
//
// **Expected: 0 allocations.** Nothing outside can reference the line, so it
// belongs on the caller's stack.
func NewLineValue(sku domain.ProductID, name string) domain.OrderLine {
	panic("TODO")
}

// NewLinePointer builds the same line and returns a pointer to it.
//
//	NewLinePointer("sku-2", "Gadget") → &OrderLine{ProductID: "sku-2", ...}
//
// **Expected: exactly 1 allocation.** The pointer outlives this call, so the
// line has to be on the heap.
//
// The lesson is not "pointers are bad". It is that returning a pointer is a
// decision with a price, and OrderLine is a value type — copying it is cheap
// and keeps it off the heap. Check what lab/layout reports for Sizeof before
// deciding which side of that trade you want.
func NewLinePointer(sku domain.ProductID, name string) *domain.OrderLine {
	panic("TODO")
}

// BoxLine puts a line into an interface and returns it.
//
//	BoxLine(OrderLine{...}) → any holding a copy of that OrderLine
//
// **Expected: more than 0 allocations** — and this is the one that surprises
// people, because no pointer appears anywhere in the code.
//
// An interface value is a (type, pointer) pair. A value that is not already
// pointer-shaped therefore has to be copied somewhere the pointer can point at,
// and that somewhere is the heap.
//
// This is why fmt.Println(line) allocates, why []any is expensive, and why
// logging inside a hot loop costs more than the logging itself.
func BoxLine(l domain.OrderLine) any {
	panic("TODO")
}

// ---- Strings ----

// FormatLineSprintf renders a line as "<sku>:<name>" using fmt.Sprintf.
//
//	line{ProductID: "sku-1", Name: "Widget"} → "sku-1:Widget"
//
// Sprintf takes ...any, so every argument is boxed exactly as in BoxLine, and
// then reflected on at run time. Convenient, and the most common accidental
// allocation in Go code.
func FormatLineSprintf(l domain.OrderLine) string {
	panic("TODO")
}

// FormatLineBuilder produces the identical string with a strings.Builder.
//
//	line{ProductID: "sku-1", Name: "Widget"} → "sku-1:Widget"
//
// **Expected: fewer allocations than FormatLineSprintf.** Predict the ratio
// before you measure it:
//
//	go test ./internal/foundation/performance -bench=Format -benchmem
//
// Builder still allocates — it has to produce a string — but it does not box,
// does not reflect, and grows one buffer instead of several. Calling Grow up
// front removes the regrowth too; try it and watch the number move.
func FormatLineBuilder(l domain.OrderLine) string {
	panic("TODO")
}

// JoinSKUsConcat joins the SKUs with "," using += in a loop.
//
//	[]OrderLine{{ProductID:"a"}, {ProductID:"b"}, {ProductID:"c"}} → "a,b,c"
//	[]OrderLine{{ProductID:"solo"}}                                → "solo"
//	nil                                                            → ""
//
// No trailing comma, and no leading one. Strings are immutable, so each +=
// builds a whole new string and copies everything so far — making the loop
// O(n²) in bytes copied. Fine for three lines, catastrophic for three thousand.
func JoinSKUsConcat(lines []domain.OrderLine) string {
	panic("TODO")
}

// JoinSKUsBuilder produces the identical output with a strings.Builder.
//
// Same three examples as above, same results. **Expected: far fewer
// allocations, and the gap widens with n.** Run the benchmark at a couple of
// sizes and look at the shape of the curve rather than one number.
func JoinSKUsBuilder(lines []domain.OrderLine) string {
	panic("TODO")
}

// SKUFromBytes converts a byte slice into a ProductID.
//
//	SKUFromBytes([]byte("sku-12345")) → ProductID("sku-12345")
//
// **Expected: exactly 1 allocation, and the result must not alias the input.**
// The test mutates the source bytes afterwards and checks the ProductID did not
// change.
//
// Converting []byte to string copies precisely because strings are immutable
// and the caller still owns the bytes. That copy is why decoding a large JSON
// body costs more than the parsing does.
//
// There is a well-known unsafe trick that skips the copy. Look it up, work out
// exactly which precondition makes it safe, and note that such things belong in
// lab/, not in a request path.
func SKUFromBytes(b []byte) domain.ProductID {
	panic("TODO")
}

// ---- Slices ----

// BuildLinesAppend builds n lines by appending to a nil slice.
//
//	BuildLinesAppend(3) → lines with SKUs "sku-0", "sku-1", "sku-2"
//	BuildLinesAppend(0) → an empty slice
//
// Name each line "Widget" and price it with domain.Zero(domain.AUD), so the two
// builders produce identical output.
//
// **Expected: several allocations.** append doubles the backing array when it
// fills, so this reallocates roughly log₂(n) times and copies what is already
// there each time.
func BuildLinesAppend(n int) []domain.OrderLine {
	panic("TODO")
}

// BuildLinesPrealloc builds the identical slice with the capacity known ahead.
//
// Same examples, same output as BuildLinesAppend.
//
// **Expected: exactly 1 allocation.** The cheapest performance habit in Go:
// when you know the final length, say so.
func BuildLinesPrealloc(n int) []domain.OrderLine {
	panic("TODO")
}

// SumCentsValue totals the prices of lines held by value.
//
//	testLines(10) → 0   (every fixture price is domain.Zero)
//	nil           → 0
//
// **Expected: 0 allocations.** The slice already exists, and each line copied
// by the loop dies at the end of its iteration.
//
// The copy is not free even though it does not allocate — an OrderLine is large
// enough that copying one per iteration shows up in the benchmark. Ranging by
// index avoids it. Measure both before deciding it matters.
func SumCentsValue(lines []domain.OrderLine) int64 {
	panic("TODO")
}

// SumCentsPointer totals the prices of lines held by pointer.
//
// Same results as SumCentsValue for equivalent data.
//
// **Expected: 0 allocations *here*** — but building that []*OrderLine cost one
// allocation per line, somewhere else. That is the trap this pair exposes:
// allocation is paid at construction and charged to whoever built the data, so
// a function that looks free can be the reason another one is not.
//
// It also costs on every read. The lines are scattered across the heap rather
// than laid out contiguously, so the CPU cannot prefetch them. Run both
// benchmarks at n=4096 and look at the gap.
func SumCentsPointer(lines []*domain.OrderLine) int64 {
	panic("TODO")
}

// ---- Closures and channels ----

// MakeLineCounter returns a closure over a counter, incrementing before it
// returns.
//
//	next := MakeLineCounter(10)
//	next() → 11
//	next() → 12
//
// **Expected: more than 0 allocations.** The captured variable outlives this
// function, so it moves to the heap even though it is a plain int and no
// pointer was ever written.
//
// Same mechanism as Order.LineProcessors in L2, seen from the allocation side.
func MakeLineCounter(start int) func() int {
	panic("TODO")
}

// SendLine builds a line and sends a pointer to it on ch.
//
//	ch := make(chan *domain.OrderLine, 1)
//	SendLine(ch, "sku-9", "Gadget")
//	<-ch → &OrderLine{ProductID: "sku-9", Name: "Gadget"}
//
// Anything sent on a channel escapes — the compiler cannot know how long the
// receiver will hold it. Every stage in platform/channel pays this, which is
// worth knowing before you build a hundred-stage pipeline.
func SendLine(ch chan<- *domain.OrderLine, sku domain.ProductID, name string) {
	panic("TODO")
}

// ---- sync.Pool ----

var linePool = sync.Pool{
	New: func() any {
		return &domain.OrderLine{}
	},
}

// GetLineFromPool takes a line from the pool and initialises it.
//
//	GetLineFromPool("sku-1", "Widget")
//	  → *OrderLine{ProductID: "sku-1", Name: "Widget", Price: A$0.00 AUD}
//
// Whatever comes back must carry the values just passed in, whether it was
// recycled or freshly made — the test checks exactly that.
//
// A pool trades allocation for reuse, and it is worth it only when the object
// is large, allocated at high frequency, and has a clear end of life: a request
// buffer, not an order.
//
// Two things people get wrong. The pool may be emptied at any GC, so it is a
// cache and never a guarantee. And New returns `any`, so every Get costs a type
// assertion.
func GetLineFromPool(sku domain.ProductID, name string) *domain.OrderLine {
	panic("TODO")
}

// PutLineToPool returns a line to the pool.
//
//	PutLineToPool(line) → line is reset and available for reuse
//	PutLineToPool(nil)  → does nothing, must not panic
//
// Reset it before putting it back. A pooled object still holding the previous
// request's SKU is how a pool turns a performance optimisation into a data leak
// between users — one of the nastiest bugs there is, because it only appears
// under load.
func PutLineToPool(l *domain.OrderLine) {
	panic("TODO")
}
