package cli

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

func uiGit(t *testing.T, root string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	cmd.Env = repositoryGitEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func uiWrite(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func uiCommit(t *testing.T, root, author, msg string) string {
	t.Helper()
	uiGit(t, root, "add", "-A", "-f")
	uiGit(t, root, "-c", "user.name="+author, "-c", "user.email="+author+"@example.com", "commit", "-q", "--no-verify", "--allow-empty", "-m", msg)
	return uiGit(t, root, "rev-parse", "HEAD")
}

// uiKey makes this test's local signing key.
func uiKey(t *testing.T) ed25519.PrivateKey {
	t.Helper()
	t.Setenv("PROOFBOUND_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path, err := pbwarranty.KeyPath()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pbwarranty.InitKey(path); err != nil {
		t.Fatal(err)
	}
	key, err := pbwarranty.LoadKey(path)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

// CLI-INV-37: `ui` flags, each at most once.
func TestParseUI(t *testing.T) {
	ok := map[string]uiOptions{
		"ui":                               {},
		"ui --open":                        {open: true},
		"ui --base v1 --port 8080":         {base: "v1", port: 8080},
		"ui --port 0 --open --base HEAD~3": {base: "HEAD~3", open: true},
	}
	for in, want := range ok {
		got, valid := parseUI(strings.Fields(in))
		if !valid || got != want {
			t.Errorf("%q = %+v %v, want %+v", in, got, valid, want)
		}
	}
	for _, in := range []string{"ui --port", "ui --port 70000", "ui --port -1", "ui --port 08", "ui --open --open", "ui --base", "ui --base --open", "ui --bundle", "ui --base v1 --bundle x", "ui extra"} {
		if _, valid := parseUI(strings.Fields(in)); valid {
			t.Errorf("%q was accepted", in)
		}
	}
	if parseCommand([]string{"ui", "--open"}) != commandUI || parseCommand([]string{"ui", "--nope"}) != commandInvalid {
		t.Fatal("parseCommand does not route ui")
	}
}

// CLI-INV-38: the git-derived parts of the audit: C6 trust paths, C7 spec order, the headline's
// counting, the empty range and drift parsing. The ledger-derived states reuse `check` and `warranty
// bundle`, which their own rows prove; the whole audit runs on the tour's repository at acceptance.
func TestAuditControls(t *testing.T) {
	ctx := context.Background()
	root := taskRepo(t)
	uiWrite(t, root, "README.md", "x\n")
	uiCommit(t, root, "a", "start")
	uiWrite(t, root, "specs/001-a/spec.md", "spec\n")
	specCommit := uiCommit(t, root, "a", "spec")
	uiWrite(t, root, "src/a.sh", "a\n")
	claimed := uiCommit(t, root, "a", "code\n\nIntent: sdd:001-a")
	uiWrite(t, root, "specs/002-b/spec.md", "spec\n")
	uiWrite(t, root, "src/b.sh", "b\n")
	together := uiCommit(t, root, "a", "spec and code together\n\nIntent: sdd:002-b")
	uiWrite(t, root, "openspec/changes/archive/2026-01-01-003-c/proposal.md", "p\n")
	archived := uiCommit(t, root, "a", "archive")
	uiWrite(t, root, "src/c.sh", "c\n")
	afterArchive := uiCommit(t, root, "a", "code\n\nIntent: sdd:003-c")
	uiWrite(t, root, ".proofbound/keys/r.pub", "ed25519 00\n")
	uiWrite(t, root, "gates/ready-1.yaml", "a\n")
	keyAndNewGate := uiCommit(t, root, "a", "declare a key, add a gate")
	uiWrite(t, root, "gates/ready-2.yaml", "a\n")
	newGate := uiCommit(t, root, "a", "a review adds a gate")
	uiWrite(t, root, "gates/ready-1.yaml", "b\n")
	gateEdit := uiCommit(t, root, "a", "edit a gate")
	uiWrite(t, root, ".github/workflows/proofbound.yml", "on: push\n")
	workflow := uiCommit(t, root, "a", "workflow")
	uiWrite(t, root, ".proofbound/sdd-profile.json", "{}\n")
	profile := uiCommit(t, root, "a", "profile")
	_ = archived

	if !specInParent(ctx, root, claimed, "001-a") || !specInParent(ctx, root, afterArchive, "003-c") {
		t.Error("a spec committed before its code (plain or dated archive) was not found in the parent")
	}
	if specInParent(ctx, root, together, "002-b") || specInParent(ctx, root, specCommit, "001-a") {
		t.Error("a spec committed with or after its code counted as existing before it")
	}
	root0 := uiGit(t, root, "rev-list", "--max-parents=0", "HEAD")
	if specInParent(ctx, root, root0, "001-a") {
		t.Error("a root commit has no parent and must fail")
	}
	want := map[string][]string{
		keyAndNewGate: {".proofbound/keys/r.pub"},
		newGate:       nil,
		gateEdit:      {"gates/ready-1.yaml"},
		workflow:      {".github/workflows/proofbound.yml"},
		profile:       {".proofbound/sdd-profile.json"},
		claimed:       nil,
	}
	for sha, paths := range want {
		got, err := trustPaths(ctx, root, sha)
		if err != nil || strings.Join(got, ",") != strings.Join(paths, ",") {
			t.Errorf("trustPaths(%s) = %v %v, want %v", sha[:7], got, err, paths)
		}
	}

	rep := auditReport{Counts: auditCounts{Total: 3}, Controls: []auditControl{{State: auditHeld}, {State: auditExcepted}, {State: auditListed}, {State: auditNone}}, Exceptions: []auditException{{ID: "a"}, {ID: "b", Outcome: &outcomeView{Verified: true}}, {ID: "c", Outcome: &outcomeView{Verified: false}}}}
	if got := auditHeadline(rep); got != "1 of 2 controls held for this range. 3 exception(s), 1 with a signed outcome." {
		t.Errorf("headline = %q", got)
	}
	if got := auditHeadline(auditReport{}); got != "The range holds no commits." {
		t.Errorf("empty headline = %q", got)
	}
	rows := driftEvidence("drift change=CI-a state=CURRENT verified_commit=aaa verdict=v current_revision=x\ndrift change=CI-b state=STALE verified_commit=bbb verdict=v verified_revision=1 current_revision=2\n  changed O-1 \"x\" -> \"y\"\nsummary warranted=2 current=1 stale=1 other=0\n")
	if len(rows) != 1 || rows[0].Commit != "bbb" || !strings.Contains(rows[0].Text, "CI-b is STALE; changed O-1") {
		t.Errorf("drift rows = %+v", rows)
	}
}

// CLI-INV-39: records verify only when exact, signed under their own tag, by the key declared at the
// range's base, committed at HEAD, about the exception they name, and not by the commit's author.
func TestOutcomeRecords(t *testing.T) {
	ctx := context.Background()
	key := uiKey(t)
	pub := key.Public().(ed25519.PublicKey)
	root := taskRepo(t)
	uiWrite(t, root, ".proofbound/keys/auditor.pub", pbwarranty.PubLine(pub))
	base := uiCommit(t, root, "auditor", "declare the auditor's key")
	uiWrite(t, root, "src/x.sh", "x\n")
	sha := uiCommit(t, root, "agent", "agent change")
	id := "uncovered-" + sha[:12]
	rep := auditReport{Base: base, Head: sha, Exceptions: []auditException{{ID: id, Kind: "uncovered", Commit: sha}}}

	a := localActor(ctx, root)
	if a.Name != "auditor" || a.key == nil {
		t.Fatalf("actor = %+v", a)
	}
	now := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	rel, fields, err := buildAct(ctx, root, rep, nil, a, actRequest{Kind: "outcome", Exception: id, Outcome: "accepted", Reason: "agreed"}, now)
	if err != nil || rel != ".proofbound/outcomes/"+id+".json" {
		t.Fatalf("buildAct = %q %v", rel, err)
	}
	data, err := signRecord(fields, key)
	if err != nil {
		t.Fatal(err)
	}
	keys := declaredKeys(ctx, root, base)
	if _, err := verifyRecord(data, outcomeSchema, keys); err != nil {
		t.Fatalf("a good record does not verify: %v", err)
	}
	altered := []byte(strings.Replace(string(data), "agreed", "agreeD", 1))
	if _, err := verifyRecord(altered, outcomeSchema, keys); err == nil {
		t.Error("an altered record verified")
	}
	if _, err := verifyRecord(append([]byte(" "), data...), outcomeSchema, keys); err == nil {
		t.Error("a non-canonical record verified")
	}
	if _, err := verifyRecord(data, attestationSchema, keys); err == nil {
		t.Error("an outcome verified as an attestation")
	}
	// Same fields and schema, signed under the attestation tag: only the tag separates them.
	unsigned, _ := canonicalJSON(fields)
	crossTag := map[string]any{}
	for k, v := range fields {
		crossTag[k] = v
	}
	crossTag["signature"] = hex.EncodeToString(ed25519.Sign(key, append(recordTag(attestationSchema), unsigned...)))
	forged, _ := canonicalJSON(crossTag)
	if _, err := verifyRecord(append(forged, '\n'), outcomeSchema, keys); err == nil {
		t.Error("a record signed under another schema's tag verified")
	}
	if _, err := verifyRecord(data, outcomeSchema, map[string]ed25519.PublicKey{}); err == nil {
		t.Error("a record verified with no declared key")
	}

	// Written but not committed: closes nothing.
	uiWrite(t, root, rel, string(data))
	attachOutcomes(ctx, root, &rep)
	if rep.Exceptions[0].Outcome != nil {
		t.Fatal("a record in the working tree only closed an exception")
	}
	head := uiCommit(t, root, "auditor", "record")
	rep.Head = head
	attachOutcomes(ctx, root, &rep)
	if o := rep.Exceptions[0].Outcome; o == nil || !o.Verified || o.Outcome != "accepted" {
		t.Fatalf("a committed record did not close its exception: %+v", o)
	}

	// A key declared inside the range cannot sign for it.
	other := taskRepo(t)
	b2 := uiCommit(t, other, "x", "start")
	uiWrite(t, other, ".proofbound/keys/auditor.pub", pbwarranty.PubLine(pub))
	uiWrite(t, other, "src/y.sh", "y\n")
	s2 := uiCommit(t, other, "agent", "key and code together")
	rep2 := auditReport{Base: b2, Head: s2, Exceptions: []auditException{{ID: "uncovered-" + s2[:12], Commit: s2}}}
	if _, _, err := buildAct(ctx, other, rep2, nil, localActor(ctx, other), actRequest{Kind: "outcome", Exception: "uncovered-" + s2[:12], Outcome: "accepted", Reason: "r"}, now); err == nil || !strings.Contains(err.Error(), "not declared before commit") {
		t.Errorf("a key declared in the judged commit could sign: %v", err)
	}

	// The commit's author cannot close it; bad outcomes and reasons are refused.
	uiWrite(t, root, "src/z.sh", "z\n")
	own := uiCommit(t, root, "auditor", "auditor's own change")
	rep.Exceptions = append(rep.Exceptions, auditException{ID: "uncovered-" + own[:12], Commit: own})
	for _, req := range []actRequest{
		{Kind: "outcome", Exception: "uncovered-" + own[:12], Outcome: "accepted", Reason: "mine"},
		{Kind: "outcome", Exception: id, Outcome: "maybe", Reason: "r"},
		{Kind: "outcome", Exception: id, Outcome: "accepted", Reason: ""},
		{Kind: "outcome", Exception: id, Outcome: "accepted", Reason: "a\u202eb"},
		{Kind: "outcome", Exception: id, Outcome: "fixed", Reason: "r", FixedBy: base},
		{Kind: "outcome", Exception: "uncovered-000000000000", Outcome: "accepted", Reason: "r"},
		{Kind: "other"},
	} {
		if _, _, err := buildAct(ctx, root, rep, nil, a, req, now); err == nil {
			t.Errorf("accepted %+v", req)
		}
	}
	if _, _, err := buildAct(ctx, root, rep, nil, a, actRequest{Kind: "outcome", Exception: id, Outcome: "fixed", Reason: "r", FixedBy: own}, now); err != nil {
		t.Errorf("a later fix was refused: %v", err)
	}
	if _, err := signTime("2026-11-22T11:40:00Z", now); err == nil {
		t.Error("a 20-minute-old preview was accepted")
	}
	if _, err := signTime("2026-11-22T11:58:00Z", now); err != nil {
		t.Errorf("a fresh preview was refused: %v", err)
	}
}

func testUIServer(t *testing.T, root string, rep auditReport) *uiServer {
	t.Helper()
	t.Setenv("PROOFBOUND_KEY", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s := newUIServer(root, "", "HEAD~1", 4242)
	s.compute = func(ctx context.Context, base string) (ledgerView, string, error) {
		return ledgerView{Report: rep}, "", nil
	}
	return s
}

func uiRequest(s *uiServer, method, path, host, origin, cookie, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Host = host
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	if cookie != "" {
		r.AddCookie(&http.Cookie{Name: s.cookieName(), Value: cookie})
	}
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	return w
}

// CLI-INV-40: Host, cookie, Origin; the token only via the terminal link; GET writes nothing.
func TestUIRequestGuards(t *testing.T) {
	root := taskRepo(t)
	uiCommit(t, root, "a", "one")
	head := uiCommit(t, root, "a", "two")
	s := testUIServer(t, root, auditReport{Schema: auditSchema, Base: head, Head: head, Headline: "h", Controls: []auditControl{}, Exceptions: []auditException{}})
	host := "127.0.0.1:4242"
	cases := []struct {
		name, method, path, host, origin, cookie string
		want                                     int
	}{
		{"foreign host", "GET", "/audit", "evil.test:4242", "", s.token, http.StatusMisdirectedRequest},
		{"rebound name", "GET", "/audit", "localhost.evil.test:4242", "", s.token, http.StatusMisdirectedRequest},
		{"other port", "GET", "/audit", "127.0.0.1:1", "", s.token, http.StatusMisdirectedRequest},
		{"no cookie", "GET", "/audit", host, "", "", http.StatusUnauthorized},
		{"wrong cookie", "GET", "/audit", host, "", "00", http.StatusUnauthorized},
		{"wrong link", "GET", "/?k=00", host, "", "", http.StatusForbidden},
		{"post without cookie", "POST", "/act/preview", host, "http://" + host, "", http.StatusUnauthorized},
		{"post foreign origin", "POST", "/act/preview", host, "http://evil.test", s.token, http.StatusForbidden},
		{"post no origin", "POST", "/act/preview", host, "", s.token, http.StatusForbidden},
		{"page", "GET", "/audit", host, "", s.token, http.StatusOK},
		{"localhost page", "GET", "/audit", "localhost:4242", "", s.token, http.StatusOK},
		{"static needs no cookie", "GET", "/static/app.css", host, "", "", http.StatusOK},
	}
	for _, c := range cases {
		w := uiRequest(s, c.method, c.path, c.host, c.origin, c.cookie, "{}")
		if w.Code != c.want {
			t.Errorf("%s: %d, want %d", c.name, w.Code, c.want)
		}
		if csp := w.Header().Get("Content-Security-Policy"); !strings.Contains(csp, "default-src 'self'") || !strings.Contains(csp, "frame-ancestors 'none'") {
			t.Errorf("%s: CSP %q", c.name, csp)
		}
	}
	w := uiRequest(s, "GET", "/?k="+s.token, host, "", "", "")
	if w.Code != http.StatusSeeOther || !strings.Contains(w.Header().Get("Set-Cookie"), "SameSite=Strict") || !strings.Contains(w.Header().Get("Set-Cookie"), "HttpOnly") {
		t.Errorf("the terminal link did not set a strict cookie: %d %q", w.Code, w.Header().Get("Set-Cookie"))
	}
	page := uiRequest(s, "GET", "/audit", host, "", s.token, "")
	if strings.Contains(page.Body.String(), s.token) {
		t.Error("the page carries the token")
	}
	if status := uiGit(t, root, "status", "--porcelain"); status != "" {
		t.Errorf("GET requests wrote to the repository:\n%s", status)
	}
	big := uiRequest(s, "POST", "/act/preview", host, "http://"+host, s.token, `{"kind":"outcome","reason":"`+strings.Repeat("x", 20000)+`"}`)
	if big.Code != http.StatusBadRequest {
		t.Errorf("an oversized body: %d", big.Code)
	}
	// A request that arrives while another runs waits for it; past the wait it gets 503.
	s.wait = 50 * time.Millisecond
	s.gate <- struct{}{}
	if busy := uiRequest(s, "GET", "/audit", host, "", s.token, ""); busy.Code != http.StatusServiceUnavailable {
		t.Errorf("a request past the wait: %d, want 503", busy.Code)
	}
	s.wait = 5 * time.Second
	done := make(chan int)
	go func() { done <- uiRequest(s, "GET", "/audit", host, "", s.token, "").Code }()
	time.Sleep(100 * time.Millisecond)
	s.release()
	if code := <-done; code != http.StatusOK {
		t.Errorf("a request that waited for the one before it: %d, want 200", code)
	}
}

// CLI-INV-41: the page and its JSON twin come from one handler and one report.
func TestUIJSONTwin(t *testing.T) {
	root := taskRepo(t)
	uiCommit(t, root, "a", "one")
	head := uiCommit(t, root, "a", "two")
	rep := auditReport{Schema: auditSchema, Base: head, Head: head, Counts: auditCounts{Total: 1}, Controls: []auditControl{{ID: "C1", Question: "q <script>", State: auditHeld, Evidence: []auditEvidence{}}}, Exceptions: []auditException{}}
	rep.Headline = auditHeadline(rep)
	s := testUIServer(t, root, rep)
	host := "127.0.0.1:4242"
	r := httptest.NewRequest("GET", "/audit", nil)
	r.Host = host
	r.Header.Set("Accept", "application/json")
	r.AddCookie(&http.Cookie{Name: s.cookieName(), Value: s.token})
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	var got uiView
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("JSON twin: %v\n%s", err, w.Body.String())
	}
	if got.Report.Headline != rep.Headline || len(got.Report.Controls) != 1 || got.Report.Controls[0].Question != "q <script>" {
		t.Errorf("JSON twin report = %+v", got.Report)
	}
	html := uiRequest(s, "GET", "/audit", host, "", s.token, "").Body.String()
	if !strings.Contains(html, rep.Headline) || !strings.Contains(html, "q &lt;script&gt;") || strings.Contains(html, "q <script>") {
		t.Error("the page did not escape the report's text")
	}
	tr := uiRequest(s, "GET", "/transcript", host, "", s.token, "")
	body, _ := io.ReadAll(tr.Body)
	want, _ := transcript(rep)
	if string(body) != string(want) {
		t.Errorf("transcript differs from the report's canonical JSON:\n%s\n%s", body, want)
	}
}

// CLI-INV-39 read-back: records committed at HEAD that break a rule are shown and close nothing; an
// attestation is read back with how far HEAD has moved.
func TestOutcomeReadBack(t *testing.T) {
	ctx := context.Background()
	key := uiKey(t)
	pub := key.Public().(ed25519.PublicKey)
	_, late, _ := ed25519.GenerateKey(nil)
	root := taskRepo(t)
	uiWrite(t, root, ".proofbound/keys/auditor.pub", pbwarranty.PubLine(pub))
	base := uiCommit(t, root, "auditor", "declare the auditor's key")
	uiWrite(t, root, "src/a.sh", "a\n")
	agent := uiCommit(t, root, "agent", "agent change")
	uiWrite(t, root, "src/b.sh", "b\n")
	own := uiCommit(t, root, "auditor", "auditor's own change")
	uiWrite(t, root, "src/c.sh", "c\n")
	other := uiCommit(t, root, "agent", "another agent change")
	uiWrite(t, root, ".proofbound/keys/late.pub", pbwarranty.PubLine(late.Public().(ed25519.PublicKey)))
	uiWrite(t, root, "src/d.sh", "d\n")
	fourth := uiCommit(t, root, "agent", "a late key and a change")
	ex := func(sha string) auditException { return auditException{ID: "uncovered-" + sha[:12], Commit: sha} }
	record := func(id, commit, by string, k ed25519.PrivateKey, extra map[string]any) {
		f := map[string]any{"schema": outcomeSchema, "exception": id, "commit": commit, "outcome": "accepted", "reason": "r", "by": by, "key": pbwarranty.Fingerprint(k.Public().(ed25519.PublicKey)), "at": "2026-11-22T12:00:00Z"}
		for k2, v := range extra {
			f[k2] = v
		}
		data, err := signRecord(f, k)
		if err != nil {
			t.Fatal(err)
		}
		uiWrite(t, root, outcomesDir+"/"+id+".json", string(data))
	}
	record("uncovered-"+agent[:12], agent, "auditor", key, nil)                                                  // good
	record("uncovered-"+own[:12], own, "auditor", key, nil)                                                      // by the author
	record("uncovered-"+other[:12], agent, "auditor", key, nil)                                                  // names another commit
	record("uncovered-"+fourth[:12], fourth, "late", late, map[string]any{"outcome": "fixed", "fixed_by": base}) // late key, bad fix
	head := uiCommit(t, root, "auditor", "records")
	rep := auditReport{Base: base, Head: head, Counts: auditCounts{Total: 5}, Exceptions: []auditException{ex(agent), ex(own), ex(other), ex(fourth)}}
	attachOutcomes(ctx, root, &rep)
	wantProblem := []string{"", "own author", "different exception", "no key declared"}
	for i, e := range rep.Exceptions {
		o := e.Outcome
		if o == nil {
			t.Fatalf("%d: no outcome read back", i)
		}
		if wantProblem[i] == "" {
			if !o.Verified {
				t.Errorf("%d: a good record did not verify: %s", i, o.Problem)
			}
			continue
		}
		if o.Verified || !strings.Contains(o.Problem, wantProblem[i]) {
			t.Errorf("%d: verified=%v problem=%q, want %q", i, o.Verified, o.Problem, wantProblem[i])
		}
	}
	if !strings.Contains(rep.Headline, "4 exception(s), 1 with a signed outcome") {
		t.Errorf("headline = %q", rep.Headline)
	}
	// The same record reads the same whatever period is viewed (owner try-out: it read invalid when
	// the period started before the keys were declared).
	wide := auditReport{Base: uiGit(t, root, "rev-list", "--max-parents=0", "HEAD"), Head: head, Counts: auditCounts{Total: 6}, Exceptions: []auditException{ex(agent)}}
	attachOutcomes(ctx, root, &wide)
	if o := wide.Exceptions[0].Outcome; o == nil || !o.Verified {
		t.Errorf("a valid outcome read invalid for a wider period: %+v", o)
	}

	att := map[string]any{"schema": attestationSchema, "base": base, "head": head, "transcript_sha256": strings.Repeat("a", 64), "headline": "h", "by": "auditor", "key": pbwarranty.Fingerprint(pub), "at": "2026-11-22T12:00:00Z"}
	data, _ := signRecord(att, key)
	uiWrite(t, root, attestationsDir+"/good.json", string(data))
	uiWrite(t, root, attestationsDir+"/bad.json", strings.Replace(string(data), `"headline":"h"`, `"headline":"H"`, 1))
	now := uiCommit(t, root, "auditor", "attestations")
	views := readAttestations(ctx, root, auditReport{Base: base, Head: now})
	if len(views) != 2 || views[0].Verified || !views[1].Verified || views[1].Moved != 1 || views[1].RecordsOnly != 1 {
		t.Errorf("attestations = %+v", views)
	}
}

// CLI-INV-40 end to end: preview then sign commits only the record; a changed record, an existing
// record and a refused commit are all refused, and a refused commit leaves nothing behind; a base that
// is not an ancestor of HEAD is a 400.
func TestUISignFlow(t *testing.T) {
	key := uiKey(t)
	root := taskRepo(t)
	uiWrite(t, root, ".proofbound/keys/auditor.pub", pbwarranty.PubLine(key.Public().(ed25519.PublicKey)))
	base := uiCommit(t, root, "auditor", "declare the auditor's key")
	uiWrite(t, root, "src/a.sh", "a\n")
	agent := uiCommit(t, root, "agent", "agent change")
	uiWrite(t, root, "src/b.sh", "b\n")
	agent2 := uiCommit(t, root, "agent", "second agent change")
	s := newUIServer(root, "", base, 4242)
	s.compute = func(ctx context.Context, b string) (ledgerView, string, error) {
		return ledgerView{Report: auditReport{Schema: auditSchema, Base: base, Head: agent2, Counts: auditCounts{Total: 2}, Controls: []auditControl{},
			Exceptions: []auditException{{ID: "uncovered-" + agent[:12], Commit: agent}, {ID: "uncovered-" + agent2[:12], Commit: agent2}}}}, "", nil
	}
	host, origin := "127.0.0.1:4242", "http://127.0.0.1:4242"
	act := func(path string, body map[string]string) (int, map[string]any) {
		raw, _ := json.Marshal(body)
		w := uiRequest(s, "POST", path, host, origin, s.token, string(raw))
		var out map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return w.Code, out
	}
	uiWrite(t, root, "unrelated.txt", "staged\n")
	uiGit(t, root, "add", "unrelated.txt")
	req := map[string]string{"kind": "outcome", "exception": "uncovered-" + agent[:12], "outcome": "accepted", "reason": "fine"}
	code, pv := act("/act/preview", req)
	if code != 200 || pv["digest"] == nil || pv["at"] == nil {
		t.Fatalf("preview: %d %v", code, pv)
	}
	req["at"] = pv["at"].(string)
	req["digest"] = strings.Repeat("0", 64)
	if code, out := act("/act/sign", req); code != http.StatusConflict {
		t.Errorf("a sign for a record other than the previewed one: %d %v", code, out)
	}
	req["digest"] = pv["digest"].(string)
	if code, out := act("/act/sign", req); code != 200 || out["committed"] != true {
		t.Fatalf("sign: %d %v", code, out)
	}
	if files := uiGit(t, root, "show", "--name-only", "--format=", "HEAD"); files != outcomesDir+"/uncovered-"+agent[:12]+".json" {
		t.Errorf("the record commit holds %q", files)
	}
	if staged := uiGit(t, root, "diff", "--cached", "--name-only"); staged != "unrelated.txt" {
		t.Errorf("other staged work was touched: %q", staged)
	}
	if code, _ := act("/act/sign", req); code != http.StatusConflict {
		t.Errorf("an existing record was replaced: %d", code)
	}
	hook := filepath.Join(root, ".git", "hooks", "commit-msg")
	uiWrite(t, root, ".git/hooks/commit-msg", "#!/bin/sh\necho refused by hook\nexit 1\n")
	if err := os.Chmod(hook, 0o755); err != nil {
		t.Fatal(err)
	}
	req2 := map[string]string{"kind": "outcome", "exception": "uncovered-" + agent2[:12], "outcome": "accepted", "reason": "fine"}
	_, pv2 := act("/act/preview", req2)
	req2["at"], req2["digest"] = pv2["at"].(string), pv2["digest"].(string)
	code, out := act("/act/sign", req2)
	rel := outcomesDir + "/uncovered-" + agent2[:12] + ".json"
	if code != http.StatusConflict || !strings.Contains(fmt.Sprint(out["error"]), "refused by hook") {
		t.Errorf("a refused commit: %d %v", code, out)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); !os.IsNotExist(err) {
		t.Error("a refused commit left its file behind")
	}
	if staged := uiGit(t, root, "diff", "--cached", "--name-only"); staged != "unrelated.txt" {
		t.Errorf("a refused commit left the index changed: %q", staged)
	}

	uiGit(t, root, "checkout", "-q", "-b", "side")
	side := uiCommit(t, root, "a", "side")
	uiGit(t, root, "checkout", "-q", "main")
	if w := uiRequest(s, "GET", "/audit?base="+side, host, "", s.token, ""); w.Code != http.StatusBadRequest {
		t.Errorf("a base that is not an ancestor of HEAD: %d", w.Code)
	}
	if w := uiRequest(s, "GET", "/audit?base=--output=x", host, "", s.token, ""); w.Code != http.StatusBadRequest {
		t.Errorf("an option-like base: %d", w.Code)
	}
	uiGit(t, root, "tag", "v1", agent)
	if got := defaultAuditBase(context.Background(), root); got != "v1" {
		t.Errorf("default base = %q, want the newest tag below HEAD", got)
	}
}

// The "Who did the work" disclosure: agents come from the author or Co-Authored-By names; each commit
// counts for its author and its agents, with its class and open exceptions.
func TestWhoDidTheWork(t *testing.T) {
	ctx := context.Background()
	root := taskRepo(t)
	base := uiCommit(t, root, "dev", "start")
	uiWrite(t, root, "src/a.sh", "a\n")
	withClaude := uiCommit(t, root, "dev", "code\n\nCo-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>")
	uiWrite(t, root, "src/b.sh", "b\n")
	bot := uiCommit(t, root, "copilot[bot]", "bot code")
	uiWrite(t, root, "docs/x.md", "x\n")
	uiCommit(t, root, "dev", "docs\n\nCo-Authored-By: Pat Person <pat@example.com>")
	infos, err := commitInfos(ctx, root, base)
	if err != nil {
		t.Fatal(err)
	}
	if got := infos[withClaude].agents; len(got) != 1 || got[0] != "Claude Opus 5.5" {
		t.Errorf("Co-Authored-By agent = %v", got)
	}
	if got := infos[bot].agents; len(got) != 1 || got[0] != "copilot[bot]" {
		t.Errorf("agent author = %v", got)
	}
	rep := auditReport{Commits: []auditCommit{
		{Commit: withClaude, Author: "dev", Class: classUncovered, Agents: infos[withClaude].agents},
		{Commit: bot, Author: "copilot[bot]", Class: classClaimed, Agents: infos[bot].agents},
	}, Exceptions: []auditException{{Commit: withClaude}, {Commit: withClaude, Outcome: &outcomeView{Verified: true}}}}
	rows := actorRows(rep)
	want := []actorRow{
		{Name: "Claude Opus 5.5", Kind: "agent", Commits: 1, Uncovered: 1, Exceptions: 2, Open: 1},
		{Name: "copilot[bot]", Kind: "agent", Commits: 1, Claimed: 1},
		{Name: "dev", Kind: "person", Commits: 1, Uncovered: 1, Exceptions: 2, Open: 1},
	}
	if len(rows) != len(want) {
		t.Fatalf("rows = %+v", rows)
	}
	for i := range want {
		if rows[i] != want[i] {
			t.Errorf("row %d = %+v, want %+v", i, rows[i], want[i])
		}
	}
}

// C8's rule: a reviewer may not be the author, an agent, the signer or a spec owner, by key name.
func TestReviewIndependence(t *testing.T) {
	c := auditCommit{Author: "Dev Person", Agents: []string{"Claude Opus 5.5"}, Signer: "agent-key"}
	cases := []struct {
		reviewers []string
		ok        bool
		note      string
	}{
		{[]string{"demo-reviewer"}, true, ""},
		{[]string{"dev-person"}, false, "its author"},
		{[]string{"claude-opus-5-5"}, false, "its agent"},
		{[]string{"agent-key"}, false, "its signer"},
		{[]string{"demo-reviewer", "spec-owner"}, false, "a spec owner"},
	}
	for _, k := range cases {
		ok, note := independence(c, k.reviewers, []string{"Spec Owner"})
		if ok != k.ok || !strings.Contains(note, k.note) {
			t.Errorf("%v: %v %q", k.reviewers, ok, note)
		}
	}
}

// `keys git-sign` (CLI-INV-44): a commit made after it verifies as the declared key's owner; an
// unsigned commit is unsigned; a key no longer declared at HEAD no longer verifies.
func TestGitSignVerifies(t *testing.T) {
	if _, err := exec.LookPath("ssh-keygen"); err != nil {
		t.Skip("ssh-keygen not installed")
	}
	ctx := context.Background()
	key := uiKey(t)
	root := taskRepo(t)
	base := uiCommit(t, root, "a", "start")
	uiWrite(t, root, ".proofbound/keys/agent-a.pub", pbwarranty.PubLine(key.Public().(ed25519.PublicKey)))
	unsigned := uiCommit(t, root, "a", "declare")
	var out strings.Builder
	if err := keysGitSign(ctx, root, &out); err != nil {
		t.Fatalf("git-sign: %v", err)
	}
	if err := keysGitSign(ctx, root, &out); err != nil {
		t.Fatalf("git-sign is not repeatable: %v", err)
	}
	if got := uiGit(t, root, "config", "--worktree", "--get", "user.signingkey"); !strings.HasSuffix(got, "-git-ssh") {
		t.Errorf("the signing key is not set for this worktree: %q", got)
	}
	if info, err := os.Stat(strings.TrimSpace(uiGit(t, root, "config", "--worktree", "--get", "user.signingkey"))); err != nil || info.Mode().Perm() != 0o600 {
		t.Errorf("the exported key is not 0600: %v", err)
	}
	uiWrite(t, root, "src/a.sh", "a\n")
	uiGit(t, root, "add", "-A")
	uiGit(t, root, "commit", "-q", "--no-verify", "-m", "signed by the agent")
	signed := uiGit(t, root, "rev-parse", "HEAD")
	sigs, err := commitSignatures(ctx, root, base)
	if err != nil {
		t.Fatal(err)
	}
	if sigs[signed] != [2]string{sigVerified, "agent-a"} || sigs[unsigned][0] != sigUnsigned {
		t.Fatalf("signatures = %v", sigs)
	}
	uiGit(t, root, "rm", "-q", ".proofbound/keys/agent-a.pub")
	uiGit(t, root, "-c", "commit.gpgsign=false", "commit", "-q", "--no-verify", "-m", "undeclare")
	sigs, _ = commitSignatures(ctx, root, base)
	if sigs[signed][0] != sigOther {
		t.Errorf("a key no longer declared still verified: %v", sigs[signed])
	}
}
