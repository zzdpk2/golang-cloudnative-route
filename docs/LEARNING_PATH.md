# Commerce to llm-d Router Learning Path

This is the repository's only required exercise order. Use the dynamic
[SvelteKit learning console](../web/learning-path/README.md), open the
zero-install [HTML tracker](learning-path.html), or run the executable source
of truth:

```bash
go run ./cmd/exercise next
```

For each reported gate, read the linked test and hints, implement the TODO
yourself, run the focused contract, and ask for `next` again. Do not search
later modules for an answer. `go run ./cmd/exercise check` is the final
required-path gate; starred material never blocks it.

Every `next`, status, or `check` run writes the currently passing gate IDs to
`docs/learning-progress.js`. The HTML tracker reloads that evidence every two
seconds; use **Sync test progress** for an immediate refresh. Passing gates are
marked test-verified and the progress bar is updated.
Manual checks remain browser-local and cannot unlock a test-verified gate.

The path also applies the engineering review lens from
[*100 Go Mistakes and How to Avoid Them*](go100/README.md). The runner prints a
targeted mistake range for every gate, and the HTML card shows the same prompt.
This is not another track: use the prompt to review the real Commerce,
concurrency, infrastructure, and Router implementation you are already doing.

```text
Phase 1             Phase 2          Phase 3                 Phase 4
Go foundations  ->  Commerce Order -> llm-d prerequisites -> llm-d Router
                                                               |
                                                               v
                                                        upstream capstone
```

## Phase 1: Go Foundations for Commerce

Learn each rule once, then apply it in the Order project. These compact gates
cover the book's language-facing risks; the later phases revisit them through
production ownership, lifetime, error, testing, and runtime decisions.

| Gate | Exercise and contract | Focused command | Exit knowledge |
|---|---|---|---|
| F0 | [`data.go`](../internal/foundation/language/data.go) / [`data_test.go`](../internal/foundation/language/data_test.go) | `go test ./internal/foundation/language -run '^TestDataFoundations$' -count=1` | ownership, UTF-8, deterministic maps, equality |
| F1 | [`interfaces.go`](../internal/foundation/language/interfaces.go) / [`interfaces_test.go`](../internal/foundation/language/interfaces_test.go) | `go test ./internal/foundation/language -run '^TestInterfaceFoundations$' -count=1` | method sets, typed nil, narrow adapters, streaming |
| F2 | [`semantics.go`](../internal/foundation/language/semantics.go) / [`semantics_test.go`](../internal/foundation/language/semantics_test.go) | `go test ./internal/foundation/language -run 'Test(ValuePartsAndAddressability\|DeferPanicAndScope\|NilMapAndChannelSemantics)$' -count=1` | value parts, defer/panic, nil map and channel rules |
| F3 | [`errors.go`](../internal/foundation/language/errors.go) / [`errors_test.go`](../internal/foundation/language/errors_test.go) | `go test ./internal/foundation/language -run 'TestDomainError_(Is\|As\|Unwrap\|DeepWrapping)$' -count=1` | stable error identity through wrapping |
| F4 | [`suite.go`](../internal/foundation/testing/suite.go) / [`testcraft_test.go`](../internal/foundation/testing/testcraft_test.go) | `go test ./internal/foundation/testing -run '^(Test(SplitSuite_(PassesTheCorrectImplementation\|CatchesEveryMutant)\|Check(SumsToTotal\|NearlyEqual\|RemainderGoesFirst)\|SplitEvenly_(Examples\|Errors))\|FuzzSplit)$' -count=1` | tables, properties, fuzzing, mutation value |

## Phase 2: Commerce Order Project

Milestone tests are the deep interface; package tests diagnose failures.

