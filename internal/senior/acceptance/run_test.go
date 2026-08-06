package acceptance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunPassesCommandEvidenceAndApproval(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "evidence/migration.md", "migration proof")
	writeTestFile(t, root, "evidence/review.md", "review proof")

	manifest := validManifest()
	manifest.Gates[0].Command = helperCommand("pass")
	t.Setenv("GO_WANT_SENIORCHECK_HELPER", "1")

	report, err := Run(context.Background(), root, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Passed {
		t.Fatalf("report did not pass: %+v", report.Results)
	}
	for _, result := range report.Results {
		if result.Status != StatusPassed {
			t.Errorf("%s status = %s", result.ID, result.Status)
		}
		if result.Duration < 0 {
			t.Errorf("%s duration = %s", result.ID, result.Duration)
		}
	}
	if report.FinishedAt.Before(report.StartedAt) {
		t.Errorf("report finished before it started")
	}
}

func TestRunFailsRequiredCommandAndKeepsOutput(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "evidence/migration.md", "migration proof")
	writeTestFile(t, root, "evidence/review.md", "review proof")

	manifest := validManifest()
	manifest.Gates[0].Command = helperCommand("fail")
	t.Setenv("GO_WANT_SENIORCHECK_HELPER", "1")

	report, err := Run(context.Background(), root, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if report.Passed {
		t.Fatal("report passed")
	}
	result := report.Results[0]
	if result.Status != StatusFailed {
		t.Fatalf("status = %s", result.Status)
	}
	if !strings.Contains(result.Output, "intentional failure") {
		t.Fatalf("output = %q", result.Output)
	}
}

func TestRunDetectsArtifactMutation(t *testing.T) {
	root := t.TempDir()
	path := writeTestFile(t, root, "evidence/migration.md", "original")
	writeTestFile(t, root, "evidence/review.md", "review proof")
	digest, err := SHA256File(path)
	if err != nil {
		t.Fatal(err)
	}

	manifest := validManifest()
	manifest.Gates[0].Command = helperCommand("pass")
	manifest.Gates[1].Artifacts[0].SHA256 = digest
	if err := os.WriteFile(path, []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GO_WANT_SENIORCHECK_HELPER", "1")

	report, err := Run(context.Background(), root, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if report.Passed {
		t.Fatal("report passed with mutated evidence")
	}
	if !strings.Contains(report.Results[1].Detail, "digest mismatch") {
		t.Fatalf("detail = %q", report.Results[1].Detail)
	}
}

func TestRunHonorsCommandTimeout(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, "evidence/migration.md", "migration proof")
	writeTestFile(t, root, "evidence/review.md", "review proof")

	manifest := validManifest()
	manifest.Gates[0].Command = helperCommand("block")
	manifest.Gates[0].Timeout = "20ms"
	t.Setenv("GO_WANT_SENIORCHECK_HELPER", "1")

	report, err := Run(context.Background(), root, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if report.Passed {
		t.Fatal("report passed")
	}
	if !strings.Contains(report.Results[0].Detail, "deadline exceeded") {
		t.Fatalf("detail = %q", report.Results[0].Detail)
	}
}

func TestCappedBufferBoundsCommandOutput(t *testing.T) {
	var buffer cappedBuffer
	buffer.limit = 5

	n, err := buffer.Write([]byte("123456789"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 9 {
		t.Fatalf("Write returned %d", n)
	}
	if got := buffer.String(); got != "12345\n...[command output truncated]" {
		t.Fatalf("got %q", got)
	}
}

func TestSeniorcheckHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_SENIORCHECK_HELPER") != "1" {
		return
	}
	mode := os.Args[len(os.Args)-1]
	switch mode {
	case "pass":
		fmt.Println("helper passed")
		os.Exit(0)
	case "fail":
		fmt.Println("intentional failure")
		os.Exit(3)
	case "block":
		time.Sleep(10 * time.Second)
		os.Exit(0)
	default:
		os.Exit(4)
	}
}

func helperCommand(mode string) []string {
	return []string{
		os.Args[0],
		"-test.run=TestSeniorcheckHelperProcess",
		"--",
		mode,
	}
}

func writeTestFile(t *testing.T, root, relative, content string) string {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
