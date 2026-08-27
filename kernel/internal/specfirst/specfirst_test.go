package specfirst

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

var proofCitation = regexp.MustCompile("\\|\\s*(?:[A-Za-z0-9]+-)?INV-[0-9]+\\s*\\|[^|]+\\|\\s*`?([A-Za-z0-9_.-]+_test\\.go)::(Test[A-Za-z0-9_]+)`?\\s*\\|")

func TestEveryInternalPackageHasSpecAndProof(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", ".."))
	internal := filepath.Join(root, "kernel", "internal")
	// The mutation harness copies the kernel module into a scratch root rather
	// than copying the entire repository. Keep the proof test valid in both
	// layouts without weakening its package walk.
	if _, err := os.Stat(internal); errors.Is(err, os.ErrNotExist) {
		scratchRoot := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
		internal = filepath.Join(scratchRoot, "internal")
	}
	err := filepath.WalkDir(internal, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			return nil
		}
		files, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		hasGo := false
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".go") && !strings.HasSuffix(file.Name(), "_test.go") {
				hasGo = true
				break
			}
		}
		if !hasGo {
			return nil
		}
		specPath := filepath.Join(path, "SPEC.md")
		spec, err := os.ReadFile(specPath)
		if err != nil {
			t.Errorf("%s: missing SPEC.md: %v", path, err)
			return nil
		}
		matches := proofCitation.FindAllStringSubmatch(string(spec), -1)
		for _, match := range matches {
			testPath := filepath.Join(path, match[1])
			if hasTestFunc(testPath, match[2]) {
				return nil
			}
		}
		t.Errorf("%s: SPEC.md has no resolvable file_test.go::TestName proving citation", path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func hasTestFunc(path, name string) bool {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return false
	}
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Recv == nil && function.Name.Name == name {
			return true
		}
	}
	return false
}
