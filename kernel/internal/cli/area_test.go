package cli

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	connectorreviews "github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// CLI-INV-52: the scope.paths grammar; any invalid line makes the whole file invalid.
func TestAreaPatterns(t *testing.T) {
	match := []struct {
		pattern, path string
		want          bool
	}{
		{"src/cart/", "src/cart/a.go", true},
		{"src/cart/", "src/cart/deep/b.go", true},
		{"src/cart/", "src/cartx/a.go", false},
		{"src/cart/", "src/cart", false},
		{"src/*.go", "src/a.go", true},
		{"src/*.go", "src/x/a.go", false},
		{"src/**/*_test.go", "src/a_test.go", true},
		{"src/**/*_test.go", "src/x/y/a_test.go", true},
		{"src/**/*_test.go", "src/x/a.go", false},
		{"**", "anything/at/all", true},
		{"README.md", "README.md", true},
		{"README.md", "docs/README.md", false},
		{"a*b*c", "aXXbYYc", true},
		{"a*b*c", "aXXcYYb", false},
	}
	for _, m := range match {
		if got := areaMatch(m.pattern, m.path); got != m.want {
			t.Errorf("areaMatch(%q, %q) = %v, want %v", m.pattern, m.path, got, m.want)
		}
	}
	got, err := parseArea([]byte("# area\n\nsrc/cart/\r\n  src/**/*_test.go  \n"))
	if err != nil || !reflect.DeepEqual(got, []string{"src/cart/", "src/**/*_test.go"}) {
		t.Errorf("parseArea = %q, %v", got, err)
	}
	for _, bad := range []string{"/abs", "a/../b", `a\b`, "a//b", "./a", "a/**b", "a\x01b", "src/\n..\n"} {
		if _, err := parseArea([]byte("src/\n" + bad + "\n")); err == nil {
			t.Errorf("parseArea accepted %q", bad)
		}
	}
	if _, err := parseArea([]byte{0xff}); err == nil {
		t.Error("parseArea accepted invalid UTF-8")
	}
}

// areaRepo is a repository whose commit `claim` claims sdd:cart, whose area (declared in the parent)
// is src/cart/ and test files. The claim changes an in-area file, a test file, two separate lines of
// src/other.go, a doc, deletes src/gone.go, and widens its own area (which must not count).
func areaRepo(t *testing.T, reviewerPub ed25519.PublicKey) (root, base, claim string) {
	t.Helper()
	root = taskRepo(t)
	if reviewerPub != nil {
		uiWrite(t, root, ".proofbound/keys/proofbound-reviewer.pub", pbwarranty.PubLine(reviewerPub))
	}
	uiWrite(t, root, "README.md", "start\n")
	base = uiCommit(t, root, "owner", "start")
	var lines []string
	for i := 1; i <= 30; i++ {
		lines = append(lines, "line")
	}
	uiWrite(t, root, "src/other.go", strings.Join(lines, "\n")+"\n")
	uiWrite(t, root, "src/gone.go", "gone\n")
	uiWrite(t, root, "openspec/changes/cart/proposal.md", "# cart\n")
	uiWrite(t, root, "openspec/changes/cart/scope.paths", "# cart\nsrc/cart/\nsrc/**/*_test.go\n")
	uiCommit(t, root, "owner", "spec")
	uiWrite(t, root, "src/cart/a.go", "a\n")
	uiWrite(t, root, "src/x/y_test.go", "t\n")
	lines[1], lines[19] = "changed", "changed"
	uiWrite(t, root, "src/other.go", strings.Join(lines, "\n")+"\n")
	uiWrite(t, root, "docs/n.md", "note\n")
	uiWrite(t, root, "openspec/changes/cart/scope.paths", "src/\n")
	if err := os.Remove(filepath.Join(root, "src", "gone.go")); err != nil {
		t.Fatal(err)
	}
	claim = uiCommit(t, root, "agent", "cart\n\nIntent: sdd:cart")
	return root, base, claim
}

var areaWant = []areaHunk{{Path: "src/gone.go", Lines: "file deleted"}, {Path: "src/other.go", Lines: "line 2"}, {Path: "src/other.go", Lines: "line 20"}}

