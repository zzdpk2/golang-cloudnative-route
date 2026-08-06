# L23: Architecture Ownership

## Mission

Own an ambiguous, cross-cutting change from problem definition through
deprecation and handoff. The implementation is only one part of the assessment.

## Scenario

The organization expects a large growth in Order volume and asks to "move to
microservices so teams can scale independently." Current evidence is incomplete,
release ownership is shared, and the highest operational cost is not yet known.

## Work

1. Interview the available evidence: repository structure, incidents, delivery
   lead time, ownership boundaries, load tests, and dependency bottlenecks.
2. Restate the problem without assuming microservices are the solution.
3. Write an RFC with at least two credible designs, including an incremental
   option.
4. Identify reversible and irreversible decisions.
5. Run adversarial architecture, security, operations, and migration reviews.
6. Implement the smallest change that tests the most important assumption.
7. Define adoption, deprecation, rollback, and ownership.
8. Present the decision and record dissent that remains unresolved.

## Candidate Outcomes

Valid outcomes include retaining a modular monolith, extracting one bounded
context, changing data ownership without a network boundary, or splitting a
service. The score depends on evidence and consequences, not on architectural
fashion.

## Required Evidence

- problem statement with non-goals and decision deadline;
- current-state dependency and ownership map;
- options compared on reliability, migration, team autonomy, security, cost,
  and reversibility;
- an ADR for the chosen hard-to-reverse decision;
- a walking skeleton or experiment for the riskiest assumption;
- compatibility and data-ownership plan;
- review comments and explicit dispositions;
- deprecation metrics, deadline, owner, and fallback;
- handoff session validated by another engineer following the runbook.

## Acceptance

- the proposal can be rejected without losing the discovery work;
- the chosen design removes a measured constraint;
- consistency and failure semantics across boundaries are explicit;
- operational ownership exists before traffic moves;
- temporary adapters and duplicate data have removal conditions;
- the reviewer can explain the trade-off without reading the implementation.

## Final Reflection

After the release has operated for two weeks, record which assumptions were
wrong, which costs appeared elsewhere, and what decision would change with the
new evidence. This delayed review is mandatory.
