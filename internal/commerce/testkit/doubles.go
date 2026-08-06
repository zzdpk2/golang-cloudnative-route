package testkit

import "context"

// This file is your work. You will build the same collaborator four times, each
// time trading away a different thing. Do them in order — the point of the
// exercise is the comparison, not any single one of them.
//
// Before you start, write down what you expect the difference to be. Then check
// your prediction against what the tests actually force you to write.

// ---- 1. Fake: a real implementation, simplified ----

// FakeInventory is a working warehouse that happens to live in a map. It has
// genuine behaviour: reserving reduces what is available, and reserving more
// than is on hand fails the same way the real warehouse would.
//
// A fake is the only double you can hand to a test that exercises a *sequence*
// of operations, because it is the only one that remembers what happened.
type FakeInventory struct {
	// The state a working in-memory warehouse needs.
	// Think about what Reserve has to change for a later Available to see it.
}

// NewFakeInventory seeds the warehouse with starting quantities, keyed by SKU.
//
// Decide whether to copy the caller's map. A test that mutates its seed data
// afterwards should not be able to reach inside your fake.
func NewFakeInventory(stock map[string]int) *FakeInventory {
	panic("TODO")
}

// Available returns the sellable quantity, or ErrUnknownSKU when the warehouse
// does not carry the SKU at all. Note that "carried but zero" and "not carried"
// are different answers.
func (f *FakeInventory) Available(ctx context.Context, sku string) (int, error) {
	panic("TODO")
}

// Reserve holds stock, reducing what a later Available call reports.
// It returns ErrUnknownSKU or ErrOutOfStock as appropriate.
func (f *FakeInventory) Reserve(ctx context.Context, sku string, qty int) error {
	panic("TODO")
}

// ---- 2. Stub: canned answers, no behaviour ----

// StubInventory returns whatever the test told it to return, and forgets
// everything. It cannot express "the second call returns something different",
// which is exactly the limitation you are here to feel.
//
// Use a stub when the collaborator's answer is an *input* to the case you are
// testing, and how it was reached is irrelevant.
type StubInventory struct {
	AvailableQty int
	AvailableErr error
	ReserveErr   error
}

// Available returns the canned quantity and error, ignoring the SKU entirely.
func (s *StubInventory) Available(ctx context.Context, sku string) (int, error) {
	panic("TODO")
}

// Reserve returns the canned error and records nothing.
func (s *StubInventory) Reserve(ctx context.Context, sku string, qty int) error {
	panic("TODO")
}

// ---- 3. Spy: records what happened ----

// NotifyCall is one recorded invocation of Notifier.Notify.
type NotifyCall struct {
	CustomerID string
	Message    string
}

// SpyNotifier delivers nothing and remembers everything.
//
// A spy answers questions a fake cannot: was this called at all, how many
// times, with what, and in what order. That power is also the trap — see the
// note at the bottom of this file.
type SpyNotifier struct {
	// What you need to answer "was Notify called, how often, and with what".
	// Also give the test a way to make Notify fail.
	Err error
}

// Notify records the call and returns whatever failure the test configured.
// It must record the call even when it is about to return an error: the test
// for "notification failed but the reservation stands" depends on that.
func (s *SpyNotifier) Notify(ctx context.Context, customerID, message string) error {
	panic("TODO")
}

// Calls returns the recorded invocations in the order they happened.
// Consider whether handing back the internal slice is safe here — you answered
// this same question in Order.Lines back in L2.
func (s *SpyNotifier) Calls() []NotifyCall {
	panic("TODO")
}

// ---- 4. Generated mock ----
//
// Run this to generate a mock for InventoryPort and Notifier:
//
//	go run go.uber.org/mock/mockgen -source=subject.go -destination=mock_ports_test.go -package=testkit
//
// The generated mock enforces expectations declared up front: which calls must
// happen, with which arguments, how many times, and in what order. Compare
// mock_ports_test.go with SpyNotifier above and ask:
//
//   - How much of what the generator wrote would you actually have written?
//   - The mock fails the test from *inside* the collaborator, at the moment of
//     the unexpected call. The spy fails later, in your assertion. Which
//     failure message would you rather debug at 3am?
//
// ---- The judgement this exercise is really about ----
//
// A mock asserts on *how* the subject talked to its collaborator. That couples
// the test to the implementation: rename a method, reorder two independent
// calls, add a cache, and a green test goes red without a single behaviour
// having changed. Tests like that are a liability — they make refactoring
// expensive, which is the opposite of what tests are for.
//
// A rough rule to argue with, not to obey: assert on *state* (what came back,
// what the fake now holds) by default, and reach for call assertions only when
// the interaction genuinely is the behaviour — "we must not charge the card
// twice" is about the call, not the state.
//
// For each of the tests in this package, ask which style it used and whether
// you agree with the choice.
