package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	connectorintent "github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	connectorreviews "github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
)

func gitRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if out, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	return root
}

// CLI-INV-18: init writes the kit, never overwrites, keeps existing content, and a second run
// changes nothing.
func TestInitOpenSpecNeverOverwritesAndIsIdempotent(t *testing.T) {
	root := gitRepo(t)
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("# My agents\nkeep me"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("node_modules\n.proofbound/*\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := initKit(context.Background(), root, "openspec", "Alice", &out); err != nil {
		t.Fatal(err)
	}
	snapshot := map[string]string{}
	for _, rel := range []string{".proofbound/sdd-profile.json", ".gitignore", "AGENTS.md", ".claude/agents/proofbound-reviewer.md", ".git/hooks/commit-msg"} {
		b, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("%s: %v\n%s", rel, err, out.String())
		}
		snapshot[rel] = string(b)
	}
	var profile struct {
		Owners []struct{ Owner string } `json:"owners"`
	}
	if err := json.Unmarshal([]byte(snapshot[".proofbound/sdd-profile.json"]), &profile); err != nil || len(profile.Owners) != 1 || profile.Owners[0].Owner != "Alice" {
		t.Fatalf("profile owner: %+v %v", profile, err)
	}
	if !strings.HasPrefix(snapshot["AGENTS.md"], "# My agents\nkeep me\n") || !strings.Contains(snapshot["AGENTS.md"], agentsBegin) {
		t.Fatalf("AGENTS.md lost its content or lacks the block:\n%s", snapshot["AGENTS.md"])
	}
	if snapshot[".gitignore"] != "node_modules\n.proofbound/*\n\n!.proofbound/sdd-profile.json\n" {
		t.Fatalf(".gitignore = %q", snapshot[".gitignore"])
	}
	if info, err := os.Stat(filepath.Join(root, ".git/hooks/commit-msg")); err != nil || info.Mode().Perm()&0o100 == 0 {
		t.Fatalf("hook not executable: %v", err)
	}

	out.Reset()
	if err := initKit(context.Background(), root, "openspec", "Bob", &out); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "wrote ") || strings.Contains(out.String(), "appended ") {
		t.Fatalf("second run changed something:\n%s", out.String())
	}
	for rel, want := range snapshot {
		if b, _ := os.ReadFile(filepath.Join(root, rel)); string(b) != want {
			t.Fatalf("second run rewrote %s", rel)
		}
	}
}

