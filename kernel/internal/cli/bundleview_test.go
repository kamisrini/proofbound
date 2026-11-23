package cli

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
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
	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

const bundleProfile = `{
  "schema": "proofbound.sdd-profile.v1",
  "kit": "speckit",
  "requirements": { "paths": ["specs/*/spec.md"] },
  "changes": { "mode": "feature-sibling", "root": "specs", "sibling": "tasks.md", "status": "accepted" },
  "owners":   [ { "pattern": "specs/**", "owner": "spec-owner" } ],
  "sponsors": [ { "pattern": "specs/**", "sponsor": "spec-owner" } ]
}
`

// signedWarranty builds a Spec Kit repository whose claimed commit is reviewed by a non-owner with a
// declared, signing key, and returns the repository, the exporter's key and a warranty that verifies
// VALID against it (the warranty package's own fixture, ported).
func signedWarranty(t *testing.T) (string, ed25519.PrivateKey, []byte, string) {
	t.Helper()
	ctx := context.Background()
	root := taskRepo(t)
	exporter := ed25519.NewKeyFromSeed(make([]byte, ed25519.SeedSize))
	reviewer := ed25519.NewKeyFromSeed(append(make([]byte, 31), 7))
	uiWrite(t, root, "README.md", "start\n")
	uiCommit(t, root, "dev", "start") // a commit before the claim, so a range can contain it
	uiWrite(t, root, ".proofbound/sdd-profile.json", bundleProfile)
	uiWrite(t, root, "specs/001-x/spec.md", "# X\n\n- **FR-001**: System MUST do A.\n- **FR-002**: System MUST do B.\n")
	uiWrite(t, root, "specs/001-x/tasks.md", "- [ ] T001\n")
	commit := uiCommit(t, root, "dev", "implement\n\nIntent: sdd:001-x")
	repo, err := gitcmd.New(root)
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
	change, reqRevs, err := p.StateRecords(ctx, commit, "001-x")
	if err != nil {
		t.Fatal(err)
	}
	var ci connectorintent.ChangeIntent
	var req connectorintent.Requirement
	if err := json.Unmarshal(change.Payload, &ci); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(reqRevs[0].Payload, &req); err != nil {
		t.Fatal(err)
	}
	rref := connectorintent.Reference{RecordKind: "requirement", Source: "intent.sdd", RecordID: req.RequirementID, ArtifactSHA256: req.ArtifactSHA256}
	short := commit[:12]
	rrPath, ovPath := "docs/verification/verdicts/RR-"+short+"-1.md", "docs/verification/verdicts/OV-"+short+"-1.md"
	render := func(build func(digest string) any) []byte {
		r := func(d string) []byte {
			front, err := json.Marshal(build(d))
			if err != nil {
				t.Fatal(err)
			}
			return []byte("---\n" + string(front) + "\n---\n\nreview\n")
		}
		return r(reviews.ArtifactSHA256(r(strings.Repeat("0", 64))))
	}
	rr := render(func(d string) any {
		r := rref
		r.Relation = "reviews"
		return reviews.RequirementReview{Schema: "proofbound.requirement-review.v1", ReviewID: "RR-" + short + "-1", Requirement: r, DeclaredReviewer: "reviewer",
			Outcomes: []reviews.RequirementReviewOutcome{{ObligationID: "O-FR-001", Outcome: "VERIFIABLE"}, {ObligationID: "O-FR-002", Outcome: "VERIFIABLE"}}, ArtifactPath: rrPath, ArtifactSHA256: d}
	})
	ov := render(func(d string) any {
		r := rref
		r.Relation = "evaluates"
		var outcomes []reviews.ObligationOutcome
		for _, id := range []string{"O-FR-001", "O-FR-002"} {
			outcomes = append(outcomes, reviews.ObligationOutcome{Source: "intent.sdd", RequirementID: req.RequirementID, ArtifactSHA256: req.ArtifactSHA256, ObligationID: id, Outcome: "SATISFIED", EvidenceEventIDs: []string{"01M7VB8XHZXNSVFATADT1XJ0KC"}})
		}
		return reviews.ObligationVerdict{Schema: "proofbound.obligation-verdict.v2", VerdictID: "OV-" + short + "-1", Status: "ACCEPTABLE", DeclaredReviewer: "reviewer", ReviewedCommit: commit,
			ChangeIntent: connectorintent.Reference{RecordKind: "change_intent", Source: "intent.sdd", RecordID: ci.IntentID, ArtifactSHA256: ci.ArtifactSHA256, Relation: "evaluates"},
			Requirements: []connectorintent.Reference{r}, Obligations: outcomes, Findings: []reviews.Finding{}, ArtifactPath: ovPath, ArtifactSHA256: d}
	})
	uiWrite(t, root, rrPath, string(rr))
	uiWrite(t, root, ovPath, string(ov))
	uiWrite(t, root, rrPath+".sig", pbwarranty.SignArtifact(rr, reviewer))
	uiWrite(t, root, ovPath+".sig", pbwarranty.SignArtifact(ov, reviewer))
	uiWrite(t, root, pbwarranty.KeysDir+"/owner.pub", pbwarranty.PubLine(exporter.Public().(ed25519.PublicKey)))
	uiWrite(t, root, pbwarranty.KeysDir+"/reviewer.pub", pbwarranty.PubLine(reviewer.Public().(ed25519.PublicKey)))
	uiCommit(t, root, "reviewer", "review and keys")
	body := pbwarranty.Body{Schema: pbwarranty.Schema, Commit: commit, ChangeName: "001-x", IssuedAt: "2026-11-19T00:00:00Z", Issuer: "spec-owner", ChangeIntent: ci, Requirements: []connectorintent.Requirement{req}}
	r1, err := reviews.ParseRequirementReview(rrPath, rr)
	if err != nil {
		t.Fatal(err)
	}
	v1, err := reviews.ParseObligationVerdict(ovPath, ov)
	if err != nil {
		t.Fatal(err)
	}
	body.Reviews, body.Verdicts = []reviews.RequirementReview{r1}, []reviews.ObligationVerdict{v1}
	data, err := pbwarranty.Sign(body, exporter)
	if err != nil {
		t.Fatal(err)
	}
	return root, exporter, data, commit
}

