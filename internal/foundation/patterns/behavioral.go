// Package patterns is the design-pattern reference.
//
// All 23 Gang of Four patterns are here, and every one is cast onto the order
// domain rather than onto light switches and chat rooms. Not because the order
// domain needs all 23 — it emphatically does not — but because a pattern you
// have only ever seen applied to a toy is a pattern you will not recognise in
// real code.
//
// # These are a reference, not a checklist
//
// Five of them already exist in the main line, done properly, against real
// requirements. Those five are kept **deliberately small here**, and each one
// points at its real counterpart:
//
//	Strategy     → domain/service.PricingStrategy
//	Observer     → adapter/eventbus
//	Iterator     → platform/fp/iter
//	State        → domain/aggregate.Order
//	Chain        → platform/minigin middleware chain
//
// Read both versions. The one here shows the skeleton; the one in the main line
// shows what the skeleton costs once it has to handle concurrency, invariants,
// and errors. That gap is the actual lesson.
//
// Several others — Interpreter, Memento, Flyweight, Visitor, Prototype — have
// **no place in this system at all**, and each says so in its own doc comment.
// Knowing a pattern and declining to use it is the skill. The most expensive
// mistake in this area is treating the catalogue as a list to be filled in.
//
// # This file: the 11 behavioural patterns
//
//	Chain of Responsibility, Command, Interpreter, Iterator, Mediator,
//	Memento, Observer, State, Strategy, Template Method, Visitor
package patterns

import (
	"fmt"
	"sort"
	"strings"
)

// ------------------------------------------------------------
// Chain of Responsibility
//
// COMPACT — the real one is the middleware chain in platform/minigin.
// ------------------------------------------------------------

// OrderApprover handles an order, or passes it along.
//
// The business rule: under $100 is auto-approved, up to $10,000 needs a
// supervisor, anything larger needs a director.
//
// The point is that no approver knows the whole ladder — each knows only its
// own limit and who comes next. Inserting a "VP" tier touches two lines and no
// existing decision.
//
// Compare with Context.Next in minigin: same shape, except that one also runs
// code *after* the downstream call, which is what makes the onion model
// possible. This version only passes forward. Work out what you would change to
// get an "after" phase, and why minigin needed one.
type OrderApprover interface {
	Approve(amountCents int64) string
	SetNext(next OrderApprover) OrderApprover
}

// BaseApprover holds the link to the next approver.
//
// Embedding it gives every concrete approver SetNext and passNext for free —
// Go's answer to an abstract base class.
type BaseApprover struct {
	next OrderApprover
}

// SetNext links the next approver and returns it, so a chain reads:
//
//	auto.SetNext(supervisor).SetNext(director)
func (h *BaseApprover) SetNext(next OrderApprover) OrderApprover {
	panic("TODO")
}

// passNext forwards to the next approver, or reports that nobody could handle it.
//
//	with a next approver → whatever that approver returns
//	at the end of chain  → "escalated: no approver for 5000000"
//
// Falling off the end of a chain is a real outcome and needs a real answer.
// Returning "" would make an unapprovable order look approved.
func (h *BaseApprover) passNext(amountCents int64) string {
	panic("TODO")
}

// AutoApprover approves anything under $100 (10000 cents).
//
//	Approve(5000)  → "auto-approved: 5000"
//	Approve(50000) → passes to the next approver
type AutoApprover struct {
	BaseApprover
}

// Approve settles anything under 10000 cents and passes the rest on.
func (h *AutoApprover) Approve(amountCents int64) string {
	panic("TODO")
}

// SupervisorApprover approves up to $10,000 (1000000 cents).
//
//	Approve(50000)   → "supervisor approved: 50000"
//	Approve(5000000) → passes on
type SupervisorApprover struct {
	BaseApprover
}

// Approve settles up to 1000000 cents and passes the rest on.
func (h *SupervisorApprover) Approve(amountCents int64) string {
	panic("TODO")
}

// DirectorApprover approves anything.
//
//	Approve(5000000) → "director approved: 5000000"
type DirectorApprover struct {
	BaseApprover
}

// Approve settles everything; nothing is passed on from here.
func (h *DirectorApprover) Approve(amountCents int64) string {
	panic("TODO")
}

