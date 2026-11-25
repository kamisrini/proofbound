package cli

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// ui-reviewer-spec R2.1: one change, its commits and the viewer's one action.
func TestChangePage(t *testing.T) {
	key := uiKey(t)
	root := taskRepo(t)
	uiWrite(t, root, ".proofbound/keys/demo-reviewer.pub", pbwarranty.PubLine(key.Public().(ed25519.PublicKey)))
	base := uiCommit(t, root, "owner", "keys")
	uiWrite(t, root, "src/a.sh", "a\n")
	blocked := uiCommit(t, root, "Claude", "first try\n\nIntent: sdd:001-x")
	uiWrite(t, root, "src/a.sh", "b\n")
	waiting := uiCommit(t, root, "Claude", "second try\n\nIntent: sdd:001-x")
	s := newUIServer(root, "", base, 4242)
	gate := map[string]string{blocked: "BLOCKED: the reviewer recorded NEEDS_WORK (OV-1)", waiting: "NO REVIEW: x"}
	author, reviewers, uncovered := "Claude", []string{"demo-reviewer"}, false
	s.compute = func(ctx context.Context, b string) (ledgerView, string, error) {
		ev := []auditEvidence{{Commit: blocked, Text: gate[blocked]}, {Commit: waiting, Text: gate[waiting]}}
		exc := []auditException{}
		if uncovered {
			exc = append(exc, auditException{ID: "uncovered-" + base[:12], Kind: "uncovered", Control: "C2", Commit: base, Text: `"x" changed code with no spec claim: src/b.sh`})
		}
		return ledgerView{Report: auditReport{Schema: auditSchema, Base: base, Head: waiting,
			Commits:  []auditCommit{{Commit: blocked, Author: author, Class: classClaimed}, {Commit: waiting, Author: author, Class: classClaimed}},
			Controls: []auditControl{{ID: "C1", Evidence: ev}}, Exceptions: exc, Roles: []auditRole{{Commit: blocked, Reviewers: reviewers}}}}, "", nil
	}
	get := func(path string) changePage {
		w := uiRequest(s, "GET", path, "127.0.0.1:4242", "", s.token, "")
		if w.Code != 200 {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
		}
		var p changePage
		_ = json.Unmarshal(w.Body.Bytes(), &p)
		return p
	}
	p := get("/change?claim=001-x&format=json")
	if p.One == nil || len(p.One.Commits) != 2 || p.One.Commits[0].State != "BLOCKED" || p.One.Action != "Review this" || p.One.Link != "/review?commit="+waiting {
		t.Fatalf("change: %+v", p.One)
	}
	// Review this is for any waiting commit of the change, not only the last (O-7).
	gate[blocked], gate[waiting] = "NO REVIEW: x", "BLOCKED: the reviewer recorded NEEDS_WORK (OV-2)"
	if p = get("/change?claim=001-x&format=json"); p.One.Action != "Review this" || p.One.Link != "/review?commit="+blocked {
		t.Errorf("an earlier waiting commit: %+v", p.One)
	}
	// The author of the waiting commits is offered nothing.
	author = "demo-reviewer"
	if p = get("/change?claim=001-x&format=json"); p.One.Action != "" {
		t.Errorf("the author was offered %q", p.One.Action)
	}
	author = "Claude"
	// Once the whole period passes: a reviewer of any of its changes is not offered Issue ...
	gate[blocked], gate[waiting] = "PASS: sdd:001-x", "PASS: sdd:001-x"
	if p = get("/change?claim=001-x&format=json"); p.One.Action != "" {
		t.Errorf("a reviewer was offered %q", p.One.Action)
	}
	// ... a key that reviewed nothing is ...
	reviewers = nil
	if p = get("/change?claim=001-x&format=json"); p.One.Action != "Issue the warranty" {
		t.Errorf("a non-reviewing key was not offered Issue: %+v", p.One)
	}
	// ... unless code in the period claims no spec (the check is then NOT WARRANTED).
	uncovered = true
	if p = get("/change?claim=001-x&format=json"); p.One.Action != "" || !strings.Contains(p.One.Note, "changed code with no spec claim") {
		t.Errorf("Issue offered over uncovered code: %+v", p.One)
	}
	if w := uiRequest(s, "GET", "/change?claim=nope", "127.0.0.1:4242", "", s.token, ""); w.Code != 404 {
		t.Errorf("an unknown change: %d", w.Code)
	}
	if w := uiRequest(s, "GET", "/change", "127.0.0.1:4242", "", s.token, ""); w.Code != 200 || !strings.Contains(w.Body.String(), "001-x") {
		t.Errorf("the list: %d", w.Code)
	}
}

