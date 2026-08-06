package concurrency

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestDetachWindowDoesNotRetainWritableBacking(t *testing.T) {
	source := make([]byte, 1<<20)
	copy(source[100:104], []byte("data"))

	window, err := DetachWindow(source, 100, 104)
	if err != nil {
		t.Fatal(err)
	}
	source[100] = 'X'
	if string(window) != "data" {
		t.Errorf("window aliases source: %q", window)
	}
	if cap(window) != len(window) {
		t.Errorf("detached capacity = %d, want %d", cap(window), len(window))
	}

	if _, err := DetachWindow(source, -1, 4); !errors.Is(err, ErrInvalidWindow) {
		t.Errorf("invalid window error = %v", err)
	}
}

func TestDeleteAtClearsTail(t *testing.T) {
	a, b, c := 1, 2, 3
	values := make([]*int, 3, 3)
	values[0], values[1], values[2] = &a, &b, &c

	got, err := DeleteAt(values, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []*int{&a, &c}) {
		t.Errorf("DeleteAt() = %v", got)
	}
	if backing := got[:cap(got)]; backing[2] != nil {
		t.Error("tail still retains a pointer")
	}

	if _, err := DeleteAt(got, 9); !errors.Is(err, ErrInvalidWindow) {
		t.Errorf("invalid index error = %v", err)
	}
}

func TestWorkerStopsOnCancellationEvenWhenInputStaysOpen(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	input := make(chan int)
	done := Worker(ctx, input, func(int) {})

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker leaked after cancellation")
	}
}

func TestWorkerProcessesUntilInputCloses(t *testing.T) {
	ctx := context.Background()
	input := make(chan int, 3)
	input <- 1
	input <- 2
	input <- 3
	close(input)

	var total int
	done := Worker(ctx, input, func(value int) {
		total += value
	})
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after input closed")
	}
	if total != 6 {
		t.Errorf("total = %d", total)
	}
}