// ------------------------------------------------------------
// Command
// ------------------------------------------------------------

// OrderCommand is an operation packaged as an object, so it can be logged,
// queued, retried, and undone.
//
// This is the pattern behind every undo you have used, and behind job queues
// generally. It is also the shape of a txdsl step from L9: an action paired
// with its compensation. Having built both, notice that a saga is a list of
// Commands with the undo made mandatory.
type OrderCommand interface {
	Execute() string
	Undo() string
}

// OrderDraft is the receiver the commands act on.
type OrderDraft struct {
	SKUs      []string
	Cancelled bool
}

// AddLineCommand adds a SKU to the draft.
type AddLineCommand struct {
	draft *OrderDraft
	sku   string
}

func NewAddLineCommand(draft *OrderDraft, sku string) *AddLineCommand { panic("TODO") }

// Execute appends the SKU.
//
//	Execute() → "added sku-1"
func (c *AddLineCommand) Execute() string {
	panic("TODO")
}

// Undo removes the SKU this command added.
//
//	Undo() → "removed sku-1"
//
// Remove only what *this* command added, and only once. "Drop the last element"
// works until two commands are undone out of order — which the console below
// permits.
func (c *AddLineCommand) Undo() string {
	panic("TODO")
}

// CancelCommand cancels the draft.
type CancelCommand struct {
	draft *OrderDraft
}

func NewCancelCommand(draft *OrderDraft) *CancelCommand { panic("TODO") }

// Execute marks the draft cancelled.
//
//	Execute() → "cancelled"
func (c *CancelCommand) Execute() string {
	panic("TODO")
}

// Undo un-cancels it.
//
//	Undo() → "uncancelled"
func (c *CancelCommand) Undo() string {
	panic("TODO")
}

// OrderConsole runs commands and remembers them so they can be undone.
type OrderConsole struct {
	history []OrderCommand
}

// Run executes a command and records it.
//
//	Run(addSKU1) → "added sku-1"
func (r *OrderConsole) Run(command OrderCommand) string {
	panic("TODO")
}

// UndoLast undoes the most recent command and forgets it.
//
//	after Run(a), Run(b): UndoLast() → b.Undo(), then UndoLast() → a.Undo()
//	with nothing to undo: UndoLast() → "nothing to undo"
//
// Last-in-first-out, for the same reason defers unwind that way: later
// operations may depend on earlier ones, so they have to come off first.
func (r *OrderConsole) UndoLast() string {
	panic("TODO")
}

// ------------------------------------------------------------
// Interpreter
//
// NOT USED IN THIS SYSTEM — see the note below.
// ------------------------------------------------------------

// PriceExpr is a node in a tiny language for promotion formulas such as
// "base * qty + surcharge".
//
// Interpreter builds a syntax tree from small node types, each of which knows
// how to evaluate itself. It is genuinely right when *users* need to write
// rules: a promotion engine with a formula editor, a search query language, a
// spreadsheet.
//
// **It is wrong here.** The promotion engine in L15.1 takes rules as Go values
// configured by developers, and a Go value is already a better language than
// anything you would build in an afternoon — typed, tooled, debuggable. Reach
// for Interpreter only when the rules must come from outside the codebase, and
// then reach for a real parser, not this.
//
// Build it anyway so you can recognise it, then notice how much machinery it
// took for arithmetic Go does in one line.
type PriceExpr interface {
	Eval(vars map[string]int) int
}

// LiteralExpr is a constant.
//
//	LiteralExpr{Value: 42}.Eval(nil) → 42
type LiteralExpr struct {
	Value int
}

// Eval returns the constant, ignoring the environment entirely.
func (e LiteralExpr) Eval(vars map[string]int) int {
	panic("TODO")
}

// VarExpr looks a name up in the environment.
//
//	VarExpr{Name: "qty"}.Eval(map[string]int{"qty": 3}) → 3
//	VarExpr{Name: "nope"}.Eval(map[string]int{})        → 0
//
// An undefined variable evaluating to 0 is a decision, and a questionable one:
// it turns a typo into a silently wrong price. Note what the interface would
// have to look like to report it instead.
type VarExpr struct {
	Name string
}

