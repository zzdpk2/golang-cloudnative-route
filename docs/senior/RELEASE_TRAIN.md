# Release Train

The same Order service moves through every release. A release is complete only
when its code, data change, operational evidence, and recovery procedure have
all passed.

## Baseline: V1

Freeze V1 immediately after M6 and M7 pass.

Preserve:

- representative Order API request and response fixtures;
- database or repository fixtures for each Order status;
- emitted event fixtures and their schema identifiers;
- one command that starts the service;
- one command that runs the baseline acceptance tests;
- observed latency and allocation baselines;
- the exact commit or archive used for the baseline.

V1 is not rewritten when later releases expose weaknesses. Record the weakness
and migrate forward.

## V2: Durable Orders

Replace process memory as the source of truth. Introduce schema migrations,
optimistic concurrency, idempotent commands, and a transactional outbox.

The release must demonstrate:

- restart without losing an accepted Order;
- concurrent updates without silent overwrite;
- no event loss between a committed Order and outbox publication;
- duplicate command and event handling;
- forward and backward database compatibility during rollout.

## V3: Identity and Tenant Boundaries

Add authenticated identities, authorization, tenant ownership, audit records,
and secret rotation. Existing V2 data must be assigned deliberately; a default
tenant hidden in application code is not a migration.

## V4: Fulfillment and Refund Evolution

Allow warehouse-level fulfillment and independently progressing Sub-orders.
Preserve V1 Order clients while exposing the new behavior. Exercise a partial
refund during an in-flight shipment and define which subsystem owns each fact.

## V5: Event and API Evolution

Change at least one event and one public API representation while old data,
consumers, and clients still exist. Introduce explicit versioning, compatibility
tests, usage telemetry, and a deprecation deadline.

## Release Gates

Every release answers these questions with evidence:

1. What user or operational problem is being solved?
2. What will remain deliberately unchanged?
3. Which old and new versions can coexist?
4. What happens if execution stops halfway through?
5. How is partial success detected and repaired?
6. What signal stops the rollout?
7. How is rollback executed after a migration has started?
8. Which temporary code or data can be deleted, by whom, and when?

## Fault Injection Matrix

At minimum, run each relevant release with:

| Fault | Required observation |
|---|---|
| Process termination after commit | State is durable and unpublished work is recoverable |
| Duplicate request | Business effect occurs at most once |
| Dependency timeout | Work is bounded; retries do not amplify overload |
| Dependency returns stale data | Invariant remains protected or risk is explicit |
| Partial migration failure | Rerun is safe and progress is observable |
| Old and new instances overlap | Shared data and contracts remain compatible |
| Telemetry backend unavailable | Business path behavior is intentional and bounded |

## History Policy

Completed release evidence is append-only. Corrections are new records that
reference the original. This policy is part of the exercise: future decisions
must deal with what was actually shipped, not a cleaned-up story.
