package operations

import (
	"errors"
	"testing"
	"time"
)

var base = time.Date(2026, time.March, 15, 12, 0, 0, 0, time.UTC)

func at(minutes int) time.Time { return base.Add(time.Duration(minutes) * time.Minute) }

// A stream: placed, two lines added, confirmed.
func testEvents() []Event {
	return []Event{
		OrderPlaced{OrderID: "ord-1", CustomerID: "cust-1", At: at(0)},
		LineAdded{SKU: "sku-1", Quantity: 2, AmountCent: 1000, At: at(1)},
		LineAdded{SKU: "sku-2", Quantity: 1, AmountCent: 500, At: at(2)},
		OrderConfirmed{At: at(3)},
	}
}

// ---- Apply ----

func TestApply_BuildsStateStepByStep(t *testing.T) {
	var state OrderState
	var err error

	state, err = Apply(state, OrderPlaced{OrderID: "ord-1", CustomerID: "cust-1", At: at(0)})
	if err != nil {
		t.Fatalf("OrderPlaced: %v", err)
	}
	if state.OrderID != "ord-1" || state.CustomerID != "cust-1" {
		t.Errorf("state = %+v, want the order and customer ids set", state)
	}
	if state.Status != "placed" {
		t.Errorf("Status = %q, want %q", state.Status, "placed")
	}
	if state.Version != 1 {
		t.Errorf("Version = %d, want 1", state.Version)
	}

	state, err = Apply(state, LineAdded{SKU: "sku-1", Quantity: 2, AmountCent: 1000, At: at(1)})
	if err != nil {
		t.Fatalf("LineAdded: %v", err)
	}
	if state.Lines["sku-1"] != 2 {
		t.Errorf("Lines[sku-1] = %d, want 2", state.Lines["sku-1"])
	}
	if state.TotalCent != 2000 {
		t.Errorf("TotalCent = %d, want 2000 (2 x 1000)", state.TotalCent)
	}

	// Adding the same SKU again accumulates.
	state, err = Apply(state, LineAdded{SKU: "sku-1", Quantity: 1, AmountCent: 1000, At: at(2)})
	if err != nil {
		t.Fatalf("second LineAdded: %v", err)
	}
	if state.Lines["sku-1"] != 3 {
		t.Errorf("Lines[sku-1] = %d, want 3", state.Lines["sku-1"])
	}
	if state.TotalCent != 3000 {
		t.Errorf("TotalCent = %d, want 3000", state.TotalCent)
	}

	state, err = Apply(state, LineRemoved{SKU: "sku-1", At: at(3)})
	if err != nil {
		t.Fatalf("LineRemoved: %v", err)
	}
	if _, present := state.Lines["sku-1"]; present {
		t.Errorf("Lines still holds sku-1: %v", state.Lines)
	}
	if state.TotalCent != 0 {
		t.Errorf("TotalCent = %d, want 0", state.TotalCent)
	}
	if state.Version != 4 {
		t.Errorf("Version = %d, want 4", state.Version)
	}
}

// Apply must not reach back into the state it was handed.
func TestApply_DoesNotMutateTheInput(t *testing.T) {
	placed, err := Apply(OrderState{}, OrderPlaced{OrderID: "ord-1", CustomerID: "c", At: at(0)})
	if err != nil {
		t.Fatal(err)
	}
	withLine, err := Apply(placed, LineAdded{SKU: "sku-1", Quantity: 2, AmountCent: 1000, At: at(1)})
	if err != nil {
		t.Fatal(err)
	}

	if len(placed.Lines) != 0 {
		t.Errorf("the earlier state gained lines: %v. Apply shared the map, so "+
			"replaying to an earlier version now returns a later one.", placed.Lines)
	}
	if placed.Version != 1 || withLine.Version != 2 {
		t.Errorf("versions = %d and %d, want 1 and 2", placed.Version, withLine.Version)
	}
}