// Eval looks the name up in vars, yielding 0 when it is absent.
func (e VarExpr) Eval(vars map[string]int) int {
	panic("TODO")
}

// AddExpr sums two sub-expressions.
//
//	AddExpr{LiteralExpr{2}, LiteralExpr{3}}.Eval(nil) → 5
type AddExpr struct {
	Left  PriceExpr
	Right PriceExpr
}

// Eval evaluates both sides and adds them.
func (e AddExpr) Eval(vars map[string]int) int {
	panic("TODO")
}

// MulExpr multiplies two sub-expressions.
//
//	MulExpr{VarExpr{"base"}, VarExpr{"qty"}} with base=100, qty=3 → 300
//
// Composition is the trick: because both sides are PriceExpr, a multiplication
// can contain an addition can contain a variable, to any depth, with no node
// knowing anything about the others.
type MulExpr struct {
	Left  PriceExpr
	Right PriceExpr
}

// Eval evaluates both sides and multiplies them.
func (e MulExpr) Eval(vars map[string]int) int {
	panic("TODO")
}

// ------------------------------------------------------------
// Iterator
//
// COMPACT — the real one is platform/fp/iter.
// ------------------------------------------------------------

// SKUIterator walks a list of SKUs.
//
// The classic HasNext/Next shape, here only so you can compare it with
// platform/fp/iter, which solves the same problem with a closure:
//
//	this      HasNext() then Next()   external state, two calls, eager
//	fp/iter   seq() → (T, bool)       one call, lazy
//
// The fp version can express an infinite sequence and this one cannot, which is
// the concrete reason to prefer it. Go 1.23's range-over-func settled the
// argument for the whole language — look at what iter.Seq chose, and why.
type SKUIterator interface {
	HasNext() bool
	Next() string
}

type sliceSKUIterator struct {
	items []string
	pos   int
}

// NewSKUIterator returns an iterator over items.
func NewSKUIterator(items []string) SKUIterator { panic("TODO") }

// HasNext reports whether Next would return a real value.
func (it *sliceSKUIterator) HasNext() bool {
	panic("TODO")
}

// Next returns the next SKU and advances.
//
//	over ["a","b"]: Next() → "a", Next() → "b", Next() → ""
//
// Calling Next past the end returns the zero value rather than panicking.
// Decide whether you agree: a silent "" is how an off-by-one becomes a blank
// line on an invoice.
func (it *sliceSKUIterator) Next() string {
	panic("TODO")
}

// ------------------------------------------------------------
// Mediator
// ------------------------------------------------------------

// FulfillmentDesk coordinates the parties in a fulfillment so none of them has
// to know about the others.
//
// Without a mediator, warehouse talks to payment talks to carrier, and you get
// n² relationships that all need rewiring when one changes. With one, each
// party knows only the desk.
//
// The cost is that the desk accumulates every interaction in the system and
// becomes the file nobody wants to open. That failure mode has a name — the god
// object — and it is why the main line uses an *event bus* instead: with
// events, the coordinator does not need to know who is listening.
//
// Build this, then look at adapter/eventbus and name the difference precisely.
type FulfillmentDesk struct {
	parties map[string]*Party
}

func NewFulfillmentDesk() *FulfillmentDesk { panic("TODO") }

// Register adds a party the desk can route to, and points the party back at
// the desk.
func (d *FulfillmentDesk) Register(p *Party) {
	panic("TODO")
}

// Send routes a message from one party to another.
//
//	Send("warehouse", "payment", "reserved sku-1")
//	  → payment's inbox gains "warehouse: reserved sku-1"
//
// A message to an unregistered party is dropped rather than erroring. Decide
// whether that is right, and what a real system would do instead.
func (d *FulfillmentDesk) Send(from, to, message string) {
	panic("TODO")
}

// Party is one participant. It knows the desk and nothing else.
type Party struct {
	Name string
	desk *FulfillmentDesk
	msgs []string
}

func NewParty(name string) *Party { panic("TODO") }

// Send asks the desk to deliver a message.
//
//	warehouse.Send("payment", "reserved sku-1")
func (p *Party) Send(to, message string) {
	panic("TODO")
}

