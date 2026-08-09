// Versioned allowlist for the destructive reset control plane.
//
// A reset is intentionally scoped to one learning gate. Keep the paths explicit:
// adding an exercise file must be a reviewed catalog change, never an automatic
// directory expansion.
export const gateResetPaths = Object.freeze({
  F0: ["internal/foundation/language/data.go"],
  F1: ["internal/foundation/language/interfaces.go"],
  F2: ["internal/foundation/language/semantics.go"],
  F3: ["internal/foundation/language/errors.go"],
  F4: [
    "internal/foundation/testing/suite.go",
    "internal/foundation/testing/testcraft_test.go",
  ],

  E0: [
    "internal/commerce/domain/address.go",
    "internal/commerce/domain/money.go",
    "internal/commerce/domain/quantity.go",
  ],
  E1: [
    "internal/commerce/domain/events.go",
    "internal/commerce/domain/order.go",
  ],
  E2: ["internal/commerce/application/order_service.go"],
  E3: ["internal/commerce/persistence/memory_repo.go"],
  E4: ["internal/commerce/eventbus/eventbus.go"],
  E5: [
    "internal/commerce/http/handler.go",
    "internal/commerce/http/middleware.go",
    "internal/commerce/http/server.go",
  ],
  E6: ["internal/commerce/http/handler.go"],

  C0: [
    "internal/foundation/concurrency/concurrency.go",
    "internal/foundation/concurrency/advanced.go",
  ],
  C1: ["internal/foundation/concurrency/advanced.go"],
  C2: ["internal/foundation/concurrency/advanced.go"],
  C3: ["internal/foundation/concurrency/memorymodel.go"],
  K0: ["internal/inference/prefixcache/prefix.go"],

  Y0: [
    "deploy/inference-lab/base/router.yaml",
    "deploy/inference-lab/base/router-config.yaml",
    "deploy/inference-lab/base/models.yaml",
    "deploy/inference-lab/base/availability.yaml",
    "deploy/inference-lab/base/network-policy.yaml",
    "deploy/inference-lab/overlays/learning/router-replicas.yaml",
  ],
  P0: [
    "deploy/inference-lab/Dockerfile",
    "deploy/inference-lab/compose.yaml",
  ],
  H0: [
    "deploy/inference-lab/chart/templates/_helpers.tpl",
    "deploy/inference-lab/chart/templates/router.yaml",
    "deploy/inference-lab/chart/templates/models.yaml",
    "deploy/inference-lab/chart/templates/router-config.yaml",
    "deploy/inference-lab/chart/templates/router-pdb.yaml",
    "deploy/inference-lab/chart/templates/router-hpa.yaml",
    "deploy/inference-lab/chart/templates/network-policy.yaml",
    "deploy/inference-lab/chart/templates/serviceaccount.yaml",
    "deploy/inference-lab/chart/templates/tests/router-ready.yaml",
    "deploy/inference-lab/chart/templates/NOTES.txt",
  ],
  T0: [
    "test/acceptance/inference/00_smoke.robot",
    "test/acceptance/inference/01_routing.robot",
    "test/acceptance/inference/02_negative.robot",
    "test/acceptance/inference/03_production.robot",
    "test/acceptance/inference/resources/inference.resource",
  ],
  I0: [
    "deploy/inference-lab/istio/ingress.yaml",
    "deploy/inference-lab/istio/security.yaml",
    "deploy/inference-lab/istio/kustomization.yaml",
    "deploy/inference-lab/istio/namespace-sidecar-patch.yaml",
  ],
  O0: ["internal/inference/observability/metrics.go"],
  O1: [
    "deploy/inference-lab/observability/kustomization.yaml",
    "deploy/inference-lab/observability/operator.yaml",
    "deploy/inference-lab/observability/grafana/dashboards/inference-router.json",
  ],
  O2: [
    "deploy/inference-lab/observability/compose.yaml",
    "deploy/inference-lab/observability/prometheus/prometheus.yml",
    "deploy/inference-lab/observability/grafana/provisioning/datasources/prometheus.yaml",
    "deploy/inference-lab/observability/grafana/provisioning/dashboards/inference.yaml",
  ],
  O3: ["deploy/inference-lab/observability/prometheus/rules.yml"],
  G0: ["deploy/inference-lab/gaie/resources.yaml"],

  R0: [
    "internal/inference/datalayer/store.go",
    "internal/inference/datalayer/source.go",
  ],
  R1: [
    "internal/inference/routing/types.go",
    "internal/inference/routing/scheduler.go",
    "internal/inference/routing/plugins.go",
  ],
  R2: [
    "internal/inference/flowcontrol/queue.go",
    "internal/inference/flowcontrol/controller.go",
  ],
  R3: [
    "internal/inference/configuration/manager.go",
    "configs/endpoint-picker-learning.yaml",
  ],
  R4: ["internal/inference/epp/request_handler.go"],
  R5: ["internal/inference/epp/production_picker.go"],
  R6: [
    "internal/inference/httpapi/router.go",
    "internal/inference/httpapi/simulator.go",
  ],
  R7: ["internal/inference/runtime/runtime.go"],
  R8: ["cmd/inference-lab/main.go"],

  "OPT-SYNTAX": [
    "internal/foundation/language/optional_syntax_test.go",
    "internal/foundation/language/compile_rules_test.go",
    "internal/foundation/performance/bce.go",
    "internal/foundation/performance/unsafe.go",
  ],
  "OPT-CONCURRENCY": [
    "internal/foundation/concurrency/advanced.go",
    "internal/foundation/concurrency/channel.go",
  ],
  "OPT-PERFORMANCE": [
    "internal/foundation/performance/escape.go",
    "internal/foundation/performance/layout.go",
    "internal/foundation/performance/profiler.go",
    "internal/foundation/performance/reflectx.go",
  ],
  "OPT-FP": [
    "internal/foundation/language/fp.go",
    "internal/foundation/collection/collection.go",
    "internal/foundation/resilience/pipeline.go",
    "internal/foundation/resilience/retry.go",
    "internal/foundation/patterns/behavioral.go",
    "internal/foundation/patterns/creational.go",
    "internal/foundation/patterns/middleware.go",
    "internal/foundation/patterns/minigin.go",
    "internal/foundation/patterns/structural.go",
    "internal/foundation/fp/compose.go",
    "internal/foundation/fp/result.go",
    "internal/foundation/fp/iter/iter.go",
  ],
  "OPT-WORKFLOW": [
    "internal/commerce/workflows/fulfillment.go",
    "internal/commerce/workflows/inventory.go",
    "internal/commerce/workflows/saga.go",
    "internal/commerce/workflows/txdsl.go",
  ],
  "OPT-COMMERCE": [
    "internal/commerce/pricing/policy.go",
    "internal/commerce/pricing/pricing.go",
    "internal/commerce/pricing/promotion.go",
    "internal/commerce/pricing/spec.go",
    "internal/commerce/operations/eventsourcing.go",
    "internal/commerce/operations/inventory.go",
    "internal/commerce/operations/settlement.go",
    "internal/commerce/integrations/controller.go",
    "internal/commerce/integrations/order_grpc.go",
    "internal/commerce/integrations/reconciler.go",
    "internal/commerce/integrations/server.go",
    "internal/commerce/integrations/types.go",
  ],
});

export const gateIds = Object.freeze(Object.keys(gateResetPaths));

export function pathsForGate(gateId) {
  return gateResetPaths[String(gateId ?? "").toUpperCase()];
}

export function sharedGatesFor(gateId) {
  const normalized = String(gateId ?? "").toUpperCase();
  const paths = pathsForGate(normalized);
  if (!paths) return [];
  const owned = new Set(paths);
  return gateIds.filter((candidate) =>
    candidate !== normalized && pathsForGate(candidate).some((path) => owned.has(path)),
  );
}
