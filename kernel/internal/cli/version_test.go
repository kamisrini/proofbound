package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/store"
)

// CLI-INV-25: `version` prints the stamped version and exits 0 outside any repository without
// opening a store; any extra argument is a usage error.
func TestVersionRunsAnywhere(t *testing.T) {
	oldWD, oldOpen, oldVersion := workingDirectory, openStore, Version
	workingDirectory = func() (string, error) { return "/proofbound-no-repository/child", nil }
	opened := false
	openStore = func(context.Context, string, string) (*store.Store, error) {
		opened = true
		return nil, nil
	}
	Version = "v9.8.7-test"
	t.Cleanup(func() { workingDirectory, openStore, Version = oldWD, oldOpen, oldVersion })

	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), []string{"version"}, &stdout, &stderr); code != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
	if got := stdout.String(); got != "proofbound v9.8.7-test\n" || stderr.Len() != 0 || opened {
		t.Fatalf("stdout=%q stderr=%q opened=%v", got, stderr.String(), opened)
	}

	stdout.Reset()
	stderr.Reset()
	if code := run(context.Background(), []string{"version", "x"}, &stdout, &stderr); code != 2 {
		t.Fatalf("version x: code=%d", code)
	}
	if stdout.Len() != 0 || !bytes.Contains(stderr.Bytes(), []byte("usage: proofbound version")) {
		t.Fatalf("version x: stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}
