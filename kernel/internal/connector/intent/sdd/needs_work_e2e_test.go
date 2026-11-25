package sdd_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/store"
)

// reviewWith seals x like review, after edit has changed the draft.
func (m *m3Tenant) reviewWith(x, outcome string, edit func(map[string]any)) {
	t := m.t
	t.Helper()
	path := filepath.Join(m.root, ".proofbound", "drafts", x, "draft.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var draft map[string]any
	if err := json.Unmarshal(raw, &draft); err != nil {
		t.Fatal(err)
	}
	edit(draft)
	edited, err := json.Marshal(draft)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, edited, 0o644); err != nil {
		t.Fatal(err)
	}
	m.review(x, outcome)
}

func decideOutOfArea(decision, reason string) func(map[string]any) {
	return func(d map[string]any) {
		for _, h := range d["out_of_area"].([]any) {
			h.(map[string]any)["decision"] = decision
			h.(map[string]any)["reason"] = reason
		}
	}
}

// ledgerExec runs SQL on the tenant's ledger between commands, as an older Proofbound left it.
func (m *m3Tenant) ledgerExec(sql string) {
	m.t.Helper()
	ctx := context.Background()
	ledger, err := store.Open(ctx, store.Config{Root: filepath.Join(m.root, ".proofbound")})
	if err != nil {
		m.t.Fatal(err)
	}
	defer func() {
		if err := ledger.Close(); err != nil {
			m.t.Fatal(err)
		}
	}()
	if err := ledger.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error { _, err := tx.Exec(ctx, sql); return err }); err != nil {
		m.t.Fatal(err)
	}
}

// GATE-INV-17, P-INV-35, W-INV-3 (docs/plans/needs-work-blocks-spec.md) end to end: every promise
// SATISFIED but a line outside the area rejected seals NEEDS_WORK, and then nothing warrants the
// commit: not the check, not the bundle, not drift, not the intent report. A ledger from before the
// fix is migrated in place, and a re-seal of the same verdict as ACCEPTABLE lets it through (D1).
func TestNeedsWorkBlocksEndToEnd(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("PROOFBOUND_KEY", "")
	m := &m3Tenant{t: t, root: t.TempDir()}
	gitIn(t, m.root, "init", "-q", "-b", "main")
	gitIn(t, m.root, "config", "user.name", "owner-a")
	m.mustRun("init", "speckit")
	if err := removeHooks(m.root); err != nil {
		t.Fatal(err)
	}
	m.write("specs/001-export/spec.md", "# Feature: Export\n\n- **FR-001**: System MUST export reports as CSV.\n")
	m.write("specs/001-export/tasks.md", "- [ ] T001 Export CSV\n")
	m.write("specs/001-export/scope.paths", "src/x\n")
	m.mustRun("keys", "init", "--name", "owner-a")
	gitIn(t, m.root, "add", "-A", "--", ".gitignore", ".proofbound/keys")
	base := m.commit("specify 001-export and declare the exporter key")
	m.write("src/x", "x\n")
	m.write("src/other", "outside the area\n")
	x := m.commit("implement 001-export", "Intent: sdd:001-export")
	m.mustRun("sync", "all")
	m.mustRun("review", "draft", x)
	reviewerKey := filepath.Join(t.TempDir(), "reviewer-key")
	t.Setenv("PROOFBOUND_KEY", reviewerKey)
	m.mustRun("keys", "init", "--name", "proofbound-reviewer")
	m.reviewWith(x, "SATISFIED", decideOutOfArea("reject", ""))
	t.Setenv("PROOFBOUND_KEY", "")
	gitIn(t, m.root, "add", "-A", "-f", "--", "docs", "gates", ".proofbound/keys")
	m.commit("record the review")
	m.mustRun("sync", "all")

	verdict := "OV-" + x[:12] + "-1"
	blocked := func(when string) {
		t.Helper()
		code, out := runCLI(t, m.root, "check", base)
		mustContain(t, out, "NOT WARRANTED", x[:12]+" sdd:001-export BLOCKED", "the reviewer recorded NEEDS_WORK ("+verdict+"); 1 of 1 obligations verified SATISFIED")
		if code == 0 {
			t.Fatalf("%s: check passed a NEEDS_WORK commit:\n%s", when, out)
		}
	}
	blocked("after the review")
	dir := filepath.Join(t.TempDir(), "bundle")
	if code, out := runCLI(t, m.root, "warranty", "bundle", base, "-o", dir); code == 0 || strings.Contains(out, "VALID,") && strings.Contains(out, "1 of 1 claims verified") {
		t.Fatalf("bundle warranted a NEEDS_WORK commit: exit %d\n%s", code, out)
	}
	if _, err := os.Stat(filepath.Join(dir, "warranty-"+x[:12]+".json")); err == nil {
		t.Fatal("bundle wrote a warranty for a NEEDS_WORK commit")
	}
	if out := m.mustRun("report", "drift"); !strings.Contains(out, "summary warranted=0") {
		t.Fatalf("drift lists a NEEDS_WORK commit as warranted:\n%s", out)
	}
	if out := m.mustRun("report", "intent", "sdd:001-export"); !strings.Contains(out, "IMPLEMENTED_UNVERIFIED") || strings.Contains(out, "state=SATISFIED") {
		t.Fatalf("report intent does not read IMPLEMENTED_UNVERIFIED:\n%s", out)
	}

	// A ledger written before the fix has no verdict_status: the next command migrates it in place,
	// with no new event, and still blocks.
	m.ledgerExec(`ALTER TABLE obligation_verdicts_view DROP COLUMN verdict_status`)
	blocked("after the migration")
	// A status that is neither ACCEPTABLE nor NEEDS_WORK fails closed.
	m.ledgerExec(`UPDATE obligation_verdicts_view SET verdict_status=''`)
	if code, out := runCLI(t, m.root, "check", base); code == 0 || !strings.Contains(out, "verdict status unknown; run proofbound rebuild") {
		t.Fatalf("an unknown status did not fail closed: exit %d\n%s", code, out)
	}
	m.mustRun("rebuild")
	blocked("after a rebuild")

	// D1: the reviewer re-seals the same commit, accepting the line with a reason. The latest revision
	// of the verdict id decides, as it does for outcomes.
	matches, _ := filepath.Glob(filepath.Join(m.root, "docs", "verification", "verdicts", "*-"+x[:12]+"*"))
	matches = append(matches, filepath.Join(m.root, "gates", "ready-"+x[:12]+".yaml"))
	for _, f := range matches {
		gitIn(t, m.root, "rm", "-q", "--", f)
	}
	m.commit("withdraw the review")
	m.mustRun("sync", "all")
	_ = os.RemoveAll(filepath.Join(m.root, ".proofbound", "drafts", x))
	m.mustRun("review", "draft", x)
	t.Setenv("PROOFBOUND_KEY", reviewerKey)
	m.reviewWith(x, "SATISFIED", decideOutOfArea("accept", "the helper is generated and reviewed upstream"))
	t.Setenv("PROOFBOUND_KEY", "")
	gitIn(t, m.root, "add", "-A", "-f", "--", "docs", "gates")
	m.commit("record the second review")
	m.mustRun("sync", "all")
	if code, out := runCLI(t, m.root, "check", base); code != 0 || !strings.Contains(out, "WARRANTED (1 of 1") || strings.Contains(out, "NEEDS_WORK") {
		t.Fatalf("a re-seal as ACCEPTABLE did not decide: exit %d\n%s", code, out)
	}
}