// receive records an incoming message.
func (p *Party) receive(from, message string) {
	panic("TODO")
}

// Inbox returns the messages received, oldest first.
//
// Return a copy — by now this question should be automatic.
func (p *Party) Inbox() []string {
	panic("TODO")
}

// ------------------------------------------------------------
// Memento
//
// NOT USED IN THIS SYSTEM — see the note below.
// ------------------------------------------------------------

// DraftSnapshot is an opaque saved state of an OrderEditor.
//
// Its field is unexported on purpose: the caller can hold a snapshot and hand
// it back without being able to read or forge it. Encapsulation survives the
// round trip, which is the entire idea.
type DraftSnapshot struct {
	skus []string
}

// OrderEditor is a draft order being edited, with undo.
//
// **This system does not use Memento.** The order aggregate protects its state
// through methods and records domain events; undo there means issuing a
// compensating command, not restoring a blob. Event sourcing in L15.5 gets the
// same capability for free, because the event log *is* the history.
//
// Memento earns its place where state is large, opaque, and has no meaningful
// intermediate operations — a canvas, a text buffer. An order has meaningful
// operations, so it gets commands instead.
type OrderEditor struct {
	skus []string
}

// AddSKU appends to the draft.
func (e *OrderEditor) AddSKU(sku string) {
	panic("TODO")
}

// SKUs returns the current draft contents.
func (e *OrderEditor) SKUs() []string {
	panic("TODO")
}

// Save captures the current state.
//
//	AddSKU("a"); s := Save(); AddSKU("b"); Restore(s) → back to ["a"]
//
// The snapshot must be independent of the editor. Storing the live slice lets a
// later append reach through and change the past, which defeats the pattern —
// the same aliasing trap as Order.Lines in L2.
func (e *OrderEditor) Save() DraftSnapshot {
	panic("TODO")
}

// Restore rewinds to a saved state.
func (e *OrderEditor) Restore(s DraftSnapshot) {
	panic("TODO")
}

// ------------------------------------------------------------
// Observer
//
// COMPACT — the real one is adapter/eventbus.
// ------------------------------------------------------------

// OrderObserver is notified when something happens to an order.
type OrderObserver interface {
	OnEvent(event string)
}

// OrderSubject notifies its observers.
//
// Deliberately minimal, because the real version in adapter/eventbus has to
// answer the questions this one ducks: what happens when an observer is slow,
// when one panics, when delivery must survive a restart, when the same observer
// subscribes twice.
//
// This version notifies synchronously and in order, so a slow observer blocks
// everyone and a panicking one takes the publisher down. Write it, then list
// what you would add to make it safe. That list is L8.
type OrderSubject struct {
	observers []OrderObserver
}

// Attach registers an observer.
func (s *OrderSubject) Attach(observer OrderObserver) {
	panic("TODO")
}

// Notify delivers an event to every observer, in registration order.
func (s *OrderSubject) Notify(event string) {
	panic("TODO")
}

// RecordingObserver remembers what it was told.
type RecordingObserver struct {
	Events []string
}

// OnEvent appends the event to Events, oldest first.
func (o *RecordingObserver) OnEvent(event string) {
	panic("TODO")
}

// ------------------------------------------------------------
// State
//
// COMPACT — the real one is domain/aggregate.Order.
// ------------------------------------------------------------

// ParcelState is one state of a parcel, and knows which transitions it allows.
//
// The pattern replaces a switch on a status field with one type per state, so
// each state answers for itself. It shines when the states behave genuinely
// differently and the transition table is large.
//
// aggregate.Order does *not* use it — it uses a status enum with explicit
// guards, because its transitions also record events, stamp timestamps, and
// hold a mutex, and spreading that across five types would hide the state
// machine rather than clarify it.
//
// Build this, then re-read Order.Confirm/Ship/Deliver/Cancel and decide which
// you would rather maintain. There is a real argument for both; what matters is
// that you can make it.
type ParcelState interface {
	Pack(p *Parcel) string
	Ship(p *Parcel) string
	Name() string
}

// Parcel delegates its behaviour to its current state.
type Parcel struct {
	state ParcelState
}

