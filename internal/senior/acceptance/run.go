package acceptance

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const maxCommandOutput = 128 << 10

type GateStatus string

const (
	StatusPassed GateStatus = "passed"
	StatusFailed GateStatus = "failed"
)

type GateResult struct {
	ID          string        `json:"id"`
	Description string        `json:"description"`
	Kind        GateKind      `json:"kind"`
	Category    string        `json:"category"`
	Required    bool          `json:"required"`
	Status      GateStatus    `json:"status"`
	Duration    time.Duration `json:"duration"`
	Detail      string        `json:"detail,omitempty"`
	Output      string        `json:"output,omitempty"`
}

type Report struct {
	SchemaVersion int          `json:"schema_version"`
	Release       string       `json:"release"`
	Level         string       `json:"level"`
	StartedAt     time.Time    `json:"started_at"`
	FinishedAt    time.Time    `json:"finished_at"`
	Passed        bool         `json:"passed"`
	Results       []GateResult `json:"results"`
}

func Run(ctx context.Context, root string, manifest Manifest) (Report, error) {
	if err := manifest.Validate(); err != nil {
		return Report{}, fmt.Errorf("invalid manifest: %w", err)
	}

	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return Report{}, fmt.Errorf("resolve module root: %w", err)
	}
	info, err := os.Stat(absoluteRoot)
	if err != nil {
		return Report{}, fmt.Errorf("stat module root: %w", err)
	}
	if !info.IsDir() {
		return Report{}, fmt.Errorf("module root is not a directory: %s", absoluteRoot)
	}

	report := Report{
		SchemaVersion: CurrentSchemaVersion,
		Release:       manifest.Release,
		Level:         manifest.Level,
		StartedAt:     time.Now().UTC(),
		Passed:        true,
		Results:       make([]GateResult, 0, len(manifest.Gates)),
	}

	for _, gate := range manifest.Gates {
		result := runGate(ctx, absoluteRoot, gate)
		report.Results = append(report.Results, result)
		if gate.Required && result.Status != StatusPassed {
			report.Passed = false
		}
	}
	report.FinishedAt = time.Now().UTC()
	return report, nil
}

func runGate(parent context.Context, root string, gate Gate) (result GateResult) {
	started := time.Now()
	result = GateResult{
		ID:          gate.ID,
		Description: gate.Description,
		Kind:        gate.Kind,
		Category:    gate.Category,
		Required:    gate.Required,
		Status:      StatusFailed,
	}
	defer func() {
		result.Duration = time.Since(started)
	}()

	switch gate.Kind {
	case GateCommand:
		result.Detail, result.Output = runCommand(parent, root, gate)
		if result.Detail != "" {
			return result
		}
	case GateEvidence:
	case GateApproval:
		if gate.Approval.Decision != "approved" {
			result.Detail = fmt.Sprintf("review decision is %q", gate.Approval.Decision)
			return result
		}
	}

	if err := verifyArtifacts(root, gate.Artifacts); err != nil {
		result.Detail = err.Error()
		return result
	}
	result.Status = StatusPassed
	return result
}

func runCommand(parent context.Context, root string, gate Gate) (detail, output string) {
	timeout := 2 * time.Minute
	if gate.Timeout != "" {
		timeout, _ = time.ParseDuration(gate.Timeout)
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	workdir := root
	if gate.WorkingDirectory != "" {
		var err error
		workdir, err = securePath(root, gate.WorkingDirectory)
		if err != nil {
			return err.Error(), ""
		}
		info, err := os.Stat(workdir)
		if err != nil {
			return fmt.Sprintf("working directory: %v", err), ""
		}
		if !info.IsDir() {
			return "working directory is not a directory", ""
		}
		resolved, err := filepath.EvalSymlinks(workdir)
		if err != nil {
			return fmt.Sprintf("working directory: resolve symlink: %v", err), ""
		}
		if err := ensureWithinRoot(root, resolved); err != nil {
			return fmt.Sprintf("working directory: %v", err), ""
		}
		workdir = resolved
	}

	cmd := exec.CommandContext(ctx, gate.Command[0], gate.Command[1:]...)
	cmd.Dir = workdir
	var captured cappedBuffer
	captured.limit = maxCommandOutput
	cmd.Stdout = &captured
	cmd.Stderr = &captured
	err := cmd.Run()
	output = strings.TrimSpace(captured.String())
	if ctx.Err() != nil {
		return fmt.Sprintf("command stopped: %v", ctx.Err()), output
	}
	if err != nil {
		return fmt.Sprintf("command failed: %v", err), output
	}
	return "", output
}

func verifyArtifacts(root string, artifacts []Artifact) error {
	for _, artifact := range artifacts {
		path, err := securePath(root, artifact.Path)
		if err != nil {
			return fmt.Errorf("artifact %q: %w", artifact.Path, err)
		}
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("artifact %q: %w", artifact.Path, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("artifact %q is not a regular file", artifact.Path)
		}

		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			return fmt.Errorf("artifact %q: resolve symlink: %w", artifact.Path, err)
		}
		if err := ensureWithinRoot(root, resolved); err != nil {
			return fmt.Errorf("artifact %q: %w", artifact.Path, err)
		}

		if artifact.SHA256 != "" {
			actual, err := SHA256File(resolved)
			if err != nil {
				return fmt.Errorf("artifact %q: %w", artifact.Path, err)
			}
			if actual != artifact.SHA256 {
				return fmt.Errorf(
					"artifact %q digest mismatch: got %s, want %s",
					artifact.Path, actual, artifact.SHA256)
			}
		}
	}
	return nil
}

type cappedBuffer struct {
	mu        sync.Mutex
	buffer    bytes.Buffer
	limit     int
	truncated bool
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	original := len(p)
	remaining := b.limit - b.buffer.Len()
	if remaining > 0 {
		if len(p) > remaining {
			p = p[:remaining]
			b.truncated = true
		}
		_, _ = b.buffer.Write(p)
	} else if len(p) > 0 {
		b.truncated = true
	}
	return original, nil
}

func (b *cappedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	value := b.buffer.String()
	if b.truncated {
		value += "\n...[command output truncated]"
	}
	return value
}

func SHA256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("sha256:%x", hash.Sum(nil)), nil
}

func securePath(root, relative string) (string, error) {
	path, err := filepath.Abs(filepath.Join(root, relative))
	if err != nil {
		return "", err
	}
	if err := ensureWithinRoot(root, path); err != nil {
		return "", err
	}
	return path, nil
}

func ensureWithinRoot(root, path string) error {
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return err
	}
	if relative == ".." ||
		strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path escapes module root")
	}
	return nil
}
