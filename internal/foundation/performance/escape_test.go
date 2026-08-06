package performance

import (
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"strings"
	"testing"
)

// Package-level sinks. Assigning results here stops the compiler deleting the
// call entirely — without them an unused result makes the whole thing vanish,
// and every allocation count reads zero for the wrong reason.
var (
	sinkLine    domain.OrderLine
	sinkLinePtr *domain.OrderLine
	sinkAny     any
	sinkString  string
	sinkInt64   int64
	sinkSlice   []domain.OrderLine
	sinkSKU     domain.ProductID
)

func testLines(n int) []domain.OrderLine {
	lines := make([]domain.OrderLine, n)
	for i := range lines {
		lines[i] = domain.OrderLine{
			ProductID: domain.ProductID("sku-" + string(rune('a'+i%26))),
			Name:      "Widget",
			Price:     domain.Zero(domain.AUD),
		}
	}
	return lines
}

func testLinePtrs(n int) []*domain.OrderLine {
	lines := testLines(n)
	ptrs := make([]*domain.OrderLine, n)
	for i := range lines {
		ptrs[i] = &lines[i]
	}
	return ptrs
}

// ---- Behaviour first ----
//
// These pin down what each function returns. Get them green before looking at
// the allocation tests below — an allocation assertion on a function that
// returns the wrong thing tells you nothing.

func TestNewLineValue_FillsTheFields(t *testing.T) {
	got := NewLineValue("sku-1", "Widget")
	if got.ProductID != "sku-1" {
		t.Errorf("ProductID = %q, want %q", got.ProductID, "sku-1")
	}
	if got.Name != "Widget" {
		t.Errorf("Name = %q, want %q", got.Name, "Widget")
	}
	if !got.Price.IsZero() {
		t.Errorf("Price = %s, want a zero amount", got.Price)
	}
	if got.Price.CurrencyType() != domain.AUD {
		t.Errorf("Currency = %s, want AUD", got.Price.CurrencyType())
	}
}

func TestNewLinePointer_FillsTheFields(t *testing.T) {
	got := NewLinePointer("sku-2", "Gadget")
	if got == nil {
		t.Fatal("NewLinePointer returned nil")
	}
	if got.ProductID != "sku-2" || got.Name != "Gadget" {
		t.Errorf("got %+v, want ProductID=sku-2 Name=Gadget", *got)
	}
}

func TestFormatLine_Format(t *testing.T) {
	line := NewLineValue("sku-1", "Widget")

	const want = "sku-1:Widget"
	if got := FormatLineSprintf(line); got != want {
		t.Errorf("FormatLineSprintf = %q, want %q", got, want)
	}
	if got := FormatLineBuilder(line); got != want {
		t.Errorf("FormatLineBuilder = %q, want %q", got, want)
	}
}

func TestJoinSKUs_Format(t *testing.T) {
	lines := []domain.OrderLine{
		{ProductID: "a", Price: domain.Zero(domain.AUD)},
		{ProductID: "b", Price: domain.Zero(domain.AUD)},
		{ProductID: "c", Price: domain.Zero(domain.AUD)},
	}

	const want = "a,b,c"
	if got := JoinSKUsConcat(lines); got != want {
		t.Errorf("JoinSKUsConcat = %q, want %q", got, want)
	}
	if got := JoinSKUsBuilder(lines); got != want {
		t.Errorf("JoinSKUsBuilder = %q, want %q", got, want)
	}
}

func TestJoinSKUs_EmptyAndSingle(t *testing.T) {
	for _, fn := range []struct {
		name string
		join func([]domain.OrderLine) string
	}{
		{"concat", JoinSKUsConcat},
		{"builder", JoinSKUsBuilder},
	} {
		t.Run(fn.name, func(t *testing.T) {
			if got := fn.join(nil); got != "" {
				t.Errorf("no lines = %q, want %q", got, "")
			}
			one := []domain.OrderLine{{ProductID: "solo", Price: domain.Zero(domain.AUD)}}
			if got := fn.join(one); got != "solo" {
				t.Errorf("one line = %q, want %q (no trailing comma)", got, "solo")
			}
		})
	}
}

func TestBuildLines_LengthAndContents(t *testing.T) {
	for _, fn := range []struct {
		name  string
		build func(int) []domain.OrderLine
	}{
		{"append", BuildLinesAppend},
		{"prealloc", BuildLinesPrealloc},
	} {
		t.Run(fn.name, func(t *testing.T) {
			got := fn.build(3)
			if len(got) != 3 {
				t.Fatalf("len = %d, want 3", len(got))
			}
			// Line i must carry SKU "sku-<i>", so the two builders agree.
			if got[0].ProductID != "sku-0" || got[2].ProductID != "sku-2" {
				t.Errorf("SKUs = %q..%q, want sku-0..sku-2",
					got[0].ProductID, got[2].ProductID)
			}
			if empty := fn.build(0); len(empty) != 0 {
				t.Errorf("build(0) returned %d lines, want 0", len(empty))
			}
		})
	}
}

