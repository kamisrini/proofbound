package cli

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	connectorgit "github.com/kamisrini/proofbound/kernel/internal/connector/git"
	connectorintent "github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	connectorreviews "github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

func ev(t *testing.T, seq int64, source core.Source, kind core.Kind, v any) promiseEvent {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return promiseEvent{seq: seq, source: source, kind: kind, payload: raw}
}

func promiseFixture(t *testing.T) []promiseEvent {
	sha0, sha1 := strings.Repeat("0", 64), strings.Repeat("1", 64)
	req := func(sha string) connectorintent.Requirement {
		return connectorintent.Requirement{RequirementID: "REQ-001-x-aaaaaa", DeclaredOwner: "spec-owner", ArtifactSHA256: sha, Obligations: []connectorintent.Obligation{
			{ID: "O-1", Statement: "one"}, {ID: "O-2", Statement: "two"}, {ID: "O-3", Statement: "three, revision " + sha[:1]}, {ID: "O-4", Statement: "four"}, {ID: "O-5", Statement: "five"}}}
	}
	ci := func(id, path, sha string) connectorintent.ChangeIntent {
		return connectorintent.ChangeIntent{IntentID: id, ArtifactPath: path, Targets: []connectorintent.Reference{{Source: "intent.sdd", RecordID: "REQ-001-x-aaaaaa", ArtifactSHA256: sha, ObligationIDs: []string{"O-1", "O-2", "O-3", "O-4", "O-5"}}}}
	}
	return []promiseEvent{
		ev(t, 1, core.SourceIntentSDD, core.KindRequirement, req(sha0)),
		ev(t, 2, core.SourceIntentSDD, core.KindChangeIntent, ci("CI-001-x-aaaaaa", "specs/001-x", sha0)),
		ev(t, 3, core.SourceReviews, core.KindReviewVerdict, connectorreviews.ObligationVerdict{Schema: "proofbound.obligation-verdict.v2", DeclaredReviewer: "old-reviewer", ChangeIntent: connectorintent.Reference{RecordID: "CI-001-x-aaaaaa"}, Obligations: []connectorreviews.ObligationOutcome{{Source: "intent.sdd", RequirementID: "REQ-001-x-aaaaaa", ArtifactSHA256: sha0, ObligationID: "O-3", Outcome: "SATISFIED"}, {Source: "intent.sdd", RequirementID: "REQ-001-x-aaaaaa", ArtifactSHA256: sha0, ObligationID: "O-5", Outcome: "SATISFIED"}}}),
		ev(t, 4, core.SourceIntentSDD, core.KindRequirement, req(sha1)),
		ev(t, 5, core.SourceIntentSDD, core.KindChangeIntent, ci("CI-001-x-aaaaaa", "specs/001-x", sha1)),
		ev(t, 6, core.SourceReviews, core.KindReviewVerdict, connectorreviews.ObligationVerdict{Schema: "proofbound.obligation-verdict.v2", DeclaredReviewer: "demo-reviewer", ChangeIntent: connectorintent.Reference{RecordID: "CI-001-x-aaaaaa"}, Obligations: []connectorreviews.ObligationOutcome{
			{Source: "intent.sdd", RequirementID: "REQ-001-x-aaaaaa", ArtifactSHA256: sha1, ObligationID: "O-1", Outcome: "SATISFIED"},
			{Source: "intent.sdd", RequirementID: "REQ-001-x-aaaaaa", ArtifactSHA256: sha1, ObligationID: "O-2", Outcome: "NOT_SATISFIED"}}}),
		ev(t, 7, core.SourceReviews, core.KindRequirementReview, connectorreviews.RequirementReview{DeclaredReviewer: "demo-reviewer", Requirement: connectorintent.Reference{Source: "intent.sdd", RecordID: "REQ-001-x-aaaaaa", ArtifactSHA256: sha1},
			Outcomes: []connectorreviews.RequirementReviewOutcome{{ObligationID: "O-1", Outcome: "AMBIGUOUS", Finding: "of what?"}, {ObligationID: "O-2", Outcome: "VERIFIABLE"}}}),
		ev(t, 8, core.SourceGit, core.KindCommitRecorded, connectorgit.Commit{SHA: strings.Repeat("c", 40), IntentRefs: []connectorgit.IntentRef{{Provider: "sdd", RecordID: "CI-001-x-aaaaaa"}}}),
		ev(t, 9, core.SourceIntentSDD, core.KindChangeIntent, ci("CI-002-y-bbbbbb", "specs/002-y", sha1)),
	}
}

