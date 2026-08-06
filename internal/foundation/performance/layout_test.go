package performance

import (
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"sync"
	"testing"
)

// ---- The basics ----

func TestSizeOf_Primitives(t *testing.T) {
	tests := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"bool", SizeOf[bool](), 1},
		{"int8", SizeOf[int8](), 1},
		{"int32", SizeOf[int32](), 4},
		{"int64", SizeOf[int64](), 8},
		{"string header", SizeOf[string](), 16},
		{"slice header", SizeOf[[]byte](), 24},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("SizeOf = %d, want %d", tt.got, tt.want)
			}
		})
	}
}

// A string header is 16 bytes whatever the text. The text lives elsewhere.
func TestSizeOf_StringHeaderIsIndependentOfContent(t *testing.T) {
	type short struct{ S string }
	type long struct{ S string }

	if SizeOf[short]() != SizeOf[long]() {
		t.Fatal("two structs each holding one string must be the same size")
	}
	if got := SizeOf[short](); got != 16 {
		t.Errorf("a struct holding one string = %d bytes, want 16: "+
			"the header is a pointer plus a length; the bytes are not inside", got)
	}
}

func TestSizeOf_FieldOrderChangesSize(t *testing.T) {
	bad, good := SizeOf[BadLayout](), SizeOf[GoodLayout]()

	if bad != 32 {
		t.Errorf("SizeOf[BadLayout] = %d, want 32", bad)
	}
	if good != 24 {
		t.Errorf("SizeOf[GoodLayout] = %d, want 24", good)
	}
	if good >= bad {
		t.Errorf("reordering saved nothing: bad=%d good=%d", bad, good)
	}
}

func TestAlignOf(t *testing.T) {
	if got := AlignOf[bool](); got != 1 {
		t.Errorf("AlignOf[bool] = %d, want 1", got)
	}
	if got := AlignOf[int64](); got != 8 {
		t.Errorf("AlignOf[int64] = %d, want 8", got)
	}
	if got := AlignOf[BadLayout](); got != 8 {
		t.Errorf("AlignOf[BadLayout] = %d, want 8: a struct takes the widest "+
			"alignment among its fields", got)
	}
}

func TestFieldOffsets(t *testing.T) {
	bad := FieldOffsets[BadLayout]()
	if want := []uintptr{0, 8, 16, 24}; !equalOffsets(bad, want) {
		t.Errorf("FieldOffsets[BadLayout] = %v, want %v", bad, want)
	}

	good := FieldOffsets[GoodLayout]()
	if want := []uintptr{0, 8, 16, 17}; !equalOffsets(good, want) {
		t.Errorf("FieldOffsets[GoodLayout] = %v, want %v: the two bools should "+
			"sit next to each other", good, want)
	}

	if got := FieldOffsets[int64](); got != nil {
		t.Errorf("FieldOffsets on a non-struct = %v, want nil", got)
	}
}

func TestPaddingWaste(t *testing.T) {
	if got := PaddingWaste[BadLayout](); got != 14 {
		t.Errorf("PaddingWaste[BadLayout] = %d, want 14 (32 bytes for 18 of fields)", got)
	}
	if got := PaddingWaste[GoodLayout](); got != 6 {
		t.Errorf("PaddingWaste[GoodLayout] = %d, want 6", got)
	}
	if got := PaddingWaste[int64](); got != 0 {
		t.Errorf("PaddingWaste on a non-struct = %d, want 0", got)
	}
}

// ---- The real order types ----

// No exact number is asserted here: it depends on the architecture, and pinning
// it would make the test brittle for no teaching value. What is asserted is the
// relationship, and the numbers are logged so you can look at them.
func TestOrderLineSize_MatchesUnsafe(t *testing.T) {
	got := OrderLineSize()
	want := SizeOf[domain.OrderLine]()

	if got != want {
		t.Errorf("OrderLineSize() = %d, want %d", got, want)
	}
	t.Logf("OrderLine = %d bytes, of which %d are padding",
		got, PaddingWaste[domain.OrderLine]())
	t.Logf("Money     = %d bytes, of which %d are padding",
		SizeOf[domain.Money](), PaddingWaste[domain.Money]())
	t.Logf("Order     = %d bytes", SizeOf[domain.Order]())
}

func TestMoneySize_MatchesUnsafe(t *testing.T) {
	if got, want := MoneySize(), SizeOf[domain.Money](); got != want {
		t.Errorf("MoneySize() = %d, want %d", got, want)
	}
}

// An OrderLine is large enough that copying one is not free — which is the
// number to weigh against the heap allocation a pointer would cost. See
// lab/escape.
func TestOrderLineIsWorthCopying(t *testing.T) {
	size := OrderLineSize()
	if size == 0 {
		t.Fatal("OrderLineSize() = 0")
	}
	if size > 128 {
		t.Errorf("OrderLine = %d bytes. Above roughly a couple of cache lines, "+
			"passing by value starts to cost more than the allocation you avoided. "+
			"If this fires, the type has grown and the trade needs revisiting.", size)
	}
}

