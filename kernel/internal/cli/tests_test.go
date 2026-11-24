package cli

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

const testsJSON = `{"schema":"proofbound.tests.v1","run":"sh tests/t.sh","paths":["tests/"],"timeout_seconds":20}`

// testsRepo: a parent that declares tests and a spec; then a claiming commit that changes two
// separate lines of src/a.sh (the test checks only the first) and adds a test under tests/.
func testsRepo(t *testing.T, declare bool) (root, base, claim string) {
	t.Helper()
	return testsRepoWith(t, declare, nil)
}

// testsRepoWith also commits extra files (for example declared keys) in the first commit.
func testsRepoWith(t *testing.T, declare bool, extra map[string]string) (root, base, claim string) {
	t.Helper()
	root = taskRepo(t)
	for rel, body := range extra {
		uiWrite(t, root, rel, body)
	}
	lines := []string{"#!/bin/sh", "echo one", "echo two", "echo three", "echo four", "echo five", "echo six"}
	uiWrite(t, root, "src/a.sh", strings.Join(lines, "\n")+"\n")
	base = uiCommit(t, root, "owner", "start")
	if declare {
		uiWrite(t, root, testsConfigPath, testsJSON+"\n")
	}
	uiWrite(t, root, "specs/001-a/spec.md", "# a\n")
	uiCommit(t, root, "owner", "spec")
	lines[1], lines[6] = "echo ONE", "echo SIX"
	uiWrite(t, root, "src/a.sh", strings.Join(lines, "\n")+"\n")
	uiWrite(t, root, "tests/t.sh", "sh src/a.sh | grep -q '^ONE$'\n")
	uiWrite(t, root, "docs/n.md", "note\n")
	claim = uiCommit(t, root, "Claude", "change a\n\nIntent: sdd:001-a")
	return root, base, claim
}

