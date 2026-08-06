# Upstream llm-d Reading Map

Upstream documentation and schemas evolve. Pin a release and revalidate every
artifact version before installation.

## Read One Request End to End

1. [llm-d architecture](https://llm-d.ai/docs/architecture)
2. [Router](https://llm-d.ai/docs/dev/architecture/core/router)
3. [Endpoint Picker](https://llm-d.ai/docs/dev/architecture/core/router/epp)
4. [Request Handler](https://llm-d.ai/docs/dev/architecture/core/router/epp/request-handling)
5. [Request Scheduling](https://llm-d.ai/docs/dev/architecture/core/router/epp/scheduling)
6. [Data Layer](https://llm-d.ai/docs/dev/architecture/core/router/epp/datalayer)
7. [EndpointPickerConfig](https://llm-d.ai/docs/dev/architecture/core/router/epp/configuration)
8. [Router operations](https://llm-d.ai/docs/dev/operations/router)

Repositories and artifacts:

- [llm-d](https://github.com/llm-d/llm-d)
- [Go inference scheduler](https://github.com/llm-d/llm-d-inference-scheduler)
- [GPU-free inference simulator](https://github.com/llm-d/llm-d-inference-sim)
- [official artifact index](https://llm-d.ai/docs/api-reference/artifacts)

Gateway and Kubernetes references:

- [Kubernetes Custom Resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
- [Gateway API Inference Extension](https://github.com/kubernetes-sigs/gateway-api-inference-extension)
- [InferencePool](https://gateway-api-inference-extension.sigs.k8s.io/api-types/inferencepool/)
- [GAIE API overview](https://gateway-api-inference-extension.sigs.k8s.io/concepts/api-overview/)

Use official documentation and repositories as the source of truth. Blog posts
and old proposals may use superseded names or schemas.

## Local Versus Upstream

| Capability | Local exercise | Upstream llm-d |
|---|---|---|
| Proxy/EPP seam | in-process Go call | Envoy External Processing |
| body and streaming | bounded learning subset | upstream protocol and images |
| Parser | chat-completions subset | multiple request parsers |
| DataProducer | bounded routing headers | tokenization, prefix, load, and latency plugins |
| Flow Control | learner-built bounded priority/fairness | upstream stages and saturation behavior |
| Scheduler | learner-built Filter/Score/Pick profiles | configurable profile and plugin graph |
| Data Layer | file/event sources and Store | discovery, polling, and Endpoint attributes |
| Configuration | learning YAML | real EndpointPickerConfig lifecycle |
| GAIE | schema-learning manifests | installed CRDs and compatible Gateway |
| Protocol compatibility | intentionally incompatible | official EPP and Model Server contracts |

The local behavior is the exercise goal, not a reference implementation.
`configs/endpoint-picker-learning.yaml` is a learning configuration, not a CRD.
