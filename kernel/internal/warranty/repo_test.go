package warranty

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/connector/git/gitcmd"
	connectorintent "github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	"github.com/kamisrini/proofbound/kernel/internal/connector/intent/sdd"
	"github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
)

const profile = `{
  "schema": "proofbound.sdd-profile.v1",
  "kit": "speckit",
  "requirements": { "paths": ["specs/*/spec.md"] },
  "changes": { "mode": "feature-sibling", "root": "specs", "sibling": "tasks.md", "status": "accepted" },
  "owners":   [ { "pattern": "specs/**", "owner": "spec-owner" } ],
  "sponsors": [ { "pattern": "specs/**", "sponsor": "spec-owner" } ]
}
`

type fixture struct {
	t        *testing.T
	root     string
	commit   string
	body     Body
	key      ed25519.PrivateKey // the exporter's
	reviewer ed25519.PrivateKey // "reviewer", declared at KeysDir/reviewer.pub; signs every artifact (A)
}

func (x *fixture) git(args ...string) string {
	x.t.Helper()
	cmd := exec.Command("git", append([]string{"-C", x.root, "-c", "user.name=t", "-c", "user.email=t@example.invalid"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		x.t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func (x *fixture) write(rel, body string) {
	x.t.Helper()
	path := filepath.Join(x.root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		x.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		x.t.Fatal(err)
	}
}

// artifact renders a review artifact exactly as `review seal` does: front matter with its own
// self-digest.
func artifact(t *testing.T, build func(digest string) any) string {
	t.Helper()
	render := func(digest string) string {
		front, err := json.Marshal(build(digest))
		if err != nil {
			t.Fatal(err)
		}
		return "---\n" + string(front) + "\n---\n\nreview\n"
	}
	return render(reviews.ArtifactSHA256([]byte(render(strings.Repeat("0", 64)))))
}

// newFixture builds a Spec Kit project whose commit is claimed, reviewed by a non-owner, and whose
// key is declared at HEAD, with a signed body that Verify accepts against it.
func newFixture(t *testing.T) *fixture {
	x := &fixture{t: t, root: t.TempDir(), key: testKey(t), reviewer: ed25519.NewKeyFromSeed(append(make([]byte, 31), 7))}
	x.git("init", "-q", "-b", "main")
	x.write(".proofbound/sdd-profile.json", profile)
	x.write("specs/001-x/spec.md", "# X\n\n- **FR-001**: System MUST do A.\n- **FR-002**: System MUST do B.\n")
	x.write("specs/001-x/tasks.md", "- [ ] T001\n")
	x.git("add", "-A")
	x.git("commit", "-qm", "implement", "-m", "Intent: sdd:001-x")
	x.commit = x.git("rev-parse", "HEAD")

	repo, err := gitcmd.New(x.root)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := sdd.NewGitReader(repo)
	if err != nil {
		t.Fatal(err)
	}
	p, err := sdd.New(reader, time.Unix(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	change, reqRevs, err := p.StateRecords(context.Background(), x.commit, "001-x")
	if err != nil {
		t.Fatal(err)
	}
	var ci connectorintent.ChangeIntent
	if err := json.Unmarshal(change.Payload, &ci); err != nil {
		t.Fatal(err)
	}
	var req connectorintent.Requirement
	if err := json.Unmarshal(reqRevs[0].Payload, &req); err != nil {
		t.Fatal(err)
	}
	rref := connectorintent.Reference{RecordKind: "requirement", Source: "intent.sdd", RecordID: req.RequirementID, ArtifactSHA256: req.ArtifactSHA256}
	short := x.commit[:12]
	rrPath, ovPath := "docs/verification/verdicts/RR-"+short+"-1.md", "docs/verification/verdicts/OV-"+short+"-1.md"
	x.write(rrPath, artifact(t, func(d string) any {
		r := rref
		r.Relation = "reviews"
		return reviews.RequirementReview{Schema: "proofbound.requirement-review.v1", ReviewID: "RR-" + short + "-1", Requirement: r, DeclaredReviewer: "reviewer",
			Outcomes: []reviews.RequirementReviewOutcome{{ObligationID: "O-FR-001", Outcome: "VERIFIABLE"}, {ObligationID: "O-FR-002", Outcome: "VERIFIABLE"}}, ArtifactPath: rrPath, ArtifactSHA256: d}
	}))
	x.write(ovPath, artifact(t, func(d string) any {
		r := rref
		r.Relation = "evaluates"
		var outcomes []reviews.ObligationOutcome
		for _, id := range []string{"O-FR-001", "O-FR-002"} {
			outcomes = append(outcomes, reviews.ObligationOutcome{Source: "intent.sdd", RequirementID: req.RequirementID, ArtifactSHA256: req.ArtifactSHA256, ObligationID: id, Outcome: "SATISFIED", EvidenceEventIDs: []string{"01M7VB8XHZXNSVFATADT1XJ0KC"}})
		}
		return reviews.ObligationVerdict{Schema: "proofbound.obligation-verdict.v2", VerdictID: "OV-" + short + "-1", Status: "ACCEPTABLE", DeclaredReviewer: "reviewer", ReviewedCommit: x.commit,
			ChangeIntent: connectorintent.Reference{RecordKind: "change_intent", Source: "intent.sdd", RecordID: ci.IntentID, ArtifactSHA256: ci.ArtifactSHA256, Relation: "evaluates"},
			Requirements: []connectorintent.Reference{r}, Obligations: outcomes, Findings: []reviews.Finding{}, ArtifactPath: ovPath, ArtifactSHA256: d}
	}))
	x.write(KeysDir+"/owner.pub", PubLine(x.key.Public().(ed25519.PublicKey)))
	x.write(KeysDir+"/reviewer.pub", PubLine(x.reviewer.Public().(ed25519.PublicKey)))
	for _, rel := range []string{rrPath, ovPath} {
		data, err := os.ReadFile(filepath.Join(x.root, rel))
		if err != nil {
			t.Fatal(err)
		}
		x.write(rel+".sig", SignArtifact(data, x.reviewer))
	}
	x.git("add", "-A")
	x.git("commit", "-qm", "review and key")

	x.body = Body{Schema: Schema, Commit: x.commit, ChangeName: "001-x", IssuedAt: "2026-11-19T00:00:00Z", Issuer: "spec-owner", ChangeIntent: ci, Requirements: []connectorintent.Requirement{req}}
	for _, rel := range []string{rrPath, ovPath} {
		data, err := os.ReadFile(filepath.Join(x.root, rel))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(rel, "/RR-") {
			r, err := reviews.ParseRequirementReview(rel, data)
			if err != nil {
				t.Fatal(err)
			}
			x.body.Reviews = append(x.body.Reviews, r)
		} else {
			v, err := reviews.ParseObligationVerdict(rel, data)
			if err != nil {
				t.Fatal(err)
			}
			x.body.Verdicts = append(x.body.Verdicts, v)
		}
	}
	return x
}

func (x *fixture) verify(b Body) error {
	x.t.Helper()
	return Verify(context.Background(), signParse(x.t, b), x.root, "")
}

// W-INV-5 and W-INV-6 inside this package (code review r3-warranty-code-review F7): against a real
// repository, the fixture verifies VALID, and every repository check rejects its own tampering.
func TestVerifyAgainstARepository(t *testing.T) {
	x := newFixture(t)
	if err := x.verify(x.body); err != nil {
		t.Fatalf("calibration: the fixture does not verify: %v", err)
	}
	pub := x.key.Public().(ed25519.PublicKey)
	if err := Verify(context.Background(), signParse(t, x.body), x.root, Fingerprint(pub)); err != nil {
		t.Fatalf("calibration: a matching pin fails: %v", err)
	}

	bodyCases := map[string]struct {
		edit func(*Body)
		want string
	}{
		"reworded obligation":   {func(b *Body) { b.Requirements[0].Obligations[0].Statement = "System MUST do A, later." }, "requirements differ"},
		"other owner":           {func(b *Body) { b.Requirements[0].DeclaredOwner = "someone" }, "requirements differ"},
		"change intent status":  {func(b *Body) { b.ChangeIntent.Status = "proposed" }, "change intent differs"},
		"unknown change name":   {func(b *Body) { b.ChangeName = "002-y" }, "spec at"},
		"commit not in clone":   {func(b *Body) { b.Commit = strings.Repeat("a", 40); b.Verdicts[0].ReviewedCommit = b.Commit }, "not in clone"},
		"verdict field differs": {func(b *Body) { b.Verdicts[0].Status = "NEEDS_WORK" }, "differs from the warranty"},
		"review field differs":  {func(b *Body) { b.Reviews[0].ReviewID = "RR-other-1" }, "differs from the warranty"},
		"artifact missing":      {func(b *Body) { b.Reviews[0].ArtifactPath = "docs/verification/verdicts/RR-none.md" }, "not committed at HEAD"},
	}
	for name, c := range bodyCases {
		raw, _ := json.Marshal(x.body)
		var b Body
		if err := json.Unmarshal(raw, &b); err != nil {
			t.Fatal(err)
		}
		c.edit(&b)
		if err := x.verify(b); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v, want %q", name, err, c.want)
		}
	}

	repoCases := map[string]struct {
		change func()
		want   string
	}{
		"key removed": {func() { x.git("rm", "-q", "-r", KeysDir) }, "key not declared"},
		"key in a subdirectory": {func() {
			x.git("mv", KeysDir+"/owner.pub", KeysDir+"/x.pub.tmp")
			x.write(KeysDir+"/sub/owner.pub", PubLine(pub))
			x.git("rm", "-q", "--cached", KeysDir+"/x.pub.tmp")
			x.git("add", "-A")
		}, "key not declared"},
		"key as a symlink": {func() {
			x.git("mv", KeysDir+"/owner.pub", "elsewhere.pub")
			if err := os.Symlink("../../elsewhere.pub", filepath.Join(x.root, KeysDir, "owner.pub")); err != nil {
				t.Fatal(err)
			}
			x.git("add", "-A")
		}, "key not declared"},
		"artifact edited": {func() {
			rel := x.body.Reviews[0].ArtifactPath
			data, _ := os.ReadFile(filepath.Join(x.root, rel))
			x.write(rel, strings.Replace(string(data), "VERIFIABLE", "AMBIGUOUS", 1))
			x.git("add", "-A")
		}, "artifact"},
		"artifact as a symlink": {func() {
			rel := x.body.Verdicts[0].ArtifactPath
			x.git("mv", rel, "ov.md")
			if err := os.Symlink("../../../ov.md", filepath.Join(x.root, filepath.FromSlash(rel))); err != nil {
				t.Fatal(err)
			}
			x.git("add", "-A")
		}, "not committed at HEAD"},
		"HEAD not descended": {func() {
			x.git("checkout", "-q", "--orphan", "other")
			x.git("add", "-A")
		}, "does not descend"},
	}
	head := x.git("rev-parse", "HEAD")
	for name, c := range repoCases {
		c.change()
		x.git("commit", "-q", "--allow-empty", "-m", name)
		if err := x.verify(x.body); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v, want %q", name, err, c.want)
		}
		x.git("checkout", "-q", "-f", "main")
		x.git("reset", "-q", "--hard", head)
		x.git("clean", "-qfdx")
		if err := x.verify(x.body); err != nil {
			t.Fatalf("after %s: the restored fixture no longer verifies: %v", name, err)
		}
	}
	if err := Verify(context.Background(), signParse(t, x.body), x.root, "0000000000000000"); err == nil || errors.Is(err, ErrUnverified) {
		t.Fatalf("a wrong pin passed: %v", err)
	}
}

// Mutation-sweep kills for verify (W-INV-6): a `.pub` holding another key is not a declaration, and
// a symlink whose target text is exactly the key line is not a regular file.
func TestKeyDeclarationIsExact(t *testing.T) {
	x := newFixture(t)
	pub := x.key.Public().(ed25519.PublicKey)
	other := ed25519.NewKeyFromSeed(append(make([]byte, 31), 1)).Public().(ed25519.PublicKey)
	x.write(KeysDir+"/owner.pub", PubLine(other))
	x.git("commit", "-qam", "another key")
	if err := x.verify(x.body); err == nil || !strings.Contains(err.Error(), "key not declared") {
		t.Fatalf("another key's .pub declared ours: %v", err)
	}
	x.git("rm", "-q", KeysDir+"/owner.pub")
	if err := os.MkdirAll(filepath.Join(x.root, KeysDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(PubLine(pub), filepath.Join(x.root, KeysDir, "owner.pub")); err != nil {
		t.Fatal(err)
	}
	x.git("add", "-A")
	x.git("commit", "-qm", "symlink whose target is the key line")
	if err := x.verify(x.body); err == nil || !strings.Contains(err.Error(), "key not declared") {
		t.Fatalf("a symlink declared the key: %v", err)
	}
}
