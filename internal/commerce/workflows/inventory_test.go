package workflows

import (
	"sync"
	"testing"
	"time"

	"github.com/asynkron/protoactor-go/actor"
)

const askTimeout = 2 * time.Second

func newSystem(t *testing.T) *actor.ActorSystem {
	t.Helper()
	system := actor.NewActorSystem()
	t.Cleanup(func() { system.Shutdown() })
	return system
}

func TestInventoryActor_ReportsInitialStock(t *testing.T) {
	system := newSystem(t)
	pid := system.Root.Spawn(NewInventoryProps(10))

	got, err := AskAvailable(system, pid, askTimeout)
	if err != nil {
		t.Fatalf("AskAvailable() = %v", err)
	}
	if got != 10 {
		t.Errorf("Available = %d, want 10", got)
	}
}

func TestInventoryActor_ReserveReducesStock(t *testing.T) {
	system := newSystem(t)
	pid := system.Root.Spawn(NewInventoryProps(10))

	res, err := AskReserve(system, pid, 4, askTimeout)
	if err != nil {
		t.Fatalf("AskReserve() = %v", err)
	}
	if !res.OK {
		t.Fatal("Reserve(4 of 10) was rejected")
	}
	if res.Available != 6 {
		t.Errorf("Available after reserve = %d, want 6", res.Available)
	}
}

func TestInventoryActor_RejectsOversell(t *testing.T) {
	system := newSystem(t)
	pid := system.Root.Spawn(NewInventoryProps(3))

	res, err := AskReserve(system, pid, 5, askTimeout)
	if err != nil {
		t.Fatalf("AskReserve() = %v", err)
	}
	if res.OK {
		t.Error("Reserve(5 of 3) was accepted, want rejected")
	}
	if res.Available != 3 {
		t.Errorf("Available = %d, want 3: a rejected reserve must not change stock", res.Available)
	}
}

func TestInventoryActor_ReleaseRestoresStock(t *testing.T) {
	system := newSystem(t)
	pid := system.Root.Spawn(NewInventoryProps(10))

	if _, err := AskReserve(system, pid, 6, askTimeout); err != nil {
		t.Fatalf("AskReserve() = %v", err)
	}
	system.Root.Send(pid, &Release{Qty: 2})

	got, err := AskAvailable(system, pid, askTimeout)
	if err != nil {
		t.Fatalf("AskAvailable() = %v", err)
	}
	if got != 6 {
		t.Errorf("Available = %d, want 6", got)
	}
}

// This is the exercise's centrepiece. A thousand goroutines hammer one SKU with
// no lock anywhere in the actor, and exactly the available units are sold.
//
// Run it with -race. There is nothing for the detector to find, and that is the
// point: the messages are serialised, so the state was never shared.
func TestInventoryActor_NoOversellUnderConcurrency(t *testing.T) {
	system := newSystem(t)

	const stock = 100
	const attempts = 1000
	pid := system.Root.Spawn(NewInventoryProps(stock))

	var mu sync.Mutex
	granted := 0

	var wg sync.WaitGroup
	for range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := AskReserve(system, pid, 1, askTimeout)
			if err != nil {
				return
			}
			if res.OK {
				mu.Lock()
				granted++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if granted != stock {
		t.Errorf("granted %d reservations, want exactly %d", granted, stock)
	}

	left, err := AskAvailable(system, pid, askTimeout)
	if err != nil {
		t.Fatalf("AskAvailable() = %v", err)
	}
	if left != 0 {
		t.Errorf("Available = %d, want 0", left)
	}
}

// Messages from one sender are processed in the order they were sent.
func TestInventoryActor_ProcessesMessagesInOrder(t *testing.T) {
	system := newSystem(t)
	pid := system.Root.Spawn(NewInventoryProps(10))

	for range 5 {
		system.Root.Send(pid, &Reserve{Qty: 1})
	}

	got, err := AskAvailable(system, pid, askTimeout)
	if err != nil {
		t.Fatalf("AskAvailable() = %v", err)
	}
	if got != 5 {
		t.Errorf("Available = %d, want 5", got)
	}
}

// A supervised actor survives a panic: the system restarts it and it keeps
// serving requests.
//
// Look closely at what the stock level is afterwards, and decide whether that
// is acceptable for inventory. The answer shapes where state belongs in a real
// actor system.
func TestInventoryActor_RestartsOnPanic(t *testing.T) {
	system := newSystem(t)
	pid := system.Root.Spawn(NewSupervisedInventoryProps(10))

	if _, err := AskAvailable(system, pid, askTimeout); err != nil {
		t.Fatalf("AskAvailable() before panic = %v", err)
	}

	system.Root.Send(pid, &PanicOn{})

	// The actor must still answer after being restarted.
	if _, err := AskAvailable(system, pid, askTimeout); err != nil {
		t.Fatalf("AskAvailable() after panic = %v: the actor did not come back", err)
	}
}
