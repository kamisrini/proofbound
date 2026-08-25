package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRejectsUnknownCommand(t *testing.T) {
	for _, args := range [][]string{{"other", "checks"}, {"sync"}, {"sync", "sessions"}} {
		var stdout, stderr bytes.Buffer
		code := run(context.Background(), args, &stdout, &stderr)
		if args[1:] != nil && len(args) == 2 && args[1] == "sessions" {
			if code != 1 || !strings.Contains(stderr.String(), "not implemented") {
				t.Fatalf("args=%v code=%d stdout=%q stderr=%q", args, code, stdout.String(), stderr.String())
			}
			continue
		}
		if code != 2 {
			t.Fatalf("args=%v code=%d stdout=%q stderr=%q", args, code, stdout.String(), stderr.String())
		}
		if stdout.Len() != 0 || stderr.String() != usage+"\n" {
			t.Fatalf("args=%v stdout=%q stderr=%q", args, stdout.String(), stderr.String())
		}
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		args []string
		want command
	}{
		{[]string{"sync", "git"}, commandSyncGit},
		{[]string{"sync", "checks"}, commandSyncChecks},
		{[]string{"sync", "all"}, commandSyncAll},
		{[]string{"rebuild"}, commandRebuild},
		{[]string{"verify"}, commandVerify},
		{[]string{"sync", "sessions"}, commandInvalid},
		{[]string{"sync", "git", "extra"}, commandInvalid},
	}
	for _, tt := range tests {
		if got := parseCommand(tt.args); got != tt.want {
			t.Errorf("parseCommand(%v) = %d, want %d", tt.args, got, tt.want)
		}
	}
}

func TestRepositoryRootWalksUpward(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "kernel", "cmd")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(nested); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
	got, err := repositoryRoot()
	if err != nil || got != root {
		t.Fatalf("root=%q error=%v", got, err)
	}
}

func TestRepositoryRootRejectsOutsideRepository(t *testing.T) {
	if got, err := repositoryRootFrom(string(filepath.Separator)); err == nil || got != "" || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("root=%q error=%v", got, err)
	}
}
