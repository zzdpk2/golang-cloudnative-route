// Command seniorcheck executes a senior-track release manifest.
//
//	go run ./cmd/seniorcheck -manifest evidence/releases/v2.json
//	go run ./cmd/seniorcheck -manifest evidence/releases/v2.json -validate-only
//	go run ./cmd/seniorcheck -hash evidence/load/l22-summary.json
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rex/go-ddd-tdd/internal/senior/acceptance"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("seniorcheck", flag.ContinueOnError)
	flags.SetOutput(stderr)
	manifestPath := flags.String("manifest", "", "path to a release manifest")
	root := flags.String("root", "", "module root; defaults to the nearest go.mod")
	reportPath := flags.String("report", "", "write the JSON report to this path")
	validateOnly := flags.Bool("validate-only", false, "validate without running gates")
	hashPath := flags.String("hash", "", "print the SHA-256 digest of one file")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	if *root == "" {
		var err error
		*root, err = moduleRoot()
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
	}

	if *hashPath != "" {
		path := *hashPath
		if !filepath.IsAbs(path) {
			path = filepath.Join(*root, path)
		}
		digest, err := acceptance.SHA256File(path)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		fmt.Fprintf(stdout, "%s  %s\n", digest, *hashPath)
		return 0
	}

	if *manifestPath == "" {
		fmt.Fprintln(stderr, "usage: seniorcheck -manifest <release.json>")
		return 2
	}

	manifest, err := acceptance.Load(*manifestPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if err := manifest.Validate(); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if *validateOnly {
		fmt.Fprintf(stdout, "VALID  %s %s (%d gates)\n",
			manifest.Level, manifest.Release, len(manifest.Gates))
		return 0
	}

	report, err := acceptance.Run(context.Background(), *root, manifest)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	printReport(stdout, report)

	if *reportPath != "" {
		if err := writeReport(*reportPath, report); err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
	}
	if !report.Passed {
		return 1
	}
	return 0
}

func printReport(w io.Writer, report acceptance.Report) {
	for _, result := range report.Results {
		mark := "PASS"
		if result.Status != acceptance.StatusPassed {
			mark = "FAIL"
		}
		required := ""
		if !result.Required {
			required = " (optional)"
		}
		fmt.Fprintf(w, "%s  %-20s %s%s\n",
			mark, result.ID, result.Description, required)
		if result.Detail != "" {
			fmt.Fprintf(w, "      %s\n", result.Detail)
		}
		if result.Status != acceptance.StatusPassed && result.Output != "" {
			for _, line := range strings.Split(result.Output, "\n") {
				fmt.Fprintf(w, "      | %s\n", line)
			}
		}
	}
	summary := "PASS"
	if !report.Passed {
		summary = "FAIL"
	}
	fmt.Fprintf(w, "\n%s  %s %s\n", summary, report.Level, report.Release)
}

func writeReport(path string, report acceptance.Report) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create report: %w", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	return nil
}

func moduleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod found above the working directory")
		}
		dir = parent
	}
}
