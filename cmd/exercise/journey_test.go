// Learning journey ordering and documentation synchronization tests.
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestLearningJourneyHasUniqueIDsAndFocusedChecks(t *testing.T) {
	t.Parallel()

	ids := map[string]bool{}
	concurrencyChecks := 0
	for _, check := range learningJourney {
		if ids[check.id] {
			t.Fatalf("duplicate foundation id %q", check.id)
		}
		ids[check.id] = true
		goCheck := check.pkg != "" && check.run != ""
		externalCheck := check.program != "" && len(check.args) != 0 && check.command != ""
		if goCheck == externalCheck {
			t.Fatalf("check %q is not executable: %+v", check.id, check)
		}
		if strings.HasPrefix(check.id, "C") {
			concurrencyChecks++
		}
		focus, exists := mistakeFocusByGate[check.id]
		if !exists || strings.TrimSpace(focus) == "" {
			t.Errorf("required gate %s has no 100 Go Mistakes focus", check.id)
		}
		for _, excluded := range []string{"patterns", "minigin", "/fp", "txdsl", "saga", "actor"} {
			if strings.Contains(check.pkg, excluded) {
				t.Fatalf("optional topic %q leaked into required check %q", excluded, check.id)
			}
		}
	}
	if concurrencyChecks < 4 {
		t.Fatalf("concurrency checks = %d, want at least 4 focused groups", concurrencyChecks)
	}
	if len(mistakeFocusByGate) != len(learningJourney) {
		t.Fatalf("100 Go Mistakes focus entries = %d, want %d required gates", len(mistakeFocusByGate), len(learningJourney))
	}
}

func TestLearningJourneyHasRequiredOrder(t *testing.T) {
	expected := []string{
		"F0", "F1", "F2", "F3", "F4",
		"E0", "E1", "E2", "E3", "E4", "E5", "E6",
		"C0", "C1", "C2", "C3", "K0",
		"Y0", "P0", "H0", "T0", "I0",
		"O0", "O1", "O2", "O3", "G0",
		"R0", "R1", "R2", "R3", "R4", "R5", "R6", "R7", "R8",
	}
	if len(learningJourney) != len(expected) {
		t.Fatalf("required learning gates = %d, want %d", len(learningJourney), len(expected))
	}
	for index, want := range expected {
		if got := learningJourney[index].id; got != want {
			t.Fatalf("required learning gate %d = %q, want %q", index, got, want)
		}
	}
}

func TestLearningCommandIncludesInfrastructureContract(t *testing.T) {
	for _, check := range learningJourney {
		if check.contractRun == "" {
			continue
		}
		command := learningCommand(check)
		for _, required := range []string{
			" && go test ./test/contracts/infrastructure",
			"-run '^" + check.contractRun + "$'",
		} {
			if !strings.Contains(command, required) {
				t.Errorf("gate %s command %q is missing %q", check.id, command, required)
			}
		}
	}
}

func TestF4SelectsEveryTestCraftContract(t *testing.T) {
	var pattern string
	for _, check := range learningJourney {
		if check.id == "F4" {
			pattern = check.run
			break
		}
	}
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{
		"TestSplitSuite_PassesTheCorrectImplementation",
		"TestSplitSuite_CatchesEveryMutant",
		"TestCheckSumsToTotal",
		"TestCheckNearlyEqual",
		"TestCheckRemainderGoesFirst",
		"TestSplitEvenly_Examples",
		"TestSplitEvenly_Errors",
		"FuzzSplit",
	} {
		if !compiled.MatchString(name) {
			t.Errorf("F4 pattern %q does not select %s", pattern, name)
		}
	}
}

func TestMarkdownPathMirrorsRequiredGateOrder(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(root, "docs", "LEARNING_PATH.md"))
	if err != nil {
		t.Fatal(err)
	}
	guide := string(content)
	phaseStart := strings.Index(guide, "## Phase 1:")
	if phaseStart < 0 {
		t.Fatal("learning guide has no Phase 1 section")
	}
	guide = guide[phaseStart:]
	cursor := 0
	for _, check := range learningJourney {
		if got := strings.Count(guide, "| "+check.id+" |"); got != 1 {
			t.Fatalf("learning guide gate %s appears %d times, want exactly once", check.id, got)
		}
		rest := guide[cursor:]
		position := firstNonNegative(
			strings.Index(rest, "| "+check.id+" |"),
			strings.Index(rest, "### "+check.id+" "),
		)
		if position < 0 {
			t.Fatalf("learning guide does not list required gate %s after the previous gate", check.id)
		}
		cursor += position + len(check.id) + 3
	}
}

