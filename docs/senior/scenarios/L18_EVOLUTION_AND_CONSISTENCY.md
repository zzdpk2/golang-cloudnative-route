# L18: Evolution and Consistency

## Mission

Move Orders from in-memory storage to durable storage while old clients and old
data remain valid. Then evolve one public event without rewriting its history.

## Given

- V1 API and event fixtures are frozen.
- Orders can be accepted through HTTP.
- Some V1 Orders are pending, confirmed, shipped, cancelled, and malformed in
  ways the old code previously allowed.
- Deployment can temporarily run old and new application instances together.

## Work

1. Write an RFC describing the persistence model, consistency boundary,
   idempotency model, and event publication guarantee.
2. Introduce migrations that can run against populated data.
3. Implement optimistic concurrency or justify another lost-update defense.
4. Couple state changes to durable event publication using an outbox or a
   defended alternative.
5. Change one event schema and support historical events through upcasting,
   versioned consumers, or another explicit strategy.
6. Execute a multi-step change such as expand, dual-write, backfill, switch
   reads, and contract.

## Required Failure Drills

- Kill the process after the Order commit but before event publication.
- Submit the same command concurrently with the same idempotency key.
- Fail a backfill halfway through and rerun it.
- Run old and new binaries against the expanded schema.
- Replay the oldest event fixture with the newest code.
- Restore from backup into an empty environment and verify business totals.

## Acceptance Evidence

- compatibility matrix for application, schema, event, and client versions;
- migration duration and lock-impact measurements on realistic data volume;
- invariant checks before and after backfill;
- proof that duplicate delivery does not duplicate the business effect;
- rollback or forward-repair transcript for every migration phase;
- a deadline and owner for removing dual-write and legacy-read paths.

## Invalid Shortcuts

- editing V1 fixtures;
- dropping malformed historical data without a business disposition;
- wrapping database writes and message publication in separate best-effort calls;
- describing a rollback that would require reversing an irreversible migration;
- using a sleep to make concurrency tests pass.
