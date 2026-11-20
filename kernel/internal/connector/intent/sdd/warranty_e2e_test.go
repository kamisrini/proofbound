package sdd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// R3 slice 1 end to end (warranty SPEC W-INV-2, W-INV-5, W-INV-6): export is refused before the
// gate passes; after review it writes a warranty that verifies VALID against a fresh clone; a body
// re-signed by the declared key with a reworded obligation, an edited review artifact at HEAD, and
// a key removed at HEAD are each INVALID.
func TestWarrantyEndToEnd(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	m := &m3Tenant{t: t, root: t.TempDir()}
	gitIn(t, m.root, "init", "-q", "-b", "main")
	gitIn(t, m.root, "config", "user.name", "owner-a")
	m.mustRun("init", "speckit")
	if err := os.Remove(filepath.Join(m.root, ".git", "hooks", "commit-msg")); err != nil {
		t.Fatal(err)
	}
	m.write("specs/001-export/spec.md", "# Feature: Export\n\n- **FR-001**: System MUST export reports as CSV.\n- **SC-001**: An export of 10k rows completes in under 2 seconds.\n")
	m.write("specs/001-export/tasks.md", "- [ ] T001 Export CSV\n")
	m.commit("specify 001-export")
	m.write("src/x", "one\n")
	x := m.commit("implement 001-export", "Intent: sdd:001-export")
	m.mustRun("keys", "init", "--name", "owner-a")
	gitIn(t, m.root, "add", "-A", "--", ".gitignore", ".proofbound/keys")
	m.commit("declare key")
	m.mustRun("sync", "all")
	m.mustRun("review", "draft", x)

	out := filepath.Join(t.TempDir(), "w.json")
	if code, msg := runCLI(t, m.root, "warranty", "export", x, "-o", out); code == 0 {
		t.Fatalf("export before review succeeded:\n%s", msg)
	}
	if _, err := os.Stat(out); err == nil {
		t.Fatal("a refused export wrote a file")
	}

	m.review(x, "SATISFIED")
	m.commit("review x")
	m.mustRun("sync", "all")
	mustContain(t, m.mustRun("warranty", "export", x, "-o", out), "warranty ", "obligations=2")
	clone := filepath.Join(t.TempDir(), "clone")
	gitIn(t, filepath.Dir(clone), "clone", "-q", m.root, clone)
	verify := func(file, repo string) (int, string) { return runCLI(t, t.TempDir(), "warranty", "verify", file, "--repo", repo) }
	if code, msg := verify(out, clone); code != 0 || !strings.HasPrefix(msg, "VALID ") {
		t.Fatalf("a fresh warranty did not verify: exit %d\n%s", code, msg)
	}

	// Re-signed by the declared key, with a reworded obligation: only the git re-derivation can
	// catch it (the signature and the rule both pass).
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	f, err := warranty.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	keyPath, _ := warranty.KeyPath()
	key, err := warranty.LoadKey(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	body := f.Body
	body.Requirements[0].Obligations[0].Statement = "System MUST export reports as CSV, eventually."
	forged, err := warranty.Sign(body, key)
	if err != nil {
		t.Fatal(err)
	}
	forgedPath := filepath.Join(t.TempDir(), "forged.json")
	if err := os.WriteFile(forgedPath, forged, 0o644); err != nil {
		t.Fatal(err)
	}
	if code, msg := verify(forgedPath, clone); code != 1 || !strings.Contains(msg, "requirements differ from the spec") {
		t.Fatalf("a re-signed reworded obligation: exit %d\n%s", code, msg)
	}

	// The evidence must stay in the repository: an edited review artifact at HEAD is INVALID.
	rr, _ := filepath.Glob(filepath.Join(clone, "docs", "verification", "verdicts", "RR-*.md"))
	if len(rr) != 1 {
		t.Fatalf("calibration: %d review artifacts", len(rr))
	}
	edited, _ := os.ReadFile(rr[0])
	if err := os.WriteFile(rr[0], []byte(strings.Replace(string(edited), "VERIFIABLE", "AMBIGUOUS", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, clone, "-c", "user.name=x", "-c", "user.email=x@example.invalid", "commit", "-qam", "edit review")
	if code, msg := verify(out, clone); code != 1 || !strings.Contains(msg, "artifact") {
		t.Fatalf("an edited review artifact: exit %d\n%s", code, msg)
	}
	gitIn(t, clone, "reset", "-q", "--hard", "HEAD~1")

	// The key must be declared at HEAD.
	gitIn(t, clone, "rm", "-q", "-r", ".proofbound/keys")
	gitIn(t, clone, "-c", "user.name=x", "-c", "user.email=x@example.invalid", "commit", "-qm", "drop key")
	if code, msg := verify(out, clone); code != 1 || !strings.Contains(msg, "key not declared") {
		t.Fatalf("a removed key: exit %d\n%s", code, msg)
	}
}