// CLI-INV-53: exactly the out-of-area lines, from the parent's area; not declared; invalid.
func TestOutOfArea(t *testing.T) {
	ctx := context.Background()
	root, _, claim := areaRepo(t, nil)
	res, err := commitArea(ctx, root, claim)
	if err != nil || !res.Declared || !reflect.DeepEqual(res.Hunks, areaWant) {
		t.Fatalf("commitArea = %+v, %v", res, err)
	}
	if !reflect.DeepEqual(res.Files, []string{"openspec/changes/cart/scope.paths"}) {
		t.Errorf("files = %q", res.Files)
	}
	// The widened area now counts for the next claim, which changes only src/other.go: inside.
	uiWrite(t, root, "src/other.go", "all new\n")
	next := uiCommit(t, root, "agent", "more\n\nIntent: sdd:cart")
	if res, err := commitArea(ctx, root, next); err != nil || !res.Declared || len(res.Hunks) != 0 {
		t.Errorf("after widening: %+v, %v", res, err)
	}
	// A pure deletion of lines in a file outside the area, and a new file of three lines.
	uiWrite(t, root, "openspec/changes/cart/scope.paths", "src/cart/\n")
	uiCommit(t, root, "owner", "narrow")
	uiWrite(t, root, "src/other.go", "")
	uiWrite(t, root, "lib/new.go", "1\n2\n3\n")
	third := uiCommit(t, root, "agent", "x\n\nIntent: sdd:cart")
	res, err = commitArea(ctx, root, third)
	want := []areaHunk{{Path: "lib/new.go", Lines: "lines 1-3"}, {Path: "src/other.go", Lines: "removed line 1"}}
	if err != nil || !reflect.DeepEqual(res.Hunks, want) {
		t.Errorf("third = %+v, %v", res.Hunks, err)
	}
	// Not declared.
	uiWrite(t, root, "specs/nope/spec.md", "# nope\n")
	uiCommit(t, root, "owner", "nope spec")
	uiWrite(t, root, "lib/z.go", "z\n")
	nope := uiCommit(t, root, "agent", "z\n\nIntent: sdd:nope")
	if res, err := commitArea(ctx, root, nope); err != nil || res.Declared || !reflect.DeepEqual(res.Missing, []string{"nope"}) || len(res.Hunks) != 0 {
		t.Errorf("not declared: %+v, %v", res, err)
	}
	// Invalid.
	uiWrite(t, root, "specs/nope/scope.paths", "lib/\n../etc\n")
	uiCommit(t, root, "owner", "bad area")
	uiWrite(t, root, "lib/z.go", "zz\n")
	bad := uiCommit(t, root, "agent", "zz\n\nIntent: sdd:nope")
	if res, err := commitArea(ctx, root, bad); err != nil || !strings.Contains(res.Invalid, "specs/nope/scope.paths is invalid: line 2") {
		t.Errorf("invalid: %+v, %v", res, err)
	}
	// A commit that claims nothing has no area at all.
	if res, err := commitArea(ctx, root, uiCommit(t, root, "owner", "plain")); err != nil || len(res.Names) != 0 {
		t.Errorf("unclaimed: %+v, %v", res, err)
	}
}

func areaTruth(claim string) func(context.Context, string) (draftFile, error) {
	return func(context.Context, string) (draftFile, error) {
		d := sealDraft("", "")
		d.Commit, d.DeclaredReviewer = claim, ""
		for i := range d.Verdicts[0].Obligations {
			d.Verdicts[0].Obligations[i].EvidenceRefs = []string{connectorreviews.CommitEvidenceRef(claim)}
		}
		return d, nil
	}
}

