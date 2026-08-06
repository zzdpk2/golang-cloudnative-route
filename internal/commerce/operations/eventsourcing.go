// Package eventsourcing is L15.5: an order whose state is derived, not stored.
//
// A ● exercise, and the one that most changes how you think about the rest of
// the codebase.
//
// # The inversion
//
// Everywhere else, aggregate.Order *is* the state and events are a side effect
// it emits. Here that is turned around: **the events are the truth**, and the
// state is a fold over them.
//
//	state = fold(replay, events)
//
// Nothing is ever updated and nothing is ever deleted. The order was placed,
// then a line was added, then it was confirmed — and all three of those remain
// true forever. What you call "the current state" is just the last frame of a
// film you can rewind.
//
// # What that buys
//
//   - A complete audit trail, for free and by construction. Not a log somebody
//     remembered to write — the log *is* the data.
//   - Answering "what did this look like last Tuesday?" is replaying to a
//     point, not restoring a backup.
//   - Bugs become fixable retroactively: correct the fold, replay, and history
//     is right again.
//
// # What it costs
//
//   - Every query is a replay, unless you keep snapshots.
//   - Events are immutable, so a schema change cannot rewrite the old ones —
//     you have to keep reading v1 forever.
//   - "Just fix the row" is not available. Every correction is a new event, and
//     that is a discipline, not a convenience.
//   - Deleting a customer's data conflicts directly with an append-only log,
//     and the GDPR does not accept "but it is immutable" as an answer.
//
// # Answer these first
//
//  1. The customer's address was entered wrong. As an event, is that
//     AddressCorrected or AddressChanged? They store identically and mean
//     opposite things to an auditor.
//  2. A snapshot is a cache of a fold. What has to be true for it to be safe to
//     use, and what invalidates one?
//  3. Version 2 of OrderPlaced adds a field. The v1 events in the log cannot
//     have it. Where does the default come from, and who owns that decision?
//  4. If replaying gives a different answer than it did last month — because
//     you fixed the fold — was the old state wrong, or is the new one?
//  5. Which is your aggregate's real invariant boundary: the version number, or
//     the events themselves? What makes two concurrent appends conflict?
package operations

import (
	"errors"
	"time"
)

var (
	// ErrVersionConflict means the expected version did not match the stream.
	ErrVersionConflict = errors.New("version conflict")
	// ErrUnknownEvent means the fold met an event type it cannot apply.
	ErrUnknownEvent = errors.New("unknown event type")
	// ErrEmptyStream means there are no events to fold.
	ErrEmptyStream = errors.New("empty stream")
	// ErrInvalidTransition means an event cannot be applied from this state.
	ErrInvalidTransition = errors.New("invalid transition")
)

// Event is one thing that happened. Past tense, always: OrderPlaced, not
// PlaceOrder. A command may be refused; an event already happened.
type Event interface {
	EventType() string
	OccurredAt() time.Time
}

// OrderPlaced starts a stream.
type OrderPlaced struct {
	OrderID    string
	CustomerID string
	At         time.Time
}

func (e OrderPlaced) EventType() string     { panic("TODO") }
func (e OrderPlaced) OccurredAt() time.Time { panic("TODO") }

// LineAdded adds an item.
type LineAdded struct {
	SKU        string
	Quantity   int
	AmountCent int64
	At         time.Time
}

func (e LineAdded) EventType() string     { panic("TODO") }
func (e LineAdded) OccurredAt() time.Time { panic("TODO") }

// LineRemoved takes one away.
//
// Note that this does not erase LineAdded — both stay in the log forever, and
// the state is what the two of them fold to. That is the whole idea, and it is
// also why the log grows faster than the data.
type LineRemoved struct {
	SKU string
	At  time.Time
}

func (e LineRemoved) EventType() string     { panic("TODO") }
func (e LineRemoved) OccurredAt() time.Time { panic("TODO") }

// OrderConfirmed locks the order.
type OrderConfirmed struct {
	At time.Time
}

func (e OrderConfirmed) EventType() string     { panic("TODO") }
func (e OrderConfirmed) OccurredAt() time.Time { panic("TODO") }

// AddressCorrected fixes a mistake in the shipping address.
//
// **This is question 1, made concrete.** It is named "corrected" rather than
// "changed" because the two mean different things: corrected says the old value
// was always wrong, changed says the customer moved. An auditor cares. A
// database row cannot tell you which happened; an event log can, and only if
// you name it deliberately.
type AddressCorrected struct {
	Address string
	Reason  string
	At      time.Time
}

func (e AddressCorrected) EventType() string     { panic("TODO") }
func (e AddressCorrected) OccurredAt() time.Time { panic("TODO") }

// OrderState is the fold's result — a projection, not a stored object.
type OrderState struct {
	OrderID    string
	CustomerID string
	Status     string // "placed", "confirmed"
	Address    string
	Lines      map[string]int // SKU → quantity
	TotalCent  int64
	Version    int // how many events produced this state
}

