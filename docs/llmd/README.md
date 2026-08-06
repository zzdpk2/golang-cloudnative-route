# llm-d Router Module Reference

This document explains the Router modules only. It does not define exercise
order. Start from the repository-wide
[Commerce to llm-d learning path](../LEARNING_PATH.md); Phase 4 links here when
the Commerce project and llm-d prerequisites are complete.

The local implementation is deliberately GPU-free and is not wire-compatible
with upstream llm-d. It teaches ownership and failure behavior before the
pinned upstream capstone.

```text
Client -> Proxy -> Request Handler -> Production Picker -> Endpoint
                                  |          |
                                  |          +-- Scheduling Profile
                                  |          +-- Flow Control
                                  |          +-- active Config
                                  +------------- request-scoped Admission

Endpoint sources -> Data Layer -> fresh immutable state
```

## R0 — Data Layer

Files: `internal/inference/datalayer/store.go`, `source.go`, and their tests.

Own ordered Endpoint updates, freshness, out-of-order rejection, delete
tombstones, deterministic defensive snapshots, bounded strict fixtures, event
ordering, cancellation, and source failure.

The Data Layer is a deep module: its interface exposes fresh state while its
implementation owns clocks, ordering, tombstones, and source lifecycle.

## R1 — Scheduling Profile

Files: `internal/inference/routing/types.go`, `scheduler.go`, `plugins.go`, and
`scheduler_test.go`.

Preserve `Filter -> Score -> Pick`, validated weights, rejected-candidate
evidence, deterministic tie-breaking, and immutable inputs. HTTP and admission
concerns stay outside this seam.

## R2 — Flow Control

Files: `internal/inference/flowcontrol/queue.go`, `controller.go`, and tests.

Own the hard queue bound, identity, priority, tenant fairness, FIFO within a
class, TTL expiry, cancellation removal, and permit transfer. A cancelled or
expired request must never leak capacity.

## R3 — Versioned Configuration

Files: `internal/inference/configuration/manager.go`, tests, and
`configs/endpoint-picker-learning.yaml`.

Validate a candidate completely before atomic publication. Keep immutable
history, last-known-good state, bounded values, and explicit rollback. Readers
must never observe a partially updated version.

## R4 — Request Handler

Files: `internal/inference/epp/request_handler.go` and tests.

Parser, Data Producer, and Admitter own different phases. Bound and strictly
decode the supported OpenAI-style request, derive request-scoped routing state,
preserve error identity, and unwind partial admission in reverse order.

## R5 — Production Picker

Files: `internal/inference/epp/production_picker.go` and tests.

Fresh Data Layer state, the Scheduling Profile, Flow Control, and active Config
must affect the same decision. Release acquired ownership exactly once and
expose only immutable fresh snapshots.

## R6 — Proxy

Files: `internal/inference/httpapi/router.go`, `simulator.go`, and tests.

Own body limits, strict decoding, hop-by-hop headers, cancellation, connection
and streaming timeouts, SSE flushing, readiness, stable error envelopes, and
response lifecycle. Never retry after response bytes become visible.

Proxy and Production Picker remain separate modules because they have a real
seam: transport lifecycle versus routing policy.

## R7 — Runtime

Files: `internal/inference/runtime/runtime.go` and tests.

Compose the real Store, Controller, Manager, Request Handler, Production Picker,
Proxy, and metrics. The runtime implementation is intentionally shallow wiring;
the invariant-heavy modules behind it stay deep and independently testable.

## R8 — Executable Harness

Files: `cmd/inference-lab/main.go` and tests.

Own strict bounded config, explicit Router and Model Server modes, safe HTTP
timeouts, signal cancellation, one-error-fails-the-group behavior, and graceful
drain.

## Production Evidence

Before the upstream capstone, record:

- race results under Endpoint churn and concurrent configuration updates;
- bounded queue, goroutine, retry, memory, and telemetry-label growth;
- cancellation, overload, stale state, Endpoint loss, and Pod eviction;
- p50/p95/p99 queue wait, TTFT, total latency, and error rate;
- one injected failure, detection signal, recovery, and alert resolution;
- rollout and rollback commands plus residual risk ownership.

Use [PRODUCTION_LABS.md](PRODUCTION_LABS.md) for exact delivery and operations
drills, [EXERCISE_RULES.md](EXERCISE_RULES.md) for the no-answer policy, and
[UPSTREAM.md](UPSTREAM.md) for the final pinned llm-d comparison.
