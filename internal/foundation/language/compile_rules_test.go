package language

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var wantDiagnostic = regexp.MustCompile(`(?m)// want: (.+)$`)

// TestCompileFailureRules turns language rules that cannot appear in a
// buildable package into executable examples under testdata/compilefail.
func TestCompileFailureRules(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "compilefail", "*.go.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no compile-failure fixtures")
	}

	for _, path := range files {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			match := wantDiagnostic.FindSubmatch(source)
			if len(match) != 2 {
				t.Fatal("fixture has no // want: diagnostic marker")
			}
			want := string(match[1])

			fset := token.NewFileSet()
			file, parseErr := parser.ParseFile(fset, path, source, parser.AllErrors)
			var diagnostics []string
			if parseErr != nil {
				diagnostics = append(diagnostics, parseErr.Error())
			}
			if file != nil && parseErr == nil {
				config := types.Config{
					Importer: importer.Default(),
					Error: func(err error) {
						diagnostics = append(diagnostics, err.Error())
					},
				}
				_, _ = config.Check("compilefail", fset, []*ast.File{file}, nil)
			}

			all := strings.Join(diagnostics, "\n")
			if all == "" {
				t.Fatal("fixture compiled successfully")
			}
			if !strings.Contains(all, want) {
				t.Fatalf("diagnostic does not contain %q:\n%s", want, all)
			}
		})
	}
}