// CLI-INV-54: seal refuses until every out-of-area range is decided; reject fails the review; accept
// writes a signed AREA- record.
func TestSealAccountsForArea(t *testing.T) {
	ctx := context.Background()
	setup := func(t *testing.T) (string, string, ed25519.PrivateKey) {
		t.Helper()
		keyPath := filepath.Join(t.TempDir(), "reviewer-key")
		t.Setenv("PROOFBOUND_KEY", keyPath)
		pub, err := pbwarranty.InitKey(keyPath)
		if err != nil {
			t.Fatal(err)
		}
		key, err := pbwarranty.LoadKey(keyPath)
		if err != nil {
			t.Fatal(err)
		}
		root, _, claim := areaRepo(t, pub)
		return root, claim, key
	}
	seal := func(t *testing.T, root, claim string, area []areaDecision) (string, error) {
		t.Helper()
		d := sealDraft("SATISFIED", "VERIFIABLE")
		d.Commit, d.OutOfArea = claim, area
		var out bytes.Buffer
		err := reviewSeal(ctx, root, writeDraft(t, root, d), areaTruth(claim), &out)
		return out.String(), err
	}
	nothing := func(t *testing.T, root, claim string) {
		t.Helper()
		for _, rel := range []string{"gates/ready-" + claim[:12] + ".yaml", areaRecordPath(claim), "docs/verification/verdicts/OV-" + claim[:12] + "-1.md"} {
			if _, err := os.Lstat(filepath.Join(root, rel)); err == nil {
				t.Fatalf("%s was written", rel)
			}
		}
	}
	decide := func(decision, reason string) []areaDecision {
		var out []areaDecision
		for _, h := range areaWant {
			out = append(out, areaDecision{Path: h.Path, Lines: h.Lines, Decision: decision, Reason: reason})
		}
		return out
	}

	t.Run("a draft without decisions is refused", func(t *testing.T) {
		root, claim, _ := setup(t)
		if _, err := seal(t, root, claim, nil); err == nil || !strings.Contains(err.Error(), "src/other.go line 20") {
			t.Fatalf("sealed with undecided lines: %v", err)
		}
		nothing(t, root, claim)
	})
	t.Run("one undecided range is refused", func(t *testing.T) {
		root, claim, _ := setup(t)
		area := decide("accept", "needed")
		area[1].Decision = ""
		if _, err := seal(t, root, claim, area); err == nil || !strings.Contains(err.Error(), "src/other.go line 2") {
			t.Fatalf("sealed with an undecided line: %v", err)
		}
		nothing(t, root, claim)
	})
	t.Run("accept needs a reason", func(t *testing.T) {
		root, claim, _ := setup(t)
		if _, err := seal(t, root, claim, decide("accept", " ")); err == nil || !strings.Contains(err.Error(), "needs a one-line reason") {
			t.Fatalf("accepted without a reason: %v", err)
		}
		nothing(t, root, claim)
	})
	t.Run("accept writes a signed record", func(t *testing.T) {
		root, claim, key := setup(t)
		out, err := seal(t, root, claim, decide("accept", "shared helper"))
		if err != nil || !strings.Contains(out, "wrote "+areaRecordPath(claim)) {
			t.Fatalf("%v\n%s", err, out)
		}
		ov, _ := os.ReadFile(filepath.Join(root, "docs/verification/verdicts/OV-"+claim[:12]+"-1.md"))
		if !strings.Contains(string(ov), `"status":"ACCEPTABLE"`) {
			t.Errorf("OV = %s", ov)
		}
		uiCommit(t, root, "proofbound-reviewer", "review")
		by, decisions, err := readAreaRecord(ctx, root, claim)
		if err != nil || by != "proofbound-reviewer" || !reflect.DeepEqual(decisions, decide("accept", "shared helper")) {
			t.Fatalf("read back: %q %+v %v", by, decisions, err)
		}
		raw, _ := os.ReadFile(filepath.Join(root, areaRecordPath(claim)))
		if _, err := verifyRecord(raw, areaSchema, map[string]ed25519.PublicKey{"proofbound-reviewer": key.Public().(ed25519.PublicKey)}); err != nil {
			t.Errorf("the record does not verify: %v", err)
		}
	})
	t.Run("reject fails the review", func(t *testing.T) {
		root, claim, _ := setup(t)
		area := decide("accept", "fine")
		area[2] = areaDecision{Path: area[2].Path, Lines: area[2].Lines, Decision: "reject"}
		if _, err := seal(t, root, claim, area); err != nil {
			t.Fatal(err)
		}
		ov, _ := os.ReadFile(filepath.Join(root, "docs/verification/verdicts/OV-"+claim[:12]+"-1.md"))
		if !strings.Contains(string(ov), `"status":"NEEDS_WORK"`) {
			t.Errorf("a rejected range left the review acceptable: %s", ov)
		}
	})
	t.Run("an invalid area is refused", func(t *testing.T) {
		root, _, _ := setup(t)
		uiWrite(t, root, "openspec/changes/cart/scope.paths", "/abs\n")
		uiCommit(t, root, "owner", "bad")
		uiWrite(t, root, "lib/q.go", "q\n")
		bad := uiCommit(t, root, "agent", "q\n\nIntent: sdd:cart")
		d := sealDraft("SATISFIED", "VERIFIABLE")
		d.Commit = bad
		var out bytes.Buffer
		if err := reviewSeal(ctx, root, writeDraft(t, root, d), areaTruth(bad), &out); err == nil || !strings.Contains(err.Error(), "is invalid") {
			t.Fatalf("sealed with an invalid area: %v", err)
		}
	})
}

