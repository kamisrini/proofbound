package sdd_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// MVP step 4 end to end (CLI-INV-28, CLI-INV-30): `warranty bundle` writes exactly one signed file per
// VALID/UNREVIEWED claiming commit plus index.json, VERIFY.md and badge.svg; each file verifies to its
// entry's state and matches its digest; exit 3 when written but incomplete, 0 when all VALID, 1 and
// nothing written when the directory is not empty.
func TestWarrantyBundleEndToEnd(t *testing.T) {
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
	// y claims its own spec: requirement reviews are per requirement, so an unsigned review of the
	// same requirement would also be cited by x's warranty and leave x UNREVIEWED.
	m.write("specs/002-import/spec.md", "# Feature: Import\n\n- **FR-001**: System MUST import CSV files.\n")
	m.write("specs/002-import/tasks.md", "- [ ] T001 Import CSV\n")
	m.mustRun("keys", "init", "--name", "owner-a")
	gitIn(t, m.root, "add", "-A", "--", ".gitignore", ".proofbound/keys")
	base := m.commit("specify 001-export and declare the exporter key")
	m.write("src/z", "z\n")
	z := m.commit("never reviewed", "Intent: sdd:001-export")
	m.write("src/y", "y\n")
	y := m.commit("reviewed without a reviewer key", "Intent: sdd:002-import")
	m.write("src/x", "x\n")
	x := m.commit("reviewed and signed", "Intent: sdd:001-export")
	m.write("src/plain", "p\n")
	m.commit("claims nothing")
	m.mustRun("sync", "all")

	m.mustRun("review", "draft", y)
	m.review(y, "SATISFIED") // no reviewer key yet: unsigned
	reviewerKey := filepath.Join(t.TempDir(), "reviewer-key")
	t.Setenv("PROOFBOUND_KEY", reviewerKey)
	m.mustRun("keys", "init", "--name", "proofbound-reviewer")
	m.mustRun("review", "draft", x)
	m.review(x, "SATISFIED") // signed (A)
	t.Setenv("PROOFBOUND_KEY", "")
	gitIn(t, m.root, "add", "--", ".proofbound/keys")
	m.commit("reviews")
	m.mustRun("sync", "all")

	out := filepath.Join(t.TempDir(), "w")
	code, msg := runCLI(t, m.root, "warranty", "bundle", base, "-o", out)
	if code != 3 {
		t.Fatalf("an incomplete bundle: exit %d\n%s", code, msg)
	}
	mustContain(t, msg, "1 of 3 claims verified (1 VALID, 1 UNREVIEWED)", "not warranted "+z[:12])
	entries, _ := os.ReadDir(out)
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	want := []string{"VERIFY.md", "badge.svg", "index.json", "warranty-" + x[:12] + ".json", "warranty-" + y[:12] + ".json"}
	sort.Strings(want)
	if strings.Join(names, " ") != strings.Join(want, " ") {
		t.Fatalf("bundle holds %v, want %v", names, want)
	}
	var ix struct {
		Unsigned           bool `json:"unsigned"`
		NonClaimingCommits int  `json:"non_claiming_commits"`
		Entries            []struct {
			Commit, State, File string
			FileSHA256          string   `json:"file_sha256"`
			ReviewersMissing    []string `json:"reviewers_missing"`
		} `json:"entries"`
	}
	raw, _ := os.ReadFile(filepath.Join(out, "index.json"))
	if err := json.Unmarshal(raw, &ix); err != nil {
		t.Fatal(err)
	}
	if !ix.Unsigned || ix.NonClaimingCommits != 2 || len(ix.Entries) != 3 {
		t.Fatalf("index: %s", raw)
	}
	states := map[string]string{z: "NO REVIEW", y: "UNREVIEWED", x: "VALID"}
	exits := map[string]int{"UNREVIEWED": 3, "VALID": 0}
	for i, e := range ix.Entries {
		if e.Commit != []string{z, y, x}[i] || e.State != states[e.Commit] {
			t.Fatalf("entry %d: %+v", i, e)
		}
		if e.File == "" {
			continue
		}
		data, _ := os.ReadFile(filepath.Join(out, e.File))
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != e.FileSHA256 {
			t.Fatalf("%s: digest mismatch", e.File)
		}
		if c, msg := runCLI(t, t.TempDir(), "warranty", "verify", filepath.Join(out, e.File), "--repo", m.root); c != exits[e.State] {
			t.Fatalf("%s verifies with exit %d, index says %s\n%s", e.File, c, e.State, msg)
		}
	}
	if len(ix.Entries[1].ReviewersMissing) != 1 || ix.Entries[1].ReviewersMissing[0] != "proofbound-reviewer" {
		t.Fatalf("UNREVIEWED entry does not name the missing reviewer: %+v", ix.Entries[1])
	}
	page, _ := os.ReadFile(filepath.Join(out, "VERIFY.md"))
	mustContain(t, string(page), "are not signed", "2 commit(s) in this range claim none")
	badge, _ := os.ReadFile(filepath.Join(out, "badge.svg"))
	mustContain(t, string(badge), "1 of 3 claims verified", `fill="#e05d44"`)

	// The same directory again: refused, nothing changed.
	if c, msg := runCLI(t, m.root, "warranty", "bundle", base, "-o", out); c != 1 || !strings.Contains(msg, "not empty") {
		t.Fatalf("a non-empty directory: exit %d\n%s", c, msg)
	}
	// Only x in range: all VALID, exit 0, green.
	green := filepath.Join(t.TempDir(), "g")
	if c, msg := runCLI(t, m.root, "warranty", "bundle", y, "-o", green); c != 0 {
		t.Fatalf("an all-VALID bundle: exit %d\n%s", c, msg)
	}
	badge, _ = os.ReadFile(filepath.Join(green, "badge.svg"))
	mustContain(t, string(badge), "1 of 1 claims verified", `fill="#4c1"`)
}
