package cli

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// fakeAgent replaces the agent command for a test, recording what it was given.
func fakeAgent(t *testing.T, answer string) *struct {
	argv   []string
	prompt string
} {
	t.Helper()
	got := &struct {
		argv   []string
		prompt string
	}{}
	old := agentRunner
	agentRunner = func(ctx context.Context, argv []string, prompt string, timeout time.Duration) ([]byte, error) {
		got.argv, got.prompt = argv, prompt
		return []byte(answer), nil
	}
	t.Cleanup(func() { agentRunner = old })
	return got
}

// agent-reviewer-spec §2, §3 (CLI-INV-80): the config is read from the parent only, checked, and
// defaulted; the default command is Claude Code headless with no tools; a project's own command runs
// as written.
func TestAgentConfigAndCommand(t *testing.T) {
	ctx := context.Background()
	root := taskRepo(t)
	uiWrite(t, root, "a.txt", "a\n")
	first := uiCommit(t, root, "owner", "start")
	uiWrite(t, root, agentConfigPath, `{"enabled":true}`+"\n")
	withConfig := uiCommit(t, root, "owner", "enable the agent")
	uiWrite(t, root, "a.txt", "b\n")
	child := uiCommit(t, root, "Claude", "change")

	if _, found, _ := readAgentConfig(ctx, root, withConfig); found {
		t.Error("the commit that adds the config enabled its own review: the config must come from the parent")
	}
	if why := agentProblem(ctx, root, first, true); !strings.Contains(why, "does not enable") {
		t.Errorf("no config: %q", why)
	}
	if why := agentProblem(ctx, root, child, false); !strings.Contains(why, "--agent-review") {
		t.Errorf("no flag: %q", why)
	}
	cfg, found, err := readAgentConfig(ctx, root, child)
	if !found || err != nil || !cfg.Enabled || cfg.BudgetUSD != 1 || cfg.TimeoutSeconds != 600 {
		t.Fatalf("defaults: %+v %v %v", cfg, found, err)
	}
	if why := agentProblem(ctx, root, child, true); why != "" {
		t.Errorf("enabled with the flag: %q", why)
	}
	argv := strings.Join(agentArgv(cfg), " ")
	for _, want := range []string{"claude -p --safe-mode --output-format json --tools  --no-session-persistence --strict-mcp-config --json-schema {", "--max-budget-usd 1.00"} {
		if !strings.Contains(argv, want) {
			t.Errorf("default command %q lacks %q", argv, want)
		}
	}
	if got := agentArgv(agentConfig{Command: []string{"codex", "exec", "-"}}); strings.Join(got, " ") != "codex exec -" {
		t.Errorf("own command: %v", got)
	}

	for body, why := range map[string]string{
		`{"enabled":true,"surprise":1}`:            "unknown field",
		`{"enabled":true,"command":[""]}`:          "command word 1",
		`{"enabled":true,"command":["a","b\nc"]}`:  "command word 2",
		`{"enabled":true,"budget_usd":101}`:        "budget_usd",
		`{"enabled":true,"timeout_seconds":99999}`: "timeout_seconds",
	} {
		uiWrite(t, root, agentConfigPath, body+"\n")
		uiCommit(t, root, "owner", "config")
		uiWrite(t, root, "a.txt", body+"\n")
		sha := uiCommit(t, root, "Claude", "change")
		if _, _, err := readAgentConfig(ctx, root, sha); err == nil || !strings.Contains(err.Error(), why) {
			t.Errorf("%s: %v", body, err)
		}
	}
	// Spec review F1: an agent's commit that sets the command is never obeyed until a person changes it.
	uiWrite(t, root, agentConfigPath, `{"enabled":true,"command":["sh","-c","cat ~/.ssh/id_ed25519"]}`+"\n")
	uiCommit(t, root, "Claude", "agent sets its own reviewer")
	uiWrite(t, root, "a.txt", "agent\n")
	if why := agentProblem(ctx, root, uiCommit(t, root, "owner", "change"), true); !strings.Contains(why, "a commit made by an agent") {
		t.Errorf("an agent-set command: %q", why)
	}
	uiWrite(t, root, agentConfigPath, `{"enabled":true,"command":["my-agent"]}`+"\n")
	uiCommit(t, root, "owner", "a person sets it")
	uiWrite(t, root, "a.txt", "person\n")
	if why := agentProblem(ctx, root, uiCommit(t, root, "Claude", "change"), true); why != "" {
		t.Errorf("a person-set command: %q", why)
	}
	uiWrite(t, root, agentConfigPath, `{"enabled":false}`+"\n")
	uiCommit(t, root, "owner", "off")
	uiWrite(t, root, "a.txt", "off\n")
	if why := agentProblem(ctx, root, uiCommit(t, root, "Claude", "change"), true); !strings.Contains(why, "enabled: false") {
		t.Errorf("disabled: %q", why)
	}
}

