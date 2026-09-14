package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/connector/checks"
	intentrecords "github.com/kamisrini/proofbound/kernel/internal/connector/intent/records"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/gates"
	"github.com/kamisrini/proofbound/kernel/internal/store"
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

func TestRunStopsWhenRepositoryRootIsMissing(t *testing.T) {
	old := workingDirectory
	workingDirectory = func() (string, error) { return "/proofbound-no-repository/child", nil }
	t.Cleanup(func() { workingDirectory = old })
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), []string{"verify"}, &stdout, &stderr); code != 1 {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), "repository root not found") || strings.Contains(stderr.String(), "timed out") {
		t.Fatalf("stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestVerificationCommandClassification(t *testing.T) {
	if !isVerificationCommand(commandVerify) {
		t.Fatal("verify command was not classified for its deadline")
	}
	if isVerificationCommand(commandSyncAll) {
		t.Fatal("non-verification command received the verify deadline")
	}
}

func TestRunUsesProofboundIdentity(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run(context.Background(), "proofbound", []string{"unknown"}, &stdout, &stderr); code != 2 {
		t.Fatalf("code=%d", code)
	}
	if strings.Contains(strings.ToLower(stderr.String()), "usage: ve"+"ra") || !strings.Contains(stderr.String(), "usage: proofbound") {
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
	if got := productEnv("PROOFBOUND_GITHUB_OWNER"); got != "live" {
		t.Fatalf("value=%q", got)
	}
}

func TestLegacyEnvironmentIsIgnored(t *testing.T) {
	legacy := "VE" + "RA_GITHUB_OWNER"
	t.Setenv(legacy, "legacy")
	if err := os.Unsetenv("PROOFBOUND_GITHUB_OWNER"); err != nil {
		t.Fatal(err)
	}
	if got := productEnv("PROOFBOUND_GITHUB_OWNER"); got != "" {
		t.Fatalf("value=%q", got)
	}
}

func TestLegacyCommandWrapperIsAbsent(t *testing.T) {
	if _, err := os.Stat(filepath.Join("..", "..", "cmd", "ve"+"ra")); !os.IsNotExist(err) {
		t.Fatalf("legacy command wrapper still exists: %v", err)
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
		{[]string{"sync", "reviews"}, commandSyncReviews},
		{[]string{"sync", "github"}, commandSyncGitHub},
		{[]string{"sync", "all"}, commandSyncAll},
		{[]string{"sync", "intent", "records"}, commandSyncIntentRecords},
		{[]string{"sync", "intent", "specdir"}, commandSyncIntentSpecdir},
		{[]string{"sync", "intent", "all"}, commandSyncIntentAll},
		{[]string{"rebuild"}, commandRebuild},
		{[]string{"verify"}, commandVerify},
		{[]string{"report", "week"}, commandReportWeek},
		{[]string{"report", "github"}, commandReportGitHub},
		{[]string{"gates", "canary"}, commandGatesCanary},
		{[]string{"gates", "enforce"}, commandGatesEnforce},
		{[]string{"migrate", "historical-evidence"}, commandMigrateHistoricalEvidence},
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
	verdictName := filepath.Join(root, "docs", "verification", "verdicts", "example.md")
	if err := os.MkdirAll(filepath.Dir(verdictName), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(verdictName, []byte("# committed verdict\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "specs/demo/requirements.md", "docs/verification/verdicts/example.md"}, {"commit", "-m", "fixture"}} {
		if out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	items, err := (committedIntentReader{root: root}).ReadSpecArtifacts(context.Background(), "HEAD")
	if err != nil || len(items) != 1 || string(items[0].Bytes) != "# committed\n" {
		t.Fatalf("items=%+v err=%v", items, err)
	}
	verdicts, err := (committedVerdictReader{root: root}).ReadCommittedVerdicts(context.Background())
	if err != nil || len(verdicts) != 1 || verdicts[0].Path != "docs/verification/verdicts/example.md" || string(verdicts[0].Bytes) != "# committed verdict\n" {
		t.Fatalf("verdicts=%+v err=%v", verdicts, err)
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
	sha := strings.Repeat("a", 40)
	for _, args := range [][]string{
		{"other", "intent", "CI-demo-item-acde12"},
		{"report", "other", "CI-demo-item-acde12"},
		{"report", "intent", ""},
		{"other", "requirement", "BR-demo-item-acde12"},
		{"report", "other", "BR-demo-item-acde12"},
		{"report", "requirement", ""},
		{"other", "check", "--commit", sha},
		{"intent", "other", "--commit", sha},
		{"intent", "check", "--other", sha},
		{"intent", "check", "--commit", ""},
	} {
		if got := parseCommand(args); got != commandInvalid {
			t.Fatalf("invalid args=%v got=%d", args, got)
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
	if got := strings.Count(body, "if cmd == commandSyncAll {"); got != 1 {
		t.Fatalf("sync-all selection branches=%d", got)
	}
}

func TestSyncIntentChecksSpecdirProviderError(t *testing.T) {
	source, err := os.ReadFile("cli.go")
	if err != nil {
		t.Fatal(err)
	}
	body := string(source)
	start, end := strings.Index(body, "func syncIntentOnStore("), strings.Index(body, "func syncReviewsOnStore(")
	if start < 0 || end <= start {
		t.Fatal("sync intent implementation boundaries missing")
	}
	body = body[start:end]
	if !strings.Contains(body, "recordsProvider, err := intentrecords.New(reader, now)\n\tif err != nil") {
		t.Fatal("records provider error is not checked")
	}
	needle := "specdirProvider, err := intentspecdir.New(reader, now)\n\tif err != nil"
	if !strings.Contains(body, needle) {
		t.Fatal("specdir provider error is not checked")
	}
}

func TestCommitIntentResolverReturnsResolvedRevision(t *testing.T) {
	root := os.Getenv("PROOFBOUND_TEST_REPO_ROOT")
	if root == "" {
		var err error
		root, err = filepath.Abs(filepath.Join("..", "..", ".."))
		if err != nil {
			t.Fatal(err)
		}
	}
	output, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	provider, err := intentrecords.New(committedIntentReader{root: root}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := (commitIntentResolver{records: provider}).Resolve(context.Background(), strings.TrimSpace(string(output)), "records", "CI-implement-p5-0a1b2c")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Provider != "records" || resolved.RecordID != "CI-implement-p5-0a1b2c" || resolved.ArtifactSHA256 != "f90b216573f3f88eee61029f81d180c5ae1dd70dcb919d9662f97cfd76395424" {
		t.Fatalf("resolved=%+v", resolved)
	}
}

func TestVerifyChecksEveryStageError(t *testing.T) {
	source, err := os.ReadFile("cli.go")
	if err != nil {
		t.Fatal(err)
	}
	body := string(source)
	start, end := strings.Index(body, "func verify("), strings.Index(body, "func verifyStep(")
	if start < 0 || end <= start {
		t.Fatal("verify implementation boundaries missing")
	}
	body = body[start:end]
	stages := strings.Count(body, "if err := verifyStep(")
	checked := strings.Count(body, "}); err != nil {")
	if stages != 16 || checked != stages {
		t.Fatalf("verify stages=%d checked=%d", stages, checked)
	}
}

func TestVerifyStepFailsClosed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	if err := verifyStep(ctx, "cancelled", func() error {
		called = true
		return nil
	}); !errors.Is(err, context.Canceled) || called {
		t.Fatalf("cancelled step err=%v called=%t", err, called)
	}

	want := errors.New("stage failed")
	if err := verifyStep(context.Background(), "failing", func() error { return want }); !errors.Is(err, want) {
		t.Fatalf("step error=%v", err)
	}

	t.Setenv("PROOFBOUND_VERIFY_TRACE", "1")
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStderr := os.Stderr
	os.Stderr = writer
	t.Cleanup(func() {
		os.Stderr = oldStderr
		_ = writer.Close()
	})
	if err := verifyStep(context.Background(), "traced", func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	trace, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(trace) != "verify: begin traced\n" {
		t.Fatalf("trace=%q", trace)
	}
}

type fakeEventReader struct {
	records []store.Record
	err     error
}

func (f fakeEventReader) ReadEvents(_ context.Context, _ store.Filter, yield func(store.Record) error) error {
	if f.err != nil {
		return f.err
	}
	for _, record := range f.records {
		if err := yield(record); err != nil {
			return err
		}
	}
	return nil
}

func TestLatestLedgerCheckRunIDRequiresCheckEvent(t *testing.T) {
	if _, err := latestLedgerCheckRunID(context.Background(), fakeEventReader{}); err == nil || !strings.Contains(err.Error(), "no check.run event") {
		t.Fatalf("empty ledger result err=%v", err)
	}
	want := "check-run-1"
	got, err := latestLedgerCheckRunID(context.Background(), fakeEventReader{records: []store.Record{{Event: core.Event{NativeID: want}}}})
	if err != nil || got != want {
		t.Fatalf("latest=%q err=%v", got, err)
	}
}

func TestSyncGitRejectsInvalidRepository(t *testing.T) {
	ids, err := newIDs()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := syncGit(context.Background(), t.TempDir(), nil, ids); err == nil || !strings.Contains(err.Error(), "not a work tree") {
		t.Fatalf("invalid repository err=%v", err)
	}
}

func TestGateExecutionEnforcesOnlyEnforceCommand(t *testing.T) {
	source, err := os.ReadFile("cli.go")
	if err != nil {
		t.Fatal(err)
	}
	body := string(source)
	start, end := strings.Index(body, "func runCommand("), strings.Index(body, "func enforceDefinitions(")
	if start < 0 || end <= start {
		t.Fatal("gate execution boundaries missing")
	}
	body = body[start:end]
	if got := strings.Count(body, "if cmd == commandGatesEnforce {"); got != 1 {
		t.Fatalf("enforcement branches=%d", got)
	}
}

func TestResolveIntentGateHeadScope(t *testing.T) {
	root := t.TempDir()
	for _, args := range [][]string{{"init"}, {"config", "user.name", "Test"}, {"config", "user.email", "test@example.invalid"}, {"commit", "--allow-empty", "-m", "head"}} {
		if out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	definitions := []gates.Definition{{Rule: "intent-delivery-readiness", ScopeCommit: "HEAD"}, {Rule: "intent-reference-integrity"}}
	if err := resolveIntentGateHeadScope(context.Background(), root, definitions); err != nil {
		t.Fatal(err)
	}
	if len(definitions[0].ScopeCommit) != 40 || definitions[1].ScopeCommit != "" {
		t.Fatalf("definitions=%+v", definitions)
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

func TestEnforcementSelectsOnlyPromotedDefinitions(t *testing.T) {
	definitions := []gates.Definition{
		{ID: "observational", Mode: "canary"},
		{ID: "controlled-boundary", Mode: "enforce"},
	}
	got := enforceDefinitions(definitions)
	if len(got) != 1 || got[0].ID != "controlled-boundary" {
		t.Fatalf("promoted=%+v", got)
	}
	if definitions[0].Mode != "canary" {
		t.Fatal("selection mutated the canary definition")
	}
}

func TestGateCommandSelectsDefinitionsAtBoundary(t *testing.T) {
	condition := gates.Condition{Field: "exit_code", Equals: json.RawMessage("0")}
	definitions := []gates.Definition{
		{Schema: gates.Version, ID: "observational", Description: "d", Expires: "2099-01-01", Mode: "canary", Source: "checks", Kind: "check.run", Condition: condition},
		{Schema: gates.Version, ID: "controlled-boundary", Description: "d", Expires: "2099-01-01", Mode: "enforce", Source: "checks", Kind: "check.run", Condition: condition},
	}
	canary, err := gateDefinitionsForCommand(commandGatesCanary, definitions)
	if err != nil || len(canary) != 2 {
		t.Fatalf("canary=%+v err=%v", canary, err)
	}
	enforce, err := gateDefinitionsForCommand(commandGatesEnforce, definitions)
	if err != nil || len(enforce) != 1 || enforce[0].ID != "controlled-boundary" {
		t.Fatalf("enforce=%+v err=%v", enforce, err)
	}
	if _, err := gateDefinitionsForCommand(commandGatesEnforce, definitions[:1]); err == nil {
		t.Fatal("enforcement accepted an empty promoted definition set")
	}
	expired := definitions[1:]
	expired[0].Expires = "2020-01-01"
	if _, err := gateDefinitionsForCommand(commandGatesEnforce, expired); err == nil {
		t.Fatal("enforcement accepted an expired promoted definition")
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
	if err := os.WriteFile(filepath.Join(spool, newID+".json"), append(makeWitness(newID, time.Unix(2, 0)), []byte(` {}`)...), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := latestSpoolWitness(root); err == nil || !strings.Contains(err.Error(), "contains trailing JSON") {
		t.Fatalf("expected second-value error, got %v", err)
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
