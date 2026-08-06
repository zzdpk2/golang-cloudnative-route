// Package acceptance runs release gates and verifies their evidence.
//
// It deliberately does not know how an Order, migration, deployment, or SLO is
// implemented. A manifest composes the tools chosen by the learner and keeps a
// machine-readable record of what was executed.
package acceptance

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const CurrentSchemaVersion = 1

type GateKind string

const (
	GateCommand  GateKind = "command"
	GateEvidence GateKind = "evidence"
	GateApproval GateKind = "approval"
)

type Manifest struct {
	SchemaVersion  int    `json:"schema_version"`
	Release        string `json:"release"`
	Level          string `json:"level"`
	Owner          string `json:"owner"`
	SourceRevision string `json:"source_revision"`
	ArtifactDigest string `json:"artifact_digest,omitempty"`
	Gates          []Gate `json:"gates"`
}

type Gate struct {
	ID               string     `json:"id"`
	Description      string     `json:"description"`
	Kind             GateKind   `json:"kind"`
	Category         string     `json:"category"`
	Required         bool       `json:"required"`
	Command          []string   `json:"command,omitempty"`
	WorkingDirectory string     `json:"working_directory,omitempty"`
	Timeout          string     `json:"timeout,omitempty"`
	Artifacts        []Artifact `json:"artifacts,omitempty"`
	Approval         *Approval  `json:"approval,omitempty"`
}

type Artifact struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256,omitempty"`
}

type Approval struct {
	Reviewer   string `json:"reviewer"`
	ReviewedAt string `json:"reviewed_at"`
	Decision   string `json:"decision"`
}

