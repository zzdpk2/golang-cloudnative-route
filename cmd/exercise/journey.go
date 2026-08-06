// Learning journey definitions and focused gate execution.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// learningCheck is intentionally test-focused instead of package-focused.
// A package may contain many ★ optional exercises; only the named behavior is
// allowed to block the single learning journey.
type learningCheck struct {
	id    string
	title string
	pkg   string
	run   string

	program string
	args    []string
	command string

	todoPaths   []string
	contractRun string
}

var checkCatalog = []learningCheck{
	{
		id:    "F0",
		title: "Values, slices, maps, and UTF-8",
		pkg:   "./internal/foundation/language",
		run:   "^TestDataFoundations$",
	},
	{
		id:    "F1",
		title: "Interfaces, method sets, and typed nil",
		pkg:   "./internal/foundation/language",
		run:   "^TestInterfaceFoundations$",
	},
	{
		id:    "F2",
		title: "Value semantics, defer, and nil rules",
		pkg:   "./internal/foundation/language",
		run:   "Test(ValuePartsAndAddressability|DeferPanicAndScope|NilMapAndChannelSemantics)$",
	},
	{
		id:    "F3",
		title: "Error chains: Is / As / Unwrap",
		pkg:   "./internal/foundation/language",
		run:   "TestDomainError_(Is|As|Unwrap|DeepWrapping)$",
	},
	{
		id:    "F4",
		title: "Table, property, fuzz, and mutation tests",
		pkg:   "./internal/foundation/testing",
		run: "^(Test(SplitSuite_(PassesTheCorrectImplementation|CatchesEveryMutant)|" +
			"Check(SumsToTotal|NearlyEqual|RemainderGoesFirst)|SplitEvenly_(Examples|Errors))|FuzzSplit)$",
	},
	{
		id:    "E0",
		title: "Commerce values: Money, Quantity, and Address",
		pkg:   "./test/e2e",
		run:   "^TestM1_",
	},
	{
		id:    "E1",
		title: "Order aggregate lifecycle and domain events",
		pkg:   "./test/e2e",
		run:   "^TestM2_",
	},
	{
		id:    "E2",
		title: "Create Order application use case",
		pkg:   "./test/e2e",
		run:   "^TestM3_",
	},
	{
		id:    "E3",
		title: "Repository ownership and concurrent persistence",
		pkg:   "./test/e2e",
		run:   "^TestM4_",
	},
	{
		id:    "E4",
		title: "Domain event publication",
		pkg:   "./test/e2e",
		run:   "^TestM5_",
	},
	{
		id:    "E5",
		title: "Commerce HTTP request path",
		pkg:   "./test/e2e",
		run:   "^TestM6_",
	},
	{
		id:    "E6",
		title: "Stable external error mapping",
		pkg:   "./test/e2e",
		run:   "^TestM7_",
	},
	{
		id:    "C0",
		title: "Shared state: mutexes, snapshots, and ownership",
		pkg:   "./internal/foundation/concurrency",
		run:   "Test(SafeCounter|Advanced_RWMutexCacheConcurrentAccess)$",
	},
	{
		id:    "C1",
		title: "Backpressure: semaphores and bounded worker pools",
		pkg:   "./internal/foundation/concurrency",
		run: "TestAdvanced_(SemaphoreLimitsConcurrency|SemaphoreTryAcquireAndCancel|" +
			"BoundedWorkerPoolRunsJobsAndPreservesOrder|BoundedWorkerPoolLimitsConcurrency)$",
	},
	{
		id:    "C2",
		title: "Atomic state: counters and CAS flags",
		pkg:   "./internal/foundation/concurrency",
		run:   "TestAdvanced_(AtomicCounter|AtomicFlagOnlyOneWinner)$",
	},
	{
		id:    "C3",
		title: "Memory visibility: Once and channel publication",
		pkg:   "./internal/foundation/concurrency",
		run:   "Test(OnceCellPublishesOneValue|ChannelClosePublishesEarlierWrites)$",
	},
	{
		id:    "K0",
		title: "KV-cache blocks, chained hashes, and prefix matching",
		pkg:   "./internal/inference/prefixcache",
		run:   "^TestPrefixCacheFoundations$",
	},
	{
		id:    "R0",
		title: "Data Layer: freshness, ordered updates, and tombstones",
		pkg:   "./internal/inference/datalayer",
		run:   "Test(Store|FileSource|EventSource)",
		todoPaths: []string{
			"internal/inference/datalayer/store.go",
			"internal/inference/datalayer/source.go",
		},
	},
	{
		id:    "R1",
		title: "Scheduling Profile: Filter, Score, and Pick",
		pkg:   "./internal/inference/routing",
		run:   "Test(DefaultScheduler|Scheduler|MaxScore)",
		todoPaths: []string{
			"internal/inference/routing/types.go",
			"internal/inference/routing/scheduler.go",
			"internal/inference/routing/plugins.go",
		},
	},
	{
		id:    "R2",
		title: "Flow Control: bounded queue, fairness, and cancellation",
		pkg:   "./internal/inference/flowcontrol",
		run:   "Test(Queue|Controller)",
		todoPaths: []string{
			"internal/inference/flowcontrol/queue.go",
			"internal/inference/flowcontrol/controller.go",
		},
	},
	{
		id:    "R3",
		title: "Versioned Config: atomic publication and rollback",
		pkg:   "./internal/inference/configuration",
		run:   "Test(Manager|Publish|Rollback|Concurrent)",
		todoPaths: []string{
			"internal/inference/configuration/manager.go",
			"configs/endpoint-picker-learning.yaml",
		},
		contractRun: "TestRuntimeConfig",
	},
	{
		id:    "R4",
		title: "Request Handler: parsing and request-scoped admission",
		pkg:   "./internal/inference/epp",
		run:   "Test(RequestHandler|OpenAI|RoutingHeader)",
		todoPaths: []string{
			"internal/inference/epp/request_handler.go",
		},
	},
	{
		id:    "R5",
		title: "Production Picker: state + policy + admission + config",
		pkg:   "./internal/inference/epp",
		run:   "TestProductionPicker",
		todoPaths: []string{
			"internal/inference/epp/production_picker.go",
		},
	},
	{
		id:    "R6",
		title: "Proxy forwarding, cancellation, and streaming",
		pkg:   "./internal/inference/httpapi",
		run:   "Test(Router|Simulator)",
		todoPaths: []string{
			"internal/inference/httpapi/router.go",
			"internal/inference/httpapi/simulator.go",
		},
	},
	{
		id:    "R7",
		title: "Production Runtime: real modules drive HTTP behavior",
		pkg:   "./internal/inference/runtime",
		run:   "TestRuntime",
		todoPaths: []string{
			"internal/inference/runtime/runtime.go",
		},
	},
	{
		id:    "R8",
		title: "Executable harness: composition, group failure, and drain",
		pkg:   "./cmd/inference-lab",
		run:   "Test",
		todoPaths: []string{
			"cmd/inference-lab/main.go",
		},
	},
	{
		id:      "Y0",
		title:   "YAML and Kustomize rendering",
		program: "kubectl",
		args:    []string{"kustomize", "deploy/inference-lab/overlays/learning"},
		command: "kubectl kustomize deploy/inference-lab/overlays/learning",
		todoPaths: []string{
			"deploy/inference-lab/base/router.yaml",
			"deploy/inference-lab/base/router-config.yaml",
			"deploy/inference-lab/base/models.yaml",
			"deploy/inference-lab/base/availability.yaml",
			"deploy/inference-lab/base/network-policy.yaml",
			"deploy/inference-lab/overlays/learning/router-replicas.yaml",
		},
		contractRun: "TestKubernetesBase",
	},
	{
		id:      "P0",
		title:   "Compose microservice packaging",
		program: "docker",
		args:    []string{"compose", "-f", "deploy/inference-lab/compose.yaml", "config", "--quiet"},
		command: "docker compose -f deploy/inference-lab/compose.yaml config --quiet",
		todoPaths: []string{
			"deploy/inference-lab/Dockerfile",
			"deploy/inference-lab/compose.yaml",
		},
		contractRun: "TestContainerPackaging",
	},
	{
		id:      "H0",
		title:   "Helm Chart lint and templating",
		program: "helm",
		args:    []string{"lint", "deploy/inference-lab/chart", "-f", "deploy/inference-lab/chart/values-learning.yaml"},
		command: "helm lint deploy/inference-lab/chart -f deploy/inference-lab/chart/values-learning.yaml",
		todoPaths: []string{
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
		},
		contractRun: "TestHelm",
	},
	{
		id:      "T0",
		title:   "Robot black-box acceptance syntax and keywords",
		program: "robot",
		args: []string{
			"--dryrun", "--output", "NONE", "--log", "NONE", "--report", "NONE",
			"test/acceptance/inference",
		},
		command: "robot --dryrun --output NONE --log NONE --report NONE test/acceptance/inference",
		todoPaths: []string{
			"test/acceptance/inference/00_smoke.robot",
			"test/acceptance/inference/01_routing.robot",
			"test/acceptance/inference/02_negative.robot",
			"test/acceptance/inference/03_production.robot",
			"test/acceptance/inference/resources/inference.resource",
		},
		contractRun: "TestRobotContracts",
	},
	{
		id:      "I0",
		title:   "Istio Gateway API rendering and security",
		program: "kubectl",
		args:    []string{"kustomize", "deploy/inference-lab/istio"},
		command: "kubectl kustomize deploy/inference-lab/istio",
		todoPaths: []string{
			"deploy/inference-lab/istio/ingress.yaml",
			"deploy/inference-lab/istio/security.yaml",
			"deploy/inference-lab/istio/kustomization.yaml",
			"deploy/inference-lab/istio/namespace-sidecar-patch.yaml",
		},
		contractRun: "TestIstio",
	},
	{
		id:    "O0",
		title: "Router Prometheus metrics and bounded labels",
		pkg:   "./internal/inference/observability",
		run:   "Test(Metrics|Instrument|UpdateEndpoints)",
		todoPaths: []string{
			"internal/inference/observability/metrics.go",
		},
	},
	{
		id:      "O1",
		title:   "Prometheus Operator and Grafana ConfigMap rendering",
		program: "kubectl",
		args:    []string{"kustomize", "deploy/inference-lab/observability"},
		command: "kubectl kustomize deploy/inference-lab/observability",
		todoPaths: []string{
			"deploy/inference-lab/observability/kustomization.yaml",
			"deploy/inference-lab/observability/operator.yaml",
			"deploy/inference-lab/observability/grafana/dashboards/inference-router.json",
		},
		contractRun: "TestOperator",
	},
	{
		id:      "O2",
		title:   "Prometheus/Grafana Compose configuration",
		program: "docker",
		args: []string{
			"compose",
			"-f", "deploy/inference-lab/compose.yaml",
			"-f", "deploy/inference-lab/observability/compose.yaml",
			"config", "--quiet",
		},
		command: "docker compose -f deploy/inference-lab/compose.yaml -f deploy/inference-lab/observability/compose.yaml config --quiet",
		todoPaths: []string{
			"deploy/inference-lab/observability/compose.yaml",
			"deploy/inference-lab/observability/prometheus/prometheus.yml",
			"deploy/inference-lab/observability/grafana/provisioning/datasources/prometheus.yaml",
			"deploy/inference-lab/observability/grafana/provisioning/dashboards/inference.yaml",
		},
		contractRun: "TestLocalObservability",
	},
	{
		id:      "O3",
		title:   "Prometheus recording/alert rules",
		program: "promtool",
		args: []string{
			"check", "rules",
			"deploy/inference-lab/observability/prometheus/rules.yml",
		},
		command: "promtool check rules deploy/inference-lab/observability/prometheus/rules.yml",
		todoPaths: []string{
			"deploy/inference-lab/observability/prometheus/rules.yml",
		},
		contractRun: "TestPrometheusRules",
	},
	{
		id:      "G0",
		title:   "Gateway API Inference Extension resource contract",
		program: "kubectl",
		args:    []string{"kustomize", "deploy/inference-lab/gaie"},
		command: "kubectl kustomize deploy/inference-lab/gaie",
		todoPaths: []string{
			"deploy/inference-lab/gaie/resources.yaml",
		},
		contractRun: "TestGAIE",
	},
}

