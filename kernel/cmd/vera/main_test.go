package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/connector/checks"
)

func TestRunRejectsUnknownCommand(t *testing.T) {
	for _, args := range [][]string{{"other", "checks"}, {"sync"}} {
		var stdout, stderr bytes.Buffer
		code := run(context.Background(), args, &stdout, &stderr)
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
		{[]string{"sync", "sessions"}, commandSyncSessions},
		{[]string{"sync", "all"}, commandSyncAll},
		{[]string{"rebuild"}, commandRebuild},
		{[]string{"verify"}, commandVerify},
		{[]string{"gates", "canary"}, commandGatesCanary},
		{[]string{"gates", "enforce"}, commandGatesEnforce},
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

func TestLatestSpoolWitnessUsesLatestULIDAndRejectsTrailingJSON(t *testing.T) {
	root := t.TempDir()
	spool := filepath.Join(root, ".vera", "spool")
	if err := os.MkdirAll(spool, 0o755); err != nil {
		t.Fatal(err)
	}
	makeWitness := func(runID string, startedAt time.Time) []byte {
		data, err := json.Marshal(checks.Witness{
			Schema: "vera.witness.v1", RunID: runID, Command: "make check",
			StartedAt: startedAt, FinishedAt: startedAt.Add(time.Second),
		})
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	oldID := "01K00000000000000000000000"
	newID := "01K00000000000000000000001"
	if err := os.WriteFile(filepath.Join(spool, oldID+".json"), makeWitness(oldID, time.Unix(1, 0)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(spool, newID+".json"), makeWitness(newID, time.Unix(2, 0)), 0o644); err != nil {
		t.Fatal(err)
	}
	witness, err := latestSpoolWitness(root)
	if err != nil || witness.RunID != newID {
		t.Fatalf("witness=%+v error=%v", witness, err)
	}
	if err := os.WriteFile(filepath.Join(spool, newID+".json"), append(makeWitness(newID, time.Unix(2, 0)), []byte("garbage")...), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := latestSpoolWitness(root); err == nil || !strings.Contains(err.Error(), "trailing") {
		t.Fatalf("expected trailing-data error, got %v", err)
	}
}

func TestRepositoryGitEnvCannotRedirectRepository(t *testing.T) {
	t.Setenv("GIT_DIR", "/tmp/foreign.git")
	t.Setenv("GIT_WORK_TREE", "/tmp/foreign-worktree")
	for _, value := range repositoryGitEnv() {
		if strings.HasPrefix(value, "GIT_") {
			t.Fatalf("repository-selection environment leaked: %s", value)
		}
	}
}
