package sdd_test

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	connectorreviews "github.com/kamisrini/proofbound/kernel/internal/connector/reviews"
	"github.com/kamisrini/proofbound/kernel/internal/core"
	"github.com/kamisrini/proofbound/kernel/internal/projections"
	"github.com/kamisrini/proofbound/kernel/internal/store"
)

// CLI-INV-31 / R-INV-14 (DEF-VERDICT-EVIDENCE-PORTABILITY-001): a commit reviewed and sealed in one
// clone is PASS in a second clone whose ledger starts empty — the GitHub Action's situation. The
// control proves the harness can fail: the same verdict rewritten to cite clone A's ledger-local
// event id dangles on a fresh ledger at `check`, which is exactly the v0.3.0 failure.
func TestPRCheckOnAFreshLedger(t *testing.T) {
	a := &m3Tenant{t: t, root: t.TempDir()}
	gitIn(t, a.root, "init", "-q", "-b", "main")
	gitIn(t, a.root, "config", "user.name", "owner-a")
	a.mustRun("init", "speckit")
	if err := removeHooks(a.root); err != nil {
		t.Fatal(err)
	}
	a.write("specs/001-export/spec.md", "# Feature: Export\n\n- **FR-001**: System MUST export reports as CSV.\n")
	a.write("specs/001-export/tasks.md", "- [ ] T001 Export CSV\n")
	base := a.commit("specify 001-export")
	a.write("src/x", "x\n")
	x := a.commit("export", "Intent: sdd:001-export")
	a.mustRun("sync", "all")
	a.mustRun("review", "draft", x)
	a.review(x, "SATISFIED")
	a.commit("reviews")

	var localID string
	a.read(store.Filter{Source: core.SourceGit}, func(r store.Record) error {
		if r.Event.NativeID == x && localID == "" {
			localID = r.Event.ID.String()
		}
		return nil
	})
	if localID == "" {
		t.Fatal("clone A has no git event for x")
	}

	clone := func(name string) *m3Tenant {
		root := filepath.Join(t.TempDir(), name)
		gitIn(t, a.root, "clone", "-q", a.root, root)
		gitIn(t, root, "config", "user.name", "ci")
		if _, err := os.Stat(filepath.Join(root, ".proofbound", "pgdata")); err == nil {
			t.Fatal("the clone carried a ledger")
		}
		return &m3Tenant{t: t, root: root}
	}

	b := clone("b")
	b.mustRun("sync", "all")
	code, out := runCLI(t, b.root, "check", base)
	if code != 0 {
		t.Fatalf("fresh-ledger check: exit %d\n%s", code, out)
	}
	mustContain(t, out, "Proofbound: WARRANTED (1 of 1 claiming commits verified)", "  "+x[:12]+" sdd:001-export PASS  ")
	// The reference resolved to clone B's OWN git event for x, not to nothing (code review F1).
	var bLocal string
	b.read(store.Filter{Source: core.SourceGit}, func(r store.Record) error {
		if r.Event.NativeID == x && bLocal == "" {
			bLocal = r.Event.ID.String()
		}
		return nil
	})
	if bLocal == "" || bLocal == localID {
		t.Fatalf("clone B git event %q (clone A %q)", bLocal, localID)
	}
	if got := viewEvidence(t, b.root); got != `["`+bLocal+`"]` {
		t.Fatalf("clone B view evidence %s, want [%q]", got, bLocal)
	}

	// Control: rewrite the sealed verdict to the v0.3.0 shape (clone A's event id, no reference).
	c := clone("c")
	verdicts, _ := filepath.Glob(filepath.Join(c.root, "docs", "verification", "verdicts", "OV-*.md"))
	if len(verdicts) != 1 {
		t.Fatalf("verdicts=%v", verdicts)
	}
	raw, err := os.ReadFile(verdicts[0])
	if err != nil {
		t.Fatal(err)
	}
	legacy := strings.ReplaceAll(string(raw), `"evidence_event_ids":[],"evidence_refs":["`+connectorreviews.CommitEvidenceRef(x)+`"]`, `"evidence_event_ids":["`+localID+`"]`)
	if legacy == string(raw) {
		t.Fatalf("sealed verdict does not cite the portable reference:\n%s", raw)
	}
	digest := regexp.MustCompile(`"artifact_sha256":"[0-9a-f]{64}"\}\n---`)
	legacy = digest.ReplaceAllString(legacy, `"artifact_sha256":"`+strings.Repeat("0", 64)+`"}`+"\n---")
	legacy = strings.Replace(legacy, strings.Repeat("0", 64), connectorreviews.ArtifactSHA256([]byte(legacy)), 1)
	if err := os.WriteFile(verdicts[0], []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	c.commit("legacy verdict")
	c.mustRun("sync", "all") // the reviews connector ledgers it; projection (run by check) rejects it
	if code, out := runCLI(t, c.root, "check", base); code == 0 || !strings.Contains(out, "is dangling") {
		t.Fatalf("control: a ledger-local evidence id projected on a fresh ledger (exit %d)\n%s", code, out)
	}
}

// viewEvidence projects the clone's ledger and returns the one obligation_verdicts_view evidence cell.
func viewEvidence(t *testing.T, root string) string {
	t.Helper()
	ctx := context.Background()
	t.Setenv("DATABASE_URL", "")
	ledger, err := store.Open(ctx, store.Config{Root: filepath.Join(root, ".proofbound")})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := ledger.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	if err := projections.New().Apply(ctx, ledger); err != nil {
		t.Fatal(err)
	}
	var got string
	if err := ledger.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error {
		return tx.QueryRow(ctx, `SELECT evidence_event_ids::text FROM obligation_verdicts_view`).Scan(&got)
	}); err != nil {
		t.Fatal(err)
	}
	return got
}
