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
	for _, rel := range []string{".proofbound/sdd-profile.json", ".gitignore", "AGENTS.md", ".claude/agents/proofbound-reviewer.md", ".git/hooks/commit-msg", ".git/hooks/prepare-commit-msg", ".github/workflows/proofbound.yml"} {
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
	if snapshot[".gitignore"] != "node_modules\n.proofbound/*\n\n!.proofbound/sdd-profile.json\n!.proofbound/keys/\n!.proofbound/waivers/\n!.proofbound/tests.json\n!.proofbound/agent-review.json\n" {
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

// Both hook scripts only hand over to `proofbound hook` (CLI-INV-35/36 live in Go, one home); the
// commit-msg script fails closed when proofbound is missing, prepare-commit-msg never blocks.
func TestInitHookScriptsDelegate(t *testing.T) {
	if !strings.Contains(commitMsgHook, "proofbound hook commit-msg \"$1\"") || !strings.Contains(commitMsgHook, "exit 1") || !strings.Contains(commitMsgHook, "too old") {
		t.Fatalf("commit-msg hook:\n%s", commitMsgHook)
	}
	if !strings.Contains(prepareCommitMsgHook, "proofbound hook prepare-commit-msg \"$@\"") || !strings.HasSuffix(strings.TrimSpace(prepareCommitMsgHook), "exit 0") {
		t.Fatalf("prepare-commit-msg hook:\n%s", prepareCommitMsgHook)
	}
	if strings.Contains(commitMsgHook+prepareCommitMsgHook, "PROOFBOUND_NO_INTENT") {
		t.Fatal("a hook still honours the retired opt-out")
	}
}

const sealCommit = "c772ac085ec235cbcf36535701c6109c9cb0bca2"

var sealEvidence = connectorreviews.CommitEvidenceRef(sealCommit)

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
				{Source: "intent.sdd", RequirementID: req.RecordID, ArtifactSHA256: digest, ObligationID: "O-B", Outcome: outcome, EvidenceRefs: []string{sealEvidence}},
				{Source: "intent.sdd", RequirementID: req.RecordID, ArtifactSHA256: digest, ObligationID: "O-A", Outcome: outcome, EvidenceRefs: []string{sealEvidence}},
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
		d.Verdicts[0].Obligations[0].EvidenceRefs = []string{"not-an-event"}
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

// CLI-INV-20 (amended, portable evidence): evidence edited in the draft is ignored, and the sealed
// verdict cites only the reviewed commit's portable reference with no ledger-local event id.
func TestReviewSealCitesTheReviewedCommit(t *testing.T) {
	root := t.TempDir()
	d := sealDraft("SATISFIED", "VERIFIABLE")
	d.Verdicts[0].Obligations[0].EvidenceRefs = []string{connectorreviews.CommitEvidenceRef(strings.Repeat("e", 40))}
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
		if len(o.EvidenceRefs) != 1 || o.EvidenceRefs[0] != sealEvidence || o.EvidenceEventIDs == nil || len(o.EvidenceEventIDs) != 0 {
			t.Fatalf("obligation %s cites refs %v ids %v, want only %s and an empty id list", o.ObligationID, o.EvidenceRefs, o.EvidenceEventIDs, sealEvidence)
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

// CLI-INV-32: the reviewer agent's front matter is strict YAML, so editors and agent loaders that
// parse it strictly accept it. Each line is `key: value` where the value is either a double-quoted
// string (checked as JSON, a subset of YAML double-quoted scalars) or a plain scalar with nothing a
// strict parser reads as structure: no ": " or " #" inside, no leading indicator character.
func TestInitReviewerAgentFrontMatterIsStrictYAML(t *testing.T) {
	root := gitRepo(t)
	var out bytes.Buffer
	if err := initKit(context.Background(), root, "speckit", "Alice", &out); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(root, ".claude/agents/proofbound-reviewer.md"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(b), "\n")
	if lines[0] != "---" {
		t.Fatalf("no front matter: %q", lines[0])
	}
	keys := map[string]string{}
	closed := false
	for _, line := range lines[1:] {
		if line == "---" {
			closed = true
			break
		}
		key, value, ok := strings.Cut(line, ": ")
		if !ok || key == "" || strings.ContainsAny(key, " \"'") {
			t.Fatalf("front-matter line is not `key: value`: %q", line)
		}
		if strings.HasPrefix(value, `"`) {
			var s string
			if err := json.Unmarshal([]byte(value), &s); err != nil {
				t.Fatalf("%s: quoted value does not parse: %v", key, err)
			}
			value = s
		} else if strings.Contains(value, ": ") || strings.Contains(value, " #") || strings.HasSuffix(value, ":") ||
			strings.ContainsAny(value[:1], "-?:,[]{}#&*!|>'%@`\"") {
			t.Fatalf("%s: plain value a strict YAML parser rejects or misreads; quote it: %q", key, value)
		}
		keys[key] = value
	}
	if !closed {
		t.Fatal("front matter is not closed")
	}
	if keys["name"] != "proofbound-reviewer" || !strings.Contains(keys["description"], "`Intent: sdd:<change>`") {
		t.Fatalf("front matter lost its content: %+v", keys)
	}
}
