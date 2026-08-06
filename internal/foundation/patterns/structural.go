package patterns

import (
	"fmt"
	"strings"
)

// ============================================================
// The 7 structural patterns
//
//	Adapter, Bridge, Composite, Decorator, Facade, Flyweight, Proxy
//
// Three of these you have already built for real in the main line without
// necessarily naming them: every adapter/ package is Adapter, the middleware
// chains are Decorator, and OrderService is a Facade over the domain.
// ============================================================

// ------------------------------------------------------------
// Adapter
// ------------------------------------------------------------

// WarehouseClient is the interface this system wants to talk to.
//
// The Commerce persistence, HTTP, and event packages use this pattern: the domain declares the
// shape it needs, and an adapter bends some external thing to fit. The arrow
// points inward, always.
type WarehouseClient interface {
	Reserve(sku string, qty int) (string, error)
}

// LegacyWarehouse is the vendor's API. You do not get to change it.
//
// Note everything wrong with it from this system's point of view: one string
// argument, a stringly-typed response, and failure signalled by a prefix rather
// than an error. Adapting is not about the method name — it is about absorbing
// that ugliness in one place instead of spreading it through the domain.
type LegacyWarehouse struct{}

// ReserveItem is the vendor's call.
//
//	ReserveItem("sku-1|2")  → "OK:RES-sku-1"
//	ReserveItem("sku-1|0")  → "ERR:invalid quantity"
//	ReserveItem("garbage")  → "ERR:bad request"
func (w *LegacyWarehouse) ReserveItem(request string) string {
	panic("TODO")
}

// WarehouseAdapter makes LegacyWarehouse satisfy WarehouseClient.
type WarehouseAdapter struct {
	legacy *LegacyWarehouse
}

func NewWarehouseAdapter(legacy *LegacyWarehouse) *WarehouseAdapter { panic("TODO") }

// Reserve translates both directions.
//
//	Reserve("sku-1", 2) → "RES-sku-1", nil
//	Reserve("sku-1", 0) → "", error
//
// Two translations, and the second is the one people skip: encoding the
// arguments into the vendor's format, and turning its "ERR:..." string back
// into a Go error. An adapter that passes the string through has not adapted
// anything — it has moved the problem one file further in.
func (a *WarehouseAdapter) Reserve(sku string, qty int) (string, error) {
	panic("TODO")
}

// ------------------------------------------------------------
// Bridge
// ------------------------------------------------------------

// LabelPrinter is the implementation side of the bridge: *how* a label is
// emitted.
//
// Bridge separates an abstraction from its implementation so the two can vary
// independently. Here: label *kinds* (standard, express, returns) times output
// *formats* (PDF, thermal ZPL).
//
// Without it you write StandardPDF, StandardZPL, ExpressPDF, ExpressZPL — the
// class explosion. Three kinds and four formats is twelve types instead of
// seven. With the bridge you add a kind or a format, never the product.
//
// This is the same idea as io.Writer: your code formats, the Writer decides
// whether the bytes reach a file, a socket, or a buffer.
type LabelPrinter interface {
	Print(content string) string
}

// PDFPrinter renders for a laser printer.
//
//	Print("EXPRESS 1 George St") → "[PDF] EXPRESS 1 George St"
type PDFPrinter struct{}

// Print wraps the content as "[PDF] <content>".
func (p PDFPrinter) Print(content string) string {
	panic("TODO")
}

// ZPLPrinter renders for a thermal label printer.
//
//	Print("EXPRESS 1 George St") → "^XA EXPRESS 1 George St ^XZ"
type ZPLPrinter struct{}

// Print wraps the content as "^XA <content> ^XZ", the thermal-printer format.
func (p ZPLPrinter) Print(content string) string {
	panic("TODO")
}

// Label is the abstraction side: *what* a label says.
type Label interface {
	Emit() string
}

// StandardLabel is an ordinary shipment.
type StandardLabel struct {
	Address string
	printer LabelPrinter
}

func NewStandardLabel(address string, printer LabelPrinter) *StandardLabel { panic("TODO") }

// Emit builds the content and hands it to the printer.
//
//	NewStandardLabel("1 George St", PDFPrinter{}).Emit()
//	  → "[PDF] STANDARD 1 George St"
//
// Notice that the label never knows which printer it has, and the printer never
// knows what kind of label it is rendering. That mutual ignorance is the bridge.
func (l *StandardLabel) Emit() string {
	panic("TODO")
}

// ExpressLabel is a priority shipment.
//
//	NewExpressLabel("1 George St", ZPLPrinter{}).Emit()
//	  → "^XA EXPRESS 1 George St ^XZ"
type ExpressLabel struct {
	Address string
	printer LabelPrinter
}

func NewExpressLabel(address string, printer LabelPrinter) *ExpressLabel { panic("TODO") }

// Emit builds "EXPRESS <address>" and hands it to the printer.
func (l *ExpressLabel) Emit() string {
	panic("TODO")
}

// ------------------------------------------------------------
// Composite
// ------------------------------------------------------------

