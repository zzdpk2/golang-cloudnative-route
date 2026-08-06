package patterns

import (
	"reflect"
	"strings"
	"testing"
)

// ---- Behavioural ----

func TestChainOfResponsibility(t *testing.T) {
	build := func() OrderApprover {
		auto := &AutoApprover{}
		auto.SetNext(&SupervisorApprover{}).SetNext(&DirectorApprover{})
		return auto
	}

	tests := []struct {
		name   string
		amount int64
		want   string
	}{
		{"under $100 is automatic", 5_000, "auto-approved: 5000"},
		{"under $10k needs a supervisor", 50_000, "supervisor approved: 50000"},
		{"above that needs a director", 5_000_000, "director approved: 5000000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := build().Approve(tt.amount); got != tt.want {
				t.Errorf("Approve(%d) = %q, want %q", tt.amount, got, tt.want)
			}
		})
	}
}

// A chain that runs out of approvers must say so, not approve by default.
func TestChainOfResponsibility_FallsOffTheEnd(t *testing.T) {
	auto := &AutoApprover{}
	got := auto.Approve(5_000_000)
	if want := "escalated: no approver for 5000000"; got != want {
		t.Errorf("Approve past the end of the chain = %q, want %q", got, want)
	}
}

func TestCommand_ExecuteAndUndo(t *testing.T) {
	draft := &OrderDraft{}
	console := &OrderConsole{}

	if got := console.Run(NewAddLineCommand(draft, "sku-1")); got != "added sku-1" {
		t.Errorf("Run(add) = %q", got)
	}
	if got := console.Run(NewAddLineCommand(draft, "sku-2")); got != "added sku-2" {
		t.Errorf("Run(add) = %q", got)
	}
	if want := []string{"sku-1", "sku-2"}; !reflect.DeepEqual(draft.SKUs, want) {
		t.Fatalf("SKUs = %v, want %v", draft.SKUs, want)
	}

	// LIFO.
	if got := console.UndoLast(); got != "removed sku-2" {
		t.Errorf("UndoLast = %q, want %q", got, "removed sku-2")
	}
	if want := []string{"sku-1"}; !reflect.DeepEqual(draft.SKUs, want) {
		t.Errorf("SKUs after undo = %v, want %v", draft.SKUs, want)
	}
	if got := console.UndoLast(); got != "removed sku-1" {
		t.Errorf("UndoLast = %q, want %q", got, "removed sku-1")
	}
	if got := console.UndoLast(); got != "nothing to undo" {
		t.Errorf("UndoLast on an empty history = %q, want %q", got, "nothing to undo")
	}
}

func TestCommand_Cancel(t *testing.T) {
	draft := &OrderDraft{}
	console := &OrderConsole{}

	console.Run(NewCancelCommand(draft))
	if !draft.Cancelled {
		t.Error("the draft should be cancelled")
	}
	console.UndoLast()
	if draft.Cancelled {
		t.Error("undo should have un-cancelled the draft")
	}
}

func TestInterpreter(t *testing.T) {
	// base * qty + surcharge
	expr := AddExpr{
		Left:  MulExpr{Left: VarExpr{Name: "base"}, Right: VarExpr{Name: "qty"}},
		Right: LiteralExpr{Value: 500},
	}
	vars := map[string]int{"base": 1000, "qty": 3}

	if got, want := expr.Eval(vars), 3500; got != want {
		t.Errorf("Eval = %d, want %d", got, want)
	}
	if got := (VarExpr{Name: "missing"}).Eval(vars); got != 0 {
		t.Errorf("an undefined variable evaluated to %d, want 0", got)
	}
}

func TestIterator(t *testing.T) {
	it := NewSKUIterator([]string{"a", "b"})

	var got []string
	for it.HasNext() {
		got = append(got, it.Next())
	}
	if want := []string{"a", "b"}; !reflect.DeepEqual(got, want) {
		t.Errorf("iterated %v, want %v", got, want)
	}
	if it.HasNext() {
		t.Error("HasNext should be false once exhausted")
	}
	if got := it.Next(); got != "" {
		t.Errorf("Next past the end = %q, want the zero value", got)
	}
}