// Review ui-u2-review-20261122 F4: a later NOT_SATISFIED outweighs an older proof of the same words.
func TestPromiseLatestVerdictWins(t *testing.T) {
	sha := func(c string) string { return strings.Repeat(c, 64) }
	req := func(s, words string) connectorintent.Requirement {
		return connectorintent.Requirement{RequirementID: "REQ-x", ArtifactSHA256: s, Obligations: []connectorintent.Obligation{{ID: "O-1", Statement: words}}}
	}
	ci := func(s string) connectorintent.ChangeIntent {
		return connectorintent.ChangeIntent{IntentID: "CI-x", ArtifactPath: "specs/x", Targets: []connectorintent.Reference{{Source: "intent.sdd", RecordID: "REQ-x", ArtifactSHA256: s, ObligationIDs: []string{"O-1"}}}}
	}
	verdict := func(s, outcome string) connectorreviews.ObligationVerdict {
		return connectorreviews.ObligationVerdict{Schema: "proofbound.obligation-verdict.v2", ChangeIntent: connectorintent.Reference{RecordID: "CI-x"}, Obligations: []connectorreviews.ObligationOutcome{{Source: "intent.sdd", RequirementID: "REQ-x", ArtifactSHA256: s, ObligationID: "O-1", Outcome: outcome}}}
	}
	events := []promiseEvent{
		ev(t, 1, core.SourceIntentSDD, core.KindRequirement, req(sha("1"), "same")),
		ev(t, 2, core.SourceReviews, core.KindReviewVerdict, verdict(sha("1"), "SATISFIED")),
		ev(t, 3, core.SourceIntentSDD, core.KindRequirement, req(sha("2"), "same")),
		ev(t, 4, core.SourceReviews, core.KindReviewVerdict, verdict(sha("2"), "NOT_SATISFIED")),
		ev(t, 5, core.SourceIntentSDD, core.KindRequirement, req(sha("3"), "same")),
		ev(t, 6, core.SourceIntentSDD, core.KindChangeIntent, ci(sha("3"))),
	}
	if got := computePromises(events)[0].Promises[0].State; got != promiseNotKept {
		t.Errorf("an old proof outweighed a later NOT_SATISFIED: %s", got)
	}
}

// CLI-INV-45: the promise rule.
func TestPromiseStates(t *testing.T) {
	changes := computePromises(promiseFixture(t))
	if len(changes) != 2 || changes[0].Change != "001-x" || changes[1].Change != "002-y" {
		t.Fatalf("changes = %+v", changes)
	}
	want := map[string]string{"O-1": promiseProven, "O-2": promiseNotKept, "O-3": promiseStale, "O-4": promiseUnreviewed, "O-5": promiseProven}
	for _, p := range changes[0].Promises {
		if p.State != want[p.ID] {
			t.Errorf("%s = %s, want %s", p.ID, p.State, want[p.ID])
		}
		if p.ID == "O-1" && (p.Unclear != "AMBIGUOUS" || p.Finding != "of what?" || p.Reviewer != "demo-reviewer") {
			t.Errorf("O-1 = %+v", p)
		}
		if p.ID == "O-2" && p.Unclear != "" {
			t.Errorf("a VERIFIABLE review marked O-2 unclear")
		}
		if len(p.Commits) != 1 {
			t.Errorf("%s commits = %v", p.ID, p.Commits)
		}
	}
	if changes[0].Owners[0] != "spec-owner" {
		t.Errorf("owners = %v", changes[0].Owners)
	}
	if changes[1].Promises[0].State != promiseUnclaimed {
		t.Errorf("an unclaimed change: %s", changes[1].Promises[0].State)
	}
}