// CLI-INV-55: audit control C10.
func TestAuditC10(t *testing.T) {
	ctx := context.Background()
	key := uiKey(t)
	pub := key.Public().(ed25519.PublicKey)
	control := func(root, base string, commits []auditCommit) (auditControl, []auditException) {
		t.Helper()
		rep := auditReport{Commits: commits, Roles: []auditRole{}}
		var ex []auditException
		c, err := areaControl(ctx, root, base, &rep, func(kind, ctl, commit, text string) {
			ex = append(ex, auditException{Kind: kind, Control: ctl, Commit: commit, Text: text})
		})
		if err != nil {
			t.Fatal(err)
		}
		return c, ex
	}
	record := func(t *testing.T, root, claim, by string, k ed25519.PrivateKey) {
		t.Helper()
		var ds []areaDecision
		for _, h := range areaWant {
			ds = append(ds, areaDecision{Path: h.Path, Lines: h.Lines, Decision: "accept", Reason: "shared helper"})
		}
		data, err := areaRecordBytes(claim, by, ds, k, "2026-11-23T00:00:00Z")
		if err != nil {
			t.Fatal(err)
		}
		uiWrite(t, root, areaRecordPath(claim), string(data))
		uiCommit(t, root, by, "area review")
	}

	t.Run("undecided lines are an exception", func(t *testing.T) {
		root, base, claim := areaRepo(t, pub)
		c, ex := control(root, base, []auditCommit{{Commit: claim, Class: classClaimed, Author: "agent", Subject: "cart"}})
		if c.State != auditExcepted || len(ex) != 1 || ex[0].Kind != "area" || !strings.Contains(ex[0].Text, "src/other.go line 20") {
			t.Fatalf("%+v %+v", c, ex)
		}
	})
	t.Run("accepted by an independent reviewer holds", func(t *testing.T) {
		root, base, claim := areaRepo(t, pub)
		record(t, root, claim, "proofbound-reviewer", key)
		c, ex := control(root, base, []auditCommit{{Commit: claim, Class: classClaimed, Author: "agent", Subject: "cart"}})
		if c.State != auditHeld || len(ex) != 0 || !strings.Contains(c.Evidence[0].Text, "accepted by proofbound-reviewer: shared helper") {
			t.Fatalf("%+v %+v", c, ex)
		}
	})
	t.Run("accepted by the change's own author does not count", func(t *testing.T) {
		root, base, claim := areaRepo(t, pub)
		record(t, root, claim, "proofbound-reviewer", key)
		c, ex := control(root, base, []auditCommit{{Commit: claim, Class: classClaimed, Author: "proofbound-reviewer", Subject: "cart"}})
		if c.State != auditExcepted || len(ex) != 1 || !strings.Contains(ex[0].Text, "not independent") {
			t.Fatalf("%+v %+v", c, ex)
		}
	})
	t.Run("a record by an undeclared key does not count", func(t *testing.T) {
		root, base, claim := areaRepo(t, pub)
		record(t, root, claim, "proofbound-reviewer", ed25519.NewKeyFromSeed(make([]byte, 32)))
		if c, ex := control(root, base, []auditCommit{{Commit: claim, Class: classClaimed, Author: "agent"}}); c.State != auditExcepted || !strings.Contains(ex[0].Text, "does not verify") {
			t.Fatalf("%+v %+v", c, ex)
		}
	})
	t.Run("an area last set by an agent is an exception", func(t *testing.T) {
		root := taskRepo(t)
		uiWrite(t, root, "README.md", "x\n")
		base := uiCommit(t, root, "owner", "start")
		uiWrite(t, root, "specs/cart/spec.md", "# cart\n")
		uiWrite(t, root, "specs/cart/scope.paths", "**\n")
		setter := uiCommit(t, root, "claude", "spec")
		uiWrite(t, root, "lib/a.go", "a\n")
		claim := uiCommit(t, root, "claude", "a\n\nIntent: sdd:cart")
		c, ex := control(root, base, []auditCommit{{Commit: setter, Class: classHousekeeping, Agents: []string{"claude"}}, {Commit: claim, Class: classClaimed, Agents: []string{"claude"}}})
		if c.State != auditExcepted || len(ex) != 1 || !strings.Contains(ex[0].Text, "last changed by "+setter[:12]) {
			t.Fatalf("%+v %+v", c, ex)
		}
	})
	t.Run("nothing declared is NONE, not an exception", func(t *testing.T) {
		root := taskRepo(t)
		base := uiCommit(t, root, "owner", "start")
		uiWrite(t, root, "specs/cart/spec.md", "# cart\n")
		uiCommit(t, root, "owner", "spec")
		uiWrite(t, root, "lib/a.go", "a\n")
		claim := uiCommit(t, root, "agent", "a\n\nIntent: sdd:cart")
		c, ex := control(root, base, []auditCommit{{Commit: claim, Class: classClaimed}})
		if c.State != auditNone || len(ex) != 0 || !strings.Contains(c.Count, "1 not declared") || !strings.Contains(c.Evidence[0].Text, "area not declared by cart") {
			t.Fatalf("%+v %+v", c, ex)
		}
	})
}