| Gate | Implementation area | Contract command | Outcome |
|---|---|---|---|
| E0 | [`money.go`](../internal/commerce/domain/money.go), [`quantity.go`](../internal/commerce/domain/quantity.go), [`address.go`](../internal/commerce/domain/address.go) | `go test ./test/e2e -run '^TestM1_' -count=1` | Money, Quantity, and Address protect invariants |
| E1 | [`order.go`](../internal/commerce/domain/order.go), [`events.go`](../internal/commerce/domain/events.go) | `go test ./test/e2e -run '^TestM2_' -count=1` | Order lifecycle and event ownership |
| E2 | [`order_service.go`](../internal/commerce/application/order_service.go), [`doubles.go`](../internal/commerce/testkit/doubles.go) | `go test ./test/e2e -run '^TestM3_' -count=1` | Create Order orchestration through ports |
| E3 | [`memory_repo.go`](../internal/commerce/persistence/memory_repo.go) | `go test ./test/e2e -run '^TestM4_' -count=1` | defensive snapshots and concurrent persistence |
| E4 | [`eventbus.go`](../internal/commerce/eventbus/eventbus.go) | `go test ./test/e2e -run '^TestM5_' -count=1` | domain events reach every subscriber |
| E5 | [`handler.go`](../internal/commerce/http/handler.go), [`orderd/main.go`](../cmd/orderd/main.go) | `go test ./test/e2e -run '^TestM6_' -count=1` | a real HTTP Order lifecycle works |
| E6 | [`handler.go`](../internal/commerce/http/handler.go), [`milestones_test.go`](../test/e2e/milestones_test.go) | `go test ./test/e2e -run '^TestM7_' -count=1` | domain failures become stable external errors |

After E6, run the Commerce black-box suite:

```bash
robot --outputdir .robot-results test/acceptance
```

## Phase 3: llm-d Prerequisites

Only prerequisites used by the Router are required. Cluster rollout and
failure drills wait until R8, when the local Router has real behavior.

### Concurrency and KV-cache

| Gate | Exercise and contract | Focused command | Exit knowledge |
|---|---|---|---|
| C0 | [`concurrency.go`](../internal/foundation/concurrency/concurrency.go) / [`concurrency_test.go`](../internal/foundation/concurrency/concurrency_test.go) | `go test ./internal/foundation/concurrency -run 'Test(SafeCounter\|Advanced_RWMutexCacheConcurrentAccess)$' -count=1` | lock ownership and immutable snapshots |
| C1 | [`advanced.go`](../internal/foundation/concurrency/advanced.go) / [`advanced_test.go`](../internal/foundation/concurrency/advanced_test.go) | `go test ./internal/foundation/concurrency -run 'TestAdvanced_(Semaphore\|BoundedWorkerPool)' -count=1` | hard bounds, cancellation, backpressure |
| C2 | [`advanced.go`](../internal/foundation/concurrency/advanced.go) / [`advanced_test.go`](../internal/foundation/concurrency/advanced_test.go) | `go test ./internal/foundation/concurrency -run 'TestAdvanced_(AtomicCounter\|AtomicFlagOnlyOneWinner)$' -count=1` | atomic counters and one-winner CAS |
| C3 | [`memorymodel.go`](../internal/foundation/concurrency/memorymodel.go) / [`memorymodel_test.go`](../internal/foundation/concurrency/memorymodel_test.go) | `go test ./internal/foundation/concurrency -run 'Test(OnceCellPublishesOneValue\|ChannelClosePublishesEarlierWrites)$' -count=1` | happens-before publication |
| K0 | [`prefix.go`](../internal/inference/prefixcache/prefix.go) / [`prefix_test.go`](../internal/inference/prefixcache/prefix_test.go) | `go test ./internal/inference/prefixcache -run '^TestPrefixCacheFoundations$' -count=1` | KV blocks, chained hashes, collision-safe prefix matching |

Run `go test -race ./internal/foundation/concurrency -count=20` on a
CGO-enabled Linux environment before continuing.

### YAML, packaging, mesh, and operations

Every artifact gate is conjunctive: the tool must accept the artifact and its Go
infrastructure contract must pass. `go run ./cmd/exercise -v next` prints the
exact composite command.

