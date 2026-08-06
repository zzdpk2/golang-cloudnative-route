package datalayer

import (
	"sync"
	"testing"
	"time"

	"github.com/rex/go-ddd-tdd/internal/inference/routing"
)

func TestStoreRejectsInvalidConstruction(t *testing.T) {
	t.Parallel()

	if _, err := NewStore(0, fixedClock{}); err == nil {
		t.Fatal("NewStore accepted a zero freshness window")
	}
	if _, err := NewStore(time.Second, nil); err == nil {
		t.Fatal("NewStore accepted a nil clock")
	}
}

func TestStoreAppliesOnlyNewerEndpointState(t *testing.T) {
	t.Parallel()

	clock := &manualClock{now: time.Unix(100, 0)}
	store := mustStore(t, time.Minute, clock)
	newest := updateAt("model-a", 2, clock.Now(), 0.4)
	if changed, err := store.Apply(newest); err != nil || !changed {
		t.Fatalf("new update: changed=%v err=%v", changed, err)
	}
	older := updateAt("model-a", 1, clock.Now().Add(time.Second), 0.9)
	if changed, err := store.Apply(older); err != nil || changed {
		t.Fatalf("out-of-order update: changed=%v err=%v", changed, err)
	}

	snapshot := store.Snapshot()
	if len(snapshot.Fresh) != 1 || snapshot.Fresh[0].KVCacheUtilization != 0.4 {
		t.Fatalf("snapshot = %+v, want version 2 state", snapshot)
	}
}

func TestStoreDeleteTombstonePreventsResurrection(t *testing.T) {
	t.Parallel()

	clock := &manualClock{now: time.Unix(200, 0)}
	store := mustStore(t, time.Minute, clock)
	_, _ = store.Apply(updateAt("model-a", 3, clock.Now(), 0.2))
	changed, err := store.Apply(Update{
		EndpointID: "model-a", Version: 4, ObservedAt: clock.Now(), Deleted: true,
	})
	if err != nil || !changed {
		t.Fatalf("delete: changed=%v err=%v", changed, err)
	}
	if changed, err = store.Apply(updateAt("model-a", 3, clock.Now(), 0.1)); err != nil || changed {
		t.Fatalf("resurrection: changed=%v err=%v", changed, err)
	}
	if got := store.Snapshot().Fresh; len(got) != 0 {
		t.Fatalf("deleted Endpoint returned: %+v", got)
	}
}

func TestStoreSeparatesStaleEndpointsFromRoutingInput(t *testing.T) {
	t.Parallel()

	clock := &manualClock{now: time.Unix(300, 0)}
	store := mustStore(t, 10*time.Second, clock)
	_, _ = store.Apply(updateAt("model-a", 1, clock.Now(), 0.2))
	clock.Advance(11 * time.Second)

	snapshot := store.Snapshot()
	if len(snapshot.Fresh) != 0 {
		t.Fatalf("stale Endpoint remained routable: %+v", snapshot.Fresh)
	}
	if len(snapshot.StaleIDs) != 1 || snapshot.StaleIDs[0] != "model-a" {
		t.Fatalf("stale IDs = %v, want model-a", snapshot.StaleIDs)
	}
}

func TestStoreSnapshotIsDefensiveAndDeterministic(t *testing.T) {
	t.Parallel()

	clock := &manualClock{now: time.Unix(400, 0)}
	store := mustStore(t, time.Minute, clock)
	_, _ = store.Apply(updateAt("model-b", 1, clock.Now(), 0.2))
	_, _ = store.Apply(updateAt("model-a", 1, clock.Now(), 0.3))

	first := store.Snapshot()
	if first.Fresh[0].ID != "model-a" || first.Fresh[1].ID != "model-b" {
		t.Fatalf("Endpoint order = %v, want deterministic ID order", first.Fresh)
	}
	first.Fresh[0].Models[0] = "mutated"
	if got := store.Snapshot().Fresh[0].Models[0]; got != "tiny-llm" {
		t.Fatalf("snapshot leaked a mutable model slice: %q", got)
	}
}

func TestStoreConcurrentApplyAndSnapshot(t *testing.T) {
	t.Parallel()

	clock := &manualClock{now: time.Unix(500, 0)}
	store := mustStore(t, time.Minute, clock)
	var wait sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wait.Add(1)
		go func(worker int) {
			defer wait.Done()
			for version := 1; version <= 100; version++ {
				_, _ = store.Apply(updateAt(
					"model-"+string(rune('a'+worker)),
					uint64(version),
					clock.Now(),
					0.2,
				))
				_ = store.Snapshot()
			}
		}(worker)
	}
	wait.Wait()
}

func TestStoreConcurrentConflictingVersionsConvergeOnNewestTombstone(t *testing.T) {
	t.Parallel()
	clock := &manualClock{now: time.Unix(600, 0)}
	store := mustStore(t, time.Minute, clock)
	var wait sync.WaitGroup
	for version := 1; version <= 100; version++ {
		wait.Add(1)
		go func(version int) {
			defer wait.Done()
			update := updateAt("shared", uint64(version), clock.Now(), float64(version)/100)
			if version == 100 {
				update.Deleted = true
			}
			_, _ = store.Apply(update)
			_ = store.Snapshot()
		}(version)
	}
	wait.Wait()
	if snapshot := store.Snapshot(); len(snapshot.Fresh) != 0 {
		t.Fatalf("newest tombstone lost during conflicting updates: %+v", snapshot)
	}
}

func mustStore(t *testing.T, maxAge time.Duration, clock Clock) *Store {
	t.Helper()
	store, err := NewStore(maxAge, clock)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func updateAt(id string, version uint64, observedAt time.Time, kv float64) Update {
	return Update{
		EndpointID: id,
		Endpoint: routing.Endpoint{
			ID: id, URL: "http://" + id + ".example", Models: []string{"tiny-llm"},
			Ready: true, KVCacheUtilization: kv,
		},
		Version: version, ObservedAt: observedAt,
	}
}

type fixedClock struct{}

func (fixedClock) Now() time.Time { return time.Unix(1, 0) }

type manualClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *manualClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *manualClock) Advance(duration time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(duration)
	c.mu.Unlock()
}
