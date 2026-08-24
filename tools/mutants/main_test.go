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
	t.Cleanup(func() { testTags = original })

	testTags = ""
	if got, want := testArgs("./..."), []string{"test", "./..."}; !reflect.DeepEqual(got, want) {
		t.Fatalf("untagged args=%v want=%v", got, want)
	}
	if got := testTimeout(); got != 10*time.Second {
		t.Fatalf("untagged timeout=%s", got)
	}

	testTags = "integration"
	if got, want := testArgs("./..."), []string{"test", "-p=1", "-tags", "integration", "./..."}; !reflect.DeepEqual(got, want) {
		t.Fatalf("tagged args=%v want=%v", got, want)
	}
	if got := testTimeout(); got != 30*time.Second {
		t.Fatalf("tagged timeout=%s", got)
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