| Gate | Clickable work area | Tool verification | Go contract |
|---|---|---|---|
| Y0 | [`base/kustomization.yaml`](../deploy/inference-lab/base/kustomization.yaml), [`learning overlay`](../deploy/inference-lab/overlays/learning/kustomization.yaml) | `kubectl kustomize deploy/inference-lab/overlays/learning` | `TestKubernetesBase` |
| P0 | [`Dockerfile`](../deploy/inference-lab/Dockerfile), [`compose.yaml`](../deploy/inference-lab/compose.yaml) | `docker compose -f deploy/inference-lab/compose.yaml config --quiet` | `TestContainerPackaging` |
| H0 | [`Chart.yaml`](../deploy/inference-lab/chart/Chart.yaml), [`values-learning.yaml`](../deploy/inference-lab/chart/values-learning.yaml) | `helm lint deploy/inference-lab/chart -f deploy/inference-lab/chart/values-learning.yaml` | `TestHelm` |
| T0 | [`00_smoke.robot`](../test/acceptance/inference/00_smoke.robot), [`inference.resource`](../test/acceptance/inference/resources/inference.resource) | `robot --dryrun --output NONE --log NONE --report NONE test/acceptance/inference` | `TestRobotContracts` |
| I0 | [`kustomization.yaml`](../deploy/inference-lab/istio/kustomization.yaml), [`ingress.yaml`](../deploy/inference-lab/istio/ingress.yaml) | `kubectl kustomize deploy/inference-lab/istio` | `TestIstio` |
| O0 | [`metrics.go`](../internal/inference/observability/metrics.go), [`metrics_test.go`](../internal/inference/observability/metrics_test.go) | `go test ./internal/inference/observability -run 'Test(Metrics\|Instrument\|UpdateEndpoints)' -count=1` | — |
| O1 | [`Operator resources`](../deploy/inference-lab/observability/operator.yaml), [`Grafana dashboard`](../deploy/inference-lab/observability/grafana/dashboards/inference-router.json) | `kubectl kustomize deploy/inference-lab/observability` | `TestOperator` |
| O2 | [`local observability`](../deploy/inference-lab/observability/compose.yaml) | `docker compose -f deploy/inference-lab/compose.yaml -f deploy/inference-lab/observability/compose.yaml config --quiet` | `TestLocalObservability` |
| O3 | [`Prometheus rules`](../deploy/inference-lab/observability/prometheus/rules.yml) | `promtool check rules deploy/inference-lab/observability/prometheus/rules.yml` | `TestPrometheusRules` |
| G0 | [`GAIE resources`](../deploy/inference-lab/gaie/resources.yaml) | `kubectl kustomize deploy/inference-lab/gaie` | `TestGAIE` |

See [Router production labs](llmd/PRODUCTION_LABS.md) for failure drills and
evidence. It adds depth, not a competing sequence.

## Phase 4: llm-d Router

The required route uses fresh Data Layer state and the Production Picker only.

| Gate | Implementation and tests | Focused outcome |
|---|---|---|
| R0 | [`store.go`](../internal/inference/datalayer/store.go), [`source.go`](../internal/inference/datalayer/source.go), [`store_test.go`](../internal/inference/datalayer/store_test.go) | fresh ordered Endpoint state, tombstones, cancellable sources |
| R1 | [`types.go`](../internal/inference/routing/types.go), [`scheduler.go`](../internal/inference/routing/scheduler.go), [`scheduler_test.go`](../internal/inference/routing/scheduler_test.go) | deterministic Filter, Score, Pick with explainability |
| R2 | [`queue.go`](../internal/inference/flowcontrol/queue.go), [`controller.go`](../internal/inference/flowcontrol/controller.go), [`controller_test.go`](../internal/inference/flowcontrol/controller_test.go) | bounded fairness, TTL, cancellation, permit transfer |
| R3 | [`manager.go`](../internal/inference/configuration/manager.go), [`manager_test.go`](../internal/inference/configuration/manager_test.go), [`learning config`](../configs/endpoint-picker-learning.yaml) | validate, atomically publish, retain, rollback |
| R4 | [`request_handler.go`](../internal/inference/epp/request_handler.go), [`request_handler_test.go`](../internal/inference/epp/request_handler_test.go) | parse, derive state, admit, reverse partial acquisition |
| R5 | [`production_picker.go`](../internal/inference/epp/production_picker.go), [`production_picker_test.go`](../internal/inference/epp/production_picker_test.go) | one decision uses state, scheduling, flow control, and config |
| R6 | [`router.go`](../internal/inference/httpapi/router.go), [`router_test.go`](../internal/inference/httpapi/router_test.go) | Proxy limits, cancellation, headers, streaming, error mapping |
| R7 | [`runtime.go`](../internal/inference/runtime/runtime.go), [`runtime_test.go`](../internal/inference/runtime/runtime_test.go) | real composition plus config-generation capacity ownership |
| R8 | [`main.go`](../cmd/inference-lab/main.go), [`main_test.go`](../cmd/inference-lab/main_test.go) | Router/Model composition, group failure, graceful drain |

