package cli

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"os"
	"path/filepath"
	"strings"
	"testing"

	connectorreviews "github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// sealedRepo: provRepo, reviewed and sealed by an independent declared reviewer with the given
// obligation outcome, and committed; returns the repository, the claim and HEAD.
func sealedRepo(t *testing.T, outcome string) (root, claim, head string) {
	t.Helper()
	keyPath := filepath.Join(t.TempDir(), "reviewer-key")
	t.Setenv("PROOFBOUND_KEY", keyPath)
	pub, err := pbwarranty.InitKey(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	root, claim = provRepo(t)
	uiWrite(t, root, ".proofbound/keys/proofbound-reviewer.pub", pbwarranty.PubLine(pub))
	uiCommit(t, root, "owner", "declare the reviewer")
	d := sealDraft(outcome, "VERIFIABLE")
	d.Commit = claim
	var out bytes.Buffer
	if err := reviewSeal(context.Background(), root, writeDraft(t, root, d), areaTruth(claim), &out); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	head = uiCommit(t, root, "proofbound-reviewer", "review")
	return root, claim, head
}

// CLI-INV-73: one row per agent (never a person), review outcomes from the verdicts committed at the
// range's head with C8's independence, and the area, tests and line counts only from what the
// controls trusted; sessions from the trusted provenance records.
func TestTrackRecord(t *testing.T) {
	ctx := context.Background()
	root, claim, head := sealedRepo(t, "SATISFIED")
	other := strings.Repeat("e", 40)
	person := strings.Repeat("f", 40)
	rep := auditReport{Base: strings.Repeat("0", 40), Head: head,
		Commits: []auditCommit{
			{Commit: claim, Class: classClaimed, Author: "owner", Agents: []string{"Claude"}},
			{Commit: other, Class: classUncovered, Author: "Claude", Agents: []string{"Claude", "Copilot"}},
			{Commit: person, Class: classClaimed, Author: "alice"},
		},
		Roles:      []auditRole{{Commit: claim, Reviewers: []string{"proofbound-reviewer"}, Independent: true}},
		Exceptions: []auditException{{Commit: other}, {Commit: other, Outcome: &outcomeView{Verified: true}}},
	}
	tc := rep.trackFor(claim)
	tc.AreaHunks, tc.AreaRejected, tc.AreaAccepted = 3, 1, 1
	tc.TestsNotCaught, tc.TestsRejected, tc.TestsAccepted, tc.WholeNotCaught = 2, 1, 1, true
	tc.Sessions["s1"], tc.Sessions["s2"] = 3, 2
	tc.LinesTraced, tc.LinesUnrecorded = 4, 1
	tr := computeTrack(ctx, root, rep)
	if len(tr.Agents) != 2 || tr.Agents[0].Name != "Claude" || tr.Agents[1].Name != "Copilot" {
		t.Fatalf("rows: %+v", tr.Agents)
	}
	c := tr.Agents[0]
	want := trackAgent{Name: "Claude", Commits: 2, Claimed: 1, Uncovered: 1, Acceptable: 1, AreaHunks: 3, AreaRejected: 1, AreaAccepted: 1, TestsNotCaught: 2, TestsRejected: 1, TestsAccepted: 1, WholeNotCaught: 1, LinesBySession: 4, LinesUnrecorded: 1, Open: 1}
	if c != want {
		t.Errorf("Claude:\n got %+v\nwant %+v", c, want)
	}
	if cp := tr.Agents[1]; cp.Commits != 1 || cp.Uncovered != 1 || cp.Open != 1 || cp.Claimed != 0 {
		t.Errorf("Copilot (a second agent on one commit counts it too): %+v", cp)
	}
	if len(tr.Sessions) != 2 || tr.Sessions[0].Session != "s1" || tr.Sessions[0].Lines != 3 || tr.Sessions[0].Acceptable != 1 || len(tr.Sessions[1].Commits) != 1 {
		t.Errorf("sessions: %+v", tr.Sessions)
	}

	// A review that is not independent is its own column, never ACCEPTABLE.
	rep.Roles[0].Independent = false
	if a := computeTrack(ctx, root, rep).Agents[0]; a.Acceptable != 0 || a.NotIndependent != 1 {
		t.Errorf("not independent: %+v", a)
	}
	rep.Roles[0].Independent = true

	// A verdict the ledger has not seen yet (no role) is still judged by its own declared reviewer:
	// one by the commit's author is not independent (code review M2).
	unsynced := rep
	unsynced.Roles = []auditRole{}
	unsynced.Commits = []auditCommit{{Commit: claim, Class: classClaimed, Author: "proofbound-reviewer", Agents: []string{"Claude"}}}
	if a := computeTrack(ctx, root, unsynced).Agents[0]; a.Acceptable != 0 || a.NotIndependent != 1 {
		t.Errorf("a self-review the ledger has not seen: %+v", a)
	}
	if s := computeTrack(ctx, root, unsynced).Sessions; len(s) != 2 || s[0].NotIndependent != 1 || s[0].Acceptable != 0 {
		t.Errorf("sessions of a self-reviewed commit: %+v", s)
	}

	// Only the verdicts at the range's head count: at the claim itself nothing is reviewed.
	rep.Head = claim
	if a := computeTrack(ctx, root, rep).Agents[0]; a.Acceptable != 0 || a.Unreviewed != 1 {
		t.Errorf("verdicts after the head counted: %+v", a)
	}

	// A verdict file named for one commit but judging another is not counted for it.
	parent := uiGit(t, root, "rev-parse", claim+"^1")
	data, err := os.ReadFile(filepath.Join(root, "docs/verification/verdicts/OV-"+claim[:12]+"-1.md"))
	if err != nil {
		t.Fatal(err)
	}
	uiWrite(t, root, "docs/verification/verdicts/OV-"+parent[:12]+"-1.md", string(data))
	h2 := uiCommit(t, root, "owner", "misplaced verdict")
	verdicts := strings.Split(strings.TrimSpace(gitOutput(ctx, root, "ls-tree", "-r", "--name-only", h2, "--", "docs/verification/verdicts/")), "\n")
	if s, _ := reviewStatus(ctx, root, h2, parent, verdicts); s != "" {
		t.Errorf("a verdict for another commit counted: %q", s)
	}
	if s, _ := reviewStatus(ctx, root, h2, claim, verdicts); s != "ACCEPTABLE" {
		t.Errorf("claim: %q", s)
	}
}

// CLI-INV-73 (NEEDS_WORK): any NEEDS_WORK verdict makes the commit NEEDS_WORK.
func TestTrackRecordNeedsWork(t *testing.T) {
	ctx := context.Background()
	root, claim, head := sealedRepo(t, "NOT_SATISFIED")
	verdicts := strings.Split(strings.TrimSpace(gitOutput(ctx, root, "ls-tree", "-r", "--name-only", head, "--", "docs/verification/verdicts/")), "\n")
	if s, _ := reviewStatus(ctx, root, head, claim, verdicts); s != "NEEDS_WORK" {
		t.Fatalf("%q", s)
	}
}

// CLI-INV-73 (trust): the controls hand over counts only from records they trust; a record signed by
// the commit's own agent verifies but counts for nothing.
func TestTrackRecordTrustsOnlyWhatControlsTrust(t *testing.T) {
	ctx := context.Background()
	key := uiKey(t)
	for _, tc := range []struct {
		by      string
		trusted bool
	}{{"rev", true}, {"claude", false}} {
		root, claim := provRepo(t)
		uiWrite(t, root, ".proofbound/keys/"+tc.by+".pub", pbwarranty.PubLine(key.Public().(ed25519.PublicKey)))
		res, err := commitProvenance(ctx, root, claim)
		if err != nil {
			t.Fatal(err)
		}
		data, err := provRecordBytes(claim, tc.by, res, key, "2026-11-24T00:00:00Z")
		if err != nil {
			t.Fatal(err)
		}
		uiWrite(t, root, provRecordPath(claim), string(data))
		uiCommit(t, root, tc.by, "provenance review")
		rep := auditReport{Commits: []auditCommit{{Commit: claim, Subject: "x", Class: classClaimed, Author: "Claude", Agents: []string{"Claude"}}}, Roles: []auditRole{}}
		if _, err := provControl(ctx, root, &rep, func(string, string, string, string) {}); err != nil {
			t.Fatal(err)
		}
		got := rep.track[claim]
		switch {
		case tc.trusted && (got == nil || !got.Traced || got.LinesTraced != 4 || got.LinesUnrecorded != 1 || got.Sessions["sess-aaaaaaaa1"] != 3 || got.Sessions["sess-bbbbbbbb2"] != 2):
			t.Errorf("trusted record: %+v", got)
		case !tc.trusted && got != nil:
			t.Errorf("a record signed by the agent was counted: %+v", got)
		}
	}
}

// CLI-INV-74 (text): the caveats come first, then one row per agent and the provenance sessions.
func TestWriteTrack(t *testing.T) {
	tr := trackReport{Base: strings.Repeat("a", 40), Head: strings.Repeat("b", 40),
		Agents:   []trackAgent{{Name: "Claude", Commits: 1, Claimed: 1, NeedsWork: 1, AreaHunks: 1, AreaRejected: 1, LinesUnrecorded: 1}},
		Sessions: []trackSession{{Session: "s1", Commits: []string{"c"}, Lines: 4, NeedsWork: 1}}}
	var out bytes.Buffer
	if err := writeTrack(&out, tr); err != nil {
		t.Fatal(err)
	}
	s := out.String()
	if len(trackNotes) != 4 || !strings.Contains(strings.Join(trackNotes, " "), "NEEDS_WORK already includes") || !strings.Contains(strings.Join(trackNotes, " "), "do not add up to the range") || !strings.Contains(strings.Join(trackNotes, " "), "Counts, not a score") || !strings.Contains(strings.Join(trackNotes, " "), "trusted") {
		t.Errorf("the caveats changed: %q", trackNotes)
	}
	for _, want := range append([]string{"Review outcomes by agent, aaaaaaaaaaaa..bbbbbbbbbbbb", "Claude (agent)\n  commits 1: 1 claimed a spec", "hunks outside the spec's area: 1 (1 rejected, 0 accepted)", "review of the claimed: 0 ACCEPTABLE, 1 NEEDS_WORK", "Sessions in the provenance records", "s1"}, trackNotes...) {
		if !strings.Contains(s, want) {
			t.Errorf("lacks %q:\n%s", want, s)
		}
	}
	out.Reset()
	if err := writeTrack(&out, trackReport{Base: tr.Base, Head: tr.Head}); err != nil || !strings.Contains(out.String(), "No commit in this range names a coding agent.") {
		t.Fatalf("%v %s", err, out.String())
	}
}

// CLI-INV-73 (trust, C10 and C11): the area and tests counts come from a record only when the control
// trusts it; the same record signed by the change's author counts for nothing.
func TestTrackRecordAreaAndTestsTrust(t *testing.T) {
	ctx := context.Background()
	key := uiKey(t)
	pub := key.Public().(ed25519.PublicKey)
	for _, author := range []string{"agent", "proofbound-reviewer"} {
		root, base, claim := areaRepo(t, pub)
		ds := []areaDecision{{Path: "src/gone.go", Lines: "file deleted", Decision: "accept", Reason: "dead code"}, {Path: "src/other.go", Lines: "line 2", Decision: "reject"}}
		data, err := areaRecordBytes(claim, "proofbound-reviewer", ds, key, "2026-11-24T00:00:00Z")
		if err != nil {
			t.Fatal(err)
		}
		uiWrite(t, root, areaRecordPath(claim), string(data))
		uiCommit(t, root, "proofbound-reviewer", "area review")
		rep := auditReport{Commits: []auditCommit{{Commit: claim, Class: classClaimed, Author: author, Subject: "cart"}}, Roles: []auditRole{}}
		if _, err := areaControl(ctx, root, base, &rep, func(string, string, string, string) {}); err != nil {
			t.Fatal(err)
		}
		got := rep.track[claim]
		trusted := author == "agent"
		if got == nil || got.AreaHunks != 3 || (trusted && (got.AreaAccepted != 1 || got.AreaRejected != 1)) || (!trusted && (got.AreaAccepted != 0 || got.AreaRejected != 0)) {
			t.Errorf("author %s: %+v", author, got)
		}
	}
	for _, by := range []string{"rev", "claude"} {
		root, base, claim := testsRepo(t, true)
		uiWrite(t, root, ".proofbound/keys/"+by+".pub", pbwarranty.PubLine(pub))
		res := testsResult{Declared: true, Run: "sh tests/t.sh", Baseline: testsPass, Whole: testsNotCaught, Hunks: []testsHunk{
			{Path: "src/a.sh", Lines: "line 2", Result: testsCaught},
			{Path: "src/a.sh", Lines: "line 7", Result: testsNotCaught, Decision: "accept", Reason: "cosmetic"},
		}}
		data, err := testsRecordBytes(claim, by, res, key, "2026-11-24T00:00:00Z")
		if err != nil {
			t.Fatal(err)
		}
		uiWrite(t, root, testsRecordPath(claim), string(data))
		uiCommit(t, root, by, "tests review")
		rep := auditReport{Commits: []auditCommit{{Commit: claim, Subject: "change a", Class: classClaimed, Author: "Claude", Agents: []string{"Claude"}}}, Roles: []auditRole{}}
		if _, err := testsControl(ctx, root, base, &rep, func(string, string, string, string) {}); err != nil {
			t.Fatal(err)
		}
		got := rep.track[claim]
		if by == "rev" && (got == nil || got.TestsNotCaught != 1 || got.TestsAccepted != 1 || got.TestsRejected != 0 || !got.WholeNotCaught) {
			t.Errorf("trusted tests record: %+v", got)
		}
		if by == "claude" && got != nil {
			t.Errorf("a tests record signed by the agent was counted: %+v", got)
		}
	}
}

// CLI-INV-73 (NEEDS_WORK wins): a commit with several verdicts is NEEDS_WORK when any of them is,
// whatever their order.
func TestTrackRecordAnyNeedsWork(t *testing.T) {
	ctx := context.Background()
	keyPath := filepath.Join(t.TempDir(), "reviewer-key")
	t.Setenv("PROOFBOUND_KEY", keyPath)
	pub, err := pbwarranty.InitKey(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	root, claim := provRepo(t)
	uiWrite(t, root, ".proofbound/keys/proofbound-reviewer.pub", pbwarranty.PubLine(pub))
	uiCommit(t, root, "owner", "declare the reviewer")
	two := func(first, second string) draftFile {
		d := sealDraft(first, "VERIFIABLE")
		d.Commit = claim
		v := d.Verdicts[0]
		v.ChangeIntent.RecordID = "CI-add-auth-zzzzzz"
		v.Obligations = append([]draftVerdictObligation{}, v.Obligations...)
		for i := range v.Obligations {
			v.Obligations[i].Outcome = second
		}
		d.Verdicts = append(d.Verdicts, v)
		for i := range d.Verdicts {
			for j := range d.Verdicts[i].Obligations {
				d.Verdicts[i].Obligations[j].EvidenceRefs = []string{connectorreviews.CommitEvidenceRef(claim)}
			}
		}
		return d
	}
	truth := func(context.Context, string) (draftFile, error) {
		d := two("", "")
		d.DeclaredReviewer = ""
		return d, nil
	}
	var out bytes.Buffer
	if err := reviewSeal(ctx, root, writeDraft(t, root, two("NOT_SATISFIED", "SATISFIED")), truth, &out); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	head := uiCommit(t, root, "proofbound-reviewer", "review")
	verdicts := strings.Split(strings.TrimSpace(gitOutput(ctx, root, "ls-tree", "-r", "--name-only", head, "--", "docs/verification/verdicts/")), "\n")
	if s, _ := reviewStatus(ctx, root, head, claim, verdicts); s != "NEEDS_WORK" {
		t.Fatalf("%q (verdicts %v)", s, verdicts)
	}
}