// learningJourney is the only required order. The catalog keeps verbose check
// definitions readable; this list is the one sequencing interface consumed by
// next, check, the Markdown guide, and the HTML tracker.
var learningJourney = mustOrderChecks(checkCatalog, []string{
	"F0", "F1", "F2", "F3", "F4",
	"E0", "E1", "E2", "E3", "E4", "E5", "E6",
	"C0", "C1", "C2", "C3", "K0",
	"Y0", "P0", "H0", "T0", "I0", "O0", "O1", "O2", "O3", "G0",
	"R0", "R1", "R2", "R3", "R4", "R5", "R6", "R7", "R8",
})

// mistakeFocusByGate keeps the book review lens attached to the real exercise
// path. These are review prompts, not implementation recipes: learners still
// have to discover the design from the contract and prove it with tests.
var mistakeFocusByGate = map[string]string{
	"F0": "#20-29, #33, #36-41: slice/map ownership, deterministic iteration, and UTF-8",
	"F1": "#5-10, #42, #45-46: consumer-owned interfaces, method sets, typed nil, and Reader inputs",
	"F2": "#1-3, #30-35, #47: shadowing, range copies, loop defer, and evaluation timing",
	"F3": "#48-54: panic boundaries, wrapping, identity, and handling each error once",
	"F4": "#82-90 + community fuzzing: test categories, race mode, tables, controlled time, and benchmarks",
	"E0": "#18-19, #29, #42: numeric bounds, exact monetary equality, and receiver choice",
	"E1": "#24-26, #30, #42: defensive event copies, range values, and aggregate receiver semantics",
	"E2": "#5-8, #46, #49-54: use-case-owned ports, Reader/Writer boundaries, and error ownership",
	"E3": "#24-28, #57-58, #69-70, #74: snapshot copies, maps under locks, races, and sync ownership",
	"E4": "#62, #64-67, #71: goroutine termination, channel semantics, capacity, and group completion",
	"E5": "#75-81: typed durations, strict JSON, body closure, response flow, and explicit HTTP servers",
	"E6": "#48-54, #77, #80: stable error identity, strict JSON envelopes, and terminal responses",
	"C0": "#57-58, #69-70, #74: choose locks deliberately; never leak aliases or copy sync values",
	"C1": "#55-67: bounded concurrency, cancellation, notification, and explicit channel capacity",
	"C2": "#58-59, #91-94: race-free atomics, workload evidence, cache lines, and alignment",
	"C3": "#58, #65, #74: happens-before publication, close notification, and non-copyable state",
	"K0": "#20-29, #39-41, #89: allocation ownership, retained backing storage, and honest benchmarks",
	"Y0": "#100: make CPU, memory, probes, shutdown, and configuration explicit in Kubernetes",
	"P0": "#79, #81, #100: own process resources, HTTP shutdown, and container runtime limits",
	"H0": "#75, #81, #100: typed operational values, safe defaults, resources, and rollout behavior",
	"T0": "#82-90: black-box test categories, deterministic waits, utilities, and useful failure evidence",
	"I0": "#81, #100: explicit network timeouts, trust boundaries, probes, and sidecar resource cost",
	"O0": "#27, #68-70, #96-99: bounded labels, synchronized maps, allocations, diagnostics, and GC cost",
	"O1": "#82, #98, #100: production diagnostics must be deployable, queryable, and resource-aware",
	"O2": "#81, #98, #100: explicit service configuration and reproducible local diagnostics",
	"O3": "#82, #86-87, #98: deterministic rule tests, controlled time, and actionable diagnostics",
	"G0": "#12, #15, #100: clear resource ownership, documented contracts, and Kubernetes consequences",
	"R0": "#24-28, #58, #60-62, #69-70: immutable snapshots, context lifetime, and locked ownership",
	"R1": "#20-29, #33, #57-59, #89: deterministic candidates, no aliasing, and benchmarkable policy",
	"R2": "#57-67, #70-76: bounded queues, cancellation, channel ownership, timers, and no permit leaks",
	"R3": "#24-28, #58, #70, #74: defensive configuration copies and atomic immutable publication",
	"R4": "#5-8, #48-54, #60-62, #79: narrow ports, error ownership, context, and exact cleanup",
	"R5": "#5-8, #49-54, #57-62: compose narrow contracts without hiding error or concurrency policy",
	"R6": "#60-62, #75-81: request context, streaming-safe timeouts, strict JSON, and body closure",
	"R7": "#3, #5-8, #62, #73, #81: explicit construction, concrete ownership, group failure, and shutdown",
	"R8": "#3, #52-54, #62, #73, #79, #81, #100: composition-root errors, process lifetime, and drain",
}