// CLI-INV-66: the declaration is read from the parent; each code hunk (never test paths or
// housekeeping) is undone alone, then all at once, in a worktree that is removed afterwards.
func TestCommitTests(t *testing.T) {
	ctx := context.Background()
	t.Run("each hunk and the whole change", func(t *testing.T) {
		root, _, claim := testsRepo(t, true)
		res, err := commitTests(ctx, root, claim)
		if err != nil {
			t.Fatal(err)
		}
		if !res.Declared || res.Baseline != testsPass || res.Whole != testsCaught || res.Run != "sh tests/t.sh" {
			t.Fatalf("result %+v", res)
		}
		want := []testsHunk{{Path: "src/a.sh", Lines: "line 2", Result: testsCaught}, {Path: "src/a.sh", Lines: "line 7", Result: testsNotCaught}}
		if len(res.Hunks) != 2 || res.Hunks[0] != want[0] || res.Hunks[1] != want[1] {
			t.Fatalf("hunks %+v, want %+v", res.Hunks, want)
		}
		if wt := gitOutput(ctx, root, "worktree", "list"); strings.Count(wt, "\n") != 1 {
			t.Errorf("a worktree was left behind:\n%s", wt)
		}
		if st := gitOutput(ctx, root, "status", "--porcelain"); st != "" {
			t.Errorf("the checkout changed:\n%s", st)
		}
	})
	t.Run("not declared runs nothing", func(t *testing.T) {
		root, _, claim := testsRepo(t, false)
		ran := 0
		old := testsRunner
		t.Cleanup(func() { testsRunner = old })
		testsRunner = func(context.Context, string, string, time.Duration) string { ran++; return testsPass }
		res, err := commitTests(ctx, root, claim)
		if err != nil || res.Declared || ran != 0 {
			t.Fatalf("declared=%v ran=%d err=%v", res.Declared, ran, err)
		}
	})
	t.Run("a declaration in the commit itself does not count", func(t *testing.T) {
		root := taskRepo(t)
		uiWrite(t, root, "specs/001-a/spec.md", "# a\n")
		uiCommit(t, root, "owner", "spec")
		uiWrite(t, root, testsConfigPath, testsJSON+"\n")
		uiWrite(t, root, "src/a.sh", "echo x\n")
		claim := uiCommit(t, root, "Claude", "x\n\nIntent: sdd:001-a")
		if res, _ := commitTests(ctx, root, claim); res.Declared {
			t.Fatal("a tests.json added by the claiming commit was used")
		}
	})
	t.Run("a failing baseline judges nothing", func(t *testing.T) {
		root, _, _ := testsRepo(t, true)
		uiWrite(t, root, testsConfigPath, strings.Replace(testsJSON, "sh tests/t.sh", "exit 1", 1)+"\n")
		uiCommit(t, root, "owner", "break the tests")
		uiWrite(t, root, "src/a.sh", "echo ONE\n")
		claim := uiCommit(t, root, "Claude", "again\n\nIntent: sdd:001-a")
		res, err := commitTests(ctx, root, claim)
		if err != nil || res.Baseline != testsFail || res.Whole != testsNotRun || len(res.Hunks) == 0 || res.Hunks[0].Result != testsNotRun {
			t.Fatalf("%+v %v", res, err)
		}
	})
	t.Run("invalid declarations", func(t *testing.T) {
		for _, bad := range []string{`{"schema":"x","run":"a"}`, `{"schema":"proofbound.tests.v1","run":""}`, `{"schema":"proofbound.tests.v1","run":"a","paths":["../x"]}`, `{"schema":"proofbound.tests.v1","run":"a","extra":1}`, `{"schema":"proofbound.tests.v1","run":"a","timeout_seconds":99999}`} {
			root := taskRepo(t)
			uiWrite(t, root, testsConfigPath, bad+"\n")
			head := uiCommit(t, root, "owner", "declare")
			if _, found, err := readTestsConfig(ctx, root, head); !found || err == nil {
				t.Errorf("%s: found=%v err=%v", bad, found, err)
			}
		}
	})
	t.Run("a hanging test is a timeout, counted as caught", func(t *testing.T) {
		dir := t.TempDir()
		if got := testsRunner(ctx, dir, "sleep 5", 200*time.Millisecond); got != testsTimeout {
			t.Fatalf("got %s", got)
		}
		if caughtWord(testsTimeout) != testsCaught {
			t.Fatal("a timeout must count as the tests noticing")
		}
	})
}

