// Package actor is the L12 exercise: the actor model as an alternative answer
// to the problem you solved with a mutex in L6.
//
// Back in L6 you protected shared inventory with sync.RWMutex: many goroutines
// touch one piece of state, so you serialise access to the state. The actor
// model inverts that. State is owned by exactly one actor, nobody else can
// reach it, and the only way to interact is to send a message. Messages for one
// actor are processed one at a time, in order, by one goroutine.
//
// So there is no lock — not because locking was made cheaper, but because the
// sharing that required it was removed.
//
// That trade is not free, and naming the costs is most of the point of this
// exercise:
//
//   - a mutex blocks the caller until the work is done; a send returns
//     immediately and the work happens later. Getting an answer back means
//     Request/Response, which reintroduces waiting — and a timeout you now have
//     to choose.
//   - an actor's mailbox is an unbounded queue by default. Under sustained
//     overload it grows until the process dies. A mutex under overload just
//     makes callers slow. Which failure would you rather explain?
//   - one actor per SKU serialises that SKU, which is exactly what you want for
//     a hot item. It also means a single hot SKU is handled by a single
//     goroutine, and cannot use more than one core no matter how many you have.
//
// Work through the tests, then write down which of L6 or L12 you would actually
// ship for the seckill scenario, and why.
package workflows

import (
	"time"

	"github.com/asynkron/protoactor-go/actor"
)

// ---- Messages ----
//
// Messages are plain structs. Keep them immutable: once sent, a message may be
// read by another goroutine, and mutating it afterwards reintroduces exactly
// the data race the actor model was supposed to remove.

// Reserve asks the inventory actor to hold Qty units.
type Reserve struct {
	Qty int
}

// Release returns previously reserved units.
type Release struct {
	Qty int
}

// GetAvailable asks for the current sellable quantity.
type GetAvailable struct{}

// ReserveResult answers a Reserve.
type ReserveResult struct {
	OK        bool
	Available int
}

// AvailableResult answers a GetAvailable.
type AvailableResult struct {
	Available int
}

// ---- The actor ----

// InventoryActor owns the stock level for exactly one SKU.
//
// Note what is *not* here: no mutex, no atomic, no channel. If you find
// yourself reaching for one, stop — it means state is leaking out of the actor,
// and the design has gone wrong.
type InventoryActor struct {
	// The stock this actor owns. A plain int is correct here. Convince
	// yourself why before moving on.
}

// NewInventoryProps returns the Props that spawn an InventoryActor holding
// initial units.
//
// Props is a recipe, not an instance: the producer function runs again every
// time the actor is restarted after a failure. That has a consequence for
// state, which the restart test below will make you confront.
func NewInventoryProps(initial int) *actor.Props {
	panic("TODO")
}

// Receive handles one message at a time. This method is never entered
// concurrently for the same actor, which is the whole guarantee.
//
// Handle Reserve, Release, and GetAvailable. Use ctx.Message() with a type
// switch, and ctx.Respond() to answer a request.
//
// Two things to get right:
//   - respond to every request that expects an answer. A caller blocked in
//     RequestFuture with no response waits for the full timeout, and you will
//     debug it as "the system is slow" rather than "I forgot a Respond".
//   - actors receive lifecycle messages too — *actor.Started, *actor.Stopping,
//     *actor.Restarting. Your switch will see them. Decide whether to ignore
//     them or use them.
func (a *InventoryActor) Receive(ctx actor.Context) {
	panic("TODO")
}

// ---- Convenience wrappers ----
//
// These wrap the raw send/request calls so the tests read like business code.
// They also mark the boundary: outside this package, nobody should know that
// inventory happens to be implemented with actors.

// AskAvailable requests the current quantity and waits for the answer.
//
// RequestFuture is how you get a value back out of an actor. It needs a
// timeout, because the actor may be busy, restarting, or dead. Pick one, then
// ask yourself what the caller should do when it expires — the answer is not
// "retry forever".
func AskAvailable(system *actor.ActorSystem, pid *actor.PID, timeout time.Duration) (int, error) {
	panic("TODO")
}

// AskReserve requests a reservation and waits for the result.
func AskReserve(system *actor.ActorSystem, pid *actor.PID, qty int, timeout time.Duration) (ReserveResult, error) {
	panic("TODO")
}

// ---- Supervision ----

// PanicOn is a message that makes the actor panic, so the tests can observe
// what supervision does. Real actors do not have one of these; it is here to
// let you watch a restart happen.
type PanicOn struct{}

// NewSupervisedInventoryProps is NewInventoryProps plus a supervisor strategy
// that restarts the actor on panic instead of stopping it.
//
// Use actor.NewOneForOneStrategy. Then answer the question the restart test
// asks: after a restart, the producer from Props runs again — so what happened
// to the stock level the actor had accumulated? Is that the behaviour you want
// for inventory, and if not, where does the state have to live instead?
func NewSupervisedInventoryProps(initial int) *actor.Props {
	panic("TODO")
}