// NewParcel starts a parcel in NewParcelState.
func NewParcel() *Parcel {
	panic("TODO")
}

// Pack asks the current state to handle packing.
func (p *Parcel) Pack() string {
	panic("TODO")
}

// Ship asks the current state to handle shipping.
func (p *Parcel) Ship() string {
	panic("TODO")
}

// StateName reports the current state: "new", "packed", or "shipped".
func (p *Parcel) StateName() string {
	panic("TODO")
}

// NewParcelState is a parcel that has not been packed.
//
//	Pack() → "packed", and the parcel moves to PackedState
//	Ship() → "cannot ship an unpacked parcel"
type NewParcelState struct{}

// Pack packs the parcel and moves it to PackedState.
func (s NewParcelState) Pack(p *Parcel) string {
	panic("TODO")
}

// Ship refuses: an unpacked parcel cannot go out.
func (s NewParcelState) Ship(p *Parcel) string {
	panic("TODO")
}

// Name reports "new".
func (s NewParcelState) Name() string {
	panic("TODO")
}

// PackedState is a parcel ready to go.
//
//	Pack() → "already packed"
//	Ship() → "shipped", and the parcel moves to ShippedState
type PackedState struct{}

// Pack reports that it is already packed, and changes nothing.
func (s PackedState) Pack(p *Parcel) string {
	panic("TODO")
}

// Ship ships it and moves the parcel to ShippedState.
func (s PackedState) Ship(p *Parcel) string {
	panic("TODO")
}

// Name reports "packed".
func (s PackedState) Name() string {
	panic("TODO")
}

// ShippedState is terminal.
//
//	Pack() → "already shipped"
//	Ship() → "already shipped"
type ShippedState struct{}

// Pack reports that it has already shipped.
func (s ShippedState) Pack(p *Parcel) string {
	panic("TODO")
}

// Ship reports that it has already shipped; this state is terminal.
func (s ShippedState) Ship(p *Parcel) string {
	panic("TODO")
}

// Name reports "shipped".
func (s ShippedState) Name() string {
	panic("TODO")
}

// ------------------------------------------------------------
// Strategy
//
// COMPACT — the real one is domain/service.PricingStrategy.
// ------------------------------------------------------------

// LineSortStrategy orders the lines of an order.
//
// Strategy at its smallest: an interchangeable algorithm behind an interface.
// The real one in domain/service exists because the *business* has
// interchangeable pricing rules; this one exists so you can see the shape with
// nothing else in the way.
//
// In Go a one-method interface and a function type are near-equivalent —
// fp.Predicate and specification.SpecFunc are the same idea written as
// functions. The interface wins when the strategy needs a name for logging, or
// a second method later. Note that this one has both.
type LineSortStrategy interface {
	Sort(skus []string) []string
	Name() string
}

// ByNameAsc sorts alphabetically.
//
//	Sort(["c","a","b"]) → ["a","b","c"]
//
// Do not mutate the caller's slice. Third time this has come up —
// collection.SortBy and functional.SortBy made the same promise, and all three
// should behave identically.
type ByNameAsc struct{}

// Sort returns a new slice ordered A to Z, leaving the input untouched.
func (s ByNameAsc) Sort(skus []string) []string {
	panic("TODO")
}

func (s ByNameAsc) Name() string { panic("TODO") }

// ByNameDesc sorts in reverse.
//
//	Sort(["a","c","b"]) → ["c","b","a"]
type ByNameDesc struct{}

// Sort returns a new slice ordered Z to A, leaving the input untouched.
func (s ByNameDesc) Sort(skus []string) []string {
	panic("TODO")
}

func (s ByNameDesc) Name() string { panic("TODO") }

// LineSorter applies whichever strategy it was given.
type LineSorter struct {
	strategy LineSortStrategy
}

func NewLineSorter(strategy LineSortStrategy) *LineSorter { panic("TODO") }

// Sort delegates to the strategy.
//
// Decide what a nil strategy does. service.NoDiscount showed the null-object
// answer; a nil check here is the other. One of them moves the problem to
// construction time, where it belongs.
func (s *LineSorter) Sort(skus []string) []string {
	panic("TODO")
}

// ------------------------------------------------------------
// Template Method
// ------------------------------------------------------------