// The commit-msg hook rejects a message without an Intent trailer unless explicitly opted out.
func TestInitHookRequiresAnIntentTrailer(t *testing.T) {
	hook := filepath.Join(t.TempDir(), "commit-msg")
	if err := os.WriteFile(hook, []byte(commitMsgHook), 0o755); err != nil {
		t.Fatal(err)
	}
	msg := filepath.Join(t.TempDir(), "msg")
	run := func(text string, env ...string) error {
		if err := os.WriteFile(msg, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command("sh", hook, msg)
		cmd.Env = append(os.Environ(), env...)
		return cmd.Run()
	}
	if run("work\n\nIntent: sdd:add-auth\n") != nil {
		t.Fatal("a trailer was rejected")
	}
	if run("work\n") == nil {
		t.Fatal("a missing trailer was accepted")
	}
	if run("work\n", "PROOFBOUND_NO_INTENT=1") != nil {
		t.Fatal("the opt-out was rejected")
	}
}

const sealCommit = "c772ac085ec235cbcf36535701c6109c9cb0bca2"
const sealEvidence = "01M28TPW9C8R7ND19MNDCJ9GDG"

func sealDraft(outcome, reviewOutcome string) draftFile {
	digest := strings.Repeat("a", 64)
	req := connectorintent.Reference{RecordKind: "requirement", Source: "intent.sdd", RecordID: "BR-add-auth-login-abcdef", ArtifactSHA256: digest, Relation: "evaluates"}
	review := req
	review.Relation = "reviews"
	return draftFile{
		Commit:           sealCommit,
		DeclaredReviewer: "proofbound-reviewer",
		Reviews: []draftReview{{Requirement: review, DeclaredOwner: "alice", Obligations: []draftReviewObligation{
			{ObligationID: "O-B", Outcome: reviewOutcome}, {ObligationID: "O-A", Outcome: reviewOutcome},
		}}},
		Verdicts: []draftVerdict{{
			ChangeIntent: connectorintent.Reference{RecordKind: "change_intent", Source: "intent.sdd", RecordID: "CI-add-auth-xuuno2", ArtifactSHA256: strings.Repeat("b", 64), Relation: "evaluates"},
			Requirements: []connectorintent.Reference{req},
			Obligations: []draftVerdictObligation{
				{Source: "intent.sdd", RequirementID: req.RecordID, ArtifactSHA256: digest, ObligationID: "O-B", Outcome: outcome, EvidenceEventIDs: []string{sealEvidence}},
				{Source: "intent.sdd", RequirementID: req.RecordID, ArtifactSHA256: digest, ObligationID: "O-A", Outcome: outcome, EvidenceEventIDs: []string{sealEvidence}},
			},
		}},
	}
}

// ledgerTruth is what the ledger says about sealCommit: sealDraft with no decisions.
func ledgerTruth(context.Context, string) (draftFile, error) {
	d := sealDraft("", "")
	d.DeclaredReviewer = ""
	return d, nil
}

func writeDraft(t *testing.T, root string, d draftFile) string {
	t.Helper()
	dir := draftDir(root, d.Commit)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "draft.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// CLI-INV-19 (unit half): a sealed set re-parses with the reviews connector's parsers, its status
// follows the outcomes, and it writes the reviewed commit's own readiness gate. The end-to-end
// half (gate PASS/BLOCKED on a real ledger) is the sdd package's TestStarterKitReviewFlow.
func TestReviewSealWritesAcceptedArtifacts(t *testing.T) {
	for _, c := range []struct{ outcome, status string }{{"SATISFIED", "ACCEPTABLE"}, {"INCONCLUSIVE", "NEEDS_WORK"}} {
		root := t.TempDir()
		dir := writeDraft(t, root, sealDraft(c.outcome, "VERIFIABLE"))
		var out bytes.Buffer
		if err := reviewSeal(context.Background(), root, dir, ledgerTruth, &out); err != nil {
			t.Fatalf("%s: %v", c.outcome, err)
		}
		short := sealCommit[:12]
		vpath := "docs/verification/verdicts/OV-" + short + "-1.md"
		vdata, err := os.ReadFile(filepath.Join(root, vpath))
		if err != nil {
			t.Fatal(err)
		}
		v, err := connectorreviews.ParseObligationVerdict(vpath, vdata)
		if err != nil || v.Status != c.status || v.DeclaredReviewer != "proofbound-reviewer" || v.ReviewedCommit != sealCommit {
			t.Fatalf("%s: verdict %+v %v", c.outcome, v, err)
		}
		rpath := "docs/verification/verdicts/RR-" + short + "-1.md"
		rdata, err := os.ReadFile(filepath.Join(root, rpath))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := connectorreviews.ParseRequirementReview(rpath, rdata); err != nil {
			t.Fatalf("review: %v", err)
		}
		gate, err := os.ReadFile(filepath.Join(root, "gates", "ready-"+short+".yaml"))
		if err != nil || !strings.Contains(string(gate), `"scope_commit":"`+sealCommit+`"`) {
			t.Fatalf("gate %s %v", gate, err)
		}
	}
}

// CLI-INV-20: seal writes nothing when any check fails, and takes nothing but the reviewer's
// decisions from the editable draft: every reference, digest, owner, evidence id and obligation
// comes from the ledger (review r2-starter-kit-review-20261119 F1).
func TestReviewSealWritesNothingOnAnyFailure(t *testing.T) {
	badEvidence := func(ctx context.Context, sha string) (draftFile, error) {
		d, _ := ledgerTruth(ctx, sha)
		d.Verdicts[0].Obligations[0].EvidenceEventIDs = []string{"not-an-event"}
		return d, nil
	}
	for name, c := range map[string]struct {
		mutate func(*draftFile)
		fresh  func(context.Context, string) (draftFile, error)
	}{
		"undecided outcome":      {mutate: func(d *draftFile) { d.Verdicts[0].Obligations[1].Outcome = "" }},
		"undecided review":       {mutate: func(d *draftFile) { d.Reviews[0].Obligations[0].Outcome = "MAYBE" }},
		"blank reviewer":         {mutate: func(d *draftFile) { d.DeclaredReviewer = " " }},
		"owner reviews own work": {mutate: func(d *draftFile) { d.DeclaredReviewer = "alice" }},
		"owner edited in draft":  {mutate: func(d *draftFile) { d.DeclaredReviewer, d.Reviews[0].DeclaredOwner = "alice", "someone-else" }},
		"digest edited":          {mutate: func(d *draftFile) { d.Verdicts[0].Obligations[0].ArtifactSHA256 = strings.Repeat("c", 64) }},
		"review digest edited":   {mutate: func(d *draftFile) { d.Reviews[0].Requirement.ArtifactSHA256 = strings.Repeat("c", 64) }},
		"obligation removed":     {mutate: func(d *draftFile) { d.Verdicts[0].Obligations = d.Verdicts[0].Obligations[:1] }},
		"obligation added": {mutate: func(d *draftFile) {
			extra := d.Verdicts[0].Obligations[0]
			extra.ObligationID = "O-Z"
			d.Verdicts[0].Obligations = append(d.Verdicts[0].Obligations, extra)
		}},
		"ledger lookup fails": {fresh: func(context.Context, string) (draftFile, error) { return draftFile{}, errors.New("not in the ledger") }},
		"artifact rejected":   {fresh: badEvidence},
	} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			d := sealDraft("SATISFIED", "VERIFIABLE")
			if c.mutate != nil {
				c.mutate(&d)
			}
			fresh := c.fresh
			if fresh == nil {
				fresh = ledgerTruth
			}
			dir := writeDraft(t, root, d)
			if err := reviewSeal(context.Background(), root, dir, fresh, &bytes.Buffer{}); err == nil {
				t.Fatal("seal accepted a bad draft")
			}
			for _, sub := range []string{"docs", "gates"} {
				if _, err := os.Stat(filepath.Join(root, sub)); err == nil {
					t.Fatalf("seal wrote %s despite failing", sub)
				}
			}
		})
	}
	t.Run("existing target", func(t *testing.T) {
		root := t.TempDir()
		dir := writeDraft(t, root, sealDraft("SATISFIED", "VERIFIABLE"))
		gate := filepath.Join(root, "gates", "ready-"+sealCommit[:12]+".yaml")
		if err := os.MkdirAll(filepath.Dir(gate), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(gate, []byte("mine"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := reviewSeal(context.Background(), root, dir, ledgerTruth, &bytes.Buffer{}); err == nil {
			t.Fatal("seal overwrote an existing gate")
		}
		if _, err := os.Stat(filepath.Join(root, "docs")); err == nil {
			t.Fatal("seal wrote artifacts despite failing")
		}
		if b, _ := os.ReadFile(gate); string(b) != "mine" {
			t.Fatal("the existing gate was changed")
		}
	})
}

// Evidence edited in the draft is ignored: the sealed verdict cites the ledger's evidence only.
func TestReviewSealTakesEvidenceFromTheLedger(t *testing.T) {
	root := t.TempDir()
	d := sealDraft("SATISFIED", "VERIFIABLE")
	d.Verdicts[0].Obligations[0].EvidenceEventIDs = []string{"01ARZ3NDEKTSV4RRFFQ69G5FAV"}
	dir := writeDraft(t, root, d)
	if err := reviewSeal(context.Background(), root, dir, ledgerTruth, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	path := "docs/verification/verdicts/OV-" + sealCommit[:12] + "-1.md"
	data, err := os.ReadFile(filepath.Join(root, path))
	if err != nil {
		t.Fatal(err)
	}
	v, err := connectorreviews.ParseObligationVerdict(path, data)
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range v.Obligations {
		if len(o.EvidenceEventIDs) != 1 || o.EvidenceEventIDs[0] != sealEvidence {
			t.Fatalf("obligation %s cites %v, want only the ledger's %s", o.ObligationID, o.EvidenceEventIDs, sealEvidence)
		}
	}
}

// CLI-INV-18 for Spec Kit: `init speckit` writes a Spec Kit profile the sdd provider accepts.
func TestInitSpecKitWritesASpecKitProfile(t *testing.T) {
	root := gitRepo(t)
	var out bytes.Buffer
	if err := initKit(context.Background(), root, "speckit", "Alice", &out); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, ".proofbound/sdd-profile.json"))
	if err != nil {
		t.Fatal(err)
	}
	var profile struct {
		Kit     string `json:"kit"`
		Changes struct {
			Root, Sibling string
		} `json:"changes"`
	}
	if err := json.Unmarshal(b, &profile); err != nil || profile.Kit != "speckit" || profile.Changes.Root != "specs" || profile.Changes.Sibling != "tasks.md" {
		t.Fatalf("profile %+v %v", profile, err)
	}
	if err := initKit(context.Background(), root, "kiro", "Alice", &out); err == nil {
		t.Fatal("an unknown kit was accepted")
	}
}

// CLI-INV-24: seal accepts the draft.json path `review draft` prints, as well as its folder, and
// seals exactly what the folder form seals.
func TestReviewSealAcceptsTheDraftFile(t *testing.T) {
	root := t.TempDir()
	dir := writeDraft(t, root, sealDraft("SATISFIED", "VERIFIABLE"))
	if err := reviewSeal(context.Background(), root, filepath.Join(dir, "draft.json"), ledgerTruth, &bytes.Buffer{}); err != nil {
		t.Fatalf("seal draft.json: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "gates", "ready-"+sealCommit[:12]+".yaml")); err != nil {
		t.Fatalf("seal draft.json wrote no gate: %v", err)
	}
	other := filepath.Join(dir, "notes.json")
	if err := os.WriteFile(other, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := reviewSeal(context.Background(), t.TempDir(), other, ledgerTruth, &bytes.Buffer{}); err == nil {
		t.Fatal("seal accepted a file that is not draft.json")
	}
}