func TestHTMLPathContainsEveryRequiredGate(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(root, "docs", "learning-path.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := string(content)
	cursor := 0
	for _, check := range learningJourney {
		marker := `data-gate="` + check.id + `"`
		position := strings.Index(html[cursor:], marker)
		if position < 0 {
			t.Fatalf("HTML path is missing ordered gate %s", check.id)
		}
		cursor += position + len(marker)
	}
	if got := strings.Count(html, `data-gate="`); got != len(learningJourney) {
		t.Errorf("HTML gate cards = %d, want %d", got, len(learningJourney))
	}
	for _, required := range []string{
		"localStorage", "progress-fill", "reset", "LEARNING_PATH.md",
		"gateIDs.has(id)", "learning-progress.js", "__LEARNING_VERIFIED_GATES__",
		"go100/README.md", "__GO_MISTAKE_FOCUS__", "renderMistakeFocus",
		"setInterval(refreshVerifiedProgress", `id="sync-progress"`,
		`syncButton.addEventListener("click", refreshVerifiedProgress)`,
	} {
		if !strings.Contains(html, required) {
			t.Errorf("HTML tracker is missing %q", required)
		}
	}
}

func TestSyncLearningProgressWritesOnlyPassedGatesInJourneyOrder(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	results := []learningResult{
		{check: learningCheck{id: "C0"}, passed: true},
		{check: learningCheck{id: "F1"}, passed: false},
		{check: learningCheck{id: "F0"}, passed: true},
	}
	if err := syncLearningProgress(root, results); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(root, "docs", "learning-progress.js"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(content)
	if !strings.Contains(got, `window.__LEARNING_VERIFIED_GATES__ = ["F0","C0"];`) {
		t.Fatalf("progress file = %q", got)
	}
	progressLine := strings.SplitN(got, "\n", 3)[1]
	if strings.Contains(progressLine, "F1") {
		t.Fatalf("failed gate leaked into verified progress: %q", progressLine)
	}
	if !strings.Contains(got, `window.__GO_MISTAKE_FOCUS__ = {`) {
		t.Fatalf("progress file has no 100 Go Mistakes catalog: %q", got)
	}
	for _, id := range []string{"F0", "C0", "R8"} {
		if !strings.Contains(got, `"`+id+`":`) {
			t.Errorf("progress file is missing focus for gate %s", id)
		}
	}
}

func TestCheckedInLearningProgressBootstrapsEveryMistakeFocus(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(root, "docs", "learning-progress.js"))
	if err != nil {
		t.Fatal(err)
	}
	script := string(content)
	encoded, err := json.Marshal(mistakeFocusByGate)
	if err != nil {
		t.Fatal(err)
	}
	exactCatalog := "window.__GO_MISTAKE_FOCUS__ = " + string(encoded) + ";"
	if !strings.Contains(script, exactCatalog) {
		t.Fatal("checked-in progress bootstrap is stale relative to the gate focus catalog")
	}
	for _, check := range learningJourney {
		if !strings.Contains(script, `"`+check.id+`":`) {
			t.Errorf("checked-in progress bootstrap has no mistake focus for gate %s", check.id)
		}
	}
}

func TestGoMistakesClinicStaysTargetedAndDoesNotBecomeAnAnswerKey(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(root, "docs", "go100", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	guide := string(content)
	for _, required := range []string{
		"100 Go Mistakes and How to Avoid Them", "Commerce", "llm-d Router",
		"#17-29", "#48-54", "#55-74", "#75-81", "#82-90", "#91-100",
		"Do not implement", "Evidence to produce",
	} {
		if !strings.Contains(guide, required) {
			t.Errorf("100 Go Mistakes clinic is missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"func solution", "Answer key", "copy this implementation",
		"decimal money", "Money with binary floating point",
	} {
		if strings.Contains(guide, forbidden) {
			t.Errorf("100 Go Mistakes clinic contains answer-like text %q", forbidden)
		}
	}
}

func TestRouterDocumentationStaysConsolidated(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(root, "docs", "llmd"))
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]bool{
		"README.md":          true,
		"PRODUCTION_LABS.md": true,
		"UPSTREAM.md":        true,
		"EXERCISE_RULES.md":  true,
	}
	for _, entry := range entries {
		if entry.IsDir() || !expected[entry.Name()] {
			t.Errorf("unexpected Router documentation entry %q; consolidate it into an authoritative file", entry.Name())
		}
		delete(expected, entry.Name())
	}
	for missing := range expected {
		t.Errorf("missing authoritative Router document %q", missing)
	}
}

func firstNonNegative(values ...int) int {
	result := -1
	for _, value := range values {
		if value >= 0 && (result < 0 || value < result) {
			result = value
		}
	}
	return result
}

func TestEveryOptionalGroupIsMarkedAndExplained(t *testing.T) {
	t.Parallel()

	ids := map[string]bool{}
	for i, group := range starredOptional {
		if group.id == "" || group.phase < 1 || group.phase > 4 || group.title == "" || group.command == "" || group.reason == "" || group.source == "" || group.tag == "" || group.focus == "" || len(group.checks) == 0 {
			t.Fatalf("optional group %d is incomplete: %+v", i, group)
		}
		if ids[group.id] {
			t.Fatalf("duplicate optional group ID %q", group.id)
		}
		ids[group.id] = true
	}
}

func TestOptionalProgressIsPublishedSeparatelyFromRequiredProgress(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	required := []string{"F0", "F1"}
	optional := []string{"OPT-FP", "OPT-WORKFLOW"}
	if err := writeLearningProgress(root, required, optional); err != nil {
		t.Fatal(err)
	}

	gotRequired := readPublishedIDs(root, "__LEARNING_VERIFIED_GATES__", []string{"F0", "F1", "F2"})
	if strings.Join(gotRequired, ",") != "F0,F1" {
		t.Fatalf("required progress = %v", gotRequired)
	}
	gotOptional := readPublishedIDs(root, "__LEARNING_OPTIONAL_VERIFIED_GATES__", optionalGroupIDs())
	if strings.Join(gotOptional, ",") != "OPT-FP,OPT-WORKFLOW" {
		t.Fatalf("optional progress = %v", gotOptional)
	}

	content, err := os.ReadFile(filepath.Join(root, "docs", "learning-progress.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "window.__LEARNING_OPTIONAL_VERIFIED_GATES__") {
		t.Fatal("optional progress assignment was not published")
	}

	if err := syncLearningProgress(root, []learningResult{{
		check:  learningJourney[0],
		passed: true,
	}}); err != nil {
		t.Fatal(err)
	}
	gotOptional = readPublishedIDs(root, "__LEARNING_OPTIONAL_VERIFIED_GATES__", optionalGroupIDs())
	if strings.Join(gotOptional, ",") != "OPT-FP,OPT-WORKFLOW" {
		t.Fatalf("required refresh did not preserve optional progress: %v", gotOptional)
	}

	if err := syncOptionalProgress(root, []optionalResult{{
		group:  starredOptional[0],
		passed: true,
	}}); err != nil {
		t.Fatal(err)
	}
	gotRequired = readPublishedIDs(root, "__LEARNING_VERIFIED_GATES__", []string{"F0", "F1", "F2"})
	if strings.Join(gotRequired, ",") != "F0" {
		t.Fatalf("optional refresh did not preserve required progress: %v", gotRequired)
	}
}

func TestUnresolvedExerciseTODOsBlocksStarterSkeletons(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	pending := filepath.Join(root, "pending.yaml")
	complete := filepath.Join(root, "complete.yaml")
	if err := os.WriteFile(pending, []byte("# TODO(exercise): implement\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(complete, []byte("kind: ConfigMap\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	got := unresolvedExerciseTODOs(root, []string{"pending.yaml", "complete.yaml", "missing.yaml"})
	if strings.Join(got, ",") != "pending.yaml,missing.yaml" {
		t.Fatalf("unresolved = %v", got)
	}
}

func TestEveryExerciseStarterBelongsToExactlyOneGate(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	owners := map[string][]string{}
	for _, check := range learningJourney {
		for _, path := range check.todoPaths {
			normalized := filepath.ToSlash(path)
			owners[normalized] = append(owners[normalized], check.id)
		}
	}

	var starters []string
	for _, relative := range []string{
		"internal/inference",
		"cmd/inference-lab",
		"deploy/inference-lab",
		"test/acceptance/inference",
		"configs/endpoint-picker-learning.yaml",
	} {
		target := filepath.Join(root, filepath.FromSlash(relative))
		info, statErr := os.Stat(target)
		if statErr != nil {
			t.Fatal(statErr)
		}
		if !info.IsDir() {
			content, readErr := os.ReadFile(target)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if strings.Contains(string(content), "TODO(exercise)") {
				starters = append(starters, filepath.ToSlash(relative))
			}
			continue
		}
		walkErr := filepath.WalkDir(target, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, "_test.go") {
				return nil
			}
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			if !strings.Contains(string(content), "TODO(exercise)") {
				return nil
			}
			relativePath, relativeErr := filepath.Rel(root, path)
			if relativeErr != nil {
				return relativeErr
			}
			starters = append(starters, filepath.ToSlash(relativePath))
			return nil
		})
		if walkErr != nil {
			t.Fatal(walkErr)
		}
	}

	for _, starter := range starters {
		if got := owners[starter]; len(got) != 1 {
			t.Errorf("%s belongs to %d gates (%v), want exactly one", starter, len(got), got)
		}
	}
}
