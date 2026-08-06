package concurrency

import (
	"container/list"
	"sync"
	"time"
)

// ============================================================
// Caches
//
// Two eviction policies for the hot-product cache: expire by age (TTL), and
// expire by pressure (LRU).
//
// They answer different questions. TTL asks "is this still true?" and suits
// data that goes stale on its own — a price, an exchange rate. LRU asks "do I
// have room?" and suits data that stays valid but will not all fit.
//
// A real product cache usually wants both, and that combination is where the
// interesting bugs live.
// ============================================================

// ---- TTL Cache ----

type ttlEntry[V any] struct {
	value     V
	expiresAt time.Time
}

// TTLCache expires entries a fixed duration after they are written.
type TTLCache[K comparable, V any] struct {
	mu      sync.RWMutex
	entries map[K]ttlEntry[V]
	ttl     time.Duration
	stop    chan struct{}
}

// NewTTLCache builds a cache and starts a background sweeper that drops expired
// entries.
//
// The sweeper is why this constructor is more than an initialiser, and why
// Close exists at all. Get already ignores expired entries, so eviction is not
// needed for *correctness* — it is needed so that a key written once and never
// read again does not occupy memory forever. That distinction is worth being
// clear about before you write the goroutine.
//
// The goroutine must exit when stop is closed, or every cache you create leaks
// one for the life of the process. Sweeping at ttl/2 is a reasonable default;
// think about what sweeping far more or far less often would cost.
func NewTTLCache[K comparable, V any](ttl time.Duration) *TTLCache[K, V] {
	panic("TODO")
}

// Set stores a value with a fresh expiry.
func (c *TTLCache[K, V]) Set(key K, value V) {
	panic("TODO")
}

// Get returns the value if present and not expired.
//
// An entry past its expiry must be reported as absent even when the sweeper has
// not reached it yet — otherwise the answer depends on sweeper timing, and the
// test will catch you.
//
// Then the harder question: should Get *delete* what it finds expired? Doing so
// needs the write lock, which means every read may become a write. Decide, and
// note that this is exactly why the type holds an RWMutex rather than a Mutex.
func (c *TTLCache[K, V]) Get(key K) (V, bool) {
	panic("TODO")
}

// Delete removes an entry, whether or not it had expired.
func (c *TTLCache[K, V]) Delete(key K) {
	panic("TODO")
}

func (c *TTLCache[K, V]) Len() int { panic("TODO") }

// evictExpired drops every entry past its expiry.
//
// Deleting from a map while ranging over it is explicitly allowed in Go —
// unlike most languages. Worth confirming in the spec rather than taking my
// word for it.
func (c *TTLCache[K, V]) evictExpired() {
	panic("TODO")
}

// Close stops the sweeper.
//
// Calling it twice panics on the double close. Left as-is deliberately: decide
// whether that is acceptable for this type, and compare with how you handled
// the same question for PubSub's unsubscribe function in L6.
func (c *TTLCache[K, V]) Close() { panic("TODO") }

// ---- LRU Cache ----

type lruEntry[K comparable, V any] struct {
	key   K
	value V
}

// LRUCache evicts the least recently used entry when it is full.
//
// The design is the exercise: a map for O(1) lookup, plus a doubly-linked list
// for O(1) reordering. Neither alone is enough — a map has no order, and a list
// has no fast lookup. The map's values are *list elements*, which is what lets
// you find a node and move it without walking the list.
//
// Note that the entry stores its own key even though the map is already keyed
// by it. Work out why eviction would be impossible otherwise; it is the detail
// people miss when writing this from memory.
type LRUCache[K comparable, V any] struct {
	mu       sync.Mutex
	capacity int
	items    map[K]*list.Element
	order    *list.List // front = most recently used
}

// NewLRUCache builds a cache holding at most capacity entries.
//
// Decide what a capacity of zero or less means — a cache that stores nothing,
// or a programming error. Silently treating it as unlimited is the one answer
// that will hurt someone later.
func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	panic("TODO")
}

// Get returns the value and marks the entry most recently used.
//
// A read that mutates. That is why this type holds a plain Mutex and not an
// RWMutex: there is no such thing as a read-only Get here, so a read lock would
// be a lie.
func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	panic("TODO")
}

// Set inserts or updates, evicting the least recently used entry when full.
//
// Three cases, and the middle one is the one that gets forgotten:
//   - key absent, room available → insert at the front
//   - key already present → update the value *and* move it to the front
//   - key absent, at capacity → evict from the back, then insert
//
// Evicting must remove the entry from both the list and the map. Dropping it
// from only one gives you a map that grows forever while the cache appears to
// be at capacity — the leak looks like a memory bug, not a cache bug.
func (c *LRUCache[K, V]) Set(key K, value V) {
	panic("TODO")
}

// Delete removes an entry from both the map and the list — miss either and
// the cache leaks while appearing to be at capacity.
func (c *LRUCache[K, V]) Delete(key K) {
	panic("TODO")
}

func (c *LRUCache[K, V]) Len() int { panic("TODO") }

// Keys returns the keys from most to least recently used.
//
// Walk the list, not the map: the map has no order, and Go randomises its
// iteration deliberately. Ranging the map here would give a test that passes
// most of the time.
func (c *LRUCache[K, V]) Keys() []K {
	panic("TODO")
}