func TestSumCents_AddsEveryLine(t *testing.T) {
	lines := testLines(10)
	if got := SumCentsValue(lines); got != 0 {
		t.Errorf("SumCentsValue = %d, want 0 (every fixture price is domain.Zero)", got)
	}
	if got := SumCentsPointer(testLinePtrs(10)); got != 0 {
		t.Errorf("SumCentsPointer = %d, want 0", got)
	}
	if got := SumCentsValue(nil); got != 0 {
		t.Errorf("SumCentsValue(nil) = %d, want 0", got)
	}
}

func TestSKUFromBytes_Value(t *testing.T) {
	if got := SKUFromBytes([]byte("sku-12345")); got != "sku-12345" {
		t.Errorf("SKUFromBytes = %q, want %q", got, "sku-12345")
	}
}

// ---- Now the allocations ----

func TestNewLineValue_DoesNotAllocate(t *testing.T) {
	allocs := testing.AllocsPerRun(200, func() {
		sinkLine = NewLineValue("sku-1", "Widget")
	})
	if allocs != 0 {
		t.Errorf("NewLineValue allocated %.0f times, want 0: "+
			"nothing outside can reference the line, so it belongs on the stack", allocs)
	}
}

func TestNewLinePointer_Allocates(t *testing.T) {
	allocs := testing.AllocsPerRun(200, func() {
		sinkLinePtr = NewLinePointer("sku-1", "Widget")
	})
	if allocs != 1 {
		t.Errorf("NewLinePointer allocated %.0f times, want exactly 1: "+
			"the pointer outlives the call, so the line must be on the heap", allocs)
	}
}

// No pointer is written anywhere, and it still allocates. This is the one to
// sit with.
func TestBoxLine_AllocatesEvenWithNoPointer(t *testing.T) {
	line := NewLineValue("sku-1", "Widget")

	allocs := testing.AllocsPerRun(200, func() {
		sinkAny = BoxLine(line)
	})
	if allocs == 0 {
		t.Error("BoxLine allocated 0 times. An interface value is a (type, pointer) " +
			"pair, so a value put into one has to live somewhere the pointer can " +
			"point at. If this really is 0, work out what the compiler proved.")
	}
	t.Logf("boxing one OrderLine: %.0f allocs", allocs)
}

func TestFormatBuilder_AllocatesLessThanSprintf(t *testing.T) {
	line := NewLineValue("sku-1", "Widget")

	sprintfAllocs := testing.AllocsPerRun(200, func() {
		sinkString = FormatLineSprintf(line)
	})
	builderAllocs := testing.AllocsPerRun(200, func() {
		sinkString = FormatLineBuilder(line)
	})

	if builderAllocs >= sprintfAllocs {
		t.Errorf("builder allocated %.0f, Sprintf allocated %.0f: "+
			"Sprintf boxes every argument into ...any before formatting anything",
			builderAllocs, sprintfAllocs)
	}
	t.Logf("Sprintf %.0f allocs, Builder %.0f allocs", sprintfAllocs, builderAllocs)
}

func TestJoinBuilder_ScalesBetterThanConcat(t *testing.T) {
	lines := testLines(200)

	concatAllocs := testing.AllocsPerRun(20, func() {
		sinkString = JoinSKUsConcat(lines)
	})
	builderAllocs := testing.AllocsPerRun(20, func() {
		sinkString = JoinSKUsBuilder(lines)
	})

	if builderAllocs >= concatAllocs {
		t.Errorf("builder %.0f allocs, concat %.0f: strings are immutable, so += "+
			"rebuilds and copies everything so far on every iteration",
			builderAllocs, concatAllocs)
	}
	t.Logf("over %d lines: concat %.0f allocs, builder %.0f allocs",
		len(lines), concatAllocs, builderAllocs)
}

func TestSKUFromBytes_CopiesRatherThanAliases(t *testing.T) {
	b := []byte("sku-12345")

	allocs := testing.AllocsPerRun(200, func() {
		sinkSKU = SKUFromBytes(b)
	})
	if allocs != 1 {
		t.Errorf("SKUFromBytes allocated %.0f times, want 1: converting []byte to "+
			"string copies, because the caller still owns the bytes", allocs)
	}

	sku := SKUFromBytes(b)
	b[0] = 'X'
	if strings.HasPrefix(string(sku), "X") {
		t.Error("mutating the source bytes changed the ProductID: the conversion aliased")
	}
}

func TestBuildLinesPrealloc_AllocatesOnce(t *testing.T) {
	allocs := testing.AllocsPerRun(200, func() {
		sinkSlice = BuildLinesPrealloc(64)
	})
	if allocs != 1 {
		t.Errorf("BuildLinesPrealloc allocated %.0f times, want exactly 1: "+
			"the final length is known, so one backing array is enough", allocs)
	}
}