// Same input, same output, forever — replay depends on it.
func TestApply_IsDeterministic(t *testing.T) {
	state, _ := Apply(OrderState{}, OrderPlaced{OrderID: "ord-1", CustomerID: "c", At: at(0)})
	event := LineAdded{SKU: "sku-1", Quantity: 2, AmountCent: 1000, At: at(1)}

	first, err := Apply(state, event)
	if err != nil {
		t.Fatal(err)
	}
	for i := range 30 {
		again, err := Apply(state, event)
		if err != nil {
			t.Fatal(err)
		}
		if again.TotalCent != first.TotalCent || again.Version != first.Version {
			t.Fatalf("run %d differed: %+v vs %+v", i, again, first)
		}
	}
}

func TestApply_RejectsAnImpossibleTransition(t *testing.T) {
	state, err := Replay(testEvents()) // ends confirmed
	if err != nil {
		t.Fatal(err)
	}

	_, err = Apply(state, LineAdded{SKU: "sku-9", Quantity: 1, AmountCent: 100, At: at(9)})
	if !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("adding a line to a confirmed order = %v, want ErrInvalidTransition", err)
	}
}

// AddressCorrected is the question-1 event: it changes the projection without
// erasing what came before.
func TestApply_AddressCorrection(t *testing.T) {
	state, err := Replay([]Event{
		OrderPlaced{OrderID: "ord-1", CustomerID: "cust-1", At: at(0)},
		AddressCorrected{Address: "1 George St", Reason: "typo at entry", At: at(1)},
	})
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}
	if state.Address != "1 George St" {
		t.Errorf("Address = %q, want the corrected value", state.Address)
	}
	if state.Version != 2 {
		t.Errorf("Version = %d, want 2: the correction is an event, not an edit",
			state.Version)
	}
}

// ---- Replay ----

func TestReplay_FoldsTheWholeStream(t *testing.T) {
	state, err := Replay(testEvents())
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}
	if state.Status != "confirmed" {
		t.Errorf("Status = %q, want %q", state.Status, "confirmed")
	}
	if state.TotalCent != 2500 {
		t.Errorf("TotalCent = %d, want 2500 (2x1000 + 1x500)", state.TotalCent)
	}
	if state.Version != 4 {
		t.Errorf("Version = %d, want 4", state.Version)
	}
}

func TestReplay_EmptyStream(t *testing.T) {
	if _, err := Replay(nil); !errors.Is(err, ErrEmptyStream) {
		t.Errorf("Replay(nil) = %v, want ErrEmptyStream", err)
	}
}

func TestReplay_MustStartWithOrderPlaced(t *testing.T) {
	_, err := Replay([]Event{LineAdded{SKU: "sku-1", Quantity: 1, AmountCent: 100, At: at(0)}})
	if !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("a stream starting with LineAdded = %v, want ErrInvalidTransition", err)
	}
}

// The time machine.
func TestReplayTo_RewindsToAPointInHistory(t *testing.T) {
	events := testEvents()

	state, err := ReplayTo(events, 2)
	if err != nil {
		t.Fatalf("ReplayTo(2): %v", err)
	}
	if state.Version != 2 {
		t.Errorf("Version = %d, want 2", state.Version)
	}
	if state.TotalCent != 2000 {
		t.Errorf("TotalCent = %d, want 2000: sku-2 had not been added yet", state.TotalCent)
	}
	if state.Status != "placed" {
		t.Errorf("Status = %q, want %q: confirmation came later", state.Status, "placed")
	}

	// The present is unchanged by having looked at the past.
	now, err := Replay(events)
	if err != nil {
		t.Fatal(err)
	}
	if now.TotalCent != 2500 {
		t.Errorf("after rewinding, the full replay gives %d, want 2500", now.TotalCent)
	}
}

func TestReplayTo_Bounds(t *testing.T) {
	events := testEvents()

	if _, err := ReplayTo(events, 0); !errors.Is(err, ErrEmptyStream) {
		t.Errorf("ReplayTo(0) = %v, want ErrEmptyStream", err)
	}

	beyond, err := ReplayTo(events, 99)
	if err != nil {
		t.Fatalf("ReplayTo past the end = %v, want the full state", err)
	}
	full, _ := Replay(events)
	if beyond.Version != full.Version {
		t.Errorf("ReplayTo(99) version %d, full replay %d", beyond.Version, full.Version)
	}
}

// ---- Snapshots ----

