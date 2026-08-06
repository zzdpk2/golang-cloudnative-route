# Context Map

## Contexts

- [Commerce Order](./CONTEXT.md) — prices, reserves, fulfills, settles, and
  refunds customer Orders.
- [Inference Routing](./internal/inference/CONTEXT.md) — admits inference
  requests and selects a suitable Model Server Endpoint.

## Relationships

- **Commerce Order and Inference Routing share no domain model.** They coexist
  in this repository because both are used to learn Go engineering, TDD,
  failure handling, observability, and production delivery.
- **Training infrastructure is shared.** Test conventions, deterministic fault
  helpers, CI gates, and release evidence may be reused without sharing domain
  types.