// Apply folds one event onto a state, returning the new state.
//
// Worked sequence:
//
//	{}                              + OrderPlaced{ord-1, cust-1}
//	  → {ord-1, cust-1, "placed", Version 1}
//	that                            + LineAdded{sku-1, 2, 1000}
//	  → Lines{sku-1: 2}, TotalCent 2000, Version 2
//	that                            + LineAdded{sku-1, 1, 1000}
//	  → Lines{sku-1: 3}, TotalCent 3000, Version 3
//	that                            + LineRemoved{sku-1}
//	  → Lines{}, TotalCent 0, Version 4
//	that                            + OrderConfirmed{}
//	  → Status "confirmed", Version 5
//
// **Apply must be pure.** Same state plus same event gives the same result,
// every time, with no clock, no randomness, and no I/O. That is not a style
// preference: replay calls this on decade-old events, and anything
// non-deterministic makes history change under you. Same constraint as a
// Temporal workflow in L11, arrived at from a completely different direction.
//
// It must not mutate the state it was given, either. Lines is a map, so the
// obvious implementation shares it with every earlier version — and then
// replaying to version 3 quietly gives you version 5's lines.
//
// An event that cannot apply from this state — a line added after confirmation
// — is ErrInvalidTransition. An event type the fold does not know is
// ErrUnknownEvent.
func Apply(state OrderState, event Event) (OrderState, error) {
	panic("TODO")
}

// Replay folds an entire stream from nothing.
//
//	Replay(nil)                          → ErrEmptyStream
//	Replay([placed, added, confirmed])   → the state after all three
//
// The first event must be an OrderPlaced; a stream that starts any other way is
// ErrInvalidTransition.
//
// Version must equal the number of events applied. That number is what makes
// optimistic concurrency work in Append below.
func Replay(events []Event) (OrderState, error) {
	panic("TODO")
}

// ReplayTo folds only the first n events — the time machine.
//
//	ReplayTo(events, 2) → the state as of the second event
//	ReplayTo(events, 0) → ErrEmptyStream
//	n greater than len   → the same as Replay
//
// This is the capability an update-in-place model simply does not have, and it
// is worth pausing on: "what did this order look like before the discount was
// applied?" is a question you can now answer without having planned to.
func ReplayTo(events []Event, n int) (OrderState, error) {
	panic("TODO")
}

// Snapshot is a cached fold, so a long stream need not be replayed from zero.
type Snapshot struct {
	State   OrderState
	Version int
}

// ReplayFrom continues from a snapshot, applying only what came after it.
//
//	snapshot at version 3, a stream of 5 → apply events 4 and 5
//	snapshot newer than the stream       → ErrVersionConflict
//
// **This is question 2.** A snapshot is only safe while the fold that produced
// it is the fold you are still using: change Apply, and every snapshot becomes
// a cached wrong answer that no test will catch, because the events they came
// from still replay correctly.
//
// Real systems version their snapshots against the code that made them and
// discard on mismatch. Decide whether this one should, and note that it is
// cheaper to discard a snapshot than to explain a wrong total.
func ReplayFrom(snapshot Snapshot, events []Event) (OrderState, error) {
	panic("TODO")
}

// Stream is an append-only event log for one order.
type Stream struct {
	// Yours. Events, and whatever you need to enforce versions.
}

// NewStream returns an empty log, at version 0.
func NewStream() *Stream {
	panic("TODO")
}

// Append adds events, but only if the stream is still at expectedVersion.
//
//	empty stream, Append(0, [placed])       → ok, now at version 1
//	at version 1, Append(1, [added])        → ok, now at version 2
//	at version 2, Append(1, [added])        → ErrVersionConflict, nothing appended
//
// **This is question 5, and it is the only concurrency control an event store
// has.** Two users editing the same order both read version 4, both compute a
// change, and both try to append at 4. One wins; the other is told to re-read
// and try again.
//
// It is the same optimistic locking as the ResourceVersion in L13's Kubernetes
// model, and the version field in L6's repository. Third appearance — by now it
// should read as *the* answer to "two writers, no locks".
//
// A rejected append must leave the stream untouched. Half-appending is worse
// than refusing.
func (s *Stream) Append(expectedVersion int, events ...Event) error {
	panic("TODO")
}

// Events returns the whole log, oldest first.
//
// Return a copy. An append-only log that a caller can rewrite is not a log, and
// this is the one place in the codebase where handing out the internal slice
// would undermine the entire premise.
func (s *Stream) Events() []Event {
	panic("TODO")
}

// Version reports how many events the stream holds.
func (s *Stream) Version() int {
	panic("TODO")
}

// State replays the stream.
//
// Every call is a full fold. That is the cost the package comment warned about,
// and Snapshot is the mitigation — but measure before reaching for it. A
// thousand events fold in microseconds, and a snapshot you did not need is a
// cache-invalidation bug you did not need either.
func (s *Stream) State() (OrderState, error) {
	panic("TODO")
}