func TestMediator(t *testing.T) {
	desk := NewFulfillmentDesk()
	warehouse, payment := NewParty("warehouse"), NewParty("payment")
	desk.Register(warehouse)
	desk.Register(payment)

	warehouse.Send("payment", "reserved sku-1")

	inbox := payment.Inbox()
	if len(inbox) != 1 {
		t.Fatalf("payment inbox = %v, want one message", inbox)
	}
	if want := "warehouse: reserved sku-1"; inbox[0] != want {
		t.Errorf("message = %q, want %q", inbox[0], want)
	}
	if len(warehouse.Inbox()) != 0 {
		t.Error("the sender should not receive its own message")
	}

	// An unknown recipient is dropped rather than panicking.
	warehouse.Send("nobody", "hello")
}

func TestMemento(t *testing.T) {
	editor := &OrderEditor{}
	editor.AddSKU("a")
	snapshot := editor.Save()

	editor.AddSKU("b")
	if want := []string{"a", "b"}; !reflect.DeepEqual(editor.SKUs(), want) {
		t.Fatalf("SKUs = %v, want %v", editor.SKUs(), want)
	}

	editor.Restore(snapshot)
	if want := []string{"a"}; !reflect.DeepEqual(editor.SKUs(), want) {
		t.Errorf("after Restore = %v, want %v", editor.SKUs(), want)
	}
}

// The snapshot must not share storage with the editor.
func TestMemento_SnapshotIsIndependent(t *testing.T) {
	editor := &OrderEditor{}
	editor.AddSKU("a")
	snapshot := editor.Save()

	for i := range 10 {
		editor.AddSKU(string(rune('b' + i)))
	}

	editor.Restore(snapshot)
	if want := []string{"a"}; !reflect.DeepEqual(editor.SKUs(), want) {
		t.Errorf("after Restore = %v, want %v: the snapshot shared its backing "+
			"array with the editor, so later appends rewrote the past",
			editor.SKUs(), want)
	}
}

func TestObserver(t *testing.T) {
	subject := &OrderSubject{}
	a, b := &RecordingObserver{}, &RecordingObserver{}
	subject.Attach(a)
	subject.Attach(b)

	subject.Notify("order.created")
	subject.Notify("order.shipped")

	want := []string{"order.created", "order.shipped"}
	if !reflect.DeepEqual(a.Events, want) {
		t.Errorf("observer a saw %v, want %v", a.Events, want)
	}
	if !reflect.DeepEqual(b.Events, want) {
		t.Errorf("observer b saw %v, want %v", b.Events, want)
	}
}

func TestState(t *testing.T) {
	p := NewParcel()
	if got := p.StateName(); got != "new" {
		t.Fatalf("initial state = %q, want %q", got, "new")
	}

	if got := p.Ship(); got != "cannot ship an unpacked parcel" {
		t.Errorf("Ship from new = %q", got)
	}
	if got := p.Pack(); got != "packed" {
		t.Errorf("Pack = %q", got)
	}
	if got := p.StateName(); got != "packed" {
		t.Errorf("state after Pack = %q, want %q", got, "packed")
	}
	if got := p.Pack(); got != "already packed" {
		t.Errorf("Pack twice = %q", got)
	}
	if got := p.Ship(); got != "shipped" {
		t.Errorf("Ship = %q", got)
	}
	if got := p.StateName(); got != "shipped" {
		t.Errorf("state after Ship = %q, want %q", got, "shipped")
	}
	if got := p.Ship(); got != "already shipped" {
		t.Errorf("Ship twice = %q", got)
	}
}

func TestStrategy(t *testing.T) {
	input := []string{"c", "a", "b"}

	asc := NewLineSorter(ByNameAsc{}).Sort(input)
	if want := []string{"a", "b", "c"}; !reflect.DeepEqual(asc, want) {
		t.Errorf("ascending = %v, want %v", asc, want)
	}

	desc := NewLineSorter(ByNameDesc{}).Sort(input)
	if want := []string{"c", "b", "a"}; !reflect.DeepEqual(desc, want) {
		t.Errorf("descending = %v, want %v", desc, want)
	}

	if want := []string{"c", "a", "b"}; !reflect.DeepEqual(input, want) {
		t.Errorf("the caller's slice was reordered: %v, want %v", input, want)
	}
}

func TestTemplateMethod(t *testing.T) {
	got := NewFulfillmentRun(WarehouseFulfillment{}).Run()

	want := []string{
		"reserved from warehouse",
		"charged card",
		"dispatched via courier",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Run() = %v, want %v", got, want)
	}
}

func TestVisitor(t *testing.T) {
	visitor := PickListVisitor{}

	items := []OrderItem{
		PhysicalItem{SKU: "sku-1", Grams: 500},
		DigitalItem{SKU: "sku-2", URL: "https://example.test/d"},
	}

	var got []string
	for _, item := range items {
		got = append(got, item.Accept(visitor))
	}

	want := []string{"pick sku-1 (500g)", "no pick: sku-2 is digital"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("visited = %v, want %v", got, want)
	}
}

