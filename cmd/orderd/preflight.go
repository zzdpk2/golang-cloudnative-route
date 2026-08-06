package main

import (
	"fmt"
	domain "github.com/rex/go-ddd-tdd/internal/commerce/domain"
	"strings"

	"github.com/rex/go-ddd-tdd/internal/commerce/application"
	"github.com/rex/go-ddd-tdd/internal/commerce/eventbus"
	"github.com/rex/go-ddd-tdd/internal/commerce/persistence"
)

// The service cannot start until the levels it depends on are implemented, and
// that is the design — orderd is the *target*, not a given.
//
// What follows exists so that starting it early tells you something useful.
// Without it the first unimplemented function panics and you get a stack trace
// pointing at a line of your own scaffolding, which answers no question you
// were actually asking.

// check is one thing the service needs before it can serve a request.
type check struct {
	level string
	what  string
	probe func()
}

// probe runs fn and converts a panic("TODO") into "not implemented yet".
//
// This is fp.Catch from L4 in its most practical form: converting a panic into
// a value at a boundary where a crash would be useless to the caller.
func probe(fn func()) (ok bool, reason string) {
	defer func() {
		if r := recover(); r != nil {
			ok = false
			if s, isStr := r.(string); isStr && s == "TODO" {
				reason = "not implemented yet"
				return
			}
			reason = fmt.Sprintf("panicked: %v", r)
		}
	}()
	fn()
	return true, ""
}

// checks are ordered the way the roadmap is, so the first failure is the next
// thing to work on.
func checks() []check {
	return []check{
		{"L1", "domain.NewMoney", func() {
			if _, err := domain.NewMoney(19.99, domain.AUD); err != nil {
				panic(err)
			}
		}},
		{"L1", "domain.NewAddress", func() {
			if _, err := domain.NewAddress("1 George St", "Sydney", "NSW", "2000", "AU"); err != nil {
				panic(err)
			}
		}},
		{"L1", "domain.NewQuantity", func() {
			if _, err := domain.NewQuantity(1); err != nil {
				panic(err)
			}
		}},
		{"L2", "domain.NewOrder", func() {
			addr, err := domain.NewAddress("1 George St", "Sydney", "NSW", "2000", "AU")
			if err != nil {
				panic(err)
			}
			if _, err := domain.NewOrder("preflight", "cust", addr); err != nil {
				panic(err)
			}
		}},
		{"L2", "Order.AddLine", func() {
			addr, _ := domain.NewAddress("1 George St", "Sydney", "NSW", "2000", "AU")
			order, err := domain.NewOrder("preflight", "cust", addr)
			if err != nil {
				panic(err)
			}
			price, _ := domain.NewMoney(1, domain.AUD)
			qty, _ := domain.NewQuantity(1)
			if err := order.AddLine(domain.ProductID("sku"), "Widget", price, qty); err != nil {
				panic(err)
			}
		}},
		{"L4", "application.CreateOrder", func() {
			svc := application.NewOrderService(
				persistence.NewInMemoryOrderRepository(),
				persistence.NewInMemoryProductRepository(),
				busPublisher{bus: eventbus.NewInMemoryEventBus(8)},
			)
			_ = svc
		}},
		{"L6", "persistence.NewInMemoryOrderRepository", func() {
			_ = persistence.NewInMemoryOrderRepository()
		}},
		{"L8", "eventbus.NewInMemoryEventBus", func() {
			_ = eventbus.NewInMemoryEventBus(8)
		}},
	}
}

// preflight reports whether the service can start, and what is missing if not.
func preflight() (ready bool, report string) {
	var b strings.Builder
	var firstMissing *check

	b.WriteString("\n  preflight\n\n")

	for _, c := range checks() {
		ok, reason := probe(c.probe)
		mark := "ok  "
		if !ok {
			mark = "FAIL"
			if firstMissing == nil {
				missing := c
				firstMissing = &missing
			}
		}
		line := fmt.Sprintf("      %s  %-4s %s", mark, c.level, c.what)
		if !ok {
			line += "   " + reason
		}
		b.WriteString(line + "\n")
	}

	if firstMissing == nil {
		b.WriteString("\n  everything the service needs is implemented.\n\n")
		return true, b.String()
	}

	fmt.Fprintf(&b, `
  orderd cannot start yet, and that is expected — the service is the target you
  are building toward, not something that works out of the box.

  Next: %s, which needs %s.

      go run ./cmd/exercise next
      go test ./test/e2e -run TestM6 -v

  M6 is the milestone that makes this binary run. Once it is green, come back.

`, firstMissing.level, firstMissing.what)

	return false, b.String()
}
