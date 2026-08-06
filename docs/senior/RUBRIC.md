# Senior Track Rubric

Score each category from 0 to 4, multiply by its weight, then divide by four.
The maximum total is 100.

| Category | Weight | What strong evidence looks like |
|---|---:|---|
| Problem framing | 15 | Defines users, constraints, non-goals, unknowns, and measurable success |
| Decision quality | 15 | Compares credible options and makes trade-offs explicit |
| Correctness and data safety | 20 | Protects invariants under concurrency, retries, partial failure, and migration |
| Operability and recovery | 20 | Supplies useful telemetry, bounded failure, tested rollback, and a usable runbook |
| Evolution and compatibility | 15 | Preserves old data and clients with a time-bounded transition plan |
| Security and cost | 10 | Models trust boundaries, abuse paths, capacity, and economic limits |
| Communication and ownership | 5 | Produces reviewable records, resolves feedback, assigns owners, and closes cleanup |

## Score Meaning

| Score | Meaning |
|---:|---|
| 0 | Missing, asserted without evidence, or creates a critical defect |
| 1 | Happy path works; material risks are ignored |
| 2 | Common failure modes are handled, but evidence or recovery is incomplete |
| 3 | Production-ready for the stated scope, with tested failure and recovery paths |
| 4 | Anticipates second-order effects and leaves the system easier to operate and evolve |

## Passing Rule

A release passes at 80 or above when:

- no category scores zero;
- correctness and operability each score at least 3;
- there is no known cross-tenant access, unrecoverable data loss, unbounded
  resource growth, or unowned critical alert;
- rollback or forward repair was executed rather than only described.

## Review Questions

The reviewer should challenge:

- whether the problem is real and the scope is the smallest useful one;
- whether a simpler option was dismissed too quickly;
- what happens during version overlap rather than only before and after;
- whether retries, caches, queues, and locks introduce hidden amplification;
- which metric proves user impact rather than component activity;
- which migration step is irreversible;
- who receives the alert and what they can do at 03:00;
- when compatibility code will be removed.

## Calibration

Keep the completed evidence bundle and score it again after two later releases.
If later failures expose an earlier blind spot, update the old release's
calibration note without rewriting its original score. The purpose is to improve
judgment, not to preserve a grade.