func mustOrderChecks(catalog []learningCheck, order []string) []learningCheck {
	byID := make(map[string]learningCheck, len(catalog))
	for _, check := range catalog {
		if _, exists := byID[check.id]; exists {
			panic("duplicate learning check " + check.id)
		}
		byID[check.id] = check
	}

	ordered := make([]learningCheck, 0, len(order))
	for _, id := range order {
		check, exists := byID[id]
		if !exists {
			panic("missing learning check " + id)
		}
		ordered = append(ordered, check)
		delete(byID, id)
	}
	if len(byID) != 0 {
		panic("learning check catalog contains unordered entries")
	}
	return ordered
}

type optionalGroup struct {
	id      string
	phase   int
	title   string
	command string
	reason  string
	source  string
	tag     string
	focus   string
	checks  [][]string
}

var starredOptional = []optionalGroup{
	{
		id:      "OPT-SYNTAX",
		phase:   1,
		title:   "Niche syntax and compiler details",
		command: "go test -tags=optional ./internal/foundation/language -run '^TestOptionalSyntaxFoundations$'; go test ./internal/foundation/language -run '^TestCompileFailureRules$'; go test ./internal/foundation/performance -run 'Test(BoundsCheckExercises|FieldAndElementPointers|StringViewAndSafeCopy)$'",
		reason:  "flags, conversions, compile-fail rules, BCE, and unsafe details; learn on demand",
		source:  "internal/foundation/language/optional_syntax_test.go",
		tag:     "syntax",
		focus:   "Use compiler evidence for niche rules; do not turn trivia into production cleverness.",
		checks: [][]string{
			{"-tags=optional", "./internal/foundation/language", "-run", "^TestOptionalSyntaxFoundations$"},
			{"./internal/foundation/language", "-run", "^TestCompileFailureRules$"},
			{"./internal/foundation/performance", "-run", "Test(BoundsCheckExercises|FieldAndElementPointers|StringViewAndSafeCopy)"},
		},
	},
	{
		id:      "OPT-CONCURRENCY",
		phase:   1,
		title:   "Extended concurrency patterns",
		command: "go test ./internal/foundation/concurrency -run 'Bridge|Tee|AnyDone|ForwardUntilDone|TryMutex|Striped|SingleFlight|CircuitBreaker|CondQueue'",
		reason:  "useful, but not prerequisites for Router ownership",
		source:  "internal/foundation/concurrency/advanced_test.go",
		tag:     "concurrency",
		focus:   "Compare coordination cost, cancellation, and ownership before adopting a pattern.",
		checks:  [][]string{{"./internal/foundation/concurrency", "-run", "Bridge|Tee|AnyDone|ForwardUntilDone|TryMutex|Striped|SingleFlight|CircuitBreaker|CondQueue"}},
	},
	{
		id:      "OPT-PERFORMANCE",
		phase:   1,
		title:   "Runtime and performance details",
		command: "go test ./internal/foundation/performance",
		reason:  "use escape, layout, reflection, and profiling when evidence calls for them",
		source:  "internal/foundation/performance/escape_test.go",
		tag:     "performance",
		focus:   "Optimize from benchmarks and profiles; keep allocation ownership and measurement honest.",
		checks:  [][]string{{"./internal/foundation/performance"}},
	},
	{
		id:      "OPT-FP",
		phase:   1,
		title:   "Collections, resilience, patterns, and FP",
		command: "go test ./internal/foundation/collection ./internal/foundation/resilience ./internal/foundation/patterns ./internal/foundation/fp/...",
		reason:  "collections, retry mechanisms, GoF examples, containers, and middleware deepen selected topics",
		source:  "internal/foundation/fp/result_test.go",
		tag:     "functional",
		focus:   "Practice composition, Result and Option while judging when explicit Go is easier to maintain.",
		checks:  [][]string{{"./internal/foundation/collection", "./internal/foundation/resilience", "./internal/foundation/patterns", "./internal/foundation/fp/..."}},
	},
	{
		id:      "OPT-WORKFLOW",
		phase:   2,
		title:   "Workflow and alternative concurrency models",
		command: "go test ./internal/commerce/workflows",
		reason:  "compensation, Temporal, and Actor do not directly block Router work",
		source:  "internal/commerce/workflows/saga.go",
		tag:     "workflow",
		focus:   "Make compensation, replay, timeout, and delivery semantics explicit before choosing a workflow model.",
		checks:  [][]string{{"./internal/commerce/workflows"}},
	},
	{
		id:      "OPT-COMMERCE",
		phase:   2,
		title:   "Advanced Commerce extensions",
		command: "go test ./internal/commerce/pricing ./internal/commerce/operations ./internal/commerce/integrations",
		reason:  "specifications, promotions, inventory, settlement, event sourcing, and the senior release train remain available",
		source:  "internal/commerce/pricing/promotion_test.go",
		tag:     "commerce+",
		focus:   "Keep policy deterministic and integration reconciliation idempotent under retries and partial failure.",
		checks:  [][]string{{"./internal/commerce/pricing", "./internal/commerce/operations", "./internal/commerce/integrations"}},
	},
}