// ---- Creational ----

func TestSingleton(t *testing.T) {
	a, b := GetAppConfig(), GetAppConfig()
	if a == nil {
		t.Fatal("GetAppConfig returned nil")
	}
	if a != b {
		t.Error("GetAppConfig returned two different pointers")
	}
	if a.Name != "go-ddd-tdd" || a.Port != 8080 {
		t.Errorf("config = %+v, want Name=go-ddd-tdd Port=8080", *a)
	}
}

func TestFactoryMethod(t *testing.T) {
	card, err := NewPayment("card")
	if err != nil {
		t.Fatalf("NewPayment(card) = %v", err)
	}
	if got := card.Pay(1999); got != "paid 1999 by credit card" {
		t.Errorf("card.Pay = %q", got)
	}

	paypal, err := NewPayment("paypal")
	if err != nil {
		t.Fatalf("NewPayment(paypal) = %v", err)
	}
	if got := paypal.Pay(1999); got != "paid 1999 by paypal" {
		t.Errorf("paypal.Pay = %q", got)
	}

	unknown, err := NewPayment("bitcoin")
	if err == nil {
		t.Error("NewPayment(bitcoin) returned no error")
	}
	if unknown != nil {
		t.Error("an unknown method must return a nil Payment alongside the error")
	}
}

func TestAbstractFactory(t *testing.T) {
	domestic, err := NewCarrierKit("AU")
	if err != nil {
		t.Fatalf("NewCarrierKit(AU) = %v", err)
	}
	if got := domestic.NewLabel().Render("1 George St"); got != "AUPOST label: 1 George St" {
		t.Errorf("domestic label = %q", got)
	}
	if got := domestic.NewTracker().Track("t1"); got != "AUPOST tracking t1" {
		t.Errorf("domestic tracker = %q", got)
	}

	intl, err := NewCarrierKit("US")
	if err != nil {
		t.Fatalf("NewCarrierKit(US) = %v", err)
	}
	if got := intl.NewLabel().Render("1 George St"); !strings.Contains(got, "customs") {
		t.Errorf("international label = %q, want it to mention customs", got)
	}
	if got := intl.NewTracker().Track("t1"); got != "DHL tracking t1" {
		t.Errorf("international tracker = %q", got)
	}

	if _, err := NewCarrierKit(""); err == nil {
		t.Error(`NewCarrierKit("") returned no error`)
	}
}

func TestBuilder(t *testing.T) {
	q, err := NewOrderQueryBuilder().
		ForCustomer("c1").
		WithStatus("pending").
		OrderedBy("created_at").
		Limit(10).
		Build()
	if err != nil {
		t.Fatalf("Build() = %v", err)
	}

	want := OrderQuery{CustomerID: "c1", Status: "pending", OrderBy: "created_at", Limit: 10}
	if q != want {
		t.Errorf("query = %+v, want %+v", q, want)
	}
}

func TestBuilder_RejectsAnUnboundedQuery(t *testing.T) {
	if _, err := NewOrderQueryBuilder().Build(); err == nil {
		t.Error("a query with no filter at all was accepted; it would scan every order")
	}
}

// The error is parked at Limit and surfaces from Build, because a chaining
// method cannot return one.
func TestBuilder_DeferredError(t *testing.T) {
	_, err := NewOrderQueryBuilder().ForCustomer("c1").Limit(-1).Build()
	if err == nil {
		t.Error("Limit(-1) did not surface an error from Build")
	}
}

func TestPrototype_CloneIsDeep(t *testing.T) {
	original := &NotificationTemplate{
		Subject: "Order shipped",
		Body:    "Your order is on its way",
		Tags:    []string{"transactional"},
		Meta:    map[string]string{"locale": "en-AU"},
	}

	clone := original.Clone()
	if clone == original {
		t.Fatal("Clone returned the same pointer")
	}
	if clone.Subject != original.Subject || clone.Body != original.Body {
		t.Errorf("clone = %+v, want the same scalar fields", clone)
	}

	clone.Tags[0] = "marketing"
	if original.Tags[0] != "transactional" {
		t.Error("mutating the clone's Tags changed the original: the slice was shared")
	}

	clone.Meta["locale"] = "fr-FR"
	if original.Meta["locale"] != "en-AU" {
		t.Error("mutating the clone's Meta changed the original: the map was shared")
	}
}