// CLI-INV-46: the Now list.
func TestNowItems(t *testing.T) {
	sha := strings.Repeat("a", 40)
	rep := auditReport{Base: strings.Repeat("b", 40), Controls: []auditControl{{ID: "C1", Evidence: []auditEvidence{
		{Commit: sha, Text: "NO REVIEW: no readiness gate"}, {Commit: strings.Repeat("d", 40), Text: "PASS: sdd:x"}, {Commit: strings.Repeat("e", 40), Text: "BLOCKED: obligation not satisfied"}}}},
		Exceptions: []auditException{{Kind: "uncovered", Commit: strings.Repeat("f", 40), Text: "x"}, {Kind: "uncovered", Commit: strings.Repeat("9", 40), Outcome: &outcomeView{Verified: true}}}}
	items := nowItems(rep, computePromises(promiseFixture(t)))
	kinds := []string{}
	for _, it := range items {
		kinds = append(kinds, it.Kind)
	}
	if got := strings.Join(kinds, ","); got != "review,blocked,uncovered,unclear,not-kept,drift,unclear" {
		t.Fatalf("kinds = %s", got)
	}
	if items[0].Link != "/review?commit="+sha || !strings.Contains(items[0].Command, "review draft "+sha[:12]) {
		t.Errorf("review item = %+v", items[0])
	}
}

// CLI-INV-47: who may review.
func TestBenchEligibility(t *testing.T) {
	c := auditCommit{Author: "agent-a", Agents: []string{"Claude Opus 5.5"}}
	owners := []string{"spec-owner"}
	for name, want := range map[string]bool{"demo-reviewer": true, "agent-a": false, "claude-opus-5-5": false, "spec-owner": false} {
		if ok, why := eligible(name, c, owners); ok != want || (!ok && !strings.HasPrefix(why, "you are")) {
			t.Errorf("%s: %v %q", name, ok, why)
		}
	}
}

// CLI-INV-48: the agent's next step and the bearer token.
func TestAgentNext(t *testing.T) {
	promises := computePromises(promiseFixture(t))
	claim := strings.Repeat("c", 40)
	rep := auditReport{Head: strings.Repeat("h", 40), Controls: []auditControl{{ID: "C1", Evidence: []auditEvidence{{Commit: claim, Text: "NO REVIEW: no readiness gate"}, {Commit: strings.Repeat("z", 40), Text: "PASS: other"}}}}}
	ans := agentNext(rep, promises, "sdd:001-x")
	if ans.Change == nil || len(ans.Commits) != 1 || ans.Commits[0].Commit != claim {
		t.Fatalf("answer = %+v", ans)
	}
	if len(ans.Blocking) != 1 || ans.Blocking[0].YouMay {
		t.Errorf("an agent may review its own change: %+v", ans.Blocking)
	}
	dos := []string{}
	for _, n := range ans.Next {
		dos = append(dos, n.Do)
	}
	if got := strings.Join(dos, ","); got != "stop_and_ask_spec_owner,fix_code,hand_off_for_re_review,hand_off" {
		t.Errorf("next = %s", got)
	}
	if got := agentNext(rep, promises, "").Next[0].Do; got != "start_task" {
		t.Errorf("no task: %s", got)
	}
	if got := agentNext(rep, promises, "sdd:999-none").Next[0].Do; got != "stop_and_ask_spec_owner" {
		t.Errorf("unknown change: %s", got)
	}

	root := taskRepo(t)
	uiCommit(t, root, "a", "one")
	head := uiCommit(t, root, "a", "two")
	s := testUIServer(t, root, auditReport{Schema: auditSchema, Base: head, Head: head, Controls: []auditControl{}, Exceptions: []auditException{}})
	get := func(auth string) int {
		r := uiRequest(s, "GET", "/agent/next?task=sdd:001-x", "127.0.0.1:4242", "", "", "")
		if auth != "" {
			r = uiRequestWith(s, "GET", "/agent/next?task=sdd:001-x", auth)
		}
		return r.Code
	}
	if get("") != http.StatusUnauthorized || get("Bearer "+s.agentToken) != http.StatusOK || get("Bearer 00") != http.StatusUnauthorized || get("Bearer "+s.token) != http.StatusUnauthorized {
		t.Error("only the agent token may read by bearer")
	}
	if w := uiRequestWith(s, "POST", "/act/review-draft", "Bearer "+s.agentToken); w.Code != http.StatusUnauthorized {
		t.Errorf("an agent's bearer token acted: %d", w.Code)
	}
	// The agent token never opens the person's page (review ui-u2-review-20261122 F1).
	if w := uiRequest(s, "GET", "/?k="+s.agentToken, "127.0.0.1:4242", "", "", ""); w.Code != http.StatusForbidden {
		t.Errorf("the agent token opened the page: %d", w.Code)
	}
}