// ui-reviewer-spec R2.2: a whole release or nothing; written and committed once per HEAD.
func TestIssueWarranty(t *testing.T) {
	key := uiKey(t)
	root := taskRepo(t)
	uiWrite(t, root, ".proofbound/keys/owner.pub", pbwarranty.PubLine(key.Public().(ed25519.PublicKey)))
	base := uiCommit(t, root, "owner", "keys")
	uiWrite(t, root, "src/a.sh", "a\n")
	head := uiCommit(t, root, "Claude", "work\n\nIntent: sdd:001-x")
	s := newUIServer(root, "", base, 4242)
	uncovered := false
	report := func(h string) auditReport {
		exc := []auditException{}
		if uncovered {
			exc = append(exc, auditException{Kind: "uncovered", Commit: base, Text: "changed code with no spec claim"})
		}
		return auditReport{Schema: auditSchema, Base: base, Head: h, Commits: []auditCommit{},
			Controls: []auditControl{{ID: "C1", Evidence: []auditEvidence{{Commit: head, Text: "PASS: sdd:001-x"}}}}, Exceptions: exc}
	}
	s.compute = func(ctx context.Context, b string) (ledgerView, string, error) {
		return ledgerView{Report: report(head)}, "", nil
	}
	state := bundleNotExportable
	s.issueCompute = func(ctx context.Context, b string) (bundleIndex, map[string][]byte, error) {
		ix := bundleIndex{Base: base, Head: head, Entries: []bundleEntry{{Commit: head, Claims: "sdd:001-x", State: state, Reason: "why"}}}
		return ix, map[string][]byte{"index.json": []byte("{}\n"), "VERIFY.md": []byte("v\n")}, nil
	}
	issue := func(confirm bool) (int, map[string]any) {
		body := `{"base":"","confirm":false}`
		if confirm {
			body = `{"base":"","confirm":true}`
		}
		w := uiRequest(s, "POST", "/act/issue", "127.0.0.1:4242", "http://127.0.0.1:4242", s.token, body)
		var out map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return w.Code, out
	}
	folder := filepath.Join(root, "docs", "warranties", head[:12])
	if code, out := issue(true); code != 422 || out["not_warranted"] == nil {
		t.Fatalf("a partial release: %d %v", code, out)
	}
	if _, err := os.Stat(folder); !os.IsNotExist(err) {
		t.Fatal("a refused release wrote files")
	}
	if w := uiRequest(s, "POST", "/act/issue", "127.0.0.1:4242", "http://evil.example", s.token, `{"confirm":true}`); w.Code != 403 {
		t.Errorf("a cross-site POST: %d", w.Code)
	}
	// Every claiming commit VALID, but code with no spec claim: the check is NOT WARRANTED (F1).
	state, uncovered = bundleValid, true
	if code, out := issue(true); code != 422 || out["not_warranted"] == nil {
		t.Fatalf("issued over uncovered code: %d %v", code, out)
	}
	if _, err := os.Stat(folder); !os.IsNotExist(err) {
		t.Fatal("a refused release wrote files")
	}
	uncovered = false
	// An uncommitted index for HEAD is not an issued release (F3): the tree is dirty, so refused.
	uiWrite(t, root, "docs/warranties/"+head[:12]+"/index.json", "{}\n")
	if code, out := issue(false); code != 409 || out["issued_already"] != nil {
		t.Errorf("an uncommitted index: %d %v", code, out)
	}
	_ = os.RemoveAll(folder)
	if code, out := issue(false); code != 200 || out["committed"] != nil {
		t.Fatalf("preview: %d %v", code, out)
	}
	if _, err := os.Stat(folder); !os.IsNotExist(err) {
		t.Fatal("a preview wrote files")
	}
	if code, out := issue(true); code != 200 || out["committed"] != true {
		t.Fatalf("issue: %d %v", code, out)
	}
	if got := uiGit(t, root, "show", "--name-only", "--format=%s", "HEAD"); got != "release warranties at "+head[:12]+"\n\ndocs/warranties/"+head[:12]+"/VERIFY.md\ndocs/warranties/"+head[:12]+"/index.json" {
		t.Errorf("the release commit holds %q", got)
	}
	newHead := strings.TrimSpace(uiGit(t, root, "rev-parse", "HEAD"))
	s.compute = func(ctx context.Context, b string) (ledgerView, string, error) {
		return ledgerView{Report: report(newHead)}, "", nil
	}
	if code, out := issue(true); code != 200 || out["issued_already"] != "docs/warranties/"+head[:12] {
		t.Errorf("issued twice: %d %v", code, out)
	}
	// A commit that only claims, by its subject, to be a release is not one (F3).
	// It even adds the index, so only the rule that a release adds nothing else refuses it.
	uiWrite(t, root, "src/z.sh", "z\n")
	uiWrite(t, root, "docs/warranties/"+newHead[:12]+"/index.json", "{}\n")
	fake := uiCommit(t, root, "owner", "release warranties at "+newHead[:12])
	if got := releasedAt(context.Background(), root, fake); got != "" {
		t.Errorf("a subject alone counted as a release: %q", got)
	}
}
