package cli

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// waitJob waits for the bench's job to finish.
func waitJob(t *testing.T, s *uiServer, commit string) benchJob {
	t.Helper()
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		if job, ok := s.jobs.snapshot(commit); ok && !job.Running {
			return job
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("the job did not finish")
	return benchJob{}
}

// ui-reviewer-spec R1 §1.1 (review F1, F2, F5, F6, F10): the tests run in the bench, their result is
// kept through every save, a part they did not catch must be decided, and a seal runs them again
// before sealing, with no lock held meanwhile.
func TestBenchTestsFlow(t *testing.T) {
	key := uiKey(t)
	root, base, claim := testsRepoWith(t, true, map[string]string{".proofbound/keys/demo-reviewer.pub": pbwarranty.PubLine(key.Public().(ed25519.PublicKey))})
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
	var sealedRan *sealTestsRun
	var sealedDraft draftFile
	s.seal = func(ctx context.Context, dir string, ran *sealTestsRun, out io.Writer) error {
		sealedRan = ran
		raw, err := os.ReadFile(filepath.Join(dir, "draft.json"))
		if err != nil {
			return err
		}
		if err := json.Unmarshal(raw, &sealedDraft); err != nil {
			return err
		}
		uiWrite(t, root, "docs/verification/verdicts/OV-x.md", "sealed\n")
		_, _ = io.WriteString(out, "wrote docs/verification/verdicts/OV-x.md\n")
		return nil
	}
	act := func(path string, decisions []reviewDecision) (int, map[string]any) {
		raw, _ := json.Marshal(reviewRequest{Commit: claim, Decisions: decisions})
		w := uiRequest(s, "POST", path, "127.0.0.1:4242", "http://127.0.0.1:4242", s.token, string(raw))
		var out map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return w.Code, out
	}
	promises := []reviewDecision{{Part: "verdict", Group: 0, Index: 0, ObligationID: "O-1", Outcome: "SATISFIED"}, {Part: "review", Group: 0, Index: 0, ObligationID: "O-1", Outcome: "VERIFIABLE"}}
	missing := func(out map[string]any) string {
		b, _ := json.Marshal(out["missing"])
		return string(b)
	}

	// Before a run, nothing can be previewed; a GET or an agent's bearer request starts nothing (F5).
	if code, out := act("/act/review-preview", promises); code != 422 || !strings.Contains(missing(out), "run the tests first") {
		t.Fatalf("previewed with no test run: %d %v", code, out)
	}
	if w := uiRequest(s, "GET", "/act/review-tests", "127.0.0.1:4242", "", s.token, ""); w.Code != 404 {
		t.Errorf("GET /act/review-tests answered %d", w.Code)
	}
	if w := uiRequestWith(s, "POST", "/act/review-tests", "Bearer "+s.agentToken); w.Code != 401 {
		t.Errorf("an agent's POST answered %d", w.Code)
	}
	if _, ok := s.jobs.snapshot(""); ok {
		t.Fatal("a refused request started a job")
	}

	// Run: progress lines, then the result is written into the local draft.
	if code, out := act("/act/review-tests", promises); code != 202 {
		t.Fatalf("run: %d %v", code, out)
	}
	job := waitJob(t, s, claim)
	log := strings.Join(job.Lines, "\n")
	if job.Error != "" || !strings.Contains(log, "this runs the commit's own code") || !strings.Contains(log, "tests caught undoing src/a.sh line 2") || !strings.Contains(log, "tests did NOT catch undoing src/a.sh line 7") {
		t.Fatalf("job %+v", job)
	}
	local, has := localDraft(root, claim)
	if !has || local.Tests == nil || len(local.Tests.Hunks) != 2 || local.Verdicts[0].Obligations[0].Outcome != "SATISFIED" {
		t.Fatalf("the run was not kept with the decisions: %+v", local)
	}

	// A reload shows the saved run without running again.
	if w := uiRequest(s, "GET", "/review?commit="+claim, "127.0.0.1:4242", "", s.token, ""); w.Code != 200 || !strings.Contains(w.Body.String(), "Tests did not notice</strong> <code>src/a.sh</code> line 7") {
		t.Fatalf("the page does not show the saved run: %d", w.Code)
	}
	if again, _ := s.jobs.snapshot(claim); len(again.Lines) != len(job.Lines) || again.Running {
		t.Error("loading the page ran the tests again")
	}
	// Save keeps the run (F1); the part not caught must be decided, with a reason to accept.
	if code, out := act("/act/review-draft", promises); code != 200 {
		t.Fatalf("save: %d %v", code, out)
	}
	if local, _ := localDraft(root, claim); local.Tests == nil || len(local.Tests.Hunks) != 2 {
		t.Fatal("Save after Run dropped the tests")
	}
	if code, out := act("/act/review-preview", promises); code != 422 || !strings.Contains(missing(out), "accept or reject src/a.sh line 7") || strings.Contains(missing(out), "line 2") {
		t.Fatalf("an undecided part previewed: %d %v", code, out)
	}
	accept := append(append([]reviewDecision{}, promises...), reviewDecision{Part: "tests", Index: 1, Path: "src/a.sh", Lines: "line 7", Outcome: "accept"})
	if code, out := act("/act/review-preview", accept); code != 422 || !strings.Contains(missing(out), "why accept") {
		t.Fatalf("accepted with no reason: %d %v", code, out)
	}
	// Line 7 decided too, so the decision on the caught line 2 is the only reason left to refuse.
	if code, out := act("/act/review-preview", append(append([]reviewDecision{}, promises...),
		reviewDecision{Part: "tests", Index: 0, Path: "src/a.sh", Lines: "line 2", Outcome: "accept", Finding: "x"},
		reviewDecision{Part: "tests", Index: 1, Path: "src/a.sh", Lines: "line 7", Outcome: "reject"})); code != 422 || !strings.Contains(out["error"].(string), "caught needs no decision") {
		t.Errorf("a decision on a part the tests caught was taken: %d %v", code, out)
	}
	reject := append(append([]reviewDecision{}, promises...), reviewDecision{Part: "tests", Index: 1, Path: "src/a.sh", Lines: "line 7", Outcome: "reject"})
	code, out := act("/act/review-preview", reject)
	files, _ := json.Marshal(out["files"])
	warnings, _ := json.Marshal(out["warnings"])
	if code != 200 || !strings.Contains(string(files), "TESTS-"+claim[:12]) || !strings.Contains(string(files), "PROV-"+claim[:12]) || !strings.Contains(string(warnings), "NEEDS_WORK") {
		t.Fatalf("preview (F10): %d %v", code, out)
	}

	// One job at a time.
	hold := make(chan struct{})
	if err := s.jobs.start(s.life, "tests", claim, func(ctx context.Context, say func(string)) (map[string]any, error) {
		<-hold
		return nil, nil
	}); err != nil {
		t.Fatal(err)
	}
	if code, _ := act("/act/review-tests", promises); code != 409 {
		t.Errorf("a second job started: %d", code)
	}
	if w := uiRequest(s, "GET", "/review?commit="+claim, "127.0.0.1:4242", "", s.token, ""); w.Code != 200 {
		t.Errorf("a page did not load while a job ran: %d", w.Code)
	}
	close(hold)
	waitJob(t, s, claim)

	// Seal: the tests run again (with progress), then the seal gets that run; the decisions sealed
	// are the ones sent with the click; the review is committed.
	if code, out := act("/act/review-seal", reject); code != 202 {
		t.Fatalf("seal: %d %v", code, out)
	}
	act("/act/review-draft", promises) // a save while the seal runs does not change what is sealed
	job = waitJob(t, s, claim)
	if seal := strings.Join(job.Lines, "\n"); job.Error != "" || job.Kind != "seal" || !strings.Contains(seal, "running them again before sealing") || !strings.Contains(seal, "tests did NOT catch undoing src/a.sh line 7") {
		t.Fatalf("seal job %+v", job)
	}
	if sealedRan == nil || sealedRan.Commit != claim || len(sealedRan.Result.Hunks) != 2 {
		t.Fatalf("the seal did not get a fresh run: %+v", sealedRan)
	}
	if sealedDraft.Tests == nil || sealedDraft.Tests.Hunks[1].Decision != "reject" {
		t.Fatalf("the sealed decisions are not the click's: %+v", sealedDraft.Tests)
	}
	if got := uiGit(t, root, "log", "-1", "--format=%s"); got != "review "+claim[:12] {
		t.Errorf("the review was not committed: %q", got)
	}

	// A reviewed commit is no longer waiting: its code is not run and its draft is not touched (F2).
	s.compute = func(ctx context.Context, b string) (ledgerView, string, error) {
		return ledgerView{Report: auditReport{Schema: auditSchema, Base: base, Head: claim, Commits: []auditCommit{{Commit: claim, Author: "Claude", Class: classClaimed}},
			Controls: []auditControl{{ID: "C1", Evidence: []auditEvidence{{Commit: claim, Text: "BLOCKED: the reviewer recorded NEEDS_WORK (OV-x)"}}}}, Exceptions: []auditException{}}}, "", nil
	}
	before, _ := s.jobs.snapshot(claim)
	for _, path := range []string{"/act/review-tests", "/act/review-seal", "/act/review-draft"} {
		if code, out := act(path, reject); code != 422 || !strings.Contains(out["error"].(string), "not waiting for review") {
			t.Errorf("%s on a reviewed commit: %d %v", path, code, out)
		}
	}
	if after, _ := s.jobs.snapshot(claim); after.Kind != before.Kind || len(after.Lines) != len(before.Lines) {
		t.Error("a reviewed commit's tests ran")
	}
	w := uiRequest(s, "GET", "/review", "127.0.0.1:4242", "", s.token, "")
	if !strings.Contains(w.Body.String(), "Reviewed: needs work") || !strings.Contains(w.Body.String(), "the reviewer recorded NEEDS_WORK (OV-x)") {
		t.Error("the needs-work list does not show the reviewed commit")
	}
}

// §1.1 (code review F1): an invalid tests file shows its reason, and nothing can be run.
func TestBenchInvalidTestsFile(t *testing.T) {
	key := uiKey(t)
	root := taskRepo(t)
	uiWrite(t, root, ".proofbound/keys/demo-reviewer.pub", pbwarranty.PubLine(key.Public().(ed25519.PublicKey)))
	uiWrite(t, root, testsConfigPath, `{"schema":"proofbound.tests.v1","run":"","paths":[]}`+"\n")
	uiWrite(t, root, "specs/001-a/spec.md", "# a\n")
	base := uiCommit(t, root, "owner", "spec")
	uiWrite(t, root, "src/a.sh", "x\n")
	claim := uiCommit(t, root, "Claude", "a\n\nIntent: sdd:001-a")
	s := newUIServer(root, "", base, 4242)
	t.Cleanup(func() { s.stopLife(); s.jobs.wg.Wait() })
	s.compute = func(ctx context.Context, b string) (ledgerView, string, error) {
		return ledgerView{Report: auditReport{Schema: auditSchema, Base: base, Head: claim, Commits: []auditCommit{{Commit: claim, Author: "Claude", Class: classClaimed}},
			Controls: []auditControl{{ID: "C1", Evidence: []auditEvidence{{Commit: claim, Text: "NO REVIEW: x"}}}}, Exceptions: []auditException{}}}, "", nil
	}
	s.drafts = func(ctx context.Context, shas []string) (map[string]draftFile, error) {
		return map[string]draftFile{claim: {Commit: claim, Verdicts: []draftVerdict{{Obligations: []draftVerdictObligation{{ObligationID: "O-1"}}}}}}, nil
	}
	w := uiRequest(s, "GET", "/review?commit="+claim, "127.0.0.1:4242", "", s.token, "")
	if !strings.Contains(w.Body.String(), "tests file is invalid") || strings.Contains(w.Body.String(), `data-act="review-tests"`) {
		t.Fatalf("the page offers Run for an invalid tests file:\n%s", w.Body.String())
	}
	w = uiRequest(s, "POST", "/act/review-tests", "127.0.0.1:4242", "http://127.0.0.1:4242", s.token, `{"commit":"`+claim+`","decisions":[]}`)
	if w.Code != 422 || !strings.Contains(w.Body.String(), "invalid") {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	if _, ok := s.jobs.snapshot(""); ok {
		t.Error("a job started")
	}
}

// §1.1: the real seal refuses, writing nothing, when the second run finds a part no decision covers.
func TestSealRefusesAPartOnlyTheSecondRunFinds(t *testing.T) {
	root, _, claim := testsRepo(t, true)
	d, _ := areaTruth(claim)(context.Background(), claim)
	d.DeclaredReviewer = "demo-reviewer"
	d.Tests = &testsResult{Declared: true, Baseline: testsPass, Whole: testsCaught, Hunks: []testsHunk{{Path: "src/a.sh", Lines: "line 2", Result: testsCaught}}}
	dir := writeDraft(t, root, d)
	ran := &sealTestsRun{Commit: claim, Result: testsResult{Declared: true, Run: "sh tests/t.sh", Baseline: testsPass, Whole: testsCaught,
		Hunks: []testsHunk{{Path: "src/a.sh", Lines: "line 2", Result: testsCaught}, {Path: "src/a.sh", Lines: "line 7", Result: testsNotCaught}}}}
	err := reviewSealWith(context.Background(), root, dir, ran, areaTruth(claim), &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "not decided") || !strings.Contains(err.Error(), "line 7") {
		t.Fatalf("%v", err)
	}
	if st := uiGit(t, root, "status", "--porcelain", "--", "docs", "gates"); st != "" {
		t.Errorf("wrote %s", st)
	}
}

// §1.3: a flagged line past the 1,500-line cut is listed, not lost.
func TestBenchFlagsPastTheCut(t *testing.T) {
	root := taskRepo(t)
	uiCommit(t, root, "owner", "start")
	var b strings.Builder
	for i := 1; i <= 1600; i++ {
		b.WriteString("line\n")
	}
	uiWrite(t, root, "src/big.txt", b.String())
	sha := uiCommit(t, root, "Claude", "big")
	f := lineFlags{}
	f.add("src/big.txt", "line 1590", flagArea)
	f.add("src/elsewhere.bin", "file added", flagArea)
	_, past := benchDiff(context.Background(), root, sha, f)
	got := strings.Join(past, "; ")
	if !strings.Contains(got, "src/big.txt line 1590: outside the area") || !strings.Contains(got, "src/elsewhere.bin: flagged, not shown") {
		t.Fatalf("past the cut: %v", past)
	}
}

// §1.1: the request lock is free while a job's tests run, so pages still load (F2); stopping the
// server stops the job (F6).
func TestBenchJobLifecycle(t *testing.T) {
	s := newUIServer(t.TempDir(), "", "", 4242)
	started := make(chan struct{})
	if err := s.jobs.start(s.life, "tests", "c", func(ctx context.Context, say func(string)) (map[string]any, error) {
		close(started)
		<-ctx.Done()
		return nil, ctx.Err()
	}); err != nil {
		t.Fatal(err)
	}
	<-started
	if !s.acquire(context.Background()) {
		t.Fatal("the request lock is held while the tests run")
	}
	s.release()
	s.stopLife()
	done := make(chan struct{})
	go func() { s.jobs.wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("stopping the server did not stop the job")
	}
	if job, _ := s.jobs.snapshot("c"); job.Running || job.Error == "" {
		t.Errorf("job %+v", job)
	}
}

// §1.1 (F7): the commit's code does not see the reviewer's key path or the record.
func TestTestsEnvDropsSecrets(t *testing.T) {
	t.Setenv("PROOFBOUND_KEY", "/home/reviewer/key")
	t.Setenv("DATABASE_URL", "postgres://secret")
	for _, kv := range testsEnv(t.TempDir()) {
		if strings.HasPrefix(kv, "PROOFBOUND_") || strings.HasPrefix(kv, "DATABASE_URL=") {
			t.Errorf("the tests see %s", kv)
		}
	}
	root, _, claim := testsRepoRun(t, `test -z $PROOFBOUND_KEY && test -z $DATABASE_URL`)
	if res, err := commitTests(context.Background(), root, claim); err != nil || res.Baseline != testsPass {
		t.Fatalf("%+v %v", res, err)
	}
}

// §1.1 (F2): a seal given a run of another commit writes nothing.
func TestSealRefusesAnotherCommitsRun(t *testing.T) {
	root, base, claim := testsRepo(t, true)
	d, _ := areaTruth(claim)(context.Background(), claim)
	d.DeclaredReviewer = "demo-reviewer"
	dir := writeDraft(t, root, d)
	var out bytes.Buffer
	err := reviewSealWith(context.Background(), root, dir, &sealTestsRun{Commit: base}, areaTruth(claim), &out)
	if err == nil || !strings.Contains(err.Error(), "another commit") {
		t.Fatalf("%v", err)
	}
	if st := uiGit(t, root, "status", "--porcelain", "--", "docs", "gates"); st != "" {
		t.Errorf("wrote %s", st)
	}
}

// §1.3: the flags sit on the added lines they name, and nowhere else.
func TestBenchFlags(t *testing.T) {
	root, _, claim := testsRepo(t, true)
	d := draftFile{
		OutOfArea:  []areaDecision{{Path: "src/a.sh", Lines: "line 7"}},
		Tests:      &testsResult{Declared: true, Hunks: []testsHunk{{Path: "src/a.sh", Lines: "line 2", Result: testsCaught}, {Path: "src/a.sh", Lines: "line 7", Result: testsNotCaught}}},
		Provenance: &provResult{Files: []provFile{{Path: "src/a.sh", Unrecorded: []int{7}}}},
	}
	lines, past := benchDiff(context.Background(), root, claim, benchFlags(d))
	found := map[string][]string{}
	for _, l := range lines {
		if l.Kind == "add" {
			found[l.Text] = l.Flags
		}
	}
	if got := strings.Join(found["+echo SIX"], ","); got != flagArea+","+flagTests+","+flagProv {
		t.Errorf("the line outside the area, untested and unrecorded carries %q", got)
	}
	if got := found["+echo ONE"]; len(got) != 0 {
		t.Errorf("a caught line carries %v", got)
	}
	if len(past) != 0 {
		t.Errorf("past the cut: %v", past)
	}
	for label, want := range map[string][3]int{"line 3": {3, 3, 0}, "lines 1-39": {1, 39, 0}, "file added": {0, 0, 1}} {
		lo, hi, whole, ok := lineSpan(label)
		if !ok || lo != want[0] || hi != want[1] || whole != (want[2] == 1) {
			t.Errorf("%s: %d %d %v %v", label, lo, hi, whole, ok)
		}
	}
	if _, _, _, ok := lineSpan("removed lines 3-4"); ok {
		t.Error("a removed range was read as added lines")
	}
}

// §1.4: the spec comes from the commit's parent, not the worktree.
func TestBenchSpec(t *testing.T) {
	root := taskRepo(t)
	uiWrite(t, root, "specs/001-a/spec.md", "# the promise\n")
	uiWrite(t, root, "specs/001-a/scope.paths", "src/\n")
	uiCommit(t, root, "owner", "spec")
	uiWrite(t, root, "src/a.sh", "x\n")
	claim := uiCommit(t, root, "Claude", "a\n\nIntent: sdd:001-a")
	uiWrite(t, root, "specs/001-a/spec.md", "# changed in the worktree\n")
	area, err := commitArea(context.Background(), root, claim)
	if err != nil {
		t.Fatal(err)
	}
	got := benchSpec(context.Background(), root, claim, area)
	// The parent's text: what the area check enforces (code review F4, spec §1.4 amended).
	if len(got) != 2 || got[0].Path != "specs/001-a/scope.paths" && got[1].Path != "specs/001-a/scope.paths" {
		t.Fatalf("%+v", got)
	}
	for _, f := range got {
		if f.Path == "specs/001-a/spec.md" && f.Text != "# the promise\n" {
			t.Errorf("spec text %q", f.Text)
		}
	}
}