func TestReplayFrom_ContinuesFromASnapshot(t *testing.T) {
	events := testEvents()

	snapState, err := ReplayTo(events, 2)
	if err != nil {
		t.Fatal(err)
	}
	snap := Snapshot{State: snapState, Version: 2}

	got, err := ReplayFrom(snap, events)
	if err != nil {
		t.Fatalf("ReplayFrom: %v", err)
	}

	full, _ := Replay(events)
	if got.TotalCent != full.TotalCent || got.Version != full.Version || got.Status != full.Status {
		t.Errorf("from snapshot = %+v, full replay = %+v; the two must agree",
			got, full)
	}
}

func TestReplayFrom_RejectsASnapshotAheadOfTheStream(t *testing.T) {
	events := testEvents()
	snap := Snapshot{State: OrderState{Version: 99}, Version: 99}

	if _, err := ReplayFrom(snap, events); !errors.Is(err, ErrVersionConflict) {
		t.Errorf("a snapshot newer than the stream = %v, want ErrVersionConflict", err)
	}
}

// ---- Stream and optimistic concurrency ----

func TestStream_AppendAndReplay(t *testing.T) {
	s := NewStream()
	if s == nil {
		t.Fatal("NewStream returned nil")
	}
	if s.Version() != 0 {
		t.Fatalf("a new stream is at version %d, want 0", s.Version())
	}

	if err := s.Append(0, OrderPlaced{OrderID: "ord-1", CustomerID: "cust-1", At: at(0)}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if s.Version() != 1 {
		t.Errorf("Version = %d, want 1", s.Version())
	}

	if err := s.Append(1, LineAdded{SKU: "sku-1", Quantity: 2, AmountCent: 1000, At: at(1)}); err != nil {
		t.Fatalf("second Append: %v", err)
	}

	state, err := s.State()
	if err != nil {
		t.Fatalf("State: %v", err)
	}
	if state.TotalCent != 2000 {
		t.Errorf("TotalCent = %d, want 2000", state.TotalCent)
	}
}

// The only concurrency control an event store has.
func TestStream_AppendRejectsAStaleVersion(t *testing.T) {
	s := NewStream()
	_ = s.Append(0, OrderPlaced{OrderID: "ord-1", CustomerID: "cust-1", At: at(0)})
	_ = s.Append(1, LineAdded{SKU: "sku-1", Quantity: 1, AmountCent: 100, At: at(1)})

	// A second writer read version 1 before the line was added.
	err := s.Append(1, LineAdded{SKU: "sku-2", Quantity: 1, AmountCent: 200, At: at(2)})
	if !errors.Is(err, ErrVersionConflict) {
		t.Fatalf("stale Append = %v, want ErrVersionConflict", err)
	}
	if s.Version() != 2 {
		t.Errorf("Version = %d, want 2: a rejected append must change nothing",
			s.Version())
	}
}

func TestStream_AppendIsAllOrNothing(t *testing.T) {
	s := NewStream()
	_ = s.Append(0, OrderPlaced{OrderID: "ord-1", CustomerID: "cust-1", At: at(0)})

	before := s.Version()
	_ = s.Append(99,
		LineAdded{SKU: "a", Quantity: 1, AmountCent: 100, At: at(1)},
		LineAdded{SKU: "b", Quantity: 1, AmountCent: 100, At: at(2)},
	)
	if s.Version() != before {
		t.Errorf("Version = %d after a rejected multi-append, want %d: "+
			"half-appending is worse than refusing", s.Version(), before)
	}
}

// The log is append-only, and a caller must not be able to rewrite it.
func TestStream_EventsAreNotAliased(t *testing.T) {
	s := NewStream()
	_ = s.Append(0, OrderPlaced{OrderID: "ord-1", CustomerID: "cust-1", At: at(0)})
	_ = s.Append(1, LineAdded{SKU: "sku-1", Quantity: 2, AmountCent: 1000, At: at(1)})

	events := s.Events()
	if len(events) != 2 {
		t.Fatalf("Events() returned %d, want 2", len(events))
	}
	events[0] = OrderConfirmed{At: at(9)}

	again := s.Events()
	if again[0].EventType() != "order.placed" {
		t.Errorf("the log was rewritten from outside: first event is now %q. "+
			"An append-only log a caller can edit is not a log.",
			again[0].EventType())
	}
}
