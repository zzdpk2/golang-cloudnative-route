# Keep inference routing as a separate learning context

The repository will teach llm-d foundations through an Inference Routing
context beside, not inside, Commerce Order. A GPU-free Go Router and Model
Server simulator provide the executable slice. Deployment-tool contracts may
be learned and rendered before Router implementation, but cluster failure
drills and upstream llm-d integration begin only after the same routing
behavior is understood locally. This preserves the Order ubiquitous language
and avoids mistaking a teaching simulator for an implementation of llm-d or
the Gateway API Inference Extension.

## Considered Options

- Replacing Commerce Order would discard a useful DDD/TDD progression.
- Adding inference concepts to Order would create a false shared domain.
- Starting directly with a full GPU cluster would hide Go and Kubernetes
  fundamentals behind installation work and cost.

## Consequences

The two contexts share engineering practices but no domain types. The local
Router intentionally models the Filter, Score, Pick lifecycle without claiming
wire compatibility with the upstream Endpoint Picker.