type learningResult struct {
	check  learningCheck
	passed bool
	output string
}

type optionalResult struct {
	group  optionalGroup
	passed bool
	output string
}

func runLearningJourney(root, action string, verbose bool) error {
	if action == "optional" {
		return runOptionalJourney(root, verbose)
	}
	if action == "next" {
		results := make([]learningResult, 0, len(learningJourney))
		for _, check := range learningJourney {
			result := executeLearningCheck(root, check)
			results = append(results, result)
			if result.passed {
				continue
			}
			if err := syncLearningProgress(root, results); err != nil {
				return err
			}
			fmt.Printf("\n  Next contract · %s\n", learningPhase(result.check.id))
			fmt.Printf("  %s %s\n\n", result.check.id, result.check.title)
			fmt.Printf("      %s\n\n", learningCommand(result.check))
			fmt.Printf("      100 Go Mistakes focus: %s\n", mistakeFocusByGate[result.check.id])
			fmt.Println("      Clinic: docs/go100/README.md")
			fmt.Println()
			fmt.Println("      Guide: docs/LEARNING_PATH.md")
			fmt.Println()
			if verbose && result.output != "" {
				fmt.Println(result.output)
				fmt.Println()
			}
			printOptionalHint()
			return nil
		}
		if err := syncLearningProgress(root, results); err != nil {
			return err
		}
		fmt.Println()
		fmt.Println("  The required journey is complete. Continue with the upstream llm-d capstone.")
		fmt.Println()
		printOptionalHint()
		return nil
	}

	// Status checks are independent. Bound process fan-out so this remains
	// responsive without turning a learning helper into a local fork bomb.
	results := make([]learningResult, len(learningJourney))
	slots := make(chan struct{}, 4)
	var wait sync.WaitGroup
	for i, check := range learningJourney {
		wait.Add(1)
		go func() {
			defer wait.Done()
			slots <- struct{}{}
			defer func() { <-slots }()
			results[i] = executeLearningCheck(root, check)
		}()
	}
	wait.Wait()
	if err := syncLearningProgress(root, results); err != nil {
		return err
	}

	fmt.Println("\n  Commerce to llm-d Router learning journey  (★ does not block)")
	fmt.Println()
	passed := 0
	currentPhase := ""
	for _, result := range results {
		phase := learningPhase(result.check.id)
		if phase != currentPhase {
			currentPhase = phase
			fmt.Printf("  %s\n", phase)
		}
		mark := "○"
		if result.passed {
			mark = "✓"
			passed++
		}
		fmt.Printf("  %s %-3s %s\n", mark, result.check.id, result.check.title)
		if verbose {
			fmt.Printf("      %s\n", learningCommand(result.check))
			fmt.Printf("      100 Go Mistakes focus: %s\n", mistakeFocusByGate[result.check.id])
			if !result.passed && result.output != "" {
				fmt.Printf("      %s\n", lastNonEmptyLine(result.output))
			}
		}
	}
	fmt.Printf("\n  Required gates passed: %d/%d\n", passed, len(results))
	printOptionalHint()
	if action == "check" && passed != len(results) {
		return fmt.Errorf("%d required learning gates are incomplete", len(results)-passed)
	}
	return nil
}