// ShipmentNode is either a single item or a bundle of them, and callers do not
// have to care which.
//
// This is the tree that L15.3's order splitting works over: a kit is a bundle
// containing items and possibly other kits, and its weight is the sum of what
// it holds.
//
// The pattern's promise is that a leaf and a branch present the same interface,
// so client code never branches on "is this a group?". The price is that the
// interface has to be the union of what both need, and some operations are
// meaningless on one side — which is why real trees often need a way to ask.
type ShipmentNode interface {
	Name() string
	Grams() int
}

// ShipmentItem is a leaf: one physical thing.
type ShipmentItem struct {
	name  string
	grams int
}

func NewShipmentItem(name string, grams int) *ShipmentItem { panic("TODO") }

func (i *ShipmentItem) Name() string { panic("TODO") }

// Grams returns this item's own weight.
func (i *ShipmentItem) Grams() int { panic("TODO") }

// ShipmentBundle is a branch: a kit containing other nodes.
type ShipmentBundle struct {
	name     string
	children []ShipmentNode
}

func NewShipmentBundle(name string) *ShipmentBundle { panic("TODO") }

// Add puts a node inside the bundle. A bundle may contain bundles.
func (b *ShipmentBundle) Add(child ShipmentNode) {
	panic("TODO")
}

func (b *ShipmentBundle) Name() string { panic("TODO") }

// Grams is the sum of everything inside, at any depth.
//
//	bundle "kit" containing item "mug" (300g) and
//	  bundle "spares" containing item "lid" (50g)
//	→ Grams() == 350
//
// The recursion is the pattern: a bundle asks its children for their weight
// without knowing whether any of them is itself a bundle.
//
// An empty bundle weighs 0. And note what nothing here prevents: adding a
// bundle to itself gives you infinite recursion and a stack overflow. Real
// composite implementations need a cycle check, and this one does not have one.
func (b *ShipmentBundle) Grams() int {
	panic("TODO")
}

// ------------------------------------------------------------
// Decorator
// ------------------------------------------------------------

// Notifier tells a customer something.
//
// Decorator wraps an object in another object with the same interface, adding
// behaviour without touching the original. Stack them and the behaviours
// compose.
//
// You have already built this three times: minigin middleware, adapter/http
// middleware, and adapter/middleware. Every one of those is Decorator over a
// handler. Seeing it named here is the point — the pattern is so idiomatic in
// Go that people use it constantly without calling it anything.
type Notifier interface {
	Send(message string) string
}

// EmailNotifier is the base.
//
//	Send("shipped") → "email: shipped"
type EmailNotifier struct{}

// Send returns "email: <message>". This is the base of every decorator chain.
func (n EmailNotifier) Send(message string) string {
	panic("TODO")
}

// SMSDecorator adds an SMS to whatever it wraps.
//
//	NewSMSDecorator(EmailNotifier{}).Send("shipped")
//	  → "email: shipped + sms: shipped"
type SMSDecorator struct {
	next Notifier
}

func NewSMSDecorator(next Notifier) *SMSDecorator { panic("TODO") }

// Send delegates first, then adds its own.
//
// Delegate-then-add gives the base-first order the tests expect. Reverse the
// two statements and the chain reads backwards — the same ordering trap as
// Chain in http/middleware.go.
func (d *SMSDecorator) Send(message string) string {
	panic("TODO")
}

// SlackDecorator adds a Slack post.
//
//	NewSlackDecorator(NewSMSDecorator(EmailNotifier{})).Send("shipped")
//	  → "email: shipped + sms: shipped + slack: shipped"
type SlackDecorator struct {
	next Notifier
}

func NewSlackDecorator(next Notifier) *SlackDecorator { panic("TODO") }

// Send delegates first, then appends " + slack: <message>".
func (d *SlackDecorator) Send(message string) string {
	panic("TODO")
}

// ------------------------------------------------------------
// Facade
// ------------------------------------------------------------

// InventorySystem, PaymentSystem, and ShippingSystem are three subsystems a
// checkout has to drive.
type InventorySystem struct{}

// Reserve holds stock.
//
//	Reserve("sku-1", 2) → "reserved 2 of sku-1"
func (s InventorySystem) Reserve(sku string, qty int) string {
	panic("TODO")
}

type PaymentSystem struct{}

// Charge takes the money.
//
//	Charge("cust-1", 1999) → "charged cust-1 1999"
func (s PaymentSystem) Charge(userID string, amountCents int) string {
	panic("TODO")
}

type ShippingSystem struct{}

// Ship dispatches the parcel.
//
//	Ship("1 George St") → "shipped to 1 George St"
func (s ShippingSystem) Ship(address string) string {
	panic("TODO")
}

// CheckoutFacade gives callers one call instead of three.
//
// A facade simplifies a subsystem for the common case. It does not hide it —
// a caller with an unusual need can still reach the parts directly, and a
// facade that forbids that has become a bottleneck.
//
// application.OrderService is exactly this: one CreateOrder call over the
// aggregate, the repository, and the publisher. Which raises the question worth
// sitting with — **when is a facade an application service, and when is it just
// a function that calls three other functions?** The answer is whether it owns
// a policy. This one owns the order of operations; that is a policy.
type CheckoutFacade struct {
	inventory InventorySystem
	payment   PaymentSystem
	shipping  ShippingSystem
}