Run `go test -race ./internal/inference/... -count=20`, then follow the
[pinned upstream map](llmd/UPSTREAM.md).

Before leaving each phase, use the
[targeted 100 Go Mistakes clinics](go100/README.md) to produce review evidence.
The clinics add no implementation files and contain no solutions; they point
back to the contracts in this path.

## What Was Consolidated

- The previous `lab`, `platform`, domain-layer, adapter, workflow, and actor
  trees were merged into `foundation` and `commerce` ownership areas.
- Duplicate basic channel, collection, semantics, and delayed-review exercises
  were removed after their knowledge moved into stronger canonical contracts.
- The obsolete parallel routing scaffold was removed; R5-R7 now teach the
  production state path directly.
- Inference module boundaries remain explicit because they map to real llm-d
  responsibilities: Data Layer, scheduling, flow control, config, EPP, Proxy,
  observability, and runtime.

## Starred Optional Extensions

Optional extensions are dashed branches inside the phase where they belong.
They may be completed at any time or skipped entirely: they never change the
required `36/36` percentage and never block `next`.

| Optional branch | Phase | Source | Evidence command |
| --- | --- | --- | --- |
| `OPT-SYNTAX` · niche syntax and compiler details | 1 | [`optional_syntax_test.go`](../internal/foundation/language/optional_syntax_test.go) | `go test -tags=optional ./internal/foundation/language -run '^TestOptionalSyntaxFoundations$'` plus the compile-failure and BCE checks shown by the UI |
| `OPT-CONCURRENCY` · extended concurrency patterns | 1 | [`advanced_test.go`](../internal/foundation/concurrency/advanced_test.go) | `go test ./internal/foundation/concurrency -run 'Bridge\|Tee\|AnyDone\|ForwardUntilDone\|TryMutex\|Striped\|SingleFlight\|CircuitBreaker\|CondQueue'` |
| `OPT-PERFORMANCE` · runtime and performance details | 1 | [`performance`](../internal/foundation/performance/) | `go test ./internal/foundation/performance` |
| `OPT-FP` · collections, resilience, patterns, and functional programming | 1 | [`result_test.go`](../internal/foundation/fp/result_test.go) | `go test ./internal/foundation/collection ./internal/foundation/resilience ./internal/foundation/patterns ./internal/foundation/fp/...` |
| `OPT-WORKFLOW` · workflow and alternative concurrency models | 2 | [`workflows`](../internal/commerce/workflows/) | `go test ./internal/commerce/workflows` |
| `OPT-COMMERCE` · advanced Commerce extensions | 2 | [`promotion_test.go`](../internal/commerce/pricing/promotion_test.go) | `go test ./internal/commerce/pricing ./internal/commerce/operations ./internal/commerce/integrations` |

Run `go run ./cmd/exercise optional` to execute all six groups and publish their
independent test evidence to the UI. A failure from that explicit command says
only that an optional group is open; the required learning route remains
unblocked.

## Remote Gate Reset

Each exercise card can queue a protected gate reset through a small Vercel
control plane. The browser never receives a GitHub token and never stores the
Reset Key. A successful request dispatches `reset-gate.yml`, which restores only
the selected gate's explicit, versioned file allowlist, commits it to `main`, and
dispatches the Pages workflow to recalculate test evidence. Some older gates
share a source file; the confirmation dialog calls out that file-level boundary.

Configure the Vercel project at the repository root with these environment
variables:

- `RESET_KEY`: a random secret of at least 32 characters.
- `GITHUB_DISPATCH_TOKEN`: a fine-grained token created by the repository owner,
  limited to this repository, with Actions read/write permission. The Vercel
  endpoint verifies the token owner before every dispatch.
- `GITHUB_REPOSITORY`: `zzdpk2/golang-cloudnative-route`.
- `ALLOWED_ORIGINS`: comma-separated exact origins, normally
  `https://zzdpk2.github.io,http://127.0.0.1:3000,http://localhost:3000`.

After deploying Vercel, create the GitHub Actions repository variable
`RESET_API_URL` with the deployment origin, for example
`https://route-learning-reset.vercel.app`. Re-run the Pages workflow so the
static frontend receives that public API origin.

After a reset workflow completes, synchronize a clean local checkout with:

```powershell
git pull --ff-only
```

Git refuses the fast-forward when local uncommitted work would be overwritten.
Commit or explicitly archive that work before retrying; never automate a forced
local reset.