// CLI-INV-67: seal runs the tests again, refuses undecided NOT CAUGHT hunks, makes the verdict
// NEEDS_WORK on a reject, and writes a signed TESTS- record.
func TestSealAccountsForTests(t *testing.T) {
	ctx := context.Background()
	setup := func(t *testing.T) (string, string) {
		t.Helper()
		keyPath := filepath.Join(t.TempDir(), "reviewer-key")
		t.Setenv("PROOFBOUND_KEY", keyPath)
		pub, err := pbwarranty.InitKey(keyPath)
		if err != nil {
			t.Fatal(err)
		}
		root, _, claim := testsRepo(t, true)
		uiWrite(t, root, ".proofbound/keys/proofbound-reviewer.pub", pbwarranty.PubLine(pub))
		uiCommit(t, root, "owner", "declare the reviewer")
		return root, claim
	}
	seal := func(t *testing.T, root, claim string, decision, reason string) (string, error) {
		t.Helper()
		d := sealDraft("SATISFIED", "VERIFIABLE")
		d.Commit = claim
		if decision != "" {
			d.Tests = &testsResult{Hunks: []testsHunk{{Path: "src/a.sh", Lines: "line 7", Decision: decision, Reason: reason}}}
		}
		var out bytes.Buffer
		err := reviewSeal(ctx, root, writeDraft(t, root, d), areaTruth(claim), &out)
		return out.String(), err
	}
	status := func(t *testing.T, root, claim string) string {
		ov, _ := os.ReadFile(filepath.Join(root, "docs/verification/verdicts/OV-"+claim[:12]+"-1.md"))
		return string(ov)
	}
	t.Run("undecided is refused", func(t *testing.T) {
		root, claim := setup(t)
		if _, err := seal(t, root, claim, "", ""); err == nil || !strings.Contains(err.Error(), "src/a.sh line 7") {
			t.Fatalf("sealed with an undecided hunk: %v", err)
		}
		if _, err := os.Lstat(filepath.Join(root, testsRecordPath(claim))); err == nil {
			t.Fatal("a refused seal wrote a record")
		}
	})
	t.Run("accept without a reason is refused", func(t *testing.T) {
		root, claim := setup(t)
		if _, err := seal(t, root, claim, "accept", " "); err == nil {
			t.Fatal("accepted without a reason")
		}
	})
	t.Run("accept writes a signed record, verdict acceptable", func(t *testing.T) {
		root, claim := setup(t)
		out, err := seal(t, root, claim, "accept", "cosmetic output only")
		if err != nil || !strings.Contains(out, "wrote "+testsRecordPath(claim)) || !strings.Contains(status(t, root, claim), `"status":"ACCEPTABLE"`) {
			t.Fatalf("%v\n%s", err, out)
		}
		uiCommit(t, root, "proofbound-reviewer", "review")
		fields, hunks, err := readTestsRecord(ctx, root, claim)
		if err != nil || fields["by"] != "proofbound-reviewer" || fields["whole"] != testsCaught || len(hunks) != 2 || hunks[1].Decision != "accept" {
			t.Fatalf("%v %v %+v", err, fields, hunks)
		}
	})
	t.Run("reject makes the verdict NEEDS_WORK", func(t *testing.T) {
		root, claim := setup(t)
		if _, err := seal(t, root, claim, "reject", ""); err != nil || !strings.Contains(status(t, root, claim), `"status":"NEEDS_WORK"`) {
			t.Fatalf("%v\n%s", err, status(t, root, claim))
		}
	})
}

