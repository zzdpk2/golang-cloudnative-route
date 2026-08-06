# Router Production Labs

This file contains detailed delivery and operations gates. Follow their order
from the [main learning path](../LEARNING_PATH.md). Every artifact starts
incomplete; the runner checks `TODO(exercise)` before tool and semantic
contracts.

During Phase 3, perform only local authoring, render, lint, schema, and dry-run
checks. Execute the runtime commands and failure drills below after R8. That
separation keeps Kubernetes and upstream integration grounded in Router
behavior that already works locally.

## Y0 — Kubernetes YAML and Kustomize

Work in `deploy/inference-lab/base` and `overlays/learning`.

```bash
kubectl kustomize deploy/inference-lab/overlays/learning
go test ./test/contracts/infrastructure -run TestKubernetesBase
```

The render must contain Router and two Model Server Deployments and Services,
stable selectors, named ports, probes, resources, non-root/read-only security,
graceful termination, multi-node placement, PDB, HPA, and default-deny
networking with only required traffic reopened.

Failure drill: break a Service selector and explain why rendering still passes;
restore it, prove Service-to-Pod selection, evict a Router Pod, and record the
availability and PDB result.

## P0 — Image and Compose Packaging

Implement `deploy/inference-lab/Dockerfile` and `compose.yaml`.

```bash
docker build -f deploy/inference-lab/Dockerfile -t inference-lab:dev .
docker history inference-lab:dev
docker compose -f deploy/inference-lab/compose.yaml config
docker compose -f deploy/inference-lab/compose.yaml up --build
go test ./test/contracts/infrastructure -run TestContainerPackaging
```

Build one multi-stage, non-root runtime image without the compiler, module
cache, or source tree. Compose runs one Router and two Model Servers from that
image with explicit roles, health contracts, limits, read-only filesystems,
dropped capabilities, mounted configuration, and deliberate host exposure.

Failure drill: stop one model, observe freshness and readiness, then record
overload, cancellation, signal handling, and recovery. Do not add blind retries.

## H0 — Helm Chart

Work in `deploy/inference-lab/chart`.

```bash
helm lint deploy/inference-lab/chart \
  -f deploy/inference-lab/chart/values-learning.yaml
helm template inference-lab deploy/inference-lab/chart \
  -f deploy/inference-lab/chart/values-learning.yaml
go test ./test/contracts/infrastructure -run TestHelm
```

Render Router and model workloads/services, ConfigMap, ServiceAccount, PDB,
HPA, NetworkPolicy, stable selectors, values-driven resources, security, and a
Helm test Pod. Add a model through values only. Retain install, upgrade, test,
history, and rollback evidence from a disposable cluster.

## T0 — Robot Framework Acceptance

Implement `test/acceptance/inference` and its shared resource.

```bash
robot --dryrun --output NONE --log NONE --report NONE \
  test/acceptance/inference
robot --outputdir .robot-results/inference test/acceptance/inference
go test ./test/contracts/infrastructure -run TestRobotContracts
```

Observe only public HTTP behavior. Cover smoke/readiness, deterministic routing
evidence, stale Endpoints, overload bounds, cancellation, streaming, mid-stream
failure, Endpoint loss, and recovery. Failure messages describe what the client
sent and observed, never private Go calls.

## I0 — Istio Ingress and Mesh Security

Work in `deploy/inference-lab/istio`.

```bash
kubectl kustomize deploy/inference-lab/istio
go test ./test/contracts/infrastructure -run TestIstio
```

Select a real GatewayClass, expose only inference and deliberate liveness
paths, keep diagnostics private, state mesh enrollment, require strict mTLS,
use `ISTIO_MUTUAL` for model traffic, and constrain principals and operations.

Verify programmed Route status, a wrong backend reference, denied diagnostic
paths, mTLS policy, and proxy configuration distribution.

## O0 — Router Metrics

Implement `internal/inference/observability`.

The contracts require request count/duration, active requests, decisions,
selected score, queue depth, KV-cache utilization, and Endpoint lifecycle.
Labels remain bounded: never use request IDs, prompts, arbitrary tenant
strings, full URLs, or changing Pod UIDs. Remove label values with Endpoint
state.

## O1 — Prometheus Operator and Dashboard

Work in `deploy/inference-lab/observability/operator.yaml`, its Kustomization,
and the Grafana dashboard JSON.

```bash
kubectl kustomize deploy/inference-lab/observability
go test ./test/contracts/infrastructure -run TestOperator
```

Render a ServiceMonitor selected through stable labels and a named port, a
PrometheusRule with recording and alert rules, and a dashboard ConfigMap. Prove
Prometheus selects the resources, the target is up, and Grafana consumes the
dashboard.

## O2 — Local Prometheus and Grafana

Implement the observability Compose extension, Prometheus scrape config, and
file-provisioned Grafana datasource/dashboard.

```bash
docker compose \
  -f deploy/inference-lab/compose.yaml \
  -f deploy/inference-lab/observability/compose.yaml config
go test ./test/contracts/infrastructure -run TestLocalObservability
```

Use loopback-only host exposure. Generate traffic and distinguish request rate,
decisions, latency quantiles, errors, queueing, and Endpoint state. Do not rely
on UI-only configuration.

## O3 — PromQL and Alert Behavior

Write rules in `deploy/inference-lab/observability/prometheus/rules.yml`.

```bash
promtool check rules deploy/inference-lab/observability/prometheus/rules.yml
go test ./test/contracts/infrastructure -run TestPrometheusRules
```

Handle zero denominators and low-traffic noise. Every alert needs a positive
pending period and a runbook action. Inject its failure, observe firing,
recover, and observe resolution.

## G0 — Gateway API Inference Extension

Complete `deploy/inference-lab/gaie/resources.yaml`.

```bash
kubectl kustomize deploy/inference-lab/gaie
go test ./test/contracts/infrastructure -run TestGAIE
```

Implement the stable InferencePool selector, target ports, Endpoint Picker
reference, explicit failure mode, and an HTTPRoute backend reference to the
pool. Pin the installed Gateway and GAIE release. This is migration vocabulary,
not proof of upstream Model Server or EPP protocol compatibility.