func NewCheckoutFacade() *CheckoutFacade { panic("TODO") }

// Checkout runs the three steps and returns what each reported.
//
//	Checkout("cust-1", "sku-1", 2, 1999, "1 George St")
//	  → ["reserved 2 of sku-1", "charged cust-1 1999", "shipped to 1 George St"]
//
// Reserve before charge before ship, and that order is the policy this facade
// exists to own. Note what it does *not* do: nothing here rolls back a charge
// when shipping fails. That gap is exactly what the saga in L10 is for, and
// seeing the facade fall short is the best argument for it.
func (f *CheckoutFacade) Checkout(userID, sku string, qty, amountCents int, address string) []string {
	panic("TODO")
}

// ------------------------------------------------------------
// Flyweight
//
// NOT USED IN THIS SYSTEM — see the note below.
// ------------------------------------------------------------

// SKUSpec is catalogue data shared by every order line for the same product:
// its name, its category, its handling class.
//
// Flyweight splits an object's state into *intrinsic* (shared, immutable — the
// spec) and *extrinsic* (per-use — the quantity on this particular line). A
// million order lines across a million orders reference a few thousand specs
// rather than each carrying a copy.
//
// **This system does not use it.** Order lines here carry their own name and
// price, deliberately: an order is a *historical record*, and if it pointed at
// shared catalogue data then renaming a product would silently rewrite every
// past invoice. That is a serious bug in an accounting system, and copying the
// data is what prevents it.
//
// So Flyweight trades memory for sharing, and sharing is precisely what this
// domain must not do. Learn it for the rendering and game-engine problems where
// it belongs.
type SKUSpec struct {
	Name     string
	Category string
	Handling string
}

// SKUSpecPool hands out shared specs.
type SKUSpecPool struct {
	specs map[string]*SKUSpec
}

func NewSKUSpecPool() *SKUSpecPool { panic("TODO") }

// Get returns the spec for a key, creating it only on first request.
//
//	p.Get("mug", "kitchen", "fragile") == p.Get("mug", "kitchen", "fragile")
//	  → true; the same pointer, and Count() == 1
//
// Returning the *same pointer* is the whole pattern — an equal-but-distinct
// value would defeat it. Key on all three fields, so two genuinely different
// specs do not collide.
//
// And note the danger the shared pointer creates: every holder can mutate what
// everyone else sees. A flyweight must be treated as immutable, and nothing in
// Go enforces that here.
func (p *SKUSpecPool) Get(name, category, handling string) *SKUSpec {
	panic("TODO")
}

// Count reports how many distinct specs exist.
func (p *SKUSpecPool) Count() int {
	panic("TODO")
}

// ------------------------------------------------------------
// Proxy
// ------------------------------------------------------------

// ProductDetail is a heavy description loaded from somewhere slow.
type ProductDetail interface {
	Description() string
}

// RealProductDetail does the expensive load.
type RealProductDetail struct {
	sku      string
	body     string
	loadedAt int
}

func NewRealProductDetail(sku string) *RealProductDetail { panic("TODO") }

// load simulates the expensive fetch.
func (d *RealProductDetail) load() {
	panic("TODO")
}

// Description returns the body, loading it on first use.
//
//	NewRealProductDetail("sku-1").Description() → "detail for sku-1"
func (d *RealProductDetail) Description() string {
	panic("TODO")
}

// LazyProductDetailProxy defers the load until somebody actually asks.
//
// A proxy stands in for another object and controls access to it. Lazy loading
// is one use; the others are access control, remote calls, and caching — and
// once you notice that a gRPC client stub is a proxy for a remote server, the
// pattern stops looking academic.
//
// The listing page for a catalogue renders a hundred products and shows the
// full description for none of them. A proxy per product means a hundred
// fetches that never happen.
type LazyProductDetailProxy struct {
	sku  string
	real *RealProductDetail
}

func NewLazyProductDetailProxy(sku string) *LazyProductDetailProxy { panic("TODO") }

// Description loads on first call and reuses afterwards.
//
//	p := NewLazyProductDetailProxy("sku-1")
//	p.Loaded()        → false, nothing has been fetched
//	p.Description()   → "detail for sku-1"
//	p.Loaded()        → true
//	p.Description()   → the same string, with no second load
//
// This is not safe for concurrent use, and the fix is fp.Lazy from L4 — sync.Once
// gives you exactly this with the race removed. Work out where two goroutines
// calling Description at once would both trigger a load, then decide whether
// that matters for a read-only fetch. (Sometimes it genuinely does not. Saying
// why is the skill.)
func (p *LazyProductDetailProxy) Description() string {
	panic("TODO")
}

// Loaded reports whether the real object has been fetched yet.
func (p *LazyProductDetailProxy) Loaded() bool {
	panic("TODO")
}

var (
	_ = fmt.Sprintf
	_ = strings.Split
)
