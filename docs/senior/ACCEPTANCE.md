# Executable Acceptance

Senior-track acceptance has four layers. No single layer is sufficient.

1. **Behavior**: Go E2E and Robot tests prove externally visible business
   behavior and compatibility.
2. **Failure**: deterministic fault tests and environment drills prove behavior
   during concurrency, interruption, timeout, overload, and version overlap.
3. **Evidence**: profiles, migration transcripts, load reports, dashboards,
   restore results, and runbooks preserve what was observed.
4. **Review**: an independent reviewer challenges the decision and approves or
   rejects the release.

`seniorcheck` aggregates these layers. It does not replace Go, Robot, k6, Helm,
or Kubernetes.

## Quick Start

The repository includes a passing bootstrap manifest that demonstrates the
mechanism but does not count as senior-track evidence:

```bash
go run ./cmd/seniorcheck \
  -manifest docs/senior/examples/bootstrap.release.json \
  -report senior-report.json
```

Validate a manifest without executing its commands:

```bash
go run ./cmd/seniorcheck \
  -manifest evidence/releases/v2/manifest.json \
  -validate-only
```

Seal an evidence file after review:

```bash
go run ./cmd/seniorcheck -hash evidence/releases/v2/migration-report.md
```

Put the printed digest in the artifact's `sha256` field. A later edit then
fails the gate instead of silently rewriting release history.

## Manifest Contract

A release manifest must contain at least one required gate of each kind:

| Kind | Purpose | Pass condition |
|---|---|---|
| `command` | Automated behavior or drill | Process exits zero before its timeout |
| `evidence` | Durable observation | Every file exists and optional digest matches |
| `approval` | Independent judgment | Reviewer differs from owner, approves, and reviewed files exist |

Every command is an argument array and is executed without a shell:

```json
{
  "command": [
    "go",
    "test",
    "-race",
    "./test/senior/l18/...",
    "-count=1"
  ]
}
```

This is intentional. If a gate needs pipes, environment preparation, multiple
commands, or cleanup, put that workflow in a reviewed script and invoke the
script from the manifest. The actual acceptance procedure then lives in source
control instead of inside an opaque shell string.

Paths are relative to the Go module root and may not escape it. Command output
is capped at 128 KiB in the report. Secrets must never be placed in a manifest
or command argument.

An approval entry is a local attestation, not cryptographic identity. In a
shared repository, branch protection and code-review identity are the trust
mechanism; the manifest records the resulting decision.

## Suggested Evidence Layout

```text
evidence/
  releases/
    v2/
      manifest.json
      report.json
      rfc.md
      compatibility-matrix.md
      migration-report.md
      restore-report.md
      review.md
      runbooks/
```

Do not create every file before it has evidence to contain.

## Level Gates

### L18: Evolution and Consistency

Required automated gates:

- old API fixtures against the new binary;
- old and new event replay;
- duplicate command and event tests;
- migration, interrupted backfill, and rerun;
- old/new binary overlap;
- backup restore and invariant comparison.

Required evidence:

- compatibility matrix;
- migration timing and lock impact;
- pre/post migration invariant totals;
- restore transcript;
- removal owner and deadline for compatibility paths.

### L19: Incidents and Diagnosis

Required automated gates:

- regression for the seeded failure;
- fault injection reproduces the original symptom;
- corrected version remains bounded under the same fault.

Required evidence:

- incident timeline;
- relevant pprof files and interpretation;
- mitigation transcript;
- post-incident report and updated runbook.

### L20: Production Delivery

Required automated gates should compose tools such as:

```text
go test -> image build/scan -> helm lint/template -> deploy
-> readiness check -> traffic during rollout -> fault drill -> rollback
```

Required evidence:

- immutable image digest and source revision;
- SLO result during rollout;
- migration and rollback transcript;
- dashboard and alert links or exported definitions;
- exercised runbook.

`helm template` alone is not a deployment acceptance test.

### L21: Security and Tenancy

Required automated gates:

- cross-tenant matrix at HTTP, repository, cache, and event boundaries;
- expired, revoked, and rotated credentials;
- sensitive-value scan over logs and traces;
- request size, rate, and cardinality limits.

Required evidence:

- threat model;
- authorization decision matrix;
- audit record examples;
- emergency revocation and rotation drill.

Any known cross-tenant access fails the release.

### L22: Capacity and Cost

Required automated gates:

- steady, step, burst, hot-key, slow-dependency, and cache-cold tests;
- explicit p99 and error thresholds;
- admitted-load and overload assertions;
- telemetry-cardinality bounds.

Required evidence:

- versioned load scripts and data;
- bottleneck profiles;
- capacity model and error bounds;
- cost per successful Order at normal and peak load.

### L23: Architecture Ownership

Required automated gates:

- the walking skeleton or experiment for the riskiest assumption;
- compatibility checks for any boundary that moves;
- deprecation-usage telemetry.

Required evidence:

- RFC and ADR;
- dependency and ownership map;
- adversarial review with dispositions;
- adoption, rollback, deprecation, and handoff records;
- delayed review after two weeks of operation.

## Test Support

`internal/senior/testsupport` supplies three generic helpers.

### Deterministic Fault Point

Production code accepts a narrow hook and defaults it to `testsupport.Noop`.
A test injects a barrier:

```go
barrier := testsupport.NewBarrier()

go func() {
    _ = service.DoWork(ctx, barrier.Reach)
}()

require.NoError(t, barrier.WaitUntilReached(ctx))
// Observe committed state, cancel a dependency, or start a competing request.
barrier.Release()
```

The hook belongs at the exact boundary being tested, such as after durable
commit and before publication. Do not scatter failpoints through business code.

### Coordinated Concurrency

```go
results := testsupport.RunConcurrent(ctx, 200,
    func(ctx context.Context, worker int) (Reservation, error) {
        return ledger.Reserve(ctx, request)
    })
```

Workers begin behind one start gate. This increases overlap without relying on
arbitrary sleeps. The test must still assert business outcomes, not scheduling
order.

### Eventual Observation

```go
err := testsupport.Eventually(ctx, 20*time.Millisecond, func() error {
    return assertOutboxPublished(store, orderID)
})
```

Use this only when the system is genuinely asynchronous. A direct synchronous
assertion is stronger when it is available.

## CI Rule

Store the generated JSON report as a CI artifact. A required gate failure exits
with code 1; invalid configuration exits with code 2. Promotion must consume the
same immutable artifact digest that the manifest records.