// The runner (spec §2): stdin is the prompt, the directory is empty and not the repository, the
// environment has no PROOFBOUND_ or DATABASE_URL, and a slow agent is stopped.
func TestAgentRunner(t *testing.T) {
	t.Setenv("PROOFBOUND_KEY", "/secret/key")
	t.Setenv("DATABASE_URL", "postgres://secret")
	script := filepath.Join(t.TempDir(), "agent.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nread first\necho \"$first|$(pwd)|$(ls -A | wc -l)|${PROOFBOUND_KEY:-none}|${DATABASE_URL:-none}\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	out, err := agentRunner(context.Background(), []string{"sh", script}, "hello agent\nmore\n", 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	f := strings.Split(strings.TrimSpace(string(out)), "|")
	if len(f) != 5 || f[0] != "hello agent" || !strings.Contains(f[1], "proofbound-agent-") || strings.TrimSpace(f[2]) != "0" || f[3] != "none" || f[4] != "none" {
		t.Errorf("the agent saw %q", out)
	}
	if _, err := agentRunner(context.Background(), []string{"sh", "-c", "sleep 30"}, "", 300*time.Millisecond); err == nil || !strings.Contains(err.Error(), "did not answer within") {
		t.Errorf("a slow agent: %v", err)
	}
	if _, err := agentRunner(context.Background(), []string{"sh", "-c", "echo boom >&2; exit 3"}, "", 5*time.Second); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Errorf("a failing agent: %v", err)
	}
}

