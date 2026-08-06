package patterns

import (
	"fmt"
	"strings"
	"sync"
)

// ============================================================
// The 5 creational patterns
//
//	Singleton, Factory Method, Abstract Factory, Builder, Prototype
//
// These are the patterns Go changes the most. Two of them (Singleton,
// Prototype) are things the language and its idioms mostly argue you out of,
// and each says so below.
// ============================================================

// ------------------------------------------------------------
// Singleton
// ------------------------------------------------------------

// AppConfig is the process-wide configuration.
type AppConfig struct {
	Name string
	Port int
}

var (
	appConfigOnce     sync.Once
	appConfigInstance *AppConfig
)

// GetAppConfig returns the one config, creating it on first use.
//
//	GetAppConfig()               → &AppConfig{Name: "go-ddd-tdd", Port: 8080}
//	GetAppConfig() == GetAppConfig() → true, the same pointer every time
//
// sync.Once is the Go way, and it does more than a nil check would: a second
// caller arriving mid-initialisation *blocks* until the first finishes rather
// than seeing a half-built value. Same guarantee as fp.Lazy in L4.
//
// **Be suspicious of this pattern.** A singleton is global mutable state with a
// respectable name. It cannot be substituted in a test, it hides a dependency
// that should have been a parameter, and it makes initialisation order matter.
//
// Notice that the main line does not use one: cmd/orderd builds the config and
// *passes* it inward. That is the composition-root answer, and it is better in
// every way except keystrokes.
func GetAppConfig() *AppConfig {
	panic("TODO")
}

// ------------------------------------------------------------
// Factory Method
// ------------------------------------------------------------

// Payment settles an order.
type Payment interface {
	Pay(amountCents int) string
}

// CreditCardPayment charges a card.
//
//	Pay(1999) → "paid 1999 by credit card"
type CreditCardPayment struct{}

// Pay returns "paid <amount> by credit card".
func (p *CreditCardPayment) Pay(amountCents int) string {
	panic("TODO")
}

// PayPalPayment charges a PayPal account.
//
//	Pay(1999) → "paid 1999 by paypal"
type PayPalPayment struct{}

// Pay returns "paid <amount> by paypal".
func (p *PayPalPayment) Pay(amountCents int) string {
	panic("TODO")
}

// NewPayment picks an implementation by name.
//
//	NewPayment("card")   → *CreditCardPayment, nil
//	NewPayment("paypal") → *PayPalPayment, nil
//	NewPayment("bitcoin") → nil, error
//
// The factory exists so callers can choose at *run time* from a string that
// arrived over the wire — which is exactly when a plain constructor cannot help
// you. If the choice is known at compile time, call the constructor and skip
// this entirely.
//
// Return an error rather than a nil Payment for an unknown method. A nil
// interface returned as a valid result is the trap from lab/types: the caller's
// nil check does not fire, and the panic lands somewhere else.
func NewPayment(method string) (Payment, error) {
	panic("TODO")
}

// ------------------------------------------------------------
// Abstract Factory
// ------------------------------------------------------------

// A shipment needs a label and a tracker, and the two must come from the same
// carrier — a domestic label with an international tracking number is not a
// mistake the type system should allow.
//
// That constraint is what separates Abstract Factory from Factory Method.
// Factory Method makes one thing; Abstract Factory makes a *family* whose
// members must match. If your factory only ever produces one kind of object,
// you wanted the simpler pattern.

// ShippingLabel is what goes on the box.
type ShippingLabel interface {
	Render(address string) string
}

// Tracker reports where a parcel is.
type Tracker interface {
	Track(id string) string
}

// CarrierKit produces a matched label and tracker.
type CarrierKit interface {
	NewLabel() ShippingLabel
	NewTracker() Tracker
}

// DomesticLabel is the local carrier's format.
//
//	Render("1 George St") → "AUPOST label: 1 George St"
type DomesticLabel struct{}

// Render returns "AUPOST label: <address>".
func (l DomesticLabel) Render(address string) string {
	panic("TODO")
}

// DomesticTracker is the local carrier's tracking.
//
//	Track("t1") → "AUPOST tracking t1"
type DomesticTracker struct{}

// Track returns "AUPOST tracking <id>".
func (t DomesticTracker) Track(id string) string {
	panic("TODO")
}

// InternationalLabel carries customs information.
//
//	Render("1 George St") → "DHL label: 1 George St (customs declared)"
type InternationalLabel struct{}

// Render returns "DHL label: <address> (customs declared)".
func (l InternationalLabel) Render(address string) string {
	panic("TODO")
}

// InternationalTracker queries the global network.
//
//	Track("t1") → "DHL tracking t1"
type InternationalTracker struct{}

// Track returns "DHL tracking <id>".
func (t InternationalTracker) Track(id string) string {
	panic("TODO")
}

// DomesticCarrier produces the domestic family.
type DomesticCarrier struct{}

// NewLabel produces a DomesticLabel — matched to NewTracker below.
func (c DomesticCarrier) NewLabel() ShippingLabel {
	panic("TODO")
}

