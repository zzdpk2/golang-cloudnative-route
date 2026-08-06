# Inference Exercise Rules

## The Repository Does Not Contain Answers

Production code and deployment artifacts use `TODO(exercise)` markers.
Comments may name an invariant, failure mode, or required evidence. They must
not contain an algorithm or completed configuration.

Tests are executable contracts. Do not weaken, skip, or delete them to advance
the track.

Repository maintainers can audit the untouched starter without blocking normal
learner test runs:

```bash
go test -tags starteraudit ./cmd/exercise -run TestLearnerOwnedGoFunctionsContainNoCompletedAnswers
```

Validation infrastructure is deliberately complete. `cmd/exercise`,
`cmd/seniorcheck`, `internal/senior/acceptance`, and generic deterministic test
helpers execute or seal learner-owned work; they are not exercises and do not
implement Router, domain, migration, deployment, incident, or architecture
outcomes. Learner-owned production paths and evidence artifacts remain red.

## Completion Loop

```bash
go run ./cmd/exercise next
go test <package printed by next> -run '<focused expression>' -count=1
```

For concurrency work, also run the printed package with `-race`. For delivery
work, remove the matching `TODO(exercise)` only after the artifact is complete
and its validation command succeeds.

## Senior Evidence

A feature is incomplete until the release records:

- behavior under cancellation and overload;
- bounded memory, queue, label, and retry growth;
- deterministic tests and race results;
- p50/p95/p99 latency or queue evidence where applicable;
- one injected failure and measured recovery;
- rollout and rollback commands;
- the SLI or alert that detects regression.

Passing a render, lint, or happy-path request alone is not production evidence.

## Hint Policy

Allowed hints:

- the invariant a lock protects;
- externally observable behavior;
- relevant failure categories;
- which official concept to study;
- the command that verifies the result.

Disallowed hints:

- completed function bodies;
- exact scheduling algorithms;
- finished Kubernetes, Helm, Istio, PromQL, or Grafana configurations;
- solution directories or hidden build-tag solutions.