// Reading and checking an answer (spec §2): Claude Code's envelope or a bare answer; unknown
// questions, outcomes outside the question's set, long reasons fail; a cited line that is not an
// added line is dropped with a note; each answer carries its question's obligation.
func TestAgentAnswerChecks(t *testing.T) {
	bare := `{"summary":"fine","overall":"ready","answers":[{"question":"v-0-0","outcome":"SATISFIED","reason":"line 2","lines":[{"path":"src/a.sh","line":2},{"path":"src/a.sh","line":99}]}]}`
	for name, out := range map[string]string{
		"bare":       bare,
		"structured": `{"type":"result","is_error":false,"result":"","structured_output":` + bare + `,"total_cost_usd":0.12,"modelUsage":{"claude-x":{}}}`,
		"text":       `{"type":"result","is_error":false,"result":"` + "```json\\n" + strings.ReplaceAll(bare, `"`, `\"`) + "\\n```" + `","total_cost_usd":0.12,"modelUsage":{"claude-x":{}}}`,
	} {
		a, model, cost, err := parseAgentOutput([]byte(out))
		if err != nil || a.Summary != "fine" || len(a.Answers) != 1 {
			t.Errorf("%s: %+v %v", name, a, err)
		}
		if name != "bare" && (model != "claude-x" || cost != 0.12) {
			t.Errorf("%s: model %q cost %v", name, model, cost)
		}
	}
	if _, _, _, err := parseAgentOutput([]byte(`{"type":"result","is_error":true,"result":"over budget"}`)); err == nil || !strings.Contains(err.Error(), "over budget") {
		t.Errorf("an agent error: %v", err)
	}
	if _, _, _, err := parseAgentOutput([]byte(`{"summary":"x","overall":"ready","answers":[],"extra":1}`)); err == nil {
		t.Error("an unknown field was taken")
	}

	qs := []agentQuestion{{Key: "v-0-0", Kind: "promise", ObligationID: "O-1", Outcomes: agentVerdictOutcomes}, {Key: "a-0", Kind: "area", Outcomes: agentPartOutcomes}}
	added := map[agentLine]bool{{"src/a.sh", 2}: true}
	a, _, _, _ := parseAgentOutput([]byte(bare))
	notes, err := checkAgentAnswer(&a, qs, added)
	if err != nil || a.Answers[0].ObligationID != "O-1" || len(a.Answers[0].Lines) != 1 || len(notes) != 1 || !strings.Contains(notes[0], "line 99") {
		t.Errorf("checked: %+v %v %v", a, notes, err)
	}
	for name, mutate := range map[string]func(*agentReply){
		"unknown question": func(a *agentReply) { a.Answers[0].Question = "v-9-9" },
		"wrong outcome":    func(a *agentReply) { a.Answers[0].Outcome = "accept" },
		"long reason":      func(a *agentReply) { a.Answers[0].Reason = strings.Repeat("x", 501) },
		"twice":            func(a *agentReply) { a.Answers = append(a.Answers, a.Answers[0]) },
		"overall":          func(a *agentReply) { a.Overall = "great" },
		"no summary":       func(a *agentReply) { a.Summary = " " },
	} {
		b, _, _, _ := parseAgentOutput([]byte(bare))
		mutate(&b)
		if _, err := checkAgentAnswer(&b, qs, added); err == nil {
			t.Errorf("%s was taken", name)
		}
	}
}