// CLI-INV-68: audit control C11 reads the signed record and checks it against the hunks git lists
// and the command declared in the parent; it runs nothing.
func TestAuditC11(t *testing.T) {
	ctx := context.Background()
	key := uiKey(t)
	control := func(root, base, claim string) (auditControl, []auditException) {
		t.Helper()
		rep := auditReport{Commits: []auditCommit{{Commit: claim, Subject: "change a", Class: classClaimed, Author: "Claude", Agents: []string{"Claude"}}}, Roles: []auditRole{}}
		var ex []auditException
		c, err := testsControl(ctx, root, base, &rep, func(kind, ctl, commit, text string) {
			ex = append(ex, auditException{Kind: kind, Control: ctl, Commit: commit, Text: text})
		})
		if err != nil {
			t.Fatal(err)
		}
		return c, ex
	}
	ran := 0
	old := testsRunner
	t.Cleanup(func() { testsRunner = old })
	record := func(t *testing.T, root, claim string, res testsResult, k ed25519.PrivateKey) {
		t.Helper()
		uiWrite(t, root, ".proofbound/keys/rev.pub", pbwarranty.PubLine(key.Public().(ed25519.PublicKey)))
		data, err := testsRecordBytes(claim, "rev", res, k, "2026-11-23T00:00:00Z")
		if err != nil {
			t.Fatal(err)
		}
		uiWrite(t, root, testsRecordPath(claim), string(data))
		uiCommit(t, root, "rev", "tests review")
	}
	good := testsResult{Declared: true, Run: "sh tests/t.sh", Baseline: testsPass, Whole: testsCaught, Hunks: []testsHunk{
		{Path: "src/a.sh", Lines: "line 2", Result: testsCaught},
		{Path: "src/a.sh", Lines: "line 7", Result: testsNotCaught, Decision: "accept", Reason: "cosmetic"},
	}}
	countRuns := func() {
		testsRunner = func(context.Context, string, string, time.Duration) string { ran++; return testsPass }
	}

	t.Run("no record is an exception", func(t *testing.T) {
		root, base, claim := testsRepo(t, true)
		c, ex := control(root, base, claim)
		if c.State != auditExcepted || len(ex) != 1 || ex[0].Kind != "tests" || !strings.Contains(ex[0].Text, "no reviewer recorded") {
			t.Fatalf("%+v %+v", c, ex)
		}
	})
	t.Run("a good record holds, and nothing runs", func(t *testing.T) {
		root, base, claim := testsRepo(t, true)
		record(t, root, claim, good, key)
		countRuns()
		c, ex := control(root, base, claim)
		if c.State != auditHeld || len(ex) != 0 || ran != 0 {
			t.Fatalf("%+v %+v ran=%d", c, ex, ran)
		}
		testsRunner = old
	})
	bad := map[string]func(r *testsResult){
		"a hunk left out":            func(r *testsResult) { r.Hunks = r.Hunks[:1] },
		"not caught and not decided": func(r *testsResult) { r.Hunks[1].Decision = "" },
		"rejected":                   func(r *testsResult) { r.Hunks[1].Decision = "reject" },
		"whole change not caught":    func(r *testsResult) { r.Whole = testsNotCaught },
		"baseline fails":             func(r *testsResult) { r.Baseline = testsFail },
		"another command":            func(r *testsResult) { r.Run = "true" },
	}
	for name, mutate := range bad {
		t.Run(name, func(t *testing.T) {
			root, base, claim := testsRepo(t, true)
			r := good
			r.Hunks = append([]testsHunk{}, good.Hunks...)
			mutate(&r)
			record(t, root, claim, r, key)
			if c, ex := control(root, base, claim); c.State != auditExcepted || len(ex) != 1 {
				t.Fatalf("%+v %+v", c, ex)
			}
		})
	}
	t.Run("an unsigned record does not count", func(t *testing.T) {
		root, base, claim := testsRepo(t, true)
		record(t, root, claim, good, nil)
		if c, _ := control(root, base, claim); c.State != auditExcepted {
			t.Fatalf("%+v", c)
		}
	})
	t.Run("not declared is counted, not flagged", func(t *testing.T) {
		root, base, claim := testsRepo(t, false)
		c, ex := control(root, base, claim)
		if c.State != auditNone || len(ex) != 0 || !strings.Contains(c.Count, "1 not declared") {
			t.Fatalf("%+v %+v", c, ex)
		}
	})
	t.Run("a declaration last changed by an agent is an exception", func(t *testing.T) {
		root := taskRepo(t)
		base := uiCommit(t, root, "owner", "start")
		uiWrite(t, root, "specs/001-a/spec.md", "# a\n")
		uiWrite(t, root, testsConfigPath, testsJSON+"\n")
		setter := uiCommit(t, root, "Claude", "declare tests")
		uiWrite(t, root, "src/a.sh", "echo x\n")
		claim := uiCommit(t, root, "Claude", "x\n\nIntent: sdd:001-a")
		rep := auditReport{Commits: []auditCommit{{Commit: setter, Class: classHousekeeping, Agents: []string{"Claude"}}, {Commit: claim, Class: classClaimed, Agents: []string{"Claude"}}}, Roles: []auditRole{}}
		var ex []auditException
		c, err := testsControl(ctx, root, base, &rep, func(kind, ctl, commit, text string) { ex = append(ex, auditException{Text: text}) })
		if err != nil || c.State != auditExcepted || len(ex) != 1 || !strings.Contains(ex[0].Text, "made by an agent") {
			t.Fatalf("%+v %+v %v", c, ex, err)
		}
	})
}

// testsRepoRun is testsRepo with its own declared command.
func testsRepoRun(t *testing.T, run string) (root, base, claim string) {
	t.Helper()
	root = taskRepo(t)
	uiWrite(t, root, "src/a.sh", "echo one\n")
	base = uiCommit(t, root, "owner", "start")
	uiWrite(t, root, testsConfigPath, strings.Replace(testsJSON, "sh tests/t.sh", run, 1)+"\n")
	uiWrite(t, root, "specs/001-a/spec.md", "# a\n")
	uiCommit(t, root, "owner", "spec")
	uiWrite(t, root, "src/a.sh", "echo ONE\n")
	claim = uiCommit(t, root, "Claude", "change a\n\nIntent: sdd:001-a")
	return root, base, claim
}

