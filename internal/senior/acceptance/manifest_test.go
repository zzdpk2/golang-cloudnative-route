package acceptance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManifestValidate(t *testing.T) {
	manifest := validManifest()
	if err := manifest.Validate(); err != nil {
		t.Fatalf("valid manifest: %v", err)
	}
}

func TestManifestValidateReportsIndependentReviewAndRequiredKinds(t *testing.T) {
	manifest := validManifest()
	manifest.Gates[2].Approval.Reviewer = manifest.Owner
	manifest.Gates[1].Required = false

	err := manifest.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	message := err.Error()
	for _, expected := range []string{
		"reviewer must be different",
		"required evidence gate",
	} {
		if !strings.Contains(message, expected) {
			t.Errorf("error %q does not contain %q", message, expected)
		}
	}
}

func TestManifestValidateRejectsEscapingPath(t *testing.T) {
	manifest := validManifest()
	manifest.Gates[1].Artifacts[0].Path = "../outside.txt"

	err := manifest.Validate()
	if err == nil || !strings.Contains(err.Error(), "stay within the module root") {
		t.Fatalf("got %v", err)
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "release.json")
	data := `{
		"schema_version": 1,
		"release": "v2",
		"level": "L18",
		"owner": "learner",
		"source_revision": "abc123",
		"gates": [],
		"surprise": true
	}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("got %v", err)
	}
}

func validManifest() Manifest {
	return Manifest{
		SchemaVersion:  CurrentSchemaVersion,
		Release:        "v2",
		Level:          "L18",
		Owner:          "learner",
		SourceRevision: "abc123",
		Gates: []Gate{
			{
				ID:          "L18.AUTO.001",
				Description: "automated checks pass",
				Kind:        GateCommand,
				Category:    "correctness",
				Required:    true,
				Command:     []string{"go", "version"},
				Timeout:     "10s",
			},
			{
				ID:          "L18.EVIDENCE.001",
				Description: "migration evidence exists",
				Kind:        GateEvidence,
				Category:    "recovery",
				Required:    true,
				Artifacts:   []Artifact{{Path: "evidence/migration.md"}},
			},
			{
				ID:          "L18.REVIEW.001",
				Description: "independent review approved",
				Kind:        GateApproval,
				Category:    "decision",
				Required:    true,
				Artifacts:   []Artifact{{Path: "evidence/review.md"}},
				Approval: &Approval{
					Reviewer:   "reviewer",
					ReviewedAt: "2026-07-28T00:00:00Z",
					Decision:   "approved",
				},
			},
		},
	}
}