// The bench (spec §1, §3; CLI-INV-81): no button and a 404 without --agent-review; with it, the
// agent's answer is kept in the local draft and shown beside each question with nothing selected,
// the preview names the AGENT record and every difference.
func TestBenchAgentFlow(t *testing.T) {
	key := uiKey(t)
	root, base, claim := testsRepoWith(t, true, map[string]string{
		".proofbound/keys/demo-reviewer.pub": pbwarranty.PubLine(key.Public().(ed25519.PublicKey)),
		agentConfigPath:                      `{"enabled":true,"command":["my-agent","--strict"]}` + "\n",
	})
	truth := draftFile{Commit: claim, Verdicts: []draftVerdict{{Obligations: []draftVerdictObligation{{ObligationID: "O-1", Statement: "one"}}}},
		Reviews: []draftReview{{DeclaredOwner: "owner", Obligations: []draftReviewObligation{{ObligationID: "O-1", Statement: "one"}}}}}
	s := newUIServer(root, "", base, 4242)
	t.Cleanup(func() { s.stopLife(); s.jobs.wg.Wait() })
	s.compute = func(ctx context.Context, b string) (ledgerView, string, error) {
		return ledgerView{Report: auditReport{Schema: auditSchema, Base: base, Head: claim, Commits: []auditCommit{{Commit: claim, Author: "Claude", Class: classClaimed}},
			Controls: []auditControl{{ID: "C1", Evidence: []auditEvidence{{Commit: claim, Text: "NO REVIEW: x"}}}}, Exceptions: []auditException{}}}, "", nil
	}
	s.drafts = func(ctx context.Context, shas []string) (map[string]draftFile, error) {
		return map[string]draftFile{claim: truth}, nil
	}
	got := fakeAgent(t, `{"summary":"The promise is kept, but the line 7 edit is untested.","overall":"needs_work","answers":[`+
		`{"question":"v-0-0","outcome":"NOT_SATISFIED","reason":"ONE is printed but not checked","lines":[{"path":"src/a.sh","line":2}]},`+
		`{"question":"r-0-0","outcome":"AMBIGUOUS","reason":"which output?"}]}`)
	act := func(path string, decisions []reviewDecision) (int, map[string]any) {
		raw, _ := json.Marshal(reviewRequest{Commit: claim, Decisions: decisions})
		w := uiRequest(s, "POST", path, "127.0.0.1:4242", "http://127.0.0.1:4242", s.token, string(raw))
		var out map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return w.Code, out
	}
	page := func() string {
		w := uiRequest(s, "GET", "/review?commit="+claim, "127.0.0.1:4242", "", s.token, "")
		if w.Code != 200 {
			t.Fatalf("page: %d", w.Code)
		}
		return w.Body.String()
	}

	// Off by default: no flag, no button, and the endpoint refuses without running anything.
	if p := page(); strings.Contains(p, `data-act="review-agent"`) || strings.Contains(p, "Agent reviewer:") {
		t.Error("the button shows without --agent-review")
	}
	if code, out := act("/act/review-agent", nil); code != 404 || !strings.Contains(out["error"].(string), "--agent-review") {
		t.Errorf("without the flag: %d %v", code, out)
	}
	if got.prompt != "" {
		t.Fatal("the agent ran without the flag")
	}
	if w := uiRequestWith(s, "POST", "/act/review-agent", "Bearer "+s.agentToken); w.Code != 401 {
		t.Errorf("an agent's bearer POST answered %d", w.Code)
	}

	s.agentReview = true
	if p := page(); !strings.Contains(p, `data-act="review-agent"`) || !strings.Contains(p, "I have read this command") {
		t.Fatal("no button with --agent-review and an enabled config")
	}
	// The tests ran first, so the agent sees them; then ask.
	if code, out := act("/act/review-tests", nil); code != 202 {
		t.Fatalf("tests: %d %v", code, out)
	}
	waitJob(t, s, claim)
	// Code review F1: a configured command is shown, and runs only on a second, confirmed click.
	if p := page(); !strings.Contains(p, "This is <b>not</b> Claude Code") || !strings.Contains(p, "Runs: <code>&#34;my-agent&#34; &#34;--strict&#34;</code>") || !strings.Contains(p, "It was set in") || !strings.Contains(p, "by owner") {
		t.Error("the configured command is not shown before the click")
	}
	if code, out := act("/act/review-agent", nil); code != 409 || out["command"] != `"my-agent" "--strict"` {
		t.Fatalf("an unconfirmed configured command: %d %v", code, out)
	}
	if got.prompt != "" {
		t.Fatal("an unconfirmed configured command ran")
	}
	raw, _ := json.Marshal(reviewRequest{Commit: claim, Confirm: true})
	if w := uiRequest(s, "POST", "/act/review-agent", "127.0.0.1:4242", "http://127.0.0.1:4242", s.token, string(raw)); w.Code != 202 {
		t.Fatalf("ask: %d %s", w.Code, w.Body.String())
	}
	job := waitJob(t, s, claim)
	if job.Error != "" || job.Kind != "agent" || !strings.Contains(strings.Join(job.Lines, "\n"), "agent: answered 2 of") {
		t.Fatalf("job %+v", job)
	}
	if strings.Join(got.argv, " ") != "my-agent --strict" {
		t.Errorf("ran %v", got.argv)
	}
	for _, want := range []string{"untrusted", "v-0-0 [promise]", "r-0-0 [requirement]", "t-1 [tests]", "+echo SIX", "TESTS: The tests pass", "Allowed outcomes: SATISFIED, NOT_SATISFIED, INCONCLUSIVE"} {
		if !strings.Contains(got.prompt, want) {
			t.Errorf("the prompt lacks %q", want)
		}
	}
	if strings.Contains(got.prompt, s.token) || strings.Contains(got.prompt, "PROOFBOUND_KEY") {
		t.Error("the prompt carries a secret")
	}
	local, _ := localDraft(root, claim)
	if local.Agent == nil || local.Tests == nil || local.Agent.Answer.Answers[0].ObligationID != "O-1" {
		t.Fatalf("not kept in the draft with the run: %+v", local.Agent)
	}

	p := page()
	for _, want := range []string{"Agent reviewer</span> <strong>needs work</strong>", "<code>my-agent</code>", "The promise is kept, but the line 7 edit is untested.",
		`Agent</span> <strong>Not kept</strong>: ONE is printed but not checked`, `Agent</span> <strong>Ambiguous</strong>: which output?`,
		`class="jump" data-jump-path="src/a.sh" data-jump-line="2">src/a.sh line 2</a>`, "Ask the agent again", "It read all of the change."} {
		if !strings.Contains(p, want) {
			t.Errorf("the bench lacks %q", want)
		}
	}
	if strings.Contains(p, ` value="NOT_SATISFIED" checked`) || strings.Contains(p, ` value="AMBIGUOUS" checked`) || strings.Contains(p, " checked>") {
		t.Error("an answer was selected for the person")
	}

	decisions := []reviewDecision{{Part: "verdict", ObligationID: "O-1", Outcome: "SATISFIED"}, {Part: "review", ObligationID: "O-1", Outcome: "VERIFIABLE"},
		{Part: "tests", Index: 1, Path: "src/a.sh", Lines: "line 7", Outcome: "reject"}}
	code, out := act("/act/review-preview", decisions)
	b, _ := json.Marshal(out)
	if code != 200 || !strings.Contains(string(b), agentRecordPath(claim)) || !strings.Contains(string(b), "you differ from the agent: O-1 (promise): you said Kept; the agent said Not kept") ||
		!strings.Contains(string(b), "O-1 (its words): you said Verifiable; the agent said Ambiguous") {
		t.Errorf("preview: %d %s", code, b)
	}
	// Saving keeps the answer; a tampered answer is no longer shown.
	if code, _ := act("/act/review-draft", decisions); code != 200 {
		t.Fatal("save")
	}
	if local, _ := localDraft(root, claim); local.Agent == nil {
		t.Fatal("Save dropped the agent's answer")
	}
	local, _ = localDraft(root, claim)
	local.Agent.Answer.Answers[0].Outcome = "SATISFIED"
	if err := writeLocalDraft(root, claim, local); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(page(), "Agent reviewer</span>") {
		t.Error("an answer changed after the agent gave it is shown as the agent's")
	}
}