// CLI-INV-54 (Review bench): the bench shows the out-of-area ranges, refuses a preview until each is
// decided, and passes the decisions to seal.
func TestUIAreaDecisions(t *testing.T) {
	key := uiKey(t)
	root, base, claim := areaRepo(t, nil)
	uiWrite(t, root, ".proofbound/keys/demo-reviewer.pub", pbwarranty.PubLine(key.Public().(ed25519.PublicKey)))
	uiCommit(t, root, "owner", "keys")
	truth := draftFile{Commit: claim, Verdicts: []draftVerdict{{Obligations: []draftVerdictObligation{{ObligationID: "O-1", Statement: "one"}}}}, Reviews: []draftReview{}}
	s := newUIServer(root, "", base, 4242)
	s.compute = func(ctx context.Context, b string) (ledgerView, string, error) {
		return ledgerView{Report: auditReport{Schema: auditSchema, Base: base, Head: claim, Commits: []auditCommit{{Commit: claim, Author: "agent", Class: classClaimed}},
			Controls: []auditControl{{ID: "C1", Evidence: []auditEvidence{{Commit: claim, Text: "NO REVIEW: x"}}}}, Exceptions: []auditException{}}}, "", nil
	}
	s.drafts = func(ctx context.Context, shas []string) (map[string]draftFile, error) {
		return map[string]draftFile{claim: truth}, nil
	}
	var sealed draftFile
	s.seal = func(ctx context.Context, dir string, _ *sealTestsRun, out io.Writer) error {
		raw, err := os.ReadFile(filepath.Join(dir, "draft.json"))
		if err != nil {
			return err
		}
		return json.Unmarshal(raw, &sealed)
	}
	page := uiRequest(s, "GET", "/review?commit="+claim, "127.0.0.1:4242", "", s.token, "").Body.String()
	if !strings.Contains(page, `data-part="area"`) || !strings.Contains(page, `data-lines="line 20"`) || !strings.Contains(page, "3 line range(s) are outside") {
		t.Fatalf("the bench does not show the out-of-area ranges:\n%s", page)
	}
	act := func(path string, decisions []reviewDecision) (int, map[string]any) {
		raw, _ := json.Marshal(reviewRequest{Commit: claim, Decisions: decisions})
		w := uiRequest(s, "POST", path, "127.0.0.1:4242", "http://127.0.0.1:4242", s.token, string(raw))
		var out map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return w.Code, out
	}
	promise := reviewDecision{Part: "verdict", Group: 0, Index: 0, ObligationID: "O-1", Outcome: "SATISFIED"}
	if code, out := act("/act/review-preview", []reviewDecision{promise}); code != http.StatusUnprocessableEntity || !strings.Contains(fmt.Sprint(out["missing"]), "outside the area") {
		t.Errorf("undecided ranges previewed: %d %v", code, out)
	}
	if code, _ := act("/act/review-preview", []reviewDecision{promise, {Part: "area", Index: 0, Path: "src/other.go", Lines: "line 2", Outcome: "accept", Finding: "x"}}); code != http.StatusUnprocessableEntity {
		t.Errorf("a decision for another range was accepted: %d", code)
	}
	all := []reviewDecision{promise}
	for i, h := range areaWant {
		all = append(all, reviewDecision{Part: "area", Index: i, Path: h.Path, Lines: h.Lines, Outcome: "accept", Finding: "shared helper"})
	}
	if code, out := act("/act/review-preview", all); code != 200 || !strings.Contains(fmt.Sprint(out["files"]), areaRecordPath(claim)) {
		t.Errorf("preview: %d %v", code, out)
	}
	act("/act/review-seal", all)
	if len(sealed.OutOfArea) != 3 || sealed.OutOfArea[2].Decision != "accept" || sealed.OutOfArea[2].Reason != "shared helper" {
		t.Errorf("seal got %+v", sealed.OutOfArea)
	}
}

