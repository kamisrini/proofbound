package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/connector/checks"
	"github.com/kamisrini/proofbound/kernel/internal/gates"
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

func TestRunUsesProofboundIdentity(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run(context.Background(), "proofbound", []string{"unknown"}, &stdout, &stderr); code != 2 {
		t.Fatalf("code=%d", code)
	}
	if strings.Contains(strings.ToLower(stderr.String()), "usage: vera") || !strings.Contains(stderr.String(), "usage: proofbound") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestOpenStoreUsesProofboundState(t *testing.T) {
	root := t.TempDir()
	if got, want := stateRoot(root), filepath.Join(root, ".proofbound"); got != want {
		t.Fatalf("state root=%q want=%q", got, want)
	}
}

func TestProductEnvPrefersProofbound(t *testing.T) {
	t.Setenv("PROOFBOUND_GITHUB_OWNER", "live")
	t.Setenv("VERA_GITHUB_OWNER", "legacy")
	if got := productEnv("PROOFBOUND_GITHUB_OWNER", "VERA_GITHUB_OWNER"); got != "live" {
		t.Fatalf("value=%q", got)
	}
}

func TestProductEnvLegacyAliasWarns(t *testing.T) {
	t.Setenv("PROOFBOUND_GITHUB_OWNER", "")
	if err := os.Unsetenv("PROOFBOUND_GITHUB_OWNER"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("VERA_GITHUB_OWNER", "legacy")
	var stderr bytes.Buffer
	warnLegacyEnvironment(&stderr)
	if got := productEnv("PROOFBOUND_GITHUB_OWNER", "VERA_GITHUB_OWNER"); got != "legacy" {
		t.Fatalf("value=%q", got)
	}
	if !strings.Contains(stderr.String(), "2026-12-31") || !strings.Contains(stderr.String(), "VERA_GITHUB_OWNER -> PROOFBOUND_GITHUB_OWNER") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestRunLegacyProgramWarns(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run(context.Background(), "vera", []string{"unknown"}, &stdout, &stderr); code != 2 {
		t.Fatalf("code=%d", code)
	}
	if !strings.Contains(stderr.String(), legacyAdvisory) || !strings.Contains(stderr.String(), usage) {
		t.Fatalf("stderr=%q", stderr.String())
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
		{[]string{"sync", "github"}, commandSyncGitHub},
		{[]string{"sync", "all"}, commandSyncAll},
		{[]string{"sync", "intent", "records"}, commandSyncIntentRecords},
		{[]string{"sync", "intent", "specdir"}, commandSyncIntentSpecdir},
		{[]string{"sync", "intent", "all"}, commandSyncIntentAll},
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

func TestParseIntentCommandsAndCommittedReader(t *testing.T) {
	for _, tc := range []struct {
		provider string
		command  command
	}{{"records", commandSyncIntentRecords}, {"specdir", commandSyncIntentSpecdir}, {"all", commandSyncIntentAll}} {
		if got := parseCommand([]string{"sync", "intent", tc.provider}); got != tc.command {
			t.Fatalf("provider=%s command=%d", tc.provider, got)
		}
	}
	root := t.TempDir()
	commands := [][]string{{"init"}, {"config", "user.name", "Test"}, {"config", "user.email", "test@example.invalid"}}
	for _, args := range commands {
		if out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	name := filepath.Join(root, "specs", "demo", "requirements.md")
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte("# committed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "specs/demo/requirements.md"}, {"commit", "-m", "fixture"}} {
		if out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	items, err := (committedIntentReader{root: root}).ReadSpecArtifacts(context.Background(), "HEAD")
	if err != nil || len(items) != 1 || string(items[0].Bytes) != "# committed\n" {
		t.Fatalf("items=%+v err=%v", items, err)
	}
}

func TestParseIntentReportAndCheckCommands(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want command
	}{
		{[]string{"report", "intent", "CI-demo-item-acde12"}, commandReportIntent},
		{[]string{"report", "requirement", "BR-demo-item-acde12"}, commandReportRequirement},
		{[]string{"intent", "check", "--commit", strings.Repeat("a", 40)}, commandIntentCheck},
	} {
		if got := parseCommand(tc.args); got != tc.want {
			t.Fatalf("args=%v got=%d want=%d", tc.args, got, tc.want)
		}
	}
}

func TestSyncAllOrdersIntentBeforeGit(t *testing.T) {
	source, err := os.ReadFile("cli.go")
	if err != nil {
		t.Fatal(err)
	}
	body := string(source)
	start := strings.Index(body, "case commandSyncGit, commandSyncAll:")
	if start < 0 {
		t.Fatal("sync-all case missing")
	}
	body = body[start:]
	intentAt, gitAt := strings.Index(body, "syncIntentOnStore(ctx"), strings.Index(body, "syncGit(ctx")
	if intentAt < 0 || gitAt < 0 || intentAt > gitAt {
		t.Fatalf("sync all order intent=%d git=%d", intentAt, gitAt)
	}
}

func TestEnforceGateResultsFailsClosed(t *testing.T) {
	definition := gates.Definition{Schema: gates.Version, ID: "x", Description: "d", Expires: "2099-01-01", Mode: "enforce", Source: "checks", Kind: "check.run", Condition: gates.Condition{Field: "exit_code", Equals: json.RawMessage("0")}}
	for _, state := range []gates.State{gates.StateBlocked, gates.StateUnknown} {
		if err := enforceGateResults([]gates.Definition{definition}, []gates.Result{{GateID: "x", State: state}}); err == nil {
			t.Fatalf("state %s accepted", state)
		}
	}
	if err := enforceGateResults(nil, nil); err == nil {
		t.Fatal("empty definitions accepted")
	}
	if err := enforceGateResults([]gates.Definition{definition}, []gates.Result{{GateID: "x", State: gates.StatePass}}); err != nil {
		t.Fatal(err)
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
	spool := filepath.Join(root, ".proofbound", "spool")
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
