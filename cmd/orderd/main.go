// Command orderd runs the order service.
//
// This file is wiring, not an exercise: it is fully implemented so that the
// moment L7 goes green you can start the binary and place a real order.
//
//	go run ./cmd/orderd
//	curl -s localhost:8080/orders -H 'Content-Type: application/json' -d '{
//	  "customer_id": "cust-1",
//	  "shipping": {"street":"1 George St","city":"Sydney","state":"NSW",
//	               "postcode":"2000","country":"AU"},
//	  "lines": [{"product_id":"sku-1","name":"Widget","price":19.99,
//	             "currency":"AUD","quantity":2}]
//	}'
//
// Before L7 it will panic on the first TODO it reaches. That is the point: this
// is the target the milestones are aiming at.
//
// Read it anyway. Composition root is a real pattern — this is the one place in
// the program that knows how every piece is wired together, which is exactly
// why no other file has to.
package main

import (
	"context"
	"flag"
	"fmt"
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rex/go-ddd-tdd/internal/commerce/application"
	"github.com/rex/go-ddd-tdd/internal/commerce/eventbus"
	httpadapter "github.com/rex/go-ddd-tdd/internal/commerce/http"
	"github.com/rex/go-ddd-tdd/internal/commerce/persistence"
)

const (
	defaultAddr     = ":8080"
	shutdownTimeout = 10 * time.Second
)

// busPublisher adapts the event bus to the narrower interface the application
// layer asked for. The application package defines what it needs; the adapter
// bends to fit. Dependencies point inward — that is the whole rule.
type busPublisher struct {
	bus *eventbus.InMemoryEventBus
}

func (p busPublisher) Publish(evt domain.DomainEvent) { p.bus.Publish(evt) }

func (p busPublisher) PublishAll(events []domain.DomainEvent) {
	for _, e := range events {
		p.bus.Publish(e)
	}
}

func main() {
	preflightOnly := flag.Bool("preflight", false,
		"report which levels the service still needs, then exit")
	flag.Parse()

	// Check before wiring anything up, so an unimplemented level produces a
	// readable report instead of a panic from somewhere inside the plumbing.
	ready, report := preflight()
	if *preflightOnly || !ready {
		fmt.Print(report)
		if !ready {
			os.Exit(1)
		}
		return
	}

	addr := defaultAddr
	if v := os.Getenv("ORDERD_ADDR"); v != "" {
		addr = v
	}

	logger := log.New(os.Stdout, "orderd ", log.LstdFlags|log.Lmsgprefix)

	bus := eventbus.NewInMemoryEventBus(256)
	bus.Subscribe("order.created", func(e domain.DomainEvent) {
		logger.Printf("event %s", e.EventName())
	})

	service := application.NewOrderService(
		persistence.NewInMemoryOrderRepository(),
		persistence.NewInMemoryProductRepository(),
		busPublisher{bus: bus},
	)

	server := httpadapter.NewServer(addr, service)

	// Start serving in the background so the main goroutine can wait for a
	// signal. Without this the process would sit in ListenAndServe and Ctrl-C
	// would kill it mid-request.
	errCh := make(chan error, 1)
	go func() {
		logger.Printf("listening on %s", addr)
		errCh <- server.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		logger.Fatalf("server stopped: %v", err)
	case sig := <-stop:
		logger.Printf("received %s, draining", sig)
	}

	// Graceful shutdown: stop accepting new connections, let in-flight requests
	// finish, and give up after the timeout rather than hanging forever.
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Printf("shutdown: %v", err)
		os.Exit(1)
	}
	logger.Print("stopped cleanly")
}
