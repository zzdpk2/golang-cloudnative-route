# 100 Go Mistakes Applied Clinics

This directory is about Teiva Harsanyi's
[*100 Go Mistakes and How to Avoid Them*](https://100go.co/), not a second
beginner syntax course. The main journey remains
[`LEARNING_PATH.md`](../LEARNING_PATH.md). Each required gate reports one
book-aligned review focus, and the clinics below apply those ideas to the real
Commerce and llm-d Router code.

Do not implement a clinic as a separate toy solution. Work in the linked gate,
read its tests first, implement only its `TODO(exercise)` contracts, and use the
clinic questions during review. The book numbers are indexes for revision, not
instructions to reproduce the book's examples.

## Working Rule

For every gate:

1. Run `go run ./cmd/exercise -v next` and note its **100 Go Mistakes focus**.
2. Read the failing contract and identify ownership, lifetime, error, and
   observability boundaries before editing production code.
3. Implement the smallest production-shaped behavior that satisfies the
   contract; do not introduce a speculative interface, utility package, or
   goroutine.
4. Run the focused test, then the race detector when shared state is involved.
5. Record evidence that would let a senior reviewer distinguish correctness
   from a test-only workaround.

## Targeted Clinics

### 1. Organization and API Boundaries — #1-16

Apply this clinic at F1, E2, R4, R5, R7, and R8.

- Find the consumer that actually needs each interface. Can the producer return
  a concrete value instead?
- Look for shadowing, deep nesting, hidden `init` work, premature generics,
  embedded behavior, and vague `any` values.
- Reject a new `utils`, `common`, or catch-all package. Put behavior beside the
  domain or runtime owner that gives it meaning.
- At the Router composition root, make construction and failure explicit rather
  than relying on package globals.

Evidence to produce: a short API-boundary note naming the consumer, the minimal
method set, and one abstraction you deliberately did not add.

### 2. Data Ownership and Value Semantics — #17-29

Apply this clinic at F0, E0, E1, E3, K0, R0, R1, and R3.

- Check integer bounds and prove that the chosen Money representation preserves
  exact equality and arithmetic invariants.
- Predict slice length, capacity, append behavior, and retained backing storage
  before running the test.
- Decide whether nil and empty collections are observably different at JSON or
  API boundaries.
- Make every snapshot contract explicit: who owns the returned slice, map, URL,
  or nested reference field?
- Preallocate only when a measured or known upper bound justifies it.

Evidence to produce: one mutation test proving that a caller cannot corrupt
stored Order, Endpoint, or configuration state through an alias.

### 3. Control Flow, Strings, and Evaluation — #30-47

Apply this clinic at F0-F2, E1, K0, and request parsing in R4/R6.

- Keep the success path left-aligned and make terminal failure branches return.
- Check whether `range` gives a copy, whether map order matters, and whether a
  loop variable's address escapes the intended iteration.
- Treat strings as bytes only when the protocol is byte-oriented; otherwise
  account for runes and UTF-8.
- Keep `defer` out of unbounded loops and predict when deferred arguments and
  receivers are evaluated.
- Choose pointer or value receivers from mutation, copying, and consistency
  requirements rather than habit.

Evidence to produce: a table test with a non-ASCII case and a boundary case
that would fail under byte-count or range-copy assumptions.

### 4. Error Ownership and Cleanup — #48-54

Apply this clinic at F3, E2, E6, R4-R8.

- Reserve panic for violated programmer invariants, not request or dependency
  failures.
- Wrap only when adding useful context while preserving identity for
  `errors.Is` or `errors.As`.
- Handle an error once: return it, translate it, or log it at the owning
  boundary. Avoid doing two of those in different layers.
- Check errors from cleanup operations when their failure can change the
  outcome.
- Keep internal detail out of stable HTTP error codes.

Evidence to produce: one test proving that a deeply wrapped domain error maps
to the intended external result without duplicate reporting.

### 5. Concurrency Foundations and Practice — #55-74

Apply this clinic at C0-C3, E3-E4, and R0-R8. This is the project's most
important book section for llm-d Router work.

- State why a mutex, channel, atomic, or immutable snapshot owns each shared
  transition. Do not select a mechanism because it looks more advanced.
- Give every goroutine a start owner, stop signal, terminal error path, and join
  point.
- Define channel direction, notification semantics, and capacity from the
  backpressure contract. Never close a channel from the receiver side.
- Propagate request context only across request-scoped work; do not store it or
  use it as optional parameters.
- Separate data-race freedom from higher-level ordering and fairness
  correctness.
- Never expose a map or slice merely because access to the field was protected
  while returning it.
- Never copy a value containing a mutex, atomic no-copy marker, condition, or
  wait group.

Evidence to produce: `go test -race` output plus a cancellation or overload
test that finishes without sleeps, leaked permits, or orphaned goroutines.

### 6. Standard Library and HTTP Boundaries — #75-81

Apply this clinic at E5-E6, P0, R0, R2, R4, R6-R8.

- Use typed durations and explicit server/client/transport policies.
- Avoid repeated `time.After` allocation in long-lived loops; own reusable
  timers and their reset/drain lifecycle when needed.
- Decode one bounded JSON value, reject unknown fields where the contract calls
  for it, and distinguish absent, zero, nil, and empty states deliberately.
- Close HTTP bodies, files, rows, and other transient resources on every path.
- After writing a terminal HTTP response, return before any later write or
  mutation.
- Preserve streaming: a total request timeout is usually incompatible with a
  long token stream, while connection and header timeouts still need bounds.

Evidence to produce: an `httptest` contract covering an oversized or malformed
request, cancellation, body closure, and streamed chunks.

### 7. Testing as Production Evidence — #82-90 + Community Fuzzing

Apply this clinic at F4 and every later gate.

- Keep fast unit contracts separate from integration, acceptance, cluster, and
  optional compiler/platform checks.
- Use table tests, `t.Parallel` only with isolated state, and shuffle where
  ordering bugs are plausible. Treat fuzzing as the official catalog's
  unnumbered community addition and use it for parsers or invariant-heavy
  values.
- Replace sleeps with controlled clocks, synchronization, or observable state.
- Benchmark the real operation, reset setup cost, report allocations, and keep
  compiler elimination from invalidating the result.
- Robot tests assert public behavior; they must not inspect Go implementation
  details.

Evidence to produce: the focused command, the relevant race/fuzz/benchmark or
Robot command, and a sentence describing what failure class each catches.

### 8. Optimization and Runtime Constraints — #91-100

Apply this clinic at C2, K0, O0-O3, Y0, P0, H0, I0, G0, and R6-R8.

- Profile before optimizing and keep a before/after benchmark with allocation
  data.
- Treat cache layout, false sharing, alignment, escape, inlining, and pooling as
  evidence-driven tools rather than default design rules.
- Bound metric labels and avoid retaining request, model, or Endpoint data
  accidentally through maps, substrings, or snapshots.
- In Docker and Kubernetes, set CPU/memory expectations, probes, graceful
  shutdown, and runtime-aware concurrency. A container limit changes scheduler
  and garbage collector behavior; it is not just deployment YAML.
- Use traces, profiles, metrics, and execution evidence to explain a bottleneck
  before changing routing policy or adding concurrency.

Evidence to produce: a benchmark or profile hypothesis, the observed evidence,
and the Kubernetes resource or shutdown implication.

## Compact Command Set

```bash
go run ./cmd/exercise -v next
go test -race ./internal/foundation/concurrency -count=20
go test -race ./internal/commerce/... -count=1
go test -race ./internal/inference/... -count=1
go test ./internal/foundation/testing -fuzz=FuzzSplit -fuzztime=10s
go test ./internal/foundation/performance -bench=. -benchmem
go vet ./...
```

Compiler layout, `unsafe`, reflection, and other niche material remains starred
and optional. The review habits above are not optional: they are the quality
lens used while completing the single Commerce-to-Router path.
