# Senior Engineering Track

This track extends L0-L17 from implementation practice into system ownership.
It does not award a senior title. It trains the work products and failure modes
that a senior backend engineer is expected to handle.

## Entry Criteria

Start after M6 and M7 pass and the service can be exercised through HTTP.
Complete L15 before the migration scenarios that depend on inventory,
settlement, refund, and event history.

Do not wait for every optional lab exercise. The hard prerequisite is a working
Order path with tests you trust.

## How This Track Works

1. Read [RELEASE_TRAIN.md](RELEASE_TRAIN.md) and freeze the baseline artifacts.
2. Complete L18-L23 in order. Later levels inherit every earlier decision and
   piece of data.
3. Create the required evidence before implementation when the scenario asks
   for a proposal, risk analysis, or rollback plan.
4. Run an adversarial review with another person or an AI reviewer. Record the
   disagreement and the resulting decision.
5. Score the release using [RUBRIC.md](RUBRIC.md). A green test suite alone is
   not a passing release.

Use [ACCEPTANCE.md](ACCEPTANCE.md) to turn the required evidence into executable
release gates. The `seniorcheck` command runs automated checks, verifies sealed
evidence, and records independent approval in one report.

## Supplied Infrastructure Versus Your Work

Do not replace the completed verification infrastructure with TODO stubs. It is
the test harness for the senior track, not an implementation exercise.

| Area | Ownership |
|---|---|
| `internal/senior/acceptance` | Supplied manifest validation and gate runner |
| `internal/senior/testsupport` | Supplied deterministic concurrency helpers |
| `cmd/seniorcheck` | Supplied CLI for executing and sealing evidence |
| `docs/senior/scenarios` | Exercise prompts and required outcomes |
| production packages and `test/senior` | Your implementation and failure tests |
| `evidence/releases/<release>` | Your reports, measurements, runbooks, and reviews |

The supplied code may execute your commands and inspect your artifacts, but it
must not implement Order behavior, Router behavior, migrations, deployment
policy, incident recovery, capacity control, or architectural decisions for
you.

## Levels

| Level | Focus | Main evidence |
|---|---|---|
| [L18](scenarios/L18_EVOLUTION_AND_CONSISTENCY.md) | Data and contract evolution | Migration proof and compatibility matrix |
| [L19](scenarios/L19_INCIDENTS_AND_DIAGNOSIS.md) | Production diagnosis | Incident timeline, profiles, and corrective actions |
| [L20](scenarios/L20_PRODUCTION_DELIVERY.md) | Deployment and operability | Repeatable delivery, SLOs, rollout, and rollback |
| [L21](scenarios/L21_SECURITY_AND_TENANCY.md) | Security boundaries | Threat model, isolation tests, and audit evidence |
| [L22](scenarios/L22_CAPACITY_AND_COST.md) | Capacity and economics | Load model, saturation evidence, and cost budget |
| [L23](scenarios/L23_ARCHITECTURE_OWNERSHIP.md) | Technical leadership | RFC, decision record, review, and deprecation plan |

## Required Evidence Bundle

Every release must contain:

- a release record created from [templates/RELEASE.md](templates/RELEASE.md);
- assumptions and measurable success criteria;
- at least two credible options for consequential decisions;
- implementation and tests at the appropriate boundaries;
- migration, rollout, and rollback procedures;
- observability and a runbook for the changed path;
- review feedback and the disposition of each material concern;
- residual risks, ownership, and a removal date for temporary compatibility.

Use the other templates only when their trigger applies. More documents are not
automatically better; each document must support a decision or an operation.

## Rules

- Never edit historical fixtures, old client requests, or completed migration
  files to make a later release pass.
- Never claim exactly-once delivery. State the delivery guarantee and prove how
  duplicates are handled.
- Never use dashboards as decoration. Every chart must answer an operational
  question.
- Never use retries without a budget, idempotency analysis, and an overload
  consequence.
- Never call a rollback safe until it has been executed against realistic data.
- A security or unrecoverable data-loss defect is a failed release regardless
  of the numeric score.

## Suggested Pace

Treat each level as a one- or two-week release, not as a daily syntax exercise.
Use shorter daily review sessions for Go fundamentals while the release work
continues. The full track should span at least eight weeks so earlier decisions
have time to become constraints rather than being forgotten.