// tests-can-fail review F1..F5: each finding's regression.
func TestTestsReviewFindings(t *testing.T) {
	ctx := context.Background()
	t.Run("F1 a test cannot ask git what was undone", func(t *testing.T) {
		root, _, claim := testsRepoRun(t, "git rev-parse --git-dir")
		res, err := commitTests(ctx, root, claim)
		if err != nil || res.Baseline != testsFail {
			t.Fatalf("a test that reads git ran with a repository: %+v %v", res, err)
		}
	})
	t.Run("F1 a test that fails once it has run is unreliable", func(t *testing.T) {
		mark := filepath.Join(t.TempDir(), "ran")
		root, _, claim := testsRepoRun(t, "test ! -e "+mark+" && touch "+mark)
		res, err := commitTests(ctx, root, claim)
		if err != nil || res.Baseline != testsUnreliable || res.Hunks[0].Result != testsCaught {
			t.Fatalf("%+v %v", res, err)
		}
		if _, needs, err := decideTests(res, nil); err != nil || !needs {
			t.Fatalf("an unreliable run did not need work: %v %v", needs, err)
		}
	})
	t.Run("F1 nothing a run leaves behind reaches the next run", func(t *testing.T) {
		root, _, claim := testsRepoRun(t, "test ! -e left && touch left")
		if res, _ := commitTests(ctx, root, claim); res.Baseline != testsPass || res.Hunks[0].Result != testsNotCaught {
			t.Fatalf("a file left by one run was seen by the next: %+v", res)
		}
	})
	t.Run("F2 an added empty file is undone, not skipped", func(t *testing.T) {
		root := taskRepo(t)
		uiWrite(t, root, testsConfigPath, strings.Replace(testsJSON, "sh tests/t.sh", "true", 1)+"\n")
		uiWrite(t, root, "specs/001-a/spec.md", "# a\n")
		uiCommit(t, root, "owner", "spec")
		uiWrite(t, root, "src/empty.txt", "")
		claim := uiCommit(t, root, "Claude", "empty\n\nIntent: sdd:001-a")
		res, err := commitTests(ctx, root, claim)
		if err != nil || len(res.Hunks) != 1 || res.Hunks[0].Lines != "file added" || res.Hunks[0].Result != testsNotCaught {
			t.Fatalf("%+v %v", res, err)
		}
	})
	t.Run("F2 hunks over the limit are not run and need a decision", func(t *testing.T) {
		root := taskRepo(t)
		var lines []string
		for i := 0; i < 50; i++ {
			lines = append(lines, "x")
		}
		uiWrite(t, root, "src/a.sh", strings.Join(lines, "\n")+"\n")
		uiWrite(t, root, testsConfigPath, testsJSON+"\n")
		uiWrite(t, root, "specs/001-a/spec.md", "# a\n")
		uiCommit(t, root, "owner", "spec")
		for i := 0; i < 44; i += 2 {
			lines[i] = "y"
		}
		uiWrite(t, root, "src/a.sh", strings.Join(lines, "\n")+"\n")
		claim := uiCommit(t, root, "Claude", "many\n\nIntent: sdd:001-a")
		runs := 0
		old := testsRunner
		t.Cleanup(func() { testsRunner = old })
		testsRunner = func(_ context.Context, _ string, _ string, _ time.Duration) string {
			runs++
			if runs == 1 || runs == 2+testsMaxHunks+1 {
				return testsPass // the two runs with nothing undone
			}
			return testsFail
		}
		res, err := commitTests(ctx, root, claim)
		if err != nil || len(res.Hunks) != 22 || res.Hunks[20].Result != testsNotRun || res.Hunks[19].Result != testsCaught || runs != testsMaxHunks+3 || res.Baseline != testsPass {
			t.Fatalf("runs=%d %+v %v", runs, res, err)
		}
		if _, _, err := decideTests(res, nil); err == nil || !strings.Contains(err.Error(), res.Hunks[21].Lines) {
			t.Fatalf("hunks not run were not asked for a decision: %v", err)
		}
	})
	t.Run("a hunk that cannot be undone is never caught", func(t *testing.T) {
		for run, want := range map[string]string{testsUndoFailed: testsUndoFailed, testsPass: testsNotCaught, testsFail: testsCaught, testsTimeout: testsCaught} {
			if got := caughtWord(run); got != want {
				t.Errorf("caughtWord(%q) = %q, want %q", run, got, want)
			}
		}
		res := testsResult{Declared: true, Run: "make test", Baseline: testsPass, Whole: testsUndoFailed, Hunks: []testsHunk{{Path: "a", Lines: "line 1", Result: testsUndoFailed}}}
		if _, _, err := decideTests(res, nil); err == nil || !strings.Contains(err.Error(), "line 1") {
			t.Fatalf("a hunk that could not be undone was not asked for a decision: %v", err)
		}
		accept := &testsResult{Hunks: []testsHunk{{Path: "a", Lines: "line 1", Decision: "accept", Reason: "generated"}}}
		if _, needs, err := decideTests(res, accept); err != nil || !needs {
			t.Fatalf("a whole change that could not be undone did not need work: %v %v", needs, err)
		}
		var out bytes.Buffer
		if err := writeTestsLines(&out, res); err != nil || !strings.Contains(out.String(), "tests: could not undo a line 1") || !strings.Contains(out.String(), "tests: could not undo the whole change") {
			t.Fatalf("%q %v", out.String(), err)
		}
	})
	t.Run("F2 decisions", func(t *testing.T) {
		res := testsResult{Declared: true, Baseline: testsPass, Whole: testsCaught, Hunks: []testsHunk{{Path: "a", Lines: "line 1", Result: testsCaught}, {Path: "a", Lines: "line 9", Result: testsUndoFailed}}}
		accept := &testsResult{Hunks: []testsHunk{{Path: "a", Lines: "line 9", Decision: "accept", Reason: "generated"}}}
		if _, needs, err := decideTests(res, accept); err != nil || needs {
			t.Fatalf("accepted: %v %v", needs, err)
		}
		if _, _, err := decideTests(res, &testsResult{Hunks: []testsHunk{{Path: "a", Lines: "line 9", Decision: "accept"}}}); err == nil {
			t.Fatal("accept without a reason")
		}
		whole := res
		whole.Whole = testsNotCaught
		if _, needs, _ := decideTests(whole, accept); !needs {
			t.Fatal("a whole change not caught did not need work")
		}
	})
	t.Run("F3 the bench refuses commits whose project declares tests", func(t *testing.T) {
		root, _, claim := testsRepo(t, true)
		s := testUIServer(t, root, auditReport{})
		body := `{"commit":"` + claim + `","decisions":[]}`
		w := uiRequest(s, "POST", "/act/review-draft", "127.0.0.1:4242", "http://127.0.0.1:4242", s.token, body)
		if w.Code != 422 || !strings.Contains(w.Body.String(), "declares its tests") {
			t.Fatalf("%d %s", w.Code, w.Body.String())
		}
	})
	t.Run("F4 a timeout stops every process the tests started", func(t *testing.T) {
		dir := t.TempDir()
		start := time.Now()
		if got := testsRunner(ctx, dir, "(sleep 1; touch late) & wait", 200*time.Millisecond); got != testsTimeout {
			t.Fatalf("got %s", got)
		}
		time.Sleep(time.Until(start.Add(1500 * time.Millisecond)))
		if _, err := os.Stat(filepath.Join(dir, "late")); err == nil {
			t.Fatal("a process the tests started outlived the timeout")
		}
	})
	t.Run("F5 git's variables do not reach the tests", func(t *testing.T) {
		t.Setenv("GIT_DIR", "/nowhere")
		if got := testsRunner(ctx, t.TempDir(), `test -z "$GIT_DIR"`, 5*time.Second); got != testsPass {
			t.Fatal("GIT_DIR reached the tests")
		}
	})
	t.Run("F5 draft lines", func(t *testing.T) {
		var b bytes.Buffer
		res := testsResult{Declared: true, Run: "x", Baseline: testsPass, Whole: testsNotCaught, Hunks: []testsHunk{{Path: "a", Lines: "line 1", Result: testsCaught}, {Path: "b", Lines: "line 2", Result: testsNotCaught}}}
		if err := writeTestsLines(&b, res); err != nil {
			t.Fatal(err)
		}
		want := "tests caught undoing a line 1\ntests did NOT catch undoing b line 2\ntests did NOT catch undoing the whole change\ndecide each hunk not caught in tests.hunks: \"accept\" with a reason, or \"reject\"\n"
		if b.String() != want {
			t.Fatalf("got\n%s", b.String())
		}
	})
	t.Run("F5 C6 lists the tests declaration", func(t *testing.T) {
		root := taskRepo(t)
		uiCommit(t, root, "owner", "start")
		uiWrite(t, root, testsConfigPath, testsJSON+"\n")
		sha := uiCommit(t, root, "owner", "declare")
		if paths, err := trustPaths(ctx, root, sha); err != nil || len(paths) != 1 || paths[0] != testsConfigPath {
			t.Fatalf("%v %v", paths, err)
		}
	})
}