// Review change-area-review-20261123 F1..F5: each way around the area found by the review.
func TestAreaReviewFindings(t *testing.T) {
	ctx := context.Background()
	t.Run("F1 a submodule bump outside the area is listed despite ignore=all", func(t *testing.T) {
		root := taskRepo(t)
		uiWrite(t, root, "specs/cart/spec.md", "# cart\n")
		uiWrite(t, root, "specs/cart/scope.paths", "src/\n")
		uiWrite(t, root, ".gitmodules", "[submodule \"vendor/lib\"]\n\tpath = vendor/lib\n\turl = ./x\n\tignore = all\n")
		uiCommit(t, root, "owner", "spec")
		uiGit(t, root, "update-index", "--add", "--cacheinfo", "160000,1111111111111111111111111111111111111111,vendor/lib")
		// Commit the index as is: `add -A` would drop a gitlink that has no checkout.
		uiGit(t, root, "-c", "user.name=agent", "-c", "user.email=agent@example.com", "commit", "-q", "--no-verify", "-m", "bump\n\nIntent: sdd:cart")
		claim := uiGit(t, root, "rev-parse", "HEAD")
		res, err := commitArea(ctx, root, claim)
		if err != nil || len(res.Hunks) != 1 || res.Hunks[0].Path != "vendor/lib" {
			t.Fatalf("a submodule bump is hidden: %+v %v", res, err)
		}
	})
	t.Run("F3 an archived change with a longer name is not this change's area", func(t *testing.T) {
		root := taskRepo(t)
		uiWrite(t, root, "openspec/changes/cart/proposal.md", "# cart\n")
		uiWrite(t, root, "openspec/changes/cart/scope.paths", "src/cart/\n")
		uiWrite(t, root, "openspec/changes/archive/2026-01-01-add-cart/scope.paths", "**\n")
		uiWrite(t, root, "openspec/changes/archive/2026-01-02-cart/scope.paths", "lib/\n")
		uiCommit(t, root, "owner", "spec")
		uiWrite(t, root, "other/x.go", "x\n")
		uiWrite(t, root, "lib/y.go", "y\n")
		claim := uiCommit(t, root, "agent", "x\n\nIntent: sdd:cart")
		res, err := commitArea(ctx, root, claim)
		if err != nil || !reflect.DeepEqual(res.Hunks, []areaHunk{{Path: "other/x.go", Lines: "line 1"}}) {
			t.Fatalf("hunks = %+v %v (files %q)", res.Hunks, err, res.Files)
		}
	})
	t.Run("F4 the list is capped and the bench can send every decision", func(t *testing.T) {
		root := taskRepo(t)
		uiWrite(t, root, "specs/cart/spec.md", "# cart\n")
		uiWrite(t, root, "specs/cart/scope.paths", "src/\n")
		uiCommit(t, root, "owner", "spec")
		for i := 0; i < 250; i++ {
			uiWrite(t, root, fmt.Sprintf("lib/f%03d.go", i), "x\n")
		}
		claim := uiCommit(t, root, "agent", "many\n\nIntent: sdd:cart")
		res, err := commitArea(ctx, root, claim)
		if err != nil || len(res.Hunks) != areaMaxTotal || res.Hunks[areaMaxTotal-1].Lines != "and 51 more hunks in other files" {
			t.Fatalf("cap: %d %v %+v", len(res.Hunks), err, res.Hunks[len(res.Hunks)-1])
		}
		key := uiKey(t)
		uiWrite(t, root, ".proofbound/keys/demo-reviewer.pub", pbwarranty.PubLine(key.Public().(ed25519.PublicKey)))
		uiCommit(t, root, "owner", "keys")
		s := newUIServer(root, "", claim+"^", 4242)
		s.compute = func(ctx context.Context, b string) (ledgerView, string, error) {
			return ledgerView{Report: auditReport{Schema: auditSchema, Head: claim, Commits: []auditCommit{{Commit: claim, Author: "agent", Class: classClaimed}},
				Controls: []auditControl{{ID: "C1", Evidence: []auditEvidence{{Commit: claim, Text: "NO REVIEW: x"}}}}, Exceptions: []auditException{}}}, "", nil
		}
		s.drafts = func(ctx context.Context, shas []string) (map[string]draftFile, error) {
			return map[string]draftFile{claim: {Commit: claim, Reviews: []draftReview{}, Verdicts: []draftVerdict{{Obligations: []draftVerdictObligation{{ObligationID: "O-1"}}}}}}, nil
		}
		all := []reviewDecision{{Part: "verdict", Group: 0, Index: 0, ObligationID: "O-1", Outcome: "SATISFIED"}}
		for i, h := range res.Hunks {
			all = append(all, reviewDecision{Part: "area", Index: i, Path: h.Path, Lines: h.Lines, Outcome: "accept", Finding: strings.Repeat("r", 500)})
		}
		raw, _ := json.Marshal(reviewRequest{Commit: claim, Decisions: all})
		w := uiRequest(s, "POST", "/act/review-preview", "127.0.0.1:4242", "http://127.0.0.1:4242", s.token, string(raw))
		if w.Code != http.StatusOK {
			t.Fatalf("a full list of decisions (%d bytes) was refused: %d %s", len(raw), w.Code, w.Body.String())
		}
	})
	agentRepo := func(t *testing.T) (string, string) {
		root := taskRepo(t)
		uiWrite(t, root, "README.md", "x\n")
		base := uiCommit(t, root, "owner", "start")
		uiWrite(t, root, "specs/a/spec.md", "# a\n")
		uiWrite(t, root, "specs/a/scope.paths", "src/a/\n")
		uiWrite(t, root, "specs/b/spec.md", "# b\n")
		uiWrite(t, root, "specs/b/scope.paths", "src/b/\n")
		uiCommit(t, root, "owner", "specs")
		return root, base
	}
	c10 := func(t *testing.T, root, base string, commits []auditCommit) (auditControl, []auditException) {
		rep := auditReport{Commits: commits, Roles: []auditRole{}}
		var ex []auditException
		c, err := areaControl(ctx, root, base, &rep, func(kind, ctl, commit, text string) {
			ex = append(ex, auditException{Kind: kind, Commit: commit, Text: text})
		})
		if err != nil {
			t.Fatal(err)
		}
		return c, ex
	}
	t.Run("F2 a later human edit of one area file does not hide an agent's widening of another", func(t *testing.T) {
		root, base := agentRepo(t)
		uiWrite(t, root, "specs/a/scope.paths", "**\n")
		widen := uiCommit(t, root, "claude", "widen a")
		uiWrite(t, root, "specs/b/scope.paths", "src/b/\nsrc/c/\n")
		human := uiCommit(t, root, "owner", "edit b")
		uiWrite(t, root, "lib/x.go", "x\n")
		claim := uiCommit(t, root, "claude", "x\n\nIntent: sdd:a\nIntent: sdd:b")
		c, ex := c10(t, root, base, []auditCommit{{Commit: widen, Agents: []string{"claude"}}, {Commit: human}, {Commit: claim, Class: classClaimed, Agents: []string{"claude"}}})
		if c.State != auditExcepted || len(ex) != 1 || !strings.Contains(ex[0].Text, widen[:12]) {
			t.Fatalf("%+v %+v", c, ex)
		}
	})
	t.Run("F5 an agent deleting the area does not make its change merely not declared", func(t *testing.T) {
		root, base := agentRepo(t)
		uiGit(t, root, "rm", "-q", "specs/a/scope.paths")
		del := uiCommit(t, root, "claude", "drop area")
		uiWrite(t, root, "lib/x.go", "x\n")
		claim := uiCommit(t, root, "claude", "x\n\nIntent: sdd:a")
		c, ex := c10(t, root, base, []auditCommit{{Commit: del, Agents: []string{"claude"}}, {Commit: claim, Class: classClaimed, Agents: []string{"claude"}}})
		if c.State != auditExcepted || len(ex) != 1 || !strings.Contains(ex[0].Text, del[:12]) {
			t.Fatalf("%+v %+v", c, ex)
		}
	})
}
