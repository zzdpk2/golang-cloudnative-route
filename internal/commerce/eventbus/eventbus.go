// Package eventbus delivers domain events to whoever is interested.
//
// It is the real version of the Observer pattern in lab/patterns — the one that
// has to answer the questions the textbook version ducks: what happens when a
// subscriber is slow, when one panics, when the bus is closed mid-publish, when
// two goroutines subscribe while a third is publishing.
//
// # Two delivery modes, and the choice matters
//
//	Publish       synchronous. The caller waits for every handler.
//	PublishAsync  queued. The caller returns immediately; workers deliver.
//
// Confirming an order should not wait for an SMS gateway, which is the argument
// for async. But async means the event can be lost if the process dies before a
// worker picks it up — an in-memory bus is **not** durable, and treating it as
// though it were is how "we definitely sent that email" becomes untrue.
//
// A real system puts a broker behind this interface. The interface is the point:
// application.EventPublisher names only what the use case needs, so swapping
// this for Kafka touches one file in cmd/.
package eventbus

import (
	"context"
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"sync"
)

// EventHandler reacts to an domain.
//
// It returns nothing, which is a design decision worth noticing: a handler
// cannot report failure to the publisher, because the publisher has already
// moved on and could not do anything about it anyway. Handlers own their own
// retries — which is what platform/retry is for.
type EventHandler func(domain.DomainEvent)

// InMemoryEventBus is a publish/subscribe bus with an optional worker pool.
type InMemoryEventBus struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler // eventName → handlers

	eventCh chan domain.DomainEvent

	once sync.Once
	done chan struct{}
	wg   sync.WaitGroup
}

// NewInMemoryEventBus builds a bus whose async queue holds bufferSize events.
//
//	NewInMemoryEventBus(256) → a usable bus, no workers running yet
//
// Initialise the map, the buffered channel, and the done channel. A nil map
// panics on write, and a nil channel blocks forever — both are silent until the
// first Publish.
//
// The buffer size is the async queue depth, and it is a real capacity decision:
// too small and PublishAsync blocks the request path it was meant to protect,
// too large and a burst is absorbed into memory you did not budget for.
func NewInMemoryEventBus(bufferSize int) *InMemoryEventBus {
	panic("TODO")
}

// Subscribe registers a handler for one event name.
//
//	Subscribe("order.created", sendEmail)
//	Subscribe("order.created", updateStats)   → both are called, in this order
//
// Several handlers may share a name, and a handler may be registered for
// several names. Take the write lock: subscribing usually happens at start-up,
// but nothing here enforces that, and a map written during a publish is a hard
// crash rather than a race.
func (bus *InMemoryEventBus) Subscribe(eventName string, handler EventHandler) {
	panic("TODO")
}

// Publish delivers an event to its handlers, synchronously, before returning.
//
//	with two handlers for "order.created" → both have run when Publish returns
//	with no handler for the event name    → returns quietly, not an error
//
// An event nobody listens to is normal, not a failure — that is what decoupling
// buys, and a bus that complained would defeat it.
//
// # The locking trap
//
// Holding the read lock while calling handlers means a handler that calls
// Subscribe deadlocks, and a slow handler blocks every subscriber. Copy the
// handler slice under the lock, release it, then call. That copy also protects
// you from a handler list that changes mid-delivery.
//
// A panicking handler currently takes the publisher down with it. Decide
// whether that is acceptable — and note that lab/patterns.OrderSubject has the
// same gap, deliberately, so you could find it here.
func (bus *InMemoryEventBus) Publish(evt domain.DomainEvent) {
	panic("TODO")
}

// PublishAsync queues an event for a worker and returns immediately.
//
//	with workers running    → the handlers run shortly, on a worker goroutine
//	after Close             → returns without blocking and without panicking
//	when the queue is full  → see below
//
// Sending on a closed channel panics, so a send here has to be guarded by a
// select on bus.done. That is the standard shape for "send unless we are
// shutting down".
//
// The full-queue case is the same three-way choice as PubSub in L6: block the
// caller, drop the event, or grow without limit. Decide, and be able to say
// which failure you chose — an event silently dropped is a customer who never
// got their confirmation.
func (bus *InMemoryEventBus) PublishAsync(evt domain.DomainEvent) {
	panic("TODO")
}

// StartWorkers launches n goroutines draining the async queue.
//
//	StartWorkers(ctx, 3) → three workers, all reading the same channel
//
// They all read one channel, so Go distributes the events for you and a slow
// handler on one worker does not stall the others. No scheduling logic needed —
// the fan-out pattern from L6.
//
// Add to the WaitGroup *before* launching each goroutine, never inside it. The
// reason is the one flagged in errors.ErrorGroup.Go, and it bites the same way:
// Close's Wait can observe zero and return before anything has started.
func (bus *InMemoryEventBus) StartWorkers(ctx context.Context, n int) {
	panic("TODO")
}

// worker drains the queue until told to stop.
//
// Three ways to stop, and it must respect all of them:
//
//	the channel is closed   the two-value receive reports it
//	ctx is cancelled        the caller gave up
//	bus.done is closed      Close was called
//
// A closed channel is always ready to receive, so forgetting the `ok` check
// spins a goroutine at 100% CPU receiving zero values forever. That is the
// classic worker-loop bug and it does not announce itself — the program simply
// gets hot.
func (bus *InMemoryEventBus) worker(ctx context.Context) {
	panic("TODO")
}

// Close shuts the bus down: no more async publishing, workers finished.
//
//	Close()          → returns once every worker has stopped
//	Close(); Close() → the second call does nothing and must not panic
//
// sync.Once is why the second call is safe — closing an already-closed channel
// panics, and Close is exactly the method people both defer and call
// explicitly.
//
// Ordering is the exercise. Signal the workers, stop accepting sends, wait for
// them to drain. Get it wrong and you either panic on a send to a closed
// channel or block forever waiting for workers that were never told to stop.
//
// Then a question with no free answer: should Close deliver the events already
// queued, or drop them? Draining is what a user expects; dropping is what a
// shutdown deadline may force. Say which you chose.
func (bus *InMemoryEventBus) Close() {
	panic("TODO")
}

// PublishAll delivers a batch synchronously — the shape the application layer
// uses after CollectEvents.
func (bus *InMemoryEventBus) PublishAll(events []domain.DomainEvent) { panic("TODO") }

// PublishAllAsync queues a batch.
func (bus *InMemoryEventBus) PublishAllAsync(events []domain.DomainEvent) { panic("TODO") }