// F5: C11 refuses a record by the change's own agent, a record for another commit, and an accept
// with no reason.
func TestAuditC11Refusals(t *testing.T) {
	ctx := context.Background()
	key := uiKey(t)
	pub := pbwarranty.PubLine(key.Public().(ed25519.PublicKey))
	agentKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{9}, ed25519.SeedSize))
	agentPub := pbwarranty.PubLine(agentKey.Public().(ed25519.PublicKey))
	ok := testsResult{Declared: true, Run: "sh tests/t.sh", Baseline: testsPass, Whole: testsCaught, Hunks: []testsHunk{
		{Path: "src/a.sh", Lines: "line 2", Result: testsCaught},
		{Path: "src/a.sh", Lines: "line 7", Result: testsNotCaught, Decision: "accept", Reason: "cosmetic"},
	}}
	cases := map[string]func(root, claim string) (by, commit string, res testsResult){
		"by the agent itself": func(root, claim string) (string, string, testsResult) { return "claude", claim, ok },
		"for another commit":  func(root, claim string) (string, string, testsResult) { return "rev", strings.Repeat("d", 40), ok },
		"accept without reason": func(root, claim string) (string, string, testsResult) {
			r := ok
			r.Hunks = append([]testsHunk{}, ok.Hunks...)
			r.Hunks[1].Reason = ""
			return "rev", claim, r
		},
	}
	want := map[string]string{"by the agent itself": "not independent", "for another commit": "judges another commit", "accept without reason": "not caught"}
	for name, mk := range cases {
		t.Run(name, func(t *testing.T) {
			root, base, claim := testsRepoWith(t, true, map[string]string{".proofbound/keys/claude.pub": agentPub, ".proofbound/keys/rev.pub": pub})
			by, commit, res := mk(root, claim)
			k := key
			if by == "claude" {
				k = agentKey
			}
			data, err := testsRecordBytes(commit, by, res, k, "2026-11-23T00:00:00Z")
			if err != nil {
				t.Fatal(err)
			}
			uiWrite(t, root, testsRecordPath(claim), string(data))
			uiCommit(t, root, by, "tests review")
			rep := auditReport{Commits: []auditCommit{{Commit: claim, Subject: "change a", Class: classClaimed, Author: "Claude", Agents: []string{"Claude"}}}, Roles: []auditRole{}}
			var ex []auditException
			c, err := testsControl(ctx, root, base, &rep, func(kind, ctl, commit, text string) { ex = append(ex, auditException{Text: text}) })
			if err != nil || c.State != auditExcepted || len(ex) != 1 || !strings.Contains(ex[0].Text, want[name]) {
				t.Fatalf("%+v %+v %v", c, ex, err)
			}
		})
	}
}