func TestPrototype_NilStaysNil(t *testing.T) {
	clone := (&NotificationTemplate{Subject: "s"}).Clone()
	if clone.Tags != nil {
		t.Errorf("Tags = %v, want nil: a nil slice should not become an empty one", clone.Tags)
	}
	if clone.Meta != nil {
		t.Errorf("Meta = %v, want nil", clone.Meta)
	}
}

// ---- Structural ----

func TestAdapter(t *testing.T) {
	adapter := NewWarehouseAdapter(&LegacyWarehouse{})

	id, err := adapter.Reserve("sku-1", 2)
	if err != nil {
		t.Fatalf("Reserve = %v", err)
	}
	if id != "RES-sku-1" {
		t.Errorf("reservation id = %q, want %q", id, "RES-sku-1")
	}

	if _, err := adapter.Reserve("sku-1", 0); err == nil {
		t.Error(`an invalid quantity returned no error: the vendor's "ERR:" ` +
			"prefix has to become a Go error, not travel onward as a string")
	}
}

func TestBridge(t *testing.T) {
	tests := []struct {
		name  string
		label Label
		want  string
	}{
		{"standard on pdf", NewStandardLabel("1 George St", PDFPrinter{}), "[PDF] STANDARD 1 George St"},
		{"standard on zpl", NewStandardLabel("1 George St", ZPLPrinter{}), "^XA STANDARD 1 George St ^XZ"},
		{"express on pdf", NewExpressLabel("1 George St", PDFPrinter{}), "[PDF] EXPRESS 1 George St"},
		{"express on zpl", NewExpressLabel("1 George St", ZPLPrinter{}), "^XA EXPRESS 1 George St ^XZ"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.label.Emit(); got != tt.want {
				t.Errorf("Emit() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestComposite(t *testing.T) {
	spares := NewShipmentBundle("spares")
	spares.Add(NewShipmentItem("lid", 50))

	kit := NewShipmentBundle("kit")
	kit.Add(NewShipmentItem("mug", 300))
	kit.Add(spares)

	if got := kit.Grams(); got != 350 {
		t.Errorf("kit.Grams() = %d, want 350 (300 plus a nested bundle of 50)", got)
	}
	if got := kit.Name(); got != "kit" {
		t.Errorf("Name() = %q, want %q", got, "kit")
	}
	if got := NewShipmentBundle("empty").Grams(); got != 0 {
		t.Errorf("an empty bundle weighs %d, want 0", got)
	}
}

func TestDecorator(t *testing.T) {
	tests := []struct {
		name     string
		notifier Notifier
		want     string
	}{
		{"base only", EmailNotifier{}, "email: shipped"},
		{"one wrap", NewSMSDecorator(EmailNotifier{}), "email: shipped + sms: shipped"},
		{
			"two wraps",
			NewSlackDecorator(NewSMSDecorator(EmailNotifier{})),
			"email: shipped + sms: shipped + slack: shipped",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.notifier.Send("shipped"); got != tt.want {
				t.Errorf("Send() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFacade(t *testing.T) {
	got := NewCheckoutFacade().Checkout("cust-1", "sku-1", 2, 1999, "1 George St")

	want := []string{
		"reserved 2 of sku-1",
		"charged cust-1 1999",
		"shipped to 1 George St",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Checkout() = %v, want %v", got, want)
	}
}

func TestFlyweight_SharesInstances(t *testing.T) {
	pool := NewSKUSpecPool()

	a := pool.Get("mug", "kitchen", "fragile")
	b := pool.Get("mug", "kitchen", "fragile")
	if a != b {
		t.Error("two requests for the same spec returned different pointers; " +
			"sharing the instance is the entire pattern")
	}
	if got := pool.Count(); got != 1 {
		t.Errorf("Count() = %d, want 1", got)
	}

	pool.Get("lid", "kitchen", "fragile")
	if got := pool.Count(); got != 2 {
		t.Errorf("Count() after a second distinct spec = %d, want 2", got)
	}
}

func TestProxy_LoadsLazilyAndOnce(t *testing.T) {
	p := NewLazyProductDetailProxy("sku-1")

	if p.Loaded() {
		t.Error("the proxy loaded before anyone asked for the description")
	}

	first := p.Description()
	if first != "detail for sku-1" {
		t.Errorf("Description() = %q, want %q", first, "detail for sku-1")
	}
	if !p.Loaded() {
		t.Error("Loaded() is false after Description()")
	}
	if second := p.Description(); second != first {
		t.Errorf("second Description() = %q, want the cached %q", second, first)
	}
}