func TestBuildLinesAppend_AllocatesRepeatedly(t *testing.T) {
	appendAllocs := testing.AllocsPerRun(200, func() {
		sinkSlice = BuildLinesAppend(64)
	})
	preallocAllocs := testing.AllocsPerRun(200, func() {
		sinkSlice = BuildLinesPrealloc(64)
	})

	if appendAllocs <= preallocAllocs {
		t.Errorf("append %.0f allocs, prealloc %.0f: growing a slice reallocates "+
			"and copies each time the backing array fills", appendAllocs, preallocAllocs)
	}
	t.Logf("building 64 lines: append %.0f allocs, prealloc %.0f allocs",
		appendAllocs, preallocAllocs)
}

func TestSumCentsValue_DoesNotAllocate(t *testing.T) {
	lines := testLines(100)

	allocs := testing.AllocsPerRun(200, func() {
		sinkInt64 = SumCentsValue(lines)
	})
	if allocs != 0 {
		t.Errorf("SumCentsValue allocated %.0f times, want 0: the slice already "+
			"exists and each copied line dies at the end of its iteration", allocs)
	}
}

func TestSumCentsPointer_DoesNotAllocateEither(t *testing.T) {
	ptrs := testLinePtrs(100)

	allocs := testing.AllocsPerRun(200, func() {
		sinkInt64 = SumCentsPointer(ptrs)
	})
	if allocs != 0 {
		t.Errorf("SumCentsPointer allocated %.0f times, want 0. "+
			"The allocations were paid when the []*OrderLine was built — "+
			"which is exactly the point of this pair", allocs)
	}
}

func TestMakeLineCounter_KeepsStateAndEscapes(t *testing.T) {
	next := MakeLineCounter(10)
	if got := next(); got != 11 {
		t.Fatalf("first call = %d, want 11", got)
	}
	if got := next(); got != 12 {
		t.Fatalf("second call = %d, want 12: the closure must keep its state", got)
	}

	allocs := testing.AllocsPerRun(200, func() {
		sinkInt64 = int64(MakeLineCounter(0)())
	})
	if allocs == 0 {
		t.Error("MakeLineCounter allocated 0 times. The captured counter outlives " +
			"the call, so it cannot stay on the stack — check what you captured.")
	}
}

func TestSendLine_DeliversThePointer(t *testing.T) {
	ch := make(chan *domain.OrderLine, 1)
	SendLine(ch, "sku-9", "Gadget")

	got := <-ch
	if got == nil {
		t.Fatal("SendLine sent nil")
	}
	if got.ProductID != "sku-9" || got.Name != "Gadget" {
		t.Errorf("received %+v, want ProductID=sku-9 Name=Gadget", *got)
	}
}

func TestPool_ReusesAndResets(t *testing.T) {
	first := GetLineFromPool("sku-1", "Widget")
	if first == nil {
		t.Fatal("GetLineFromPool returned nil")
	}
	if first.ProductID != "sku-1" || first.Name != "Widget" {
		t.Errorf("got %+v, want ProductID=sku-1 Name=Widget", *first)
	}
	PutLineToPool(first)

	// A pool is a cache, not a guarantee — the GC may have emptied it. So this
	// asserts what must hold either way: whatever comes back is clean.
	second := GetLineFromPool("sku-2", "Gadget")
	if second.ProductID != "sku-2" || second.Name != "Gadget" {
		t.Errorf("pooled line = %+v, want the new values. A pooled object still "+
			"carrying the previous caller's data is a data leak between requests",
			*second)
	}
}

func TestPutLineToPool_HandlesNil(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PutLineToPool(nil) panicked: %v. Callers defer this.", r)
		}
	}()
	PutLineToPool(nil)
}

// ---- Benchmarks ----
//
//	go test ./internal/foundation/performance -bench=. -benchmem
//
// Read the B/op and allocs/op columns, not just ns/op. Two of these pairs are
// closer in time than you would expect and much further apart in allocations —
// exactly the kind of thing you cannot guess and have to measure.

func BenchmarkFormatSprintf(b *testing.B) {
	line := NewLineValue("sku-1", "Widget")
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		sinkString = FormatLineSprintf(line)
	}
}

func BenchmarkFormatBuilder(b *testing.B) {
	line := NewLineValue("sku-1", "Widget")
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		sinkString = FormatLineBuilder(line)
	}
}

func BenchmarkJoinConcat(b *testing.B) {
	lines := testLines(200)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		sinkString = JoinSKUsConcat(lines)
	}
}

func BenchmarkJoinBuilder(b *testing.B) {
	lines := testLines(200)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		sinkString = JoinSKUsBuilder(lines)
	}
}

func BenchmarkBuildAppend(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		sinkSlice = BuildLinesAppend(256)
	}
}

func BenchmarkBuildPrealloc(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		sinkSlice = BuildLinesPrealloc(256)
	}
}

// The interesting pair. Same total work, both allocate nothing at call time —
// but the pointer version chases scattered heap addresses while the value
// version walks contiguous memory the CPU can prefetch.
func BenchmarkSumValue(b *testing.B) {
	lines := testLines(4096)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		sinkInt64 = SumCentsValue(lines)
	}
}

func BenchmarkSumPointer(b *testing.B) {
	ptrs := testLinePtrs(4096)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		sinkInt64 = SumCentsPointer(ptrs)
	}
}
