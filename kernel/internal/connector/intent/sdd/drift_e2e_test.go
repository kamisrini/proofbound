package sdd_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var driftLineRE = regexp.MustCompile(`(?m)^drift change=(\S+) state=(\S+) verified_commit=(\S+) verdict=(\S+)`)

func (m *m3Tenant) drift() (state, verdict, out string) {
	m.t.Helper()
	out = m.mustRun("report", "drift")
	match := driftLineRE.FindStringSubmatch(out)
	if match == nil {
		m.t.Fatalf("no drift line:\n%s", out)
	}
	return match[2], match[4], out
}

// R2 M4, CLI-INV-21 end to end (docs/plans/R2-m4-drift-spec.md): verify a change, reword its
// obligation (STALE, with the exact diff), re-verify (CURRENT), then archive it (still CURRENT:
// an archive move changes every digest but not the promise).
func TestDriftEndToEnd(t *testing.T) {
	m := newM3Tenant(t)
	if out := m.mustRun("report", "drift"); out != "summary warranted=0 current=0 stale=0 other=0\n" {
		t.Fatalf("an unwarranted tenant reported drift:\n%s", out)
	}
	x := m.commit("add auth", "Intent: sdd:add-auth")
	m.mustRun("sync", "intent", "sdd")
	m.mustRun("sync", "git")
	m.mustRun("review", "draft", x)
	m.review(x, "SATISFIED")
	m.commit("review x")
	m.mustRun("sync", "git")
	m.mustRun("sync", "reviews")
	if state, verdict, out := m.drift(); state != "CURRENT" || !strings.HasPrefix(verdict, "OV-"+x[:12]) {
		t.Fatalf("after verification:\n%s", out)
	}

	m.write("openspec/changes/add-auth/specs/core/spec.md", "## ADDED Requirements\n\n### Requirement: Login\nThe system SHALL authenticate users with a second factor.\n")
	m.commit("reword the login requirement")
	m.mustRun("sync", "intent", "sdd")
	state, verdict, out := m.drift()
	if state != "STALE" || !strings.HasPrefix(verdict, "OV-"+x[:12]) {
		t.Fatalf("after the reword:\n%s", out)
	}
	mustContain(t, out, `"The system SHALL authenticate users." -> "The system SHALL authenticate users with a second factor."`, "summary warranted=1 current=0 stale=1 other=0")

	y := m.commit("implement the second factor", "Intent: sdd:add-auth")
	m.mustRun("sync", "git")
	m.mustRun("review", "draft", y)
	m.review(y, "SATISFIED")
	m.commit("review y")
	m.mustRun("sync", "git")
	m.mustRun("sync", "reviews")
	if state, verdict, out := m.drift(); state != "CURRENT" || !strings.HasPrefix(verdict, "OV-"+y[:12]) {
		t.Fatalf("after re-verification:\n%s", out)
	}

	if err := os.MkdirAll(filepath.Join(m.root, "openspec/changes/archive"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(m.root, "openspec/changes/add-auth"), filepath.Join(m.root, "openspec/changes/archive/2026-11-19-add-auth")); err != nil {
		t.Fatal(err)
	}
	m.commit("archive add-auth")
	// The archive move must really produce new revisions (new digests), or this step proves nothing.
	if out := m.mustRun("sync", "intent", "sdd"); !regexp.MustCompile(`appended=[1-9]`).MatchString(out) {
		t.Fatalf("the archive move appended no revision:\n%s", out)
	}
	if state, _, out := m.drift(); state != "CURRENT" {
		t.Fatalf("an archive move with the same promise was reported as drift:\n%s", out)
	}
}

// The review r2-m4-drift-review-20261119 F1 scenario: A is ledgered, B is ledgered and verified,
// then the change is reverted to A's exact bytes. The ledger dedupes A (no new revision), so ledger
// order still says B; drift must read HEAD and report STALE, since A's promise was never verified.
// Reverting back to B is CURRENT again.
func TestDriftRevertToAnUnverifiedPromise(t *testing.T) {
	m := newM3Tenant(t)
	a := "## ADDED Requirements\n\n### Requirement: Login\nThe system SHALL authenticate users.\n"
	b := "## ADDED Requirements\n\n### Requirement: Login\nThe system SHALL authenticate users with a second factor.\n"
	spec := "openspec/changes/add-auth/specs/core/spec.md"
	m.commit("propose add-auth (A)")
	m.mustRun("sync", "intent", "sdd")
	m.write(spec, b)
	y := m.commit("implement add-auth (B)", "Intent: sdd:add-auth")
	m.mustRun("sync", "intent", "sdd")
	m.mustRun("sync", "git")
	m.mustRun("review", "draft", y)
	m.review(y, "SATISFIED")
	m.commit("review y")
	m.mustRun("sync", "git")
	m.mustRun("sync", "reviews")
	if state, _, out := m.drift(); state != "CURRENT" {
		t.Fatalf("after verifying B:\n%s", out)
	}

	m.write(spec, a)
	m.commit("revert to A")
	if out := m.mustRun("sync", "intent", "sdd"); !strings.Contains(out, "appended=0 ") {
		t.Fatalf("calibration: the revert must append nothing (A is deduped), or this test proves nothing:\n%s", out)
	}
	state, _, out := m.drift()
	if state != "STALE" {
		t.Fatalf("a revert to a never-verified promise was not STALE:\n%s", out)
	}
	mustContain(t, out, `"The system SHALL authenticate users with a second factor." -> "The system SHALL authenticate users."`)

	m.write(spec, b)
	m.commit("back to B")
	m.mustRun("sync", "intent", "sdd")
	if state, _, out := m.drift(); state != "CURRENT" {
		t.Fatalf("back on the verified promise:\n%s", out)
	}
}