func runOptionalJourney(root string, verbose bool) error {
	results := make([]optionalResult, len(starredOptional))
	slots := make(chan struct{}, 2)
	var wait sync.WaitGroup
	for i, group := range starredOptional {
		wait.Add(1)
		go func() {
			defer wait.Done()
			slots <- struct{}{}
			defer func() { <-slots }()
			results[i] = executeOptionalGroup(root, group)
		}()
	}
	wait.Wait()
	if err := syncOptionalProgress(root, results); err != nil {
		return err
	}

	fmt.Println("\n  Starred optional extensions  (never block next)")
	fmt.Println()
	passed := 0
	phase := 0
	for _, result := range results {
		if result.group.phase != phase {
			phase = result.group.phase
			fmt.Printf("  Phase %d extensions\n", phase)
		}
		mark := "○"
		if result.passed {
			mark = "✓"
			passed++
		}
		fmt.Printf("  %s ★ %-18s %s\n", mark, result.group.id, result.group.title)
		if verbose {
			fmt.Printf("      %s\n", result.group.command)
			fmt.Printf("      Review lens: %s\n", result.group.focus)
			if !result.passed && result.output != "" {
				fmt.Printf("      %s\n", lastNonEmptyLine(result.output))
			}
		}
	}
	fmt.Printf("\n  Optional groups passed: %d/%d (required progress is unchanged)\n\n", passed, len(results))
	if passed != len(results) {
		return fmt.Errorf("%d optional groups are incomplete; the required journey remains unblocked", len(results)-passed)
	}
	return nil
}