// NewTracker produces a DomesticTracker.
func (c DomesticCarrier) NewTracker() Tracker {
	panic("TODO")
}

// InternationalCarrier produces the international family.
type InternationalCarrier struct{}

// NewLabel produces an InternationalLabel.
func (c InternationalCarrier) NewLabel() ShippingLabel {
	panic("TODO")
}

// NewTracker produces an InternationalTracker.
func (c InternationalCarrier) NewTracker() Tracker {
	panic("TODO")
}

// NewCarrierKit picks a family by destination.
//
//	NewCarrierKit("AU") → DomesticCarrier, nil
//	NewCarrierKit("US") → InternationalCarrier, nil
//	NewCarrierKit("")   → nil, error
//
// Every country other than "AU" is international here. A real system would read
// this from configuration, which is the usual fate of a factory like this —
// and a good reason to keep the selection logic in one place.
func NewCarrierKit(country string) (CarrierKit, error) {
	panic("TODO")
}

// ------------------------------------------------------------
// Builder
// ------------------------------------------------------------

// OrderQuery is a repository query assembled piece by piece.
//
// Builder is for objects with many optional parts, where a constructor taking
// eight arguments would be unreadable and half of them would be zero.
//
// Compare with the functional options in domain/service.NewPricingService.
// Options are the more idiomatic Go answer for *configuring* something;
// a chained builder reads better when you are *composing* something whose parts
// have an order, as a query does. Having written both, you can say why.
type OrderQuery struct {
	CustomerID string
	Status     string
	OrderBy    string
	Limit      int
}

// OrderQueryBuilder accumulates the parts of a query.
type OrderQueryBuilder struct {
	query OrderQuery
	err   error
}

func NewOrderQueryBuilder() *OrderQueryBuilder { panic("TODO") }

// ForCustomer filters by customer id.
func (b *OrderQueryBuilder) ForCustomer(id string) *OrderQueryBuilder {
	panic("TODO")
}

// WithStatus filters by order status.
func (b *OrderQueryBuilder) WithStatus(status string) *OrderQueryBuilder {
	panic("TODO")
}

// OrderedBy sets the sort column.
func (b *OrderQueryBuilder) OrderedBy(column string) *OrderQueryBuilder {
	panic("TODO")
}

// Limit caps the result count.
//
// A non-positive limit is a caller error. Because this returns *OrderQueryBuilder
// for chaining, it cannot return an error — park it in b.err and surface it
// from Build. That deferred-error pattern is the same one platform/txdsl.Step
// and errors.ValidationBuilder use; three occurrences is no longer a
// coincidence, it is the Go idiom for a fluent API.
func (b *OrderQueryBuilder) Limit(n int) *OrderQueryBuilder {
	panic("TODO")
}

// Build validates and returns the query.
//
//	NewOrderQueryBuilder().ForCustomer("c1").Limit(10).Build()
//	  → OrderQuery{CustomerID: "c1", Limit: 10}, nil
//
//	NewOrderQueryBuilder().Build()
//	  → error: a query with no filter at all would scan every order
//
//	NewOrderQueryBuilder().ForCustomer("c1").Limit(-1).Build()
//	  → the parked error
//
// Requiring at least one filter is a real safeguard, not pedantry: an unbounded
// query is how a report page takes down a database.
func (b *OrderQueryBuilder) Build() (OrderQuery, error) {
	panic("TODO")
}

// ------------------------------------------------------------
// Prototype
//
// NOT USED IN THIS SYSTEM — see the note below.
// ------------------------------------------------------------

// NotificationTemplate is a message template that gets copied and customised
// per order.
//
// Prototype means "make a new object by copying an existing one", and it earns
// its place when construction is expensive or the object's exact class is not
// known statically — neither of which is common in Go, where a struct literal
// is cheap and types are known.
//
// **This system does not use it.** Templates here would be values built by a
// constructor, and Go's assignment already copies a struct.
//
// The exercise is worth doing anyway, because Clone is where you meet the
// shallow-versus-deep copy question in its purest form — and that question
// *does* show up in the main line, every time a method hands out a slice.
type NotificationTemplate struct {
	Subject string
	Body    string
	Tags    []string
	Meta    map[string]string
}

// Clone returns an independent copy.
//
//	original := &NotificationTemplate{Subject: "Order shipped", Tags: []string{"tx"}}
//	copy := original.Clone()
//	copy.Tags[0] = "marketing"      → original.Tags[0] must still be "tx"
//	copy.Meta["k"] = "v"            → original.Meta must not gain "k"
//
// **Assigning the struct is not enough.** `*t` copies the fields, but a slice
// field copies the *header* and a map field copies the *reference*, so both
// clones end up sharing storage. The test mutates the clone and checks the
// original, which is exactly how this bug is caught in real code.
//
// A nil slice and a nil map should stay nil rather than becoming empty ones —
// the difference is visible through reflect.DeepEqual and through JSON, and
// silently changing it is its own small bug.
func (t *NotificationTemplate) Clone() *NotificationTemplate {
	panic("TODO")
}

var (
	_ = fmt.Sprintf
	_ = strings.ToUpper
)
