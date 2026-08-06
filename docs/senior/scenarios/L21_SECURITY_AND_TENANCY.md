# L21: Security and Tenancy

## Mission

Add identity and tenant isolation to an existing system without leaking old or
new Order data across trust boundaries.

## Threats to Model

- horizontal access to another Customer's Order;
- cross-tenant reads through list, cache, search, export, or event consumers;
- forged, expired, replayed, or incorrectly scoped tokens;
- key and secret rotation during live traffic;
- sensitive data in logs, traces, metrics labels, or error responses;
- privileged support actions without an audit record;
- resource abuse through unbounded requests or high-cardinality input.

## Work

1. Draw trust boundaries and identify assets, actors, entry points, and abuse
   cases.
2. Define authentication separately from authorization.
3. Assign historical Orders to tenants through an auditable migration.
4. Enforce ownership at the narrowest dependable data-access boundary.
5. Add audit records for privileged and money-affecting actions.
6. Define token, key, and secret rotation.
7. Add rate, size, and cardinality limits where an attacker controls cost.

## Required Tests

- every read and mutation using a valid identity from the wrong tenant;
- list and pagination queries with mixed-tenant seeded data;
- cache keys and invalidation under identical Order IDs in two tenants;
- event consumers receiving a mismatched tenant identity;
- expired and revoked credentials during a long request;
- rotation with old and new keys overlapping;
- logs and traces scanned for configured sensitive values.

## Acceptance Evidence

- threat model with mitigations and accepted residual risks;
- tenant-isolation tests at HTTP, application, repository, and asynchronous
  boundaries;
- audit record proving actor, action, target, time, and correlation;
- emergency revocation and rotation runbook;
- least-privilege deployment identity;
- security failure returns no protected resource metadata.

## Release Blockers

Any known cross-tenant read or mutation, unaudited privileged money operation,
or secret committed to source fails the release regardless of rubric score.