func executeOptionalGroup(root string, group optionalGroup) optionalResult {
	for _, args := range group.checks {
		commandArgs := append([]string{"test", "-buildvcs=false"}, args...)
		commandArgs = append(commandArgs, "-count=1")
		cmd := exec.Command("go", commandArgs...)
		cmd.Dir = root
		output, err := cmd.CombinedOutput()
		if err != nil {
			outputText := strings.TrimSpace(string(output))
			if outputText == "" {
				outputText = err.Error()
			}
			return optionalResult{group: group, output: outputText}
		}
	}
	return optionalResult{group: group, passed: true}
}

// syncLearningProgress publishes only test-verified gate IDs. The HTML
// tracker loads this small JavaScript assignment directly, including from a
// file:// URL where fetching JSON is commonly blocked by browser security.
func syncLearningProgress(root string, results []learningResult) error {
	passed := make(map[string]bool, len(results))
	for _, result := range results {
		if result.passed {
			passed[result.check.id] = true
		}
	}

	ids := make([]string, 0, len(passed))
	for _, check := range learningJourney {
		if passed[check.id] {
			ids = append(ids, check.id)
		}
	}
	optionalIDs := readPublishedIDs(root, "__LEARNING_OPTIONAL_VERIFIED_GATES__", optionalGroupIDs())
	return writeLearningProgress(root, ids, optionalIDs)
}

