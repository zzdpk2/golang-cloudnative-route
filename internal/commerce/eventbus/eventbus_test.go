package eventbus

import (
	"context"
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestEventBus_SyncPublish(t *testing.T) {
	bus := NewInMemoryEventBus(10)
	defer bus.Close()

	var received []string
	var mu sync.Mutex

	bus.Subscribe("order.created", func(evt domain.DomainEvent) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, evt.AggregateID())
	})

	evt := domain.NewOrderCreated("order-1", "cust-1")
	bus.Publish(evt)

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 1 || received[0] != "order-1" {
		t.Errorf("received = %v", received)
	}
}

func TestEventBus_MultipleHandlers(t *testing.T) {
	bus := NewInMemoryEventBus(10)
	defer bus.Close()

	var count int32

	for i := 0; i < 3; i++ {
		bus.Subscribe("order.created", func(evt domain.DomainEvent) {
			atomic.AddInt32(&count, 1)
		})
	}

	bus.Publish(domain.NewOrderCreated("o1", "c1"))

	if atomic.LoadInt32(&count) != 3 {
		t.Errorf("count = %d, want 3", count)
	}
}

func TestEventBus_DifferentEvents(t *testing.T) {
	bus := NewInMemoryEventBus(10)
	defer bus.Close()

	var created, confirmed int32

	bus.Subscribe("order.created", func(evt domain.DomainEvent) {
		atomic.AddInt32(&created, 1)
	})
	bus.Subscribe("order.confirmed", func(evt domain.DomainEvent) {
		atomic.AddInt32(&confirmed, 1)
	})

	bus.Publish(domain.NewOrderCreated("o1", "c1"))
	bus.Publish(domain.NewOrderConfirmed("o1", 1000))

	if atomic.LoadInt32(&created) != 1 {
		t.Errorf("created = %d", created)
	}
	if atomic.LoadInt32(&confirmed) != 1 {
		t.Errorf("confirmed = %d", confirmed)
	}
}

func TestEventBus_UnsubscribedEvent(t *testing.T) {
	bus := NewInMemoryEventBus(10)
	defer bus.Close()

	bus.Publish(domain.NewOrderCancelled("o1", "no reason"))
}

func TestEventBus_AsyncPublish(t *testing.T) {
	bus := NewInMemoryEventBus(100)

	var count int32
	bus.Subscribe("order.created", func(evt domain.DomainEvent) {
		atomic.AddInt32(&count, 1)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bus.StartWorkers(ctx, 3)

	for i := 0; i < 50; i++ {
		bus.PublishAsync(domain.NewOrderCreated("o", "c"))
	}

	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatalf("timeout: only processed %d/50", atomic.LoadInt32(&count))
		default:
			if atomic.LoadInt32(&count) == 50 {
				return // See the corresponding tests for the intended behavior.
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestEventBus_ContextCancel(t *testing.T) {
	bus := NewInMemoryEventBus(10)

	ctx, cancel := context.WithCancel(context.Background())
	bus.StartWorkers(ctx, 2)

	cancel() // See the corresponding tests for the intended behavior.

	done := make(chan struct{})
	go func() {
		bus.Close()
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(2 * time.Second):
		t.Fatal("Close() deadlocked after context cancel")
	}
}

func TestEventBus_CloseIdempotent(t *testing.T) {
	bus := NewInMemoryEventBus(10)

	bus.Close()
	bus.Close()
	bus.Close()
}

func TestEventBus_PublishAfterClose(t *testing.T) {
	bus := NewInMemoryEventBus(10)
	bus.Close()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("should not panic on publish after close, got: %v", r)
		}
	}()
	bus.PublishAsync(domain.NewOrderCreated("o1", "c1"))
}

func TestEventBus_PublishAll(t *testing.T) {
	bus := NewInMemoryEventBus(10)
	defer bus.Close()

	var count int32
	bus.Subscribe("order.created", func(evt domain.DomainEvent) {
		atomic.AddInt32(&count, 1)
	})
	bus.Subscribe("order.confirmed", func(evt domain.DomainEvent) {
		atomic.AddInt32(&count, 1)
	})

	events := []domain.DomainEvent{
		domain.NewOrderCreated("o1", "c1"),
		domain.NewOrderConfirmed("o1", 1000),
	}
	bus.PublishAll(events)

	if atomic.LoadInt32(&count) != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}
