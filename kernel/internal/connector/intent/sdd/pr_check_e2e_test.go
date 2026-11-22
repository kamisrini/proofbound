package sdd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/store"
)

// MVP step 2 end to end (CLI-INV-26, CLI-INV-33, CHK-1..3): `proofbound check <base>` gives every
// claiming commit in base..HEAD one state — NO REVIEW, PASS, BLOCKED, UNBOUND — and lists every
// commit that claims no spec but changes code as UNCOVERED; the headline and exit code follow from
// them; it appends no ledger event and writes no file.
func TestPRCheckEndToEnd(t *testing.T) {
	m := &m3Tenant{t: t, root: t.TempDir()}
	gitIn(t, m.root, "init", "-q", "-b", "main")
	gitIn(t, m.root, "config", "user.name", "owner-a")
	m.mustRun("init", "speckit")
	if err := removeHooks(m.root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(m.root, ".github", "workflows", "proofbound.yml")); err != nil {
		t.Fatalf("init did not write the workflow: %v", err)
	}
	m.write("specs/001-export/spec.md", "# Feature: Export\n\n- **FR-001**: System MUST export reports as CSV.\n- **SC-001**: An export of 10k rows completes in under 2 seconds.\n")
	m.write("specs/001-export/tasks.md", "- [ ] T001 Export CSV\n")
	base := m.commit("specify 001-export")
	m.write("src/z", "z\n")
	z := m.commit("first attempt", "Intent: sdd:001-export")
	m.write("src/x", "x\n")
	x := m.commit("second attempt", "Intent: sdd:001-export")
	check := func(want int, args ...string) string {
		t.Helper()
		code, out := runCLI(t, m.root, append([]string{"check"}, args...)...)
		if code != want {
			t.Fatalf("check %v: exit %d, want %d\n%s", args, code, want, out)
		}
		return out
	}
	line := func(sha, state string) string { return "  " + sha[:12] + " sdd:001-export " + state + "  " }

	// Before any sync: both claims are UNBOUND (not in the ledger).
	out := check(1, base)
	mustContain(t, out, "Proofbound: NOT WARRANTED (0 of 2 claiming commits verified)", line(z, "UNBOUND"), line(x, "UNBOUND"), "run `proofbound sync all`")

	m.mustRun("sync", "all")
	out = check(1, base)
	mustContain(t, out, "NOT WARRANTED (0 of 2", line(z, "NO REVIEW"), line(x, "NO REVIEW"), "review draft "+x[:12])
	if strings.Count(out, "sdd:001-export") != 2 || strings.Contains(out, "UNCOVERED") {
		t.Fatalf("a claim was listed twice, or a housekeeping commit as UNCOVERED:\n%s", out)
	}

	m.mustRun("review", "draft", z)
	m.review(z, "INCONCLUSIVE")
	m.mustRun("review", "draft", x)
	m.review(x, "SATISFIED")
	m.commit("reviews")
	m.mustRun("sync", "all")

	countEvents := func() int {
		n := 0
		m.read(store.Filter{}, func(store.Record) error { n++; return nil })
		return n
	}
	before, status := countEvents(), gitIn(t, m.root, "status", "--porcelain")
	out = check(1, base)
	mustContain(t, out, "NOT WARRANTED (1 of 2", line(z, "BLOCKED"), line(x, "PASS"), "Drift:")
	if after := countEvents(); after != before || gitIn(t, m.root, "status", "--porcelain") != status {
		t.Fatalf("check wrote something: events %d -> %d", before, after)
	}

	// Range z..HEAD holds only x among the claims, plus the unclaimed "reviews" commit, which touches
	// only docs/ and gates/: WARRANTED, exit 0, and Markdown for the summary.
	mustContain(t, check(0, z), "Proofbound: WARRANTED (1 of 1 claiming commits verified)", line(x, "PASS"))
	mustContain(t, check(0, z, "--markdown"), "### Proofbound: WARRANTED", "| `"+x[:12]+"` | `sdd:001-export` | **PASS** |")

	// No claims in range: passes, unless --require-claims; a docs-only unclaimed commit changes nothing.
	mustContain(t, check(0, "HEAD"), "Proofbound: NO CLAIMS")
	check(1, "HEAD", "--require-claims")
	m.write("docs/notes.md", "n\n")
	m.commit("notes")
	mustContain(t, check(0, "HEAD~1"), "Proofbound: NO CLAIMS")

	// The mixed pull request: a reviewed claim plus an unclaimed commit that changes code is NOT
	// WARRANTED, and the unclaimed commit is named with the path it changed (CLI-INV-33).
	m.write("src/plain", "p\n")
	plain := m.commit("unrelated")
	out = check(1, z)
	mustContain(t, out, "Proofbound: NOT WARRANTED (1 of 1 claiming commits verified; 1 commit changes code without claiming a spec)",
		line(x, "PASS"), "  "+plain[:12]+" - UNCOVERED  claims no spec but changes src/plain")
	mustContain(t, check(1, z, "--markdown"), "| `"+plain[:12]+"` | `-` | **UNCOVERED** |")
	mustContain(t, check(1, "HEAD~1"), "Proofbound: NOT WARRANTED (no commit claims a spec; 1 commit changes code without claiming a spec)")

	// A mistyped claim is not "no claims": the commit is listed UNBOUND and the PR is not warranted.
	m.write("src/typo", "t\n")
	m.commit("mistyped claim", "intent: SDD:001-export")
	out = check(1, "HEAD~1")
	mustContain(t, out, "Proofbound: NOT WARRANTED (0 of 1 claiming commits verified)", "UNBOUND", "malformed claim")

	// A bad base is an error, not a verdict.
	if code, msg := runCLI(t, m.root, "check", "no-such-ref"); code != 1 || !strings.Contains(msg, "cannot list no-such-ref..HEAD") {
		t.Fatalf("bad base: exit %d\n%s", code, msg)
	}
}