// CLI-INV-49: the bench saves a draft, previews, seals and commits only the sealed files; a refused
// commit leaves nothing; an ineligible reviewer is refused.
func TestUISealFlow(t *testing.T) {
	key := uiKey(t)
	root := taskRepo(t)
	uiWrite(t, root, ".proofbound/keys/demo-reviewer.pub", pbwarranty.PubLine(key.Public().(ed25519.PublicKey)))
	base := uiCommit(t, root, "spec-owner", "keys")
	uiWrite(t, root, "src/a.sh", "a\n")
	claim := uiCommit(t, root, "agent-a", "code\n\nIntent: sdd:001-x")
	truth := draftFile{Commit: claim, Verdicts: []draftVerdict{{Obligations: []draftVerdictObligation{{ObligationID: "O-1", Statement: "one"}}}},
		Reviews: []draftReview{{DeclaredOwner: "spec-owner", Obligations: []draftReviewObligation{{ObligationID: "O-1", Statement: "one"}}}}}
	s := newUIServer(root, "", base, 4242)
	s.compute = func(ctx context.Context, b string) (ledgerView, string, error) {
		return ledgerView{Report: auditReport{Schema: auditSchema, Base: base, Head: claim, Commits: []auditCommit{{Commit: claim, Author: "agent-a", Class: classClaimed}},
			Controls: []auditControl{{ID: "C1", Evidence: []auditEvidence{{Commit: claim, Text: "NO REVIEW: x"}}}}, Exceptions: []auditException{}}}, "", nil
	}
	s.drafts = func(ctx context.Context, shas []string) (map[string]draftFile, error) {
		return map[string]draftFile{claim: truth}, nil
	}
	var sealed draftFile
	s.seal = func(ctx context.Context, dir string, out io.Writer) error {
		raw, err := os.ReadFile(filepath.Join(dir, "draft.json"))
		if err != nil {
			return err
		}
		if err := json.Unmarshal(raw, &sealed); err != nil {
			return err
		}
		if sealed.DeclaredReviewer != "demo-reviewer" {
			return errors.New("wrong reviewer")
		}
		for _, rel := range []string{"docs/verification/verdicts/OV-x.md", "gates/ready-x.yaml"} {
			uiWrite(t, root, rel, "sealed\n")
			_, _ = io.WriteString(out, "wrote "+rel+"\n")
		}
		return nil
	}
	act := func(path string, decisions []reviewDecision) (int, map[string]any) {
		raw, _ := json.Marshal(reviewRequest{Commit: claim, Decisions: decisions})
		w := uiRequest(s, "POST", path, "127.0.0.1:4242", "http://127.0.0.1:4242", s.token, string(raw))
		var out map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return w.Code, out
	}
	half := []reviewDecision{{Part: "verdict", Group: 0, Index: 0, ObligationID: "O-1", Outcome: "SATISFIED"}}
	full := append(half, reviewDecision{Part: "review", Group: 0, Index: 0, ObligationID: "O-1", Outcome: "AMBIGUOUS", Finding: "of what?"})
	if code, out := act("/act/review-draft", half); code != 200 || !strings.Contains(out["saved"].(string), claim) {
		t.Fatalf("draft: %d %v", code, out)
	}
	if code, out := act("/act/review-preview", half); code != http.StatusUnprocessableEntity || out["missing"] == nil {
		t.Errorf("an undecided promise previewed: %d %v", code, out)
	}
	if code, _ := act("/act/review-preview", []reviewDecision{{Part: "verdict", Group: 0, Index: 0, ObligationID: "O-9", Outcome: "SATISFIED"}}); code != http.StatusUnprocessableEntity {
		t.Errorf("a decision for another obligation was accepted: %d", code)
	}
	if code, out := act("/act/review-preview", full); code != 200 || len(out["files"].([]any)) != 3 {
		t.Errorf("preview: %d %v", code, out)
	}
	uiWrite(t, root, "unrelated.txt", "x\n")
	uiGit(t, root, "add", "unrelated.txt")
	if code, out := act("/act/review-seal", full); code != 200 || out["committed"] != true {
		t.Fatalf("seal: %d %v", code, out)
	}
	if got := uiGit(t, root, "show", "--name-only", "--format=%s", "HEAD"); got != "review "+claim[:12]+"\n\ndocs/verification/verdicts/OV-x.md\ngates/ready-x.yaml" {
		t.Errorf("the review commit holds %q", got)
	}
	if staged := uiGit(t, root, "diff", "--cached", "--name-only"); staged != "unrelated.txt" {
		t.Errorf("other staged work was touched: %q", staged)
	}
	if sealed.Reviews[0].Obligations[0].Finding != "of what?" || sealed.Verdicts[0].Obligations[0].Outcome != "SATISFIED" {
		t.Errorf("sealed draft = %+v", sealed)
	}

	// A refused commit leaves nothing; then an ineligible reviewer is refused before anything is written.
	uiWrite(t, root, ".git/hooks/commit-msg", "#!/bin/sh\nexit 1\n")
	_ = os.Chmod(filepath.Join(root, ".git", "hooks", "commit-msg"), 0o755)
	s.seal = func(ctx context.Context, dir string, out io.Writer) error {
		uiWrite(t, root, "gates/ready-y.yaml", "y\n")
		_, _ = io.WriteString(out, "wrote gates/ready-y.yaml\n")
		return nil
	}
	if code, _ := act("/act/review-seal", full); code != http.StatusConflict {
		t.Errorf("a refused commit: %d", code)
	}
	if _, err := os.Stat(filepath.Join(root, "gates", "ready-y.yaml")); !os.IsNotExist(err) {
		t.Error("a refused commit left a sealed file behind")
	}
	truth.Reviews[0].DeclaredOwner = "demo-reviewer"
	if code, out := act("/act/review-preview", full); code != http.StatusForbidden || !strings.Contains(out["error"].(string), "you are a spec owner") {
		t.Errorf("an owner reviewing: %d %v", code, out)
	}
}

// uiRequestWith sends a request with an Authorization header and no cookie (an agent).
func uiRequestWith(s *uiServer, method, path, auth string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader("{}"))
	r.Host = "127.0.0.1:4242"
	r.Header.Set("Authorization", auth)
	if method == "POST" {
		r.Header.Set("Origin", "http://127.0.0.1:4242")
	}
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	return w
}