func syncOptionalProgress(root string, results []optionalResult) error {
	passed := make(map[string]bool, len(results))
	for _, result := range results {
		if result.passed {
			passed[result.group.id] = true
		}
	}
	ids := make([]string, 0, len(passed))
	for _, group := range starredOptional {
		if passed[group.id] {
			ids = append(ids, group.id)
		}
	}
	requiredIDs := make([]string, 0, len(learningJourney))
	for _, check := range learningJourney {
		requiredIDs = append(requiredIDs, check.id)
	}
	requiredIDs = readPublishedIDs(root, "__LEARNING_VERIFIED_GATES__", requiredIDs)
	return writeLearningProgress(root, requiredIDs, ids)
}

func optionalGroupIDs() []string {
	ids := make([]string, 0, len(starredOptional))
	for _, group := range starredOptional {
		ids = append(ids, group.id)
	}
	return ids
}

func readPublishedIDs(root, assignment string, orderedAllowed []string) []string {
	content, err := os.ReadFile(filepath.Join(root, "docs", "learning-progress.js"))
	if err != nil {
		return nil
	}
	prefix := "window." + assignment + " = "
	var decoded []string
	for _, line := range strings.Split(string(content), "\n") {
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		if json.Unmarshal([]byte(strings.TrimSuffix(strings.TrimPrefix(line, prefix), ";")), &decoded) != nil {
			return nil
		}
		break
	}
	seen := make(map[string]bool, len(decoded))
	for _, id := range decoded {
		seen[id] = true
	}
	ordered := make([]string, 0, len(decoded))
	for _, id := range orderedAllowed {
		if seen[id] {
			ordered = append(ordered, id)
		}
	}
	return ordered
}

