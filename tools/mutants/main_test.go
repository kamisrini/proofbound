package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestTestArgsSerializesTaggedIntegrationPackages(t *testing.T) {
	original := testTags
	originalRun := testRunPattern
	t.Cleanup(func() { testTags = original })
	t.Cleanup(func() { testRunPattern = originalRun })

	testTags = ""
	testRunPattern = ""
	if got, want := testArgs("./..."), []string{"test", "-count=1", "./..."}; !reflect.DeepEqual(got, want) {
		t.Fatalf("untagged args=%v want=%v", got, want)
	}
	if got := testTimeout(); got != time.Minute {
		t.Fatalf("untagged timeout=%s", got)
	}

	testTags = "integration"
	if got, want := testArgs("./..."), []string{"test", "-count=1", "-p=1", "-tags", "integration", "./..."}; !reflect.DeepEqual(got, want) {
		t.Fatalf("tagged args=%v want=%v", got, want)
	}
	if got := testTimeout(); got != 2*time.Minute {
		t.Fatalf("tagged timeout=%s", got)
	}
	testRunPattern = "^TestIntent"
	if got, want := testArgs("./internal/projections"), []string{"test", "-count=1", "-p=1", "-tags", "integration", "-run", "^TestIntent", "./internal/projections"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("filtered tagged args=%v want=%v", got, want)
	}
}

func TestCopyTreeIncludesRepositoryMakefile(t *testing.T) {
	repository := t.TempDir()
	kernel := filepath.Join(repository, "kernel")
	if err := os.Mkdir(kernel, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, "Makefile"), []byte("check-witnessed:\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	destination := t.TempDir()
	copyTree(kernel, destination)
	data, err := os.ReadFile(filepath.Join(destination, "Makefile"))
	if err != nil || string(data) != "check-witnessed:\n" {
		t.Fatalf("makefile=%q error=%v", data, err)
	}
}

func TestCollectStaysWithinNamedPackage(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "root.go"), []byte("package root\nvar rootValue = 1 == 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(root, "child")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(child, "child.go"), []byte("package child\nvar childValue = 1 == 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mutants := collect(root)
	if len(mutants) != 1 || filepath.Base(mutants[0].file) != "root.go" {
		t.Fatalf("mutants=%+v", mutants)
	}
}

func TestCollectSkipsInactiveBuildFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "active.go"), []byte("package root\nvar activeValue = 1 == 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "inactive_windows.go"), []byte("package root\nvar inactiveValue = 1 == 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mutants := collect(root)
	if len(mutants) != 1 || filepath.Base(mutants[0].file) != "active.go" {
		t.Fatalf("mutants=%+v", mutants)
	}
}

func TestValidateRange(t *testing.T) {
	for _, valid := range [][2]int{{1, 1}, {2, 4}, {4, 4}} {
		if err := validateRange(valid[0], valid[1], 4); err != nil {
			t.Fatalf("range %v rejected: %v", valid, err)
		}
	}
	for _, invalid := range [][2]int{{0, 1}, {5, 5}, {3, 2}, {1, 5}} {
		if err := validateRange(invalid[0], invalid[1], 4); err == nil {
			t.Fatalf("range %v accepted", invalid)
		}
	}
}