func TestOrderLineWastePerOrder_Scales(t *testing.T) {
	one := OrderLineWastePerOrder(1)
	forty := OrderLineWastePerOrder(40)

	if forty != one*40 {
		t.Errorf("waste for 40 lines = %d, want 40 x %d", forty, one)
	}
	t.Logf("an order of 40 lines carries %d bytes of padding", forty)
}

// ---- Zero-size types ----

func TestZeroSizeStruct(t *testing.T) {
	if got := SizeOf[struct{}](); got != 0 {
		t.Errorf("SizeOf[struct{}] = %d, want 0", got)
	}
	// A zero-size field in the middle costs nothing.
	if got := SizeOf[ZeroSizeDemo](); got != 2 {
		t.Errorf("SizeOf[ZeroSizeDemo] = %d, want 2: an empty struct between two "+
			"bytes adds nothing", got)
	}
}

func TestSliceMemory(t *testing.T) {
	if got := SliceMemory[int64](100); got != 800 {
		t.Errorf("SliceMemory[int64](100) = %d, want 800", got)
	}
	if got := SliceMemory[struct{}](1_000_000); got != 0 {
		t.Errorf("SliceMemory[struct{}](1e6) = %d, want 0: this is why a set is "+
			"map[K]struct{} and not map[K]bool", got)
	}
	if got := SliceMemory[byte](0); got != 0 {
		t.Errorf("SliceMemory[byte](0) = %d, want 0", got)
	}
	if got := SliceMemory[byte](-5); got != 0 {
		t.Errorf("SliceMemory[byte](-5) = %d, want 0", got)
	}

	want := uintptr(40) * SizeOf[domain.OrderLine]()
	if got := SliceMemory[domain.OrderLine](40); got != want {
		t.Errorf("SliceMemory[OrderLine](40) = %d, want %d", got, want)
	}
	t.Logf("a 40-line order occupies %d bytes of line data", want)
}

// ---- False sharing ----

func TestPaddedShard_OccupiesAWholeCacheLine(t *testing.T) {
	if got := SizeOf[PaddedShard](); got != CacheLineSize {
		t.Errorf("SizeOf[PaddedShard] = %d, want exactly %d: the padding field "+
			"has to fill the line, no more and no less", got, CacheLineSize)
	}
}

func TestShardsPerCacheLine(t *testing.T) {
	naive := ShardsPerCacheLine[NaiveShard]()
	if naive != 8 {
		t.Errorf("ShardsPerCacheLine[NaiveShard] = %d, want 8: eight int64 "+
			"counters share one 64-byte line, so eight goroutines contend in "+
			"hardware even with no lock between them", naive)
	}

	padded := ShardsPerCacheLine[PaddedShard]()
	if padded != 1 {
		t.Errorf("ShardsPerCacheLine[PaddedShard] = %d, want 1", padded)
	}
}

func TestSumShards(t *testing.T) {
	naive := []NaiveShard{{Count: 1}, {Count: 2}, {Count: 3}}
	if got := SumNaive(naive); got != 6 {
		t.Errorf("SumNaive = %d, want 6", got)
	}
	if got := SumNaive(nil); got != 0 {
		t.Errorf("SumNaive(nil) = %d, want 0", got)
	}

	padded := make([]PaddedShard, 3)
	padded[0].Count, padded[1].Count, padded[2].Count = 1, 2, 3
	if got := SumPadded(padded); got != 6 {
		t.Errorf("SumPadded = %d, want 6", got)
	}
}

// ---- Benchmarks: the whole point of the padding ----
//
//	go test ./internal/foundation/performance -bench=Shard -benchtime=2s
//
// Both benchmarks do identical work with identical correctness. The only
// difference is whether adjacent counters share a cache line.
//
// Predict the ratio before running. Then run it, and if the gap is smaller than
// you expected, try raising the goroutine count — false sharing needs real
// parallelism to show, so it barely appears on one core and hurts badly on
// sixteen.

const shardCount = 8
const incrementsPerShard = 200_000

func BenchmarkNaiveShardsContend(b *testing.B) {
	for range b.N {
		shards := make([]NaiveShard, shardCount)
		var wg sync.WaitGroup
		for i := range shardCount {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				for range incrementsPerShard {
					shards[i].Count++
				}
			}(i)
		}
		wg.Wait()
		if SumNaive(shards) != shardCount*incrementsPerShard {
			b.Fatal("lost increments")
		}
	}
}

func BenchmarkPaddedShardsDoNot(b *testing.B) {
	for range b.N {
		shards := make([]PaddedShard, shardCount)
		var wg sync.WaitGroup
		for i := range shardCount {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				for range incrementsPerShard {
					shards[i].Count++
				}
			}(i)
		}
		wg.Wait()
		if SumPadded(shards) != shardCount*incrementsPerShard {
			b.Fatal("lost increments")
		}
	}
}

func equalOffsets(a, b []uintptr) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
