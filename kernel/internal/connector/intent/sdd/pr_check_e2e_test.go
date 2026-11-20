package sdd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/store"
)

// MVP step 2 end to end (CLI-INV-26, CHK-1..3): `proofbound check <base>` gives every claiming
// commit in base..HEAD one state — NO REVIEW, PASS, BLOCKED, UNBOUND — and the headline and exit
// code follow from them; it appends no ledger event and writes no file.
func TestPRCheckEndToEnd(t *testing.T) {
	m := &m3Tenant{t: t, root: t.TempDir()}
	gitIn(t, m.root, "init", "-q", "-b", "main")
	gitIn(t, m.root, "config", "user.name", "owner-a")
	m.mustRun("init", "speckit")
	if err := os.Remove(filepath.Join(m.root, ".git", "hooks", "commit-msg")); err != nil {
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
	m.write("src/plain", "p\n")
	m.commit("unrelated")

	check := func(want int, args ...string) string {
		t.Helper()
		code, out := runCLI(t, m.root, append([]string{"check"}, args...)...)
		if code != want {
			t.Fatalf("check %v: exit %d, want %d\n%s", args, code, want, out)
		}
		return out
	}
	line := func(sha, state string) string { return "  " + sha[:12] + " sdd:001-export " + state + "  " }

	// Before any sync: both claims are UNBOUND (not in the ledger); the unrelated commit is not listed.
	out := check(1, base)
	mustContain(t, out, "Proofbound: NOT WARRANTED (0 of 2 claiming commits verified)", line(z, "UNBOUND"), line(x, "UNBOUND"), "run `proofbound sync all`")

	m.mustRun("sync", "all")
	out = check(1, base)
	mustContain(t, out, "NOT WARRANTED (0 of 2", line(z, "NO REVIEW"), line(x, "NO REVIEW"), "review draft "+x[:12])
	if strings.Count(out, "sdd:001-export") != 2 {
		t.Fatalf("a non-claiming commit was listed:\n%s", out)
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

	// Range z..HEAD holds only x among the claims: WARRANTED, exit 0, and Markdown for the summary.
	mustContain(t, check(0, z), "Proofbound: WARRANTED (1 of 1 claiming commits verified)", line(x, "PASS"))
	mustContain(t, check(0, z, "--markdown"), "### Proofbound: WARRANTED", "| `"+x[:12]+"` | `sdd:001-export` | **PASS** |")

	// No claims in range: passes, unless --require-claims.
	mustContain(t, check(0, "HEAD"), "Proofbound: NO CLAIMS")
	check(1, "HEAD", "--require-claims")

	// A bad base is an error, not a verdict.
	if code, msg := runCLI(t, m.root, "check", "no-such-ref"); code != 1 || !strings.Contains(msg, "cannot list no-such-ref..HEAD") {
		t.Fatalf("bad base: exit %d\n%s", code, msg)
	}
}