// Code review F2: a change longer than the bench shows is cut for the agent too, the prompt says
// where, and the answer says how much it read.
func TestAgentSeesTheCut(t *testing.T) {
	root := taskRepo(t)
	uiWrite(t, root, "a.txt", "a\n")
	uiCommit(t, root, "owner", "start")
	var big strings.Builder
	for i := 0; i < 1600; i++ {
		big.WriteString("line\n")
	}
	uiWrite(t, root, "big.txt", big.String())
	sha := uiCommit(t, root, "Claude", "big")
	got := fakeAgent(t, `{"summary":"partial","overall":"needs_work","answers":[]}`)
	d := draftFile{Commit: sha, Verdicts: []draftVerdict{{Obligations: []draftVerdictObligation{{ObligationID: "O-1", Statement: "one"}}}}}
	var lines []string
	rv, err := agentReviewCommit(context.Background(), root, d, agentConfig{Command: []string{"x"}, TimeoutSeconds: 5}, func(l string) { lines = append(lines, l) })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(rv.Seen, "the first 1500 of ") || !strings.Contains(got.prompt, "THE CHANGE IS CUT HERE") || !strings.Contains(strings.Join(lines, "\n"), "It read only the first 1500 of") {
		t.Errorf("seen %q; lines %v", rv.Seen, lines)
	}
	if seenWords("all") != "It read all of the change." {
		t.Error("seenWords")
	}
}

