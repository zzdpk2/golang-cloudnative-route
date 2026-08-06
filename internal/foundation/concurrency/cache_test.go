package concurrency

import (
	"testing"
	"time"
)

// ---- TTL Cache ----

func TestTTLCache_SetGet(t *testing.T) {
	c := NewTTLCache[string, int](1 * time.Second)
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)

	v, ok := c.Get("a")
	if !ok || v != 1 {
		t.Errorf("a = %d, %v", v, ok)
	}

	v, ok = c.Get("b")
	if !ok || v != 2 {
		t.Errorf("b = %d, %v", v, ok)
	}

	_, ok = c.Get("c")
	if ok {
		t.Error("c should not exist")
	}
}

func TestTTLCache_Expiration(t *testing.T) {
	c := NewTTLCache[string, int](50 * time.Millisecond)
	defer c.Close()

	c.Set("x", 42)

	v, ok := c.Get("x")
	if !ok || v != 42 {
		t.Error("should exist immediately")
	}

	time.Sleep(100 * time.Millisecond) // See the corresponding tests for the intended behavior.

	_, ok = c.Get("x")
	if ok {
		t.Error("should be expired")
	}
}

func TestTTLCache_Overwrite(t *testing.T) {
	c := NewTTLCache[string, string](1 * time.Second)
	defer c.Close()

	c.Set("key", "old")
	c.Set("key", "new")

	v, ok := c.Get("key")
	if !ok || v != "new" {
		t.Errorf("got %q", v)
	}
}

func TestTTLCache_Delete(t *testing.T) {
	c := NewTTLCache[string, int](1 * time.Second)
	defer c.Close()

	c.Set("a", 1)
	c.Delete("a")
	_, ok := c.Get("a")
	if ok {
		t.Error("should be deleted")
	}
}

func TestTTLCache_BackgroundEviction(t *testing.T) {
	c := NewTTLCache[int, string](30 * time.Millisecond)
	defer c.Close()

	for i := 0; i < 100; i++ {
		c.Set(i, "val")
	}
	if c.Len() != 100 {
		t.Errorf("len = %d", c.Len())
	}

	time.Sleep(100 * time.Millisecond)

	if c.Len() > 0 {
		t.Errorf("len = %d, should be 0 after eviction", c.Len())
	}
}

// ---- LRU Cache ----

func TestLRUCache_SetGet(t *testing.T) {
	c := NewLRUCache[string, int](3)

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	v, ok := c.Get("a")
	if !ok || v != 1 {
		t.Error("a")
	}
	v, ok = c.Get("b")
	if !ok || v != 2 {
		t.Error("b")
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	c := NewLRUCache[string, int](3)

	c.Set("a", 1) // [a]
	c.Set("b", 2) // [b, a]
	c.Set("c", 3) // [c, b, a]
	c.Set("d", 4) // See the corresponding tests for the intended behavior.

	_, ok := c.Get("a")
	if ok {
		t.Error("a should be evicted")
	}

	v, ok := c.Get("d")
	if !ok || v != 4 {
		t.Error("d should exist")
	}
}

func TestLRUCache_AccessRefreshesOrder(t *testing.T) {
	c := NewLRUCache[string, int](3)

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	c.Get("a") // See the corresponding tests for the intended behavior.

	c.Set("d", 4) // See the corresponding tests for the intended behavior.

	_, ok := c.Get("b")
	if ok {
		t.Error("b should be evicted (least recently used)")
	}

	_, ok = c.Get("a")
	if !ok {
		t.Error("a should still exist (was accessed)")
	}
}

func TestLRUCache_Update(t *testing.T) {
	c := NewLRUCache[string, int](3)

	c.Set("a", 1)
	c.Set("a", 10) // See the corresponding tests for the intended behavior.

	v, ok := c.Get("a")
	if !ok || v != 10 {
		t.Errorf("a = %d", v)
	}

	if c.Len() != 1 {
		t.Errorf("len = %d", c.Len())
	}
}

func TestLRUCache_Delete(t *testing.T) {
	c := NewLRUCache[string, int](3)
	c.Set("a", 1)
	c.Delete("a")
	_, ok := c.Get("a")
	if ok {
		t.Error("deleted")
	}
	if c.Len() != 0 {
		t.Errorf("len = %d", c.Len())
	}
}

func TestLRUCache_Keys(t *testing.T) {
	c := NewLRUCache[string, int](5)
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	c.Get("a") // See the corresponding tests for the intended behavior.

	keys := c.Keys()
	if len(keys) != 3 {
		t.Fatalf("keys = %v", keys)
	}
	if keys[0] != "a" {
		t.Errorf("keys[0] = %q, want 'a'", keys[0])
	}
}