func writeLearningProgress(root string, requiredIDs, optionalIDs []string) error {
	encoded, err := json.Marshal(requiredIDs)
	if err != nil {
		return fmt.Errorf("encode learning progress: %w", err)
	}
	optionalEncoded, err := json.Marshal(optionalIDs)
	if err != nil {
		return fmt.Errorf("encode optional learning progress: %w", err)
	}
	focus, err := json.Marshal(mistakeFocusByGate)
	if err != nil {
		return fmt.Errorf("encode 100 Go Mistakes focus: %w", err)
	}
	content := append([]byte("// Generated by cmd/exercise; do not edit manually.\nwindow.__LEARNING_VERIFIED_GATES__ = "), encoded...)
	content = append(content, ';', '\n')
	content = append(content, []byte("window.__LEARNING_OPTIONAL_VERIFIED_GATES__ = ")...)
	content = append(content, optionalEncoded...)
	content = append(content, ';', '\n')
	content = append(content, []byte("window.__GO_MISTAKE_FOCUS__ = ")...)
	content = append(content, focus...)
	content = append(content, ';', '\n')
	path := filepath.Join(root, "docs", "learning-progress.js")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write learning progress: %w", err)
	}
	return nil
}

func learningPhase(id string) string {
	if id == "" {
		return "Unknown phase"
	}
	switch id[0] {
	case 'F':
		return "Phase 1 · Go foundations for Commerce"
	case 'E':
		return "Phase 2 · Commerce project"
	case 'R':
		return "Phase 4 · llm-d Router"
	default:
		return "Phase 3 · llm-d prerequisites"
	}
}

func learningCommand(check learningCheck) string {
	var primary string
	if check.command != "" {
		primary = check.command
	} else {
		primary = fmt.Sprintf("go test %s -run '%s' -count=1", check.pkg, check.run)
	}
	if check.contractRun == "" {
		return primary
	}
	contract := fmt.Sprintf(
		"go test ./test/contracts/infrastructure -run '^%s$' -count=1",
		check.contractRun,
	)
	return primary + " && " + contract
}

func executeLearningCheck(root string, check learningCheck) learningResult {
	if unresolved := unresolvedExerciseTODOs(root, check.todoPaths); len(unresolved) > 0 {
		return learningResult{
			check:  check,
			passed: false,
			output: "unresolved TODO(exercise): " + strings.Join(unresolved, ", "),
		}
	}

	program := check.program
	args := check.args
	if program == "" {
		program = "go"
		args = []string{
			"test", "-buildvcs=false", check.pkg,
			"-run", check.run,
			"-count=1",
		}
	}
	cmd := exec.Command(program, args...)
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	outputText := strings.TrimSpace(string(output))
	if err != nil && outputText == "" {
		outputText = err.Error()
	}
	if err == nil && check.contractRun != "" {
		contract := exec.Command(
			"go", "test", "-buildvcs=false",
			"./test/contracts/infrastructure",
			"-run", "^"+check.contractRun+"$",
			"-count=1",
		)
		contract.Dir = root
		contractOutput, contractErr := contract.CombinedOutput()
		if contractErr != nil {
			err = contractErr
			outputText = strings.TrimSpace(string(contractOutput))
			if outputText == "" {
				outputText = contractErr.Error()
			}
		}
	}
	return learningResult{
		check:  check,
		passed: err == nil,
		output: outputText,
	}
}

func unresolvedExerciseTODOs(root string, paths []string) []string {
	var unresolved []string
	for _, relative := range paths {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil || strings.Contains(string(content), "TODO(exercise)") {
			unresolved = append(unresolved, relative)
		}
	}
	return unresolved
}

func printOptionalHint() {
	fmt.Println()
	fmt.Println("  ★ Optional groups are excluded from next and do not block the required journey:")
	for _, group := range starredOptional {
		fmt.Printf("    ★ %-22s %s\n", group.title, group.reason)
	}
	fmt.Println("    Track evidence: go run ./cmd/exercise optional")
	fmt.Println()
}

func lastNonEmptyLine(output string) string {
	lines := strings.Split(output, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if line := strings.TrimSpace(lines[i]); line != "" {
			return line
		}
	}
	return ""
}
