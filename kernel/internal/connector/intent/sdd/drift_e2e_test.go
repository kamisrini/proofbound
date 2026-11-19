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
	if out := m.mustRun("report", "drift"); out != "summary warranted=0 current=0 stale=0\n" {
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
	mustContain(t, out, `"The system SHALL authenticate users." -> "The system SHALL authenticate users with a second factor."`, "summary warranted=1 current=0 stale=1")

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