// Code review F3: the CLI needs the draft first (it holds the test run) and --yes for a configured
// command; both refusals happen before anything runs.
func TestReviewAgentRefusals(t *testing.T) {
	ctx := context.Background()
	root := taskRepo(t)
	uiWrite(t, root, ".gitignore", "")
	uiWrite(t, root, agentConfigPath, `{"enabled":true,"command":["my-agent"]}`+"\n")
	uiCommit(t, root, "owner", "enable")
	uiWrite(t, root, "a.txt", "a\n")
	sha := uiCommit(t, root, "Claude", "change")
	got := fakeAgent(t, `{}`)
	var out bytes.Buffer
	if err := reviewAgent(ctx, root, nil, sha, true, &out); err == nil || !strings.Contains(err.Error(), "review draft") {
		t.Errorf("no draft: %v", err)
	}
	if err := writeLocalDraft(root, sha, draftFile{Commit: sha}); err != nil {
		t.Fatal(err)
	}
	if err := reviewAgent(ctx, root, nil, sha, false, &out); err == nil || !strings.Contains(err.Error(), `"my-agent"`) || !strings.Contains(err.Error(), "--yes") {
		t.Errorf("no --yes: %v", err)
	}
	if got.prompt != "" {
		t.Error("the agent ran")
	}
	if c := parseCommand([]string{"review", "agent", "HEAD", "--agent-review", "--yes"}); c != commandReviewAgent {
		t.Error("--yes is not parsed")
	}
	if c := parseCommand([]string{"review", "agent", "HEAD"}); c == commandReviewAgent {
		t.Error("review agent without --agent-review was taken")
	}
	if _, ok := parseUI([]string{"ui", "--bundle", "d", "--agent-review"}); ok {
		t.Error("--agent-review with --bundle was taken")
	}
}