// FulfillmentSteps are the parts of a fulfillment that vary by channel.
//
// Template Method fixes the *sequence* and lets the steps vary. Fulfilling an
// order is always reserve → charge → dispatch; what each step means differs
// between a warehouse shipment and a digital download.
//
// In Go this is composition, not inheritance: the template holds an interface
// rather than being subclassed. A real improvement — the steps are testable on
// their own and there is no fragile base class.
//
// Compare with the saga in L10, which also fixes a sequence. The difference is
// that a saga's steps can *fail and be compensated*, and this one assumes they
// succeed. That difference is most of what makes distributed systems hard.
type FulfillmentSteps interface {
	Reserve() string
	Charge() string
	Dispatch() string
}

// FulfillmentRun executes the fixed sequence.
type FulfillmentRun struct {
	steps FulfillmentSteps
}

func NewFulfillmentRun(steps FulfillmentSteps) *FulfillmentRun { panic("TODO") }

// Run performs the three steps in order and returns what each reported.
//
//	Run() → ["reserved from warehouse", "charged card", "dispatched via courier"]
//
// The sequence belongs to the template. A step must not be able to reorder or
// skip another — that guarantee is what the pattern is selling.
func (t *FulfillmentRun) Run() []string {
	panic("TODO")
}

// WarehouseFulfillment ships physical goods.
type WarehouseFulfillment struct{}

// Reserve returns "reserved from warehouse".
func (f WarehouseFulfillment) Reserve() string {
	panic("TODO")
}

// Charge returns "charged card".
func (f WarehouseFulfillment) Charge() string {
	panic("TODO")
}

// Dispatch returns "dispatched via courier".
func (f WarehouseFulfillment) Dispatch() string {
	panic("TODO")
}

// ------------------------------------------------------------
// Visitor
//
// NOT USED IN THIS SYSTEM — see the note below.
// ------------------------------------------------------------

// OrderItem is something a visitor can operate on.
//
// Visitor lets you add a new *operation* over a fixed set of types without
// touching those types. The price is exactly inverted: adding a new *type*
// means changing every visitor.
//
// So it fits when the types are stable and the operations keep multiplying — a
// compiler's AST is the classic case. It does not fit here: the order model
// gains new line kinds far more often than it gains new renderers, so Visitor
// would make the common change expensive to buy a discount on the rare one.
//
// It is also unusually awkward in Go. Without generic methods the double
// dispatch has to be hand-written on every type, which is the boilerplate you
// are about to feel.
type OrderItem interface {
	Accept(visitor OrderItemVisitor) string
}

// OrderItemVisitor is one operation over the item types.
type OrderItemVisitor interface {
	VisitPhysical(p PhysicalItem) string
	VisitDigital(d DigitalItem) string
}

// PhysicalItem is a shipped good.
type PhysicalItem struct {
	SKU   string
	Grams int
}

// Accept performs the double dispatch: the item picks the visitor method.
//
// This one line is the whole mechanism, and it must be written on every item
// type. Notice that it cannot be factored out — that is the boilerplate the
// note above warned about.
func (p PhysicalItem) Accept(visitor OrderItemVisitor) string {
	panic("TODO")
}

// DigitalItem is a download.
type DigitalItem struct {
	SKU string
	URL string
}

// Accept dispatches to the visitor's digital-item method.
func (d DigitalItem) Accept(visitor OrderItemVisitor) string {
	panic("TODO")
}

// PickListVisitor renders items for a warehouse picking list.
//
//	PhysicalItem{SKU: "sku-1", Grams: 500} → "pick sku-1 (500g)"
//	DigitalItem{SKU: "sku-2"}              → "no pick: sku-2 is digital"
type PickListVisitor struct{}

// VisitPhysical renders "pick <sku> (<grams>g)".
func (v PickListVisitor) VisitPhysical(p PhysicalItem) string {
	panic("TODO")
}

// VisitDigital renders "no pick: <sku> is digital" — nothing to fetch.
func (v PickListVisitor) VisitDigital(d DigitalItem) string {
	panic("TODO")
}

var (
	_ = fmt.Sprintf
	_ = sort.Strings
	_ = strings.Join
)
