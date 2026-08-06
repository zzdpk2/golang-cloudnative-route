package performance

import (
	"errors"
	"testing"
	"unsafe"
)

func TestFieldAndElementPointers(t *testing.T) {
	pair := Pair{Flag: true, Size: 10}
	pointer := SizePointer(&pair)
	*pointer = 20
	if pair.Size != 20 {
		t.Errorf("pair.Size = %d", pair.Size)
	}

	values := []int{10, 20, 30}
	element, err := ElementPointer(values, 1)
	if err != nil {
		t.Fatal(err)
	}
	*element = 99
	if values[1] != 99 {
		t.Errorf("values = %v", values)
	}
	if _, err := ElementPointer(values, 3); !errors.Is(err, ErrIndexOutOfRange) {
		t.Errorf("out-of-range error = %v", err)
	}
}

func TestStringViewAndSafeCopy(t *testing.T) {
	source := []byte("order")
	view := StringView(source)
	if view != "order" {
		t.Errorf("StringView() = %q", view)
	}
	if unsafe.StringData(view) != unsafe.SliceData(source) {
		t.Error("StringView copied its input")
	}

	copySource := []byte("order")
	safe := SafeString(copySource)
	copySource[0] = 'X'
	if safe != "order" {
		t.Errorf("SafeString aliases input: %q", safe)
	}
}