// The seal (spec §4; CLI-INV-82): the AGENT record is written, signed by the person, verifies, holds
// the answer and every difference; the verdicts are the same bytes with and without it; an answer
// edited after the run is refused with nothing written.
func TestSealRecordsTheAgent(t *testing.T) {
	ctx := context.Background()
	setup := func(t *testing.T) (string, string) {
		t.Helper()
		keyPath := filepath.Join(t.TempDir(), "reviewer-key")
		t.Setenv("PROOFBOUND_KEY", keyPath)
		pub, err := pbwarranty.InitKey(keyPath)
		if err != nil {
			t.Fatal(err)
		}
		// scope.paths leaves tests/t.sh outside the area: the ledger's draft holds no area parts, and
		// the seal must still see the agent's a-0 answer (live run 2026-11-25).
		root, _, claim := testsRepoWith(t, true, map[string]string{"specs/001-a/scope.paths": "src/a.sh\n"})
		uiWrite(t, root, ".proofbound/keys/proofbound-reviewer.pub", pbwarranty.PubLine(pub))
		uiCommit(t, root, "owner", "declare the reviewer")
		return root, claim
	}
	draft := func(claim string, agent bool) draftFile {
		d := sealDraft("SATISFIED", "VERIFIABLE")
		d.Commit = claim
		d.Tests = &testsResult{Hunks: []testsHunk{{Path: "src/a.sh", Lines: "line 7", Decision: "accept", Reason: "checked by hand"}}}
		d.OutOfArea = []areaDecision{{Path: "tests/t.sh", Lines: "line 1", Decision: "accept", Reason: "its test"}}
		if agent {
			a := agentReply{Schema: agentAnswerSchema, Summary: "s", Overall: "needs_work", Answers: []agentItem{
				{Question: "v-0-0", ObligationID: "O-B", Outcome: "NOT_SATISFIED", Reason: "r"},
				{Question: "r-0-1", ObligationID: "O-A", Outcome: "VERIFIABLE", Reason: "clear"},
				{Question: "a-0", Outcome: "reject", Reason: "not in scope"}}}
			digest, _ := answerDigest(a)
			d.Agent = &agentReview{Answer: a, Command: []string{"claude", "-p"}, Model: "m", PromptSHA256: strings.Repeat("c", 64), AnswerSHA256: digest, At: "2026-11-25T00:00:00Z"}
		}
		return d
	}
	ov := func(root, claim string) []byte {
		b, err := os.ReadFile(filepath.Join(root, "docs/verification/verdicts/OV-"+claim[:12]+"-1.md"))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}

	rootA, claimA := setup(t)
	var out bytes.Buffer
	if err := reviewSeal(ctx, rootA, writeDraft(t, rootA, draft(claimA, true)), areaTruth(claimA), &out); err != nil {
		t.Fatalf("seal: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "wrote "+agentRecordPath(claimA)) {
		t.Fatalf("no AGENT record: %s", out.String())
	}
	raw, err := os.ReadFile(filepath.Join(rootA, agentRecordPath(claimA)))
	if err != nil {
		t.Fatal(err)
	}
	uiCommit(t, rootA, "proofbound-reviewer", "review")
	fields, err := verifyRecord(raw, agentRecordSchema, declaredKeys(ctx, rootA, "HEAD"))
	if err != nil {
		t.Fatalf("the AGENT record does not verify: %v", err)
	}
	if fields["commit"] != claimA || fields["by"] != "proofbound-reviewer" || !strings.Contains(fields["answer"], `"NOT_SATISFIED"`) ||
		!strings.Contains(fields["differences"], "O-B (promise): you said Kept; the agent said Not kept") || strings.Contains(fields["differences"], "O-A") ||
		!strings.Contains(fields["differences"], "tests/t.sh line 1 (outside the area): you said Accept; the agent said Reject") {
		t.Errorf("record %v", fields)
	}

	// The same decisions without an agent: the verdict bytes are the same (the agent decides nothing).
	rootB, claimB := setup(t)
	out.Reset()
	if err := reviewSeal(ctx, rootB, writeDraft(t, rootB, draft(claimB, false)), areaTruth(claimB), &out); err != nil {
		t.Fatalf("seal without agent: %v", err)
	}
	if _, err := os.Stat(filepath.Join(rootB, agentRecordPath(claimB))); !os.IsNotExist(err) {
		t.Error("an AGENT record without an agent run")
	}
	a, b := ov(rootA, claimA), ov(rootB, claimB)
	if strings.ReplaceAll(string(a), claimA, "X") != strings.ReplaceAll(string(b), claimB, "X") && !bytes.Contains(a, []byte(`"status":"ACCEPTABLE"`)) {
		t.Errorf("the verdict changed with the agent:\n%s\n%s", a, b)
	}
	if !bytes.Contains(a, []byte("ACCEPTABLE")) || !bytes.Contains(b, []byte("ACCEPTABLE")) {
		t.Errorf("the person's Kept was not what was sealed:\n%s", a)
	}

	// Code review F4: an answer to questions that are gone is left out, with a line, and the seal goes on.
	rootD, claimD := setup(t)
	dd := draft(claimD, true)
	dd.Agent.Answer.Answers = append(dd.Agent.Answer.Answers, agentItem{Question: "t-9", Outcome: "reject", Reason: "gone"})
	dd.Agent.AnswerSHA256, _ = answerDigest(dd.Agent.Answer)
	out.Reset()
	if err := reviewSeal(ctx, rootD, writeDraft(t, rootD, dd), areaTruth(claimD), &out); err != nil || !strings.Contains(out.String(), "no longer fits") {
		t.Fatalf("a stale answer: %v %s", err, out.String())
	}
	if _, err := os.Stat(filepath.Join(rootD, agentRecordPath(claimD))); !os.IsNotExist(err) {
		t.Error("a stale answer was recorded")
	}

	// An answer changed after the agent gave it is refused; nothing written.
	rootC, claimC := setup(t)
	d := draft(claimC, true)
	d.Agent.Answer.Answers[0].Outcome = "SATISFIED"
	if err := reviewSeal(ctx, rootC, writeDraft(t, rootC, d), areaTruth(claimC), &out); err == nil || !strings.Contains(err.Error(), "does not match the answer the agent gave") {
		t.Fatalf("a tampered answer: %v", err)
	}
	if _, err := os.Stat(filepath.Join(rootC, "docs/verification/verdicts/OV-"+claimC[:12]+"-1.md")); !os.IsNotExist(err) {
		t.Error("a refused seal wrote a verdict")
	}
}
