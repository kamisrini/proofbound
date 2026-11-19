package sdd_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// R2 starter kit, CLI-INV-19 end to end (docs/plans/R2-starter-kit-spec.md): on a tenant that ran
// `review draft` -> decide -> `review seal` -> commit -> sync, the reviewed commit's own readiness
// gate is PASS when every outcome is SATISFIED/VERIFIABLE by a non-owner, and BLOCKED with any other
// outcome. Runs through cli.Run on the embedded store (this package owns the default port).
func TestStarterKitReviewFlow(t *testing.T) {
	for _, c := range []struct{ outcome, state string }{{"SATISFIED", "PASS"}, {"INCONCLUSIVE", "BLOCKED"}} {
		t.Run(c.outcome, func(t *testing.T) {
			m := newM3Tenant(t)
			x := m.commit("add auth", "Intent: sdd:add-auth")
			m.mustRun("sync", "intent", "sdd")
			m.mustRun("sync", "git")
			mustContain(t, m.mustRun("review", "draft", x), "SHALL authenticate users", "draft ")

			m.review(x, c.outcome)
			m.commit("review " + x[:12])
			m.mustRun("sync", "git")
			m.mustRun("sync", "reviews")
			m.expectStates(map[string]string{"ready-" + x[:12]: c.state})
			if c.state == "PASS" && !strings.Contains(m.mustRun("gates", "canary"), `reason="1 of 1 obligations verified SATISFIED; 1 of 1 requirement reviews VERIFIABLE"`) {
				t.Fatal("PASS without its counts")
			}
		})
	}
}

// review drafts, decides every outcome (reviews VERIFIABLE, obligations as given) as a non-owner,
// and seals: the reviewer workflow of the starter kit, through cli.Run.
func (m *m3Tenant) review(x, outcome string) {
	t := m.t
	t.Helper()
	dir := filepath.Join(m.root, ".proofbound", "drafts", x)
	path := filepath.Join(dir, "draft.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var draft map[string]any
	if err := json.Unmarshal(raw, &draft); err != nil {
		t.Fatal(err)
	}
	draft["declared_reviewer"] = "proofbound-reviewer"
	for _, r := range draft["reviews"].([]any) {
		for _, o := range r.(map[string]any)["obligations"].([]any) {
			o.(map[string]any)["outcome"] = "VERIFIABLE"
		}
	}
	for _, v := range draft["verdicts"].([]any) {
		for _, o := range v.(map[string]any)["obligations"].([]any) {
			o.(map[string]any)["outcome"] = outcome
		}
	}
	edited, err := json.Marshal(draft)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, edited, 0o644); err != nil {
		t.Fatal(err)
	}
	m.mustRun("review", "seal", dir)
}
