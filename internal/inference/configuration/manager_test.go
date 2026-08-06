package configuration

import (
	"errors"
	"strconv"
	"sync"
	"testing"
)

func TestManagerRejectsInvalidInitialConfig(t *testing.T) {
	t.Parallel()

	invalid := validConfig(1)
	invalid.MaxInFlight = 0
	if _, err := NewManager(invalid); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("error = %v, want ErrInvalidConfig", err)
	}
}

func TestPublishFailurePreservesLastKnownGood(t *testing.T) {
	t.Parallel()

	manager := mustManager(t, validConfig(1))
	invalid := validConfig(2)
	invalid.Weights["queue"] = -1
	if err := manager.Publish(invalid); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("error = %v, want ErrInvalidConfig", err)
	}
	if got := manager.Active().Version; got != 1 {
		t.Fatalf("active version = %d, want last-known-good 1", got)
	}
}

func TestPublishRequiresMonotonicVersionAndDefensiveCopies(t *testing.T) {
	t.Parallel()

	manager := mustManager(t, validConfig(1))
	next := validConfig(2)
	if err := manager.Publish(next); err != nil {
		t.Fatal(err)
	}
	next.Weights["queue"] = 999
	if got := manager.Active().Weights["queue"]; got == 999 {
		t.Fatal("Publish retained the caller's mutable map")
	}
	active := manager.Active()
	active.Weights["queue"] = 888
	if got := manager.Active().Weights["queue"]; got == 888 {
		t.Fatal("Active leaked the Manager's mutable map")
	}
	if err := manager.Publish(validConfig(2)); !errors.Is(err, ErrStaleVersion) {
		t.Fatalf("error = %v, want ErrStaleVersion", err)
	}
}

func TestRollbackRepublishesKnownCompleteVersion(t *testing.T) {
	t.Parallel()

	manager := mustManager(t, validConfig(1))
	second := validConfig(2)
	second.SchedulerName = "canary"
	if err := manager.Publish(second); err != nil {
		t.Fatal(err)
	}
	if err := manager.Rollback(1); err != nil {
		t.Fatal(err)
	}
	active := manager.Active()
	if active.Version != 1 || active.SchedulerName != "profile-1" {
		t.Fatalf("active after rollback = %+v", active)
	}
	if err := manager.Rollback(99); !errors.Is(err, ErrUnknownVersion) {
		t.Fatalf("error = %v, want ErrUnknownVersion", err)
	}
}

func TestConcurrentReadersNeverObserveMixedConfig(t *testing.T) {
	t.Parallel()

	manager := mustManager(t, validConfig(1))
	var wait sync.WaitGroup
	for reader := 0; reader < 8; reader++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 1000 {
				active := manager.Active()
				if active.SchedulerName != schedulerName(active.Version) ||
					active.MaxInFlight != int(active.Version)*10 {
					t.Errorf("mixed configuration snapshot: %+v", active)
					return
				}
			}
		}()
	}
	for version := uint64(2); version <= 20; version++ {
		if err := manager.Publish(validConfig(version)); err != nil {
			t.Fatal(err)
		}
	}
	wait.Wait()
}

func TestConcurrentPublishersConvergeWithoutMixedSnapshots(t *testing.T) {
	t.Parallel()
	manager := mustManager(t, validConfig(1))
	start := make(chan struct{})
	var wait sync.WaitGroup
	for version := uint64(2); version <= 50; version++ {
		wait.Add(1)
		go func(version uint64) {
			defer wait.Done()
			<-start
			_ = manager.Publish(validConfig(version))
			active := manager.Active()
			if active.SchedulerName != schedulerName(active.Version) {
				t.Errorf("mixed snapshot after concurrent publish: %+v", active)
			}
		}(version)
	}
	close(start)
	wait.Wait()
	if got := manager.Active().Version; got != 50 {
		t.Fatalf("active version = %d, want highest published version 50", got)
	}
}

func mustManager(t *testing.T, initial Config) *Manager {
	t.Helper()
	manager, err := NewManager(initial)
	if err != nil {
		t.Fatal(err)
	}
	return manager
}

func validConfig(version uint64) Config {
	return Config{
		Version:       version,
		SchedulerName: schedulerName(version),
		Weights:       map[string]float64{"queue": 0.5, "kv": 0.3, "prefix": 0.2},
		MaxInFlight:   int(version) * 10,
		MaxQueued:     int(version) * 20,
	}
}

func schedulerName(version uint64) string {
	return "profile-" + strconv.FormatUint(version, 10)
}