// bundleDir writes a bundle folder holding data as the warranty for commit, with an index naming it.
func bundleDir(t *testing.T, commit string, exporter ed25519.PrivateKey, data []byte) (string, string) {
	t.Helper()
	dir := t.TempDir()
	file := "warranty-" + commit[:12] + ".json"
	sum := sha256.Sum256(data)
	ix := bundleIndex{Schema: bundleSchema, Unsigned: true, Note: bundleNote, Base: strings.Repeat("a", 40), Head: commit, ExporterKey: pbwarranty.Fingerprint(exporter.Public().(ed25519.PublicKey)),
		NonClaimingCommits: 1, Entries: []bundleEntry{{Commit: commit, Claims: "sdd:001-x", State: bundleValid, File: file, FileSHA256: hex.EncodeToString(sum[:])},
			{Commit: strings.Repeat("c", 40), Claims: "sdd:002-y", State: checkNoReview, Reason: "no review yet"}},
		Waived: []bundleWaived{{Commit: strings.Repeat("d", 40), Waiver: "vendor-lib", By: "approver", Expires: "2026-12-31", Reason: "vendored <code>"}}}
	raw, err := json.Marshal(ix)
	if err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string][]byte{"index.json": raw, file: data} {
		if err := os.WriteFile(filepath.Join(dir, name), body, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir, file
}

// CLI-INV-60: `ui --bundle` flags.
func TestParseUIBundle(t *testing.T) {
	if o, ok := parseUI([]string{"ui", "--bundle", "w", "--trust-key", "abcd", "--repo", "r", "--port", "0", "--open"}); !ok || o.bundle != "w" || o.trustKey != "abcd" || o.repo != "r" || !o.open {
		t.Errorf("parse = %+v %v", o, ok)
	}
	for _, args := range [][]string{
		{"ui", "--bundle", "w", "--base", "main"},
		{"ui", "--bundle", "w", "--bundle", "x"},
		{"ui", "--bundle"},
		{"ui", "--bundle", "--open"},
		{"ui", "--trust-key", "abcd"},
		{"ui", "--repo", "r"},
		{"ui", "--bundle", "w", "--trust-key", "a", "--trust-key", "b"},
	} {
		if _, ok := parseUI(args); ok || parseCommand(args) != commandInvalid {
			t.Errorf("parsed %q", args)
		}
	}
}

func row(t *testing.T, v bundleView, file string) bundleWarranty {
	t.Helper()
	for _, w := range v.Warranties {
		if w.File == file {
			return w
		}
	}
	t.Fatalf("no row for %s in %+v", file, v.Warranties)
	return bundleWarranty{}
}

// CLI-INV-61: each warranty is checked as `warranty verify` checks it and labelled with how far.
func TestBundleView(t *testing.T) {
	ctx := context.Background()
	root, exporter, data, commit := signedWarranty(t)
	dir, file := bundleDir(t, commit, exporter, data)
	fp := pbwarranty.Fingerprint(exporter.Public().(ed25519.PublicKey))

	v, err := loadBundle(ctx, dir, "", "", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	w := row(t, v, file)
	if w.Result != bundleSignatureOnly || w.Key != fp || w.Change != "001-x" || len(w.Promises) != 1 || len(w.Promises[0].Obligations) != 2 ||
		w.Promises[0].Obligations[0].Statement != "System MUST do A." || !strings.Contains(w.Promises[0].Obligations[0].Outcome, "SATISFIED (reviewer)") || len(w.Reviews) != 1 {
		t.Errorf("no clone: %+v", w)
	}
	if len(v.Warranties) != 2 || v.Warranties[1].Result != bundleNoWarranty || v.Warranties[1].Problem != "no review yet" {
		t.Errorf("index entry without a file: %+v", v.Warranties)
	}
	if v, _ := loadBundle(ctx, dir, "", root, time.Now()); row(t, v, file).Result != bundleValid || len(row(t, v, file).Proven) == 0 {
		t.Errorf("with the clone: %+v", row(t, v, file))
	}
	if v, _ := loadBundle(ctx, dir, "0123456789abcdef", "", time.Now()); row(t, v, file).Result != bundleInvalid || !strings.Contains(row(t, v, file).Problem, "fingerprint mismatch") {
		t.Errorf("wrong trust key: %+v", row(t, v, file))
	}
	if v, _ := loadBundle(ctx, dir, fp, "", time.Now()); row(t, v, file).Result != bundleSignatureOnly {
		t.Errorf("right trust key: %+v", row(t, v, file))
	}

	// An edited file: its bytes differ from the index, and its signature no longer holds.
	edited := strings.Replace(string(data), "2026-11-19T00:00:00Z", "2026-11-20T00:00:00Z", 1)
	if err := os.WriteFile(filepath.Join(dir, file), []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	if v, _ := loadBundle(ctx, dir, "", "", time.Now()); row(t, v, file).Result != bundleAltered || !strings.Contains(row(t, v, file).Problem, "signature") {
		t.Errorf("edited: %+v", row(t, v, file))
	}
	// A deleted file, and an extra signed file the index does not name.
	if err := os.Remove(filepath.Join(dir, file)); err != nil {
		t.Fatal(err)
	}
	extra := "warranty-" + strings.Repeat("e", 12) + ".json"
	if err := os.WriteFile(filepath.Join(dir, extra), data, 0o644); err != nil {
		t.Fatal(err)
	}
	v, _ = loadBundle(ctx, dir, "", "", time.Now())
	if row(t, v, file).Result != bundleMissing || row(t, v, extra).InIndex || row(t, v, extra).Result != bundleAltered || !strings.Contains(row(t, v, extra).Problem, "file name does not match") {
		t.Errorf("missing, and an extra file under another commit's name: %+v", v.Warranties)
	}
	// A symlinked warranty is refused; a folder without index.json does not start.
	if err := os.Symlink(filepath.Join(dir, extra), filepath.Join(dir, file)); err != nil {
		t.Fatal(err)
	}
	if v, _ := loadBundle(ctx, dir, "", "", time.Now()); row(t, v, file).Result != bundleInvalid || !strings.Contains(row(t, v, file).Problem, "not a regular file") {
		t.Errorf("symlink: %+v", row(t, v, file))
	}
	if _, err := loadBundle(ctx, t.TempDir(), "", "", time.Now()); err == nil || !strings.Contains(err.Error(), "no readable index.json") {
		t.Errorf("no index: %v", err)
	}
}

// CLI-INV-62: the bundle server is read-only behind the same guards, and carries the browser proof.
func TestBundleServer(t *testing.T) {
	ctx := context.Background()
	_, exporter, data, commit := signedWarranty(t)
	dir, file := bundleDir(t, commit, exporter, data)
	v, err := loadBundle(ctx, dir, "", "", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	s := newUIServer("", "", "", 4242)
	s.bundle = &v
	host := "127.0.0.1:4242"
	if w := uiRequest(s, "GET", "/", host, "", "", ""); w.Code != http.StatusUnauthorized {
		t.Errorf("no cookie: %d", w.Code)
	}
	if w := uiRequest(s, "GET", "/", "evil.example:4242", "", s.token, ""); w.Code != http.StatusMisdirectedRequest {
		t.Errorf("wrong host: %d", w.Code)
	}
	if w := uiRequest(s, "GET", "/?k="+s.token, host, "", "", ""); w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/" {
		t.Errorf("link: %d %s", w.Code, w.Header().Get("Location"))
	}
	if w := uiRequestWith(s, "GET", "/", "Bearer "+s.agentToken); w.Code != http.StatusUnauthorized {
		t.Errorf("an agent token read a bundle page: %d", w.Code)
	}
	for _, p := range []string{"/audit", "/review", "/agent/next", "/transcript"} {
		if w := uiRequest(s, "GET", p, host, "", s.token, ""); w.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d", p, w.Code)
		}
	}
	for _, p := range []string{"/act/sign", "/act/review-seal", "/"} {
		if w := uiRequest(s, "POST", p, host, "http://"+host, s.token, "{}"); w.Code != http.StatusNotFound {
			t.Errorf("POST %s = %d", p, w.Code)
		}
	}
	if w := uiRequest(s, "GET", "/static/app.js", host, "", "", ""); w.Code != http.StatusOK {
		t.Errorf("static: %d", w.Code)
	}
	page := uiRequest(s, "GET", "/", host, "", s.token, "").Body.String()
	for _, want := range []string{"SIGNATURE ONLY", "System MUST do A.", "Waived commits", "from the unsigned index", "vendored &lt;code&gt;", "not pinned", `data-kind="file"`} {
		if !strings.Contains(page, want) {
			t.Errorf("page lacks %q", want)
		}
	}
	// The proof the page carries verifies as the browser checks it: tag + bytes, `.sig`-style line.
	pr := row(t, v, file).Proof
	if pr == nil {
		t.Fatal("no browser proof")
	}
	raw, _ := base64.StdEncoding.DecodeString(pr.Bytes64)
	f := strings.Fields(pr.Sig)
	sig, _ := hex.DecodeString(f[2])
	pub, _ := hex.DecodeString(pr.KeyHex)
	if len(f) != 3 || f[0] != "ed25519" || f[1] != pr.KeyHex || !ed25519.Verify(pub, append([]byte(pr.Tag), raw...), sig) {
		t.Errorf("the browser proof does not verify: %+v", pr)
	}
	r := uiRequest(s, "GET", "/?format=json", host, "", s.token, "")
	var got bundleView
	if err := json.Unmarshal(r.Body.Bytes(), &got); err != nil || len(got.Warranties) != 2 || got.Warranties[0].Result != bundleSignatureOnly {
		t.Errorf("json twin: %v %s", err, r.Body.String())
	}
}

// The command runs outside any repository (CLI-INV-60): run() must not need a root.
func TestBundleRunsOutsideARepository(t *testing.T) {
	if testing.Short() {
		t.Skip("starts a server")
	}
	_, exporter, data, commit := signedWarranty(t)
	dir, _ := bundleDir(t, commit, exporter, data)
	bin := filepath.Join(t.TempDir(), "pb")
	if out, err := exec.Command("go", "build", "-o", bin, "github.com/kamisrini/proofbound/kernel/cmd/proofbound").CombinedOutput(); err != nil {
		t.Skipf("cannot build the binary here: %v %s", err, out)
	}
	cmd := exec.Command(bin, "ui", "--bundle", dir, "--port", "0")
	cmd.Dir = t.TempDir() // not a repository
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()
	buf := make([]byte, 512)
	n, _ := stdout.Read(buf)
	if !strings.Contains(string(buf[:n]), "http://127.0.0.1:") || !strings.Contains(string(buf[:n]), "bundle "+dir) {
		t.Errorf("no link printed: %q", buf[:n])
	}
}

func writeIndex(t *testing.T, dir string, ix bundleIndex) {
	t.Helper()
	raw, err := json.Marshal(ix)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
}

// Review ui-bundle-review-20261123 F1..F3: the page never says more than was checked.
func TestBundleReviewFindings(t *testing.T) {
	ctx := context.Background()
	root, _, data, commit := signedWarranty(t)
	file := "warranty-" + commit[:12] + ".json"
	sum := sha256.Sum256(data)
	digest := hex.EncodeToString(sum[:])
	base := strings.TrimSpace(uiGit(t, root, "rev-list", "--max-parents=0", "HEAD"))
	head := strings.TrimSpace(uiGit(t, root, "rev-parse", "HEAD"))
	newDir := func(entries []bundleEntry, files map[string][]byte, b, h string) string {
		dir := t.TempDir()
		writeIndex(t, dir, bundleIndex{Schema: bundleSchema, Base: b, Head: h, Entries: entries})
		for n, d := range files {
			if err := os.WriteFile(filepath.Join(dir, n), d, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return dir
	}
	t.Run("F1 an invalid file's contents are not shown", func(t *testing.T) {
		forged := []byte(strings.Replace(string(data), "System MUST do A.", "System MUST do A. Audited by Big4.", 1))
		fs := sha256.Sum256(forged)
		dir := newDir([]bundleEntry{{Commit: commit, State: bundleValid, File: file, FileSHA256: hex.EncodeToString(fs[:])}}, map[string][]byte{file: forged}, base, head)
		v, err := loadBundle(ctx, dir, "", "", time.Now())
		if err != nil {
			t.Fatal(err)
		}
		w := row(t, v, file)
		if w.Result != bundleInvalid || w.Signed || len(w.Promises) != 0 || w.Proof != nil || !strings.Contains(w.Problem, "edited or forged") {
			t.Errorf("forged: %+v", w)
		}
		s := newUIServer("", "", "", 4242)
		s.bundle = &v
		if page := uiRequest(s, "GET", "/", "127.0.0.1:4242", "", s.token, "").Body.String(); strings.Contains(page, "Big4") || !strings.Contains(page, "commit from the unsigned index") {
			t.Error("the page shows a forged warranty's contents, or does not label its commit as unsigned")
		}
		other := ed25519.NewKeyFromSeed(bytes32(9))
		dir = newDir([]bundleEntry{{Commit: commit, State: bundleValid, File: file, FileSHA256: digest}}, map[string][]byte{file: data}, base, head)
		v, _ = loadBundle(ctx, dir, pbwarranty.Fingerprint(other.Public().(ed25519.PublicKey)), "", time.Now())
		if w := row(t, v, file); w.Result != bundleInvalid || len(w.Promises) != 0 {
			t.Errorf("wrong trust key still shows contents: %+v", w)
		}
	})
	t.Run("F2 the index, the file name and the signed commit must agree; no duplicates", func(t *testing.T) {
		otherCommit := strings.Repeat("b", 40)
		dir := newDir([]bundleEntry{{Commit: otherCommit, State: bundleValid, File: file, FileSHA256: digest}}, map[string][]byte{file: data}, base, head)
		v, _ := loadBundle(ctx, dir, "", "", time.Now())
		if w := row(t, v, file); w.Result != bundleAltered || !strings.Contains(w.Problem, "the index lists it for commit bbbbbbbbbbbb") {
			t.Errorf("index commit mismatch: %+v", w)
		}
		dir = newDir([]bundleEntry{{Commit: commit, State: bundleValid, File: file, FileSHA256: digest}, {Commit: otherCommit, State: bundleValid, File: file, FileSHA256: digest}, {Commit: commit, State: bundleValid, File: file, FileSHA256: digest}}, map[string][]byte{file: data}, base, head)
		v, _ = loadBundle(ctx, dir, "", "", time.Now())
		if len(v.Warranties) != 3 || v.Warranties[0].Result != bundleSignatureOnly || v.Warranties[1].Result != bundleAltered || !strings.Contains(v.Warranties[1].Problem, "second commit") ||
			v.Warranties[2].Result != bundleAltered || !strings.Contains(v.Warranties[2].Problem, "twice") {
			t.Errorf("duplicates: %+v", v.Warranties)
		}
		for _, c := range v.Counts {
			if c.Result == bundleSignatureOnly && c.N != 1 {
				t.Errorf("a duplicated file counted %d times", c.N)
			}
		}
		// A correctly named extra file is listed as not in the index, and checked.
		dir = newDir(nil, map[string][]byte{file: data}, base, head)
		v, _ = loadBundle(ctx, dir, "", "", time.Now())
		if w := row(t, v, file); w.InIndex || w.Result != bundleSignatureOnly {
			t.Errorf("extra file: %+v", w)
		}
	})
	t.Run("F3 with a clone, a warranty outside the index's range is not shown as in it", func(t *testing.T) {
		dir := newDir([]bundleEntry{{Commit: commit, State: bundleValid, File: file, FileSHA256: digest}}, map[string][]byte{file: data}, head, head)
		v, _ := loadBundle(ctx, dir, "", root, time.Now())
		if w := row(t, v, file); w.Result != bundleAltered || !strings.Contains(w.Problem, "not in the index's range") {
			t.Errorf("out of range: %+v", w)
		}
		dir = newDir([]bundleEntry{{Commit: commit, State: bundleValid, File: file, FileSHA256: digest}}, map[string][]byte{file: data}, base, head)
		v, _ = loadBundle(ctx, dir, "", root, time.Now())
		if w := row(t, v, file); w.Result != bundleValid {
			t.Errorf("in range: %+v", w)
		}
		s := newUIServer("", "", "", 4242)
		s.bundle = &v
		if page := uiRequest(s, "GET", "/", "127.0.0.1:4242", "", s.token, "").Body.String(); !strings.Contains(page, "from the unsigned index</span> · each warranty's commit is checked to be in it") {
			t.Error("the range is not labelled as unsigned")
		}
	})
}

func bytes32(b byte) []byte {
	out := make([]byte, 32)
	for i := range out {
		out[i] = b
	}
	return out
}
