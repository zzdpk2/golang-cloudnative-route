# Foundation Exercises

`foundation` is the single home for reusable Go knowledge. Required gates use
only a compact subset; the remaining packages are starred optional extensions.

| Package | Purpose | Required? |
|---|---|---|
| `language` | data ownership, interfaces, semantics, errors | F0-F3 |
| `testing` | tables, properties, fuzzing, mutation value | F4 |
| `concurrency` | memory model, synchronization, channels, cache ownership | C0-C3 plus optional extensions |
| `collection` | sorting, grouping, chunking, frequency | optional support |
| `performance` | escape, layout, BCE, reflection, profiling, unsafe | optional |
| `patterns` | GoF examples, middleware, small router mechanisms | optional |
| `fp` | generic Option, Result, containers, iterators | optional |
| `resilience` | pipelines, retries, circuit breaking | optional |

Basic syntax is not repeated in one-function exercises. Stronger scenarios
combine related rules, while niche syntax and compiler topics stay optional.
The two cancellation helpers intentionally teach different contracts:
`AnyDone` combines multiple stop signals; `ForwardUntilDone` forwards typed
values until cancellation.

Useful commands:

```bash
go test ./internal/foundation/language ./internal/foundation/testing
go test -race ./internal/foundation/concurrency -count=20
go test ./internal/foundation/performance -bench=. -benchmem
```

Production packages must not import `internal/foundation` merely to reuse an
exercise implementation. Promote a mechanism only when a real production
interface requires it.
