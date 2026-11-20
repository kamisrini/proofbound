package sdd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// R2 Spec Kit end to end (docs/plans/R2-speckit-spec.md): the profile `init speckit` writes is one
// the provider maps, and a Spec Kit feature goes claim -> review -> gate PASS -> drift exactly as an
// OpenSpec change does, through cli.Run on the embedded store.
func TestSpecKitEndToEnd(t *testing.T) {
	m := &m3Tenant{t: t, root: t.TempDir()}
	gitIn(t, m.root, "init", "-q", "-b", "main")
	gitIn(t, m.root, "config", "user.name", "owner-a")
	mustContain(t, m.mustRun("init", "speckit"), "wrote ")
	if err := os.Remove(filepath.Join(m.root, ".git", "hooks", "commit-msg")); err != nil {
		t.Fatal(err)
	}
	m.write("specs/001-export/spec.md", "# Feature: Export\n\n- **FR-001**: System MUST export reports as CSV.\n- **SC-001**: An export of 10k rows completes in under 2 seconds.\n")
	m.commit("specify 001-export")
	m.write("specs/001-export/tasks.md", "- [ ] T001 Export CSV\n")
	m.write("src/x", "one\n")
	x := m.commit("implement 001-export", "Intent: sdd:001-export")
	// The documented user path: `sync all` alone. With a committed profile it reads the specs
	// (CLI-INV-23), and it must not trip over the P5 specdir convention, also rooted at specs/
	// (CLI-INV-22).
	out := m.mustRun("sync", "all")
	mustContain(t, out, "sdd listed=")
	if strings.Contains(out, "claim-skip") {
		t.Fatalf("the Spec Kit claim did not bind:\n%s", out)
	}
	mustContain(t, m.mustRun("report", "intent", "sdd:001-export"), "O-FR-001", "O-SC-001", "System MUST export reports as CSV.")
	m.mustRun("review", "draft", x)
	m.review(x, "SATISFIED")
	m.commit("review x")
	m.mustRun("sync", "all")
	m.expectStates(map[string]string{"ready-" + x[:12]: "PASS"})
	if state, _, out := m.drift(); state != "CURRENT" {
		t.Fatalf("drift after verification:\n%s", out)
	}

	m.write("specs/001-export/spec.md", "# Feature: Export\n\n- **FR-001**: System MUST export reports as CSV and JSON.\n- **SC-001**: An export of 10k rows completes in under 2 seconds.\n")
	m.commit("edit FR-001 in place")
	m.mustRun("sync", "all")
	state, _, out := m.drift()
	if state != "STALE" {
		t.Fatalf("an in-place Spec Kit edit was not drift:\n%s", out)
	}
	mustContain(t, out, `changed O-FR-001 "System MUST export reports as CSV." -> "System MUST export reports as CSV and JSON."`)
	if strings.Contains(out, "O-SC-001") {
		t.Fatalf("an unchanged obligation was reported:\n%s", out)
	}
}