var (
	gateIDPattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9._-]*$`)
	levelPattern  = regexp.MustCompile(`^L(18|19|20|21|22|23)$`)
	digestPattern = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)
)

func Load(path string) (Manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("open manifest: %w", err)
	}
	defer f.Close()

	var manifest Manifest
	decoder := json.NewDecoder(f)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode manifest: %w", err)
	}

	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Manifest{}, errors.New("decode manifest: multiple JSON values")
		}
		return Manifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	return manifest, nil
}

func (m Manifest) Validate() error {
	var problems []error

	if m.SchemaVersion != CurrentSchemaVersion {
		problems = append(problems, fmt.Errorf(
			"schema_version must be %d", CurrentSchemaVersion))
	}
	if strings.TrimSpace(m.Release) == "" {
		problems = append(problems, errors.New("release is required"))
	}
	if !levelPattern.MatchString(m.Level) {
		problems = append(problems, errors.New("level must be one of L18 through L23"))
	}
	if strings.TrimSpace(m.Owner) == "" {
		problems = append(problems, errors.New("owner is required"))
	}
	if strings.TrimSpace(m.SourceRevision) == "" {
		problems = append(problems, errors.New("source_revision is required"))
	}
	if m.ArtifactDigest != "" && !digestPattern.MatchString(m.ArtifactDigest) {
		problems = append(problems, errors.New(
			"artifact_digest must use sha256:<64 lowercase hex characters>"))
	}
	if len(m.Gates) == 0 {
		problems = append(problems, errors.New("at least one gate is required"))
	}

	seen := make(map[string]struct{}, len(m.Gates))
	requiredKinds := map[GateKind]bool{}
	for i, gate := range m.Gates {
		prefix := fmt.Sprintf("gates[%d]", i)
		if gate.ID == "" || !gateIDPattern.MatchString(gate.ID) {
			problems = append(problems, fmt.Errorf(
				"%s.id must contain uppercase letters, digits, dot, underscore, or dash", prefix))
		} else if _, exists := seen[gate.ID]; exists {
			problems = append(problems, fmt.Errorf("%s.id %q is duplicated", prefix, gate.ID))
		} else {
			seen[gate.ID] = struct{}{}
		}
		if strings.TrimSpace(gate.Description) == "" {
			problems = append(problems, fmt.Errorf("%s.description is required", prefix))
		}
		if strings.TrimSpace(gate.Category) == "" {
			problems = append(problems, fmt.Errorf("%s.category is required", prefix))
		}
		if gate.Required {
			requiredKinds[gate.Kind] = true
		}
		if err := validateGate(m.Owner, gate); err != nil {
			problems = append(problems, fmt.Errorf("%s: %w", prefix, err))
		}
	}

	for _, kind := range []GateKind{GateCommand, GateEvidence, GateApproval} {
		if !requiredKinds[kind] {
			problems = append(problems, fmt.Errorf(
				"at least one required %s gate is required", kind))
		}
	}

	return errors.Join(problems...)
}

func validateGate(owner string, gate Gate) error {
	var problems []error

	switch gate.Kind {
	case GateCommand:
		if len(gate.Command) == 0 || strings.TrimSpace(gate.Command[0]) == "" {
			problems = append(problems, errors.New("command gate requires command argv"))
		}
		if gate.Approval != nil {
			problems = append(problems, errors.New("command gate cannot contain approval"))
		}
	case GateEvidence:
		if len(gate.Artifacts) == 0 {
			problems = append(problems, errors.New("evidence gate requires artifacts"))
		}
		if len(gate.Command) != 0 || gate.Approval != nil {
			problems = append(problems, errors.New(
				"evidence gate cannot contain command or approval"))
		}
	case GateApproval:
		if gate.Approval == nil {
			problems = append(problems, errors.New("approval gate requires approval"))
			break
		}
		if len(gate.Command) != 0 {
			problems = append(problems, errors.New("approval gate cannot contain command"))
		}
		if len(gate.Artifacts) == 0 {
			problems = append(problems, errors.New(
				"approval gate requires the reviewed artifacts"))
		}
		if strings.TrimSpace(gate.Approval.Reviewer) == "" {
			problems = append(problems, errors.New("approval reviewer is required"))
		}
		if strings.EqualFold(
			strings.TrimSpace(gate.Approval.Reviewer),
			strings.TrimSpace(owner),
		) {
			problems = append(problems, errors.New(
				"approval reviewer must be different from the release owner"))
		}
		if _, err := time.Parse(time.RFC3339, gate.Approval.ReviewedAt); err != nil {
			problems = append(problems, errors.New(
				"approval reviewed_at must be RFC3339"))
		}
		switch gate.Approval.Decision {
		case "approved", "rejected":
		default:
			problems = append(problems, errors.New(
				"approval decision must be approved or rejected"))
		}
	default:
		problems = append(problems, fmt.Errorf("unknown gate kind %q", gate.Kind))
	}

	if gate.Timeout != "" {
		timeout, err := time.ParseDuration(gate.Timeout)
		if err != nil || timeout <= 0 {
			problems = append(problems, errors.New(
				"timeout must be a positive Go duration such as 30s or 5m"))
		}
	}
	if gate.WorkingDirectory != "" {
		if err := validateRelativePath(gate.WorkingDirectory); err != nil {
			problems = append(problems, fmt.Errorf("working_directory: %w", err))
		}
	}
	for i, artifact := range gate.Artifacts {
		if err := validateArtifact(artifact); err != nil {
			problems = append(problems, fmt.Errorf("artifacts[%d]: %w", i, err))
		}
	}

	return errors.Join(problems...)
}

func validateArtifact(artifact Artifact) error {
	if err := validateRelativePath(artifact.Path); err != nil {
		return err
	}
	if artifact.SHA256 != "" && !digestPattern.MatchString(artifact.SHA256) {
		return errors.New("sha256 must use sha256:<64 lowercase hex characters>")
	}
	return nil
}

func validateRelativePath(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("path is required")
	}
	if filepath.IsAbs(path) {
		return errors.New("path must be relative to the module root")
	}
	clean := filepath.Clean(path)
	if clean == "." || clean == ".." ||
		strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return errors.New("path must stay within the module root")
	}
	return nil
}
