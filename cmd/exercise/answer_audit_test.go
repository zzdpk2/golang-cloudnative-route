//go:build starteraudit

package main

// This maintenance-only test protects the red starter distributed to learners.
// It is deliberately excluded from normal test runs so a learner's completed
// implementation is not treated as a regression.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

var completedImplementationAllowlist = []string{
	"cmd/exercise/", // exercise runner and maintenance checks
	"cmd/orderd/main.go",
	"cmd/orderd/preflight.go", // documented composition-root scaffolding
	"cmd/seniorcheck/main.go", // senior evidence CLI
	"internal/senior/acceptance/manifest.go",
	"internal/senior/acceptance/run.go",      // release-gate validation infrastructure
	"internal/senior/testsupport/support.go", // deterministic test infrastructure
	"internal/foundation/testing/harness.go",
	"internal/foundation/testing/subject.go", // mutation-test subject supplied to the learner
	"internal/commerce/testkit/subject.go",   // collaborator subject supplied to the learner
}

var forbiddenCommentSolutionMarkers = []string{
	"completed answer:",
	"completed solution:",
	"copy this implementation",
	"exact implementation:",
	"solution code:",
}

func TestLearnerOwnedGoFunctionsContainNoCompletedAnswers(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, relativeRoot := range []string{"internal", "cmd"} {
		walkRoot := filepath.Join(root, filepath.FromSlash(relativeRoot))
		err := filepath.WalkDir(walkRoot, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			relative, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			normal := filepath.ToSlash(relative)
			if entry.IsDir() {
				if entry.Name() == ".attempts" {
					t.Errorf("saved solution directory is forbidden: %s", normal)
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(normal, ".go") || strings.HasSuffix(normal, "_test.go") ||
				isCompletedImplementationAllowed(normal) {
				return nil
			}
			file, parseErr := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if parseErr != nil {
				return parseErr
			}
			for _, group := range file.Comments {
				comment := strings.ToLower(group.Text())
				for _, marker := range forbiddenCommentSolutionMarkers {
					if strings.Contains(comment, marker) {
						position := fset.Position(group.Pos())
						t.Errorf("solution marker %q in learner comment: %s:%d", marker, normal, position.Line)
					}
				}
			}
			for _, declaration := range file.Decls {
				function, ok := declaration.(*ast.FuncDecl)
				if !ok || function.Body == nil || bodyIsTODOPanic(function.Body) {
					continue
				}
				position := fset.Position(function.Pos())
				t.Errorf("completed learner implementation: %s:%d %s", normal, position.Line, function.Name.Name)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func isCompletedImplementationAllowed(path string) bool {
	for _, allowed := range completedImplementationAllowlist {
		if strings.HasSuffix(allowed, "/") && strings.HasPrefix(path, allowed) {
			return true
		}
		if path == allowed {
			return true
		}
	}
	return false
}

func bodyIsTODOPanic(body *ast.BlockStmt) bool {
	if len(body.List) != 1 {
		return false
	}
	expression, ok := body.List[0].(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := expression.X.(*ast.CallExpr)
	if !ok || len(call.Args) != 1 {
		return false
	}
	identifier, ok := call.Fun.(*ast.Ident)
	if !ok || identifier.Name != "panic" {
		return false
	}
	literal, ok := call.Args[0].(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return false
	}
	value, err := strconv.Unquote(literal.Value)
	return err == nil && strings.Contains(value, "TODO")
}
